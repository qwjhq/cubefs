// Copyright 2024 The CubeFS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package catalog

import (
	"context"
	"encoding/json"
	goerrors "errors"

	"github.com/google/uuid"

	cmapi "github.com/cubefs/cubefs/blobstore/api/clustermgr"
	"github.com/cubefs/cubefs/blobstore/api/shardnode"
	"github.com/cubefs/cubefs/blobstore/clustermgr/base"
	"github.com/cubefs/cubefs/blobstore/clustermgr/cluster"
	"github.com/cubefs/cubefs/blobstore/clustermgr/persistence/catalogdb"
	apierrors "github.com/cubefs/cubefs/blobstore/common/errors"
	"github.com/cubefs/cubefs/blobstore/common/proto"
	"github.com/cubefs/cubefs/blobstore/common/sharding"
	"github.com/cubefs/cubefs/blobstore/common/trace"
	"github.com/cubefs/cubefs/blobstore/util/errors"
)

// ListShardUnitInfo return disk's shard unit infos, it uses index-table disk-suid as index
// this API is lightly operation, it only called when broken disk or some else, so here just get data from db
func (c *CatalogMgr) ListShardUnitInfo(ctx context.Context, args *cmapi.ListShardUnitArgs) ([]cmapi.ShardUnitInfo, error) {
	unitPrefixes, err := c.catalogTbl.ListShardUnit(args.DiskID)
	if err != nil {
		return nil, errors.Info(err, "list shardUnit from tbl failed")
	}

	ret := make([]cmapi.ShardUnitInfo, 0, len(unitPrefixes))
	for _, unitPrefix := range unitPrefixes {
		shard := c.allShards.getShard(unitPrefix.ShardID())
		shard.withRLocked(func() error {
			unit := shard.info.Units[unitPrefix.Index()]
			ret = append(ret, shardUnitToShardUnitInfo(unit, shard.info.RouteVersion, shard.info.Range, shard.info.LeaderDiskID))
			return nil
		})
	}
	return ret, nil
}

func (c *CatalogMgr) AllocShardUnit(ctx context.Context, args *cmapi.AllocShardUnitArgs) (*cmapi.AllocShardUnitRet, error) {
	span := trace.SpanFromContextSafe(ctx)
	suid := args.Suid
	shardID := suid.ShardID()
	shard := c.allShards.getShard(shardID)
	if shard == nil {
		return nil, ErrShardNotExist
	}

	var nextEpoch uint32
	err := shard.withRLocked(func() error {
		index := suid.Index()
		if index >= uint8(len(shard.unitEpochs)) {
			return ErrShardUnitNotExist
		}

		unitEpoch := shard.unitEpochs[index]
		if _suid := proto.EncodeSuid(shardID, index, unitEpoch.epoch); _suid != suid {
			span.Errorf("request suid[%d] not equal with record suid[%d]", suid, _suid)
			return ErrOldSuidNotMatch
		}
		nextEpoch = unitEpoch.nextEpoch + 1
		return nil
	})
	if err != nil {
		return nil, err
	}

	pendingSuidKey := uuid.New().String()
	c.pendingEntries.Store(pendingSuidKey, proto.Suid(0))
	// clear pending entry key
	defer c.pendingEntries.Delete(pendingSuidKey)

	data, err := json.Marshal(&allocShardUnitCtx{Suid: suid, NextEpoch: nextEpoch, PendingSuidKey: pendingSuidKey})
	if err != nil {
		return nil, errors.Info(err, "json marshal failed")
	}

	err = c.raftServer.Propose(ctx, base.EncodeProposeInfo(c.GetModuleName(), OperTypeAllocShardUnit, data, base.ProposeContext{ReqID: span.TraceID()}))
	if err != nil {
		return nil, errors.Info(err, "propose failed")
	}
	newSuid, _ := c.pendingEntries.Load(pendingSuidKey)
	if newSuid.(proto.Suid) == 0 {
		return nil, apierrors.ErrConcurrentAllocShardUnit
	}

	shardRange := sharding.Range{}
	targetDiskID := proto.DiskID(0)
	routeVersion := proto.RouteVersion(0)
	excludeDisks := args.ExcludeDiskIDs
	repairUnits := make([]cmapi.ShardUnit, 0, c.CodeMode.GetShardNum())
	shard.withRLocked(func() error {
		shardRange = shard.info.Range
		targetDiskID = shard.info.Units[suid.Index()].DiskID
		for i := range shard.info.Units {
			excludeDisks = append(excludeDisks, shard.info.Units[i].DiskID)
		}
		routeVersion = shard.info.RouteVersion
		repairUnits = shard.info.Units
		return nil
	})

	diskInfo, err := c.diskMgr.GetDiskInfo(ctx, targetDiskID)
	if err != nil {
		return nil, errors.Info(err, "get disk info failed")
	}

	// The shard routeVersion recorded by shardnode will be corrected when reporting shards
	policy := cluster.AllocShardsPolicy{
		DiskType:     proto.DiskTypeNVMeSSD,
		Suids:        []proto.Suid{newSuid.(proto.Suid)},
		Range:        shardRange,
		RouteVersion: routeVersion,
		RepairUnits:  repairUnits,
		ExcludeDisks: excludeDisks,
		DiskSetID:    diskInfo.DiskSetID,
		Idc:          diskInfo.Idc,
	}
	span.Infof("AllocShards policy %+v", policy)

	allocDiskID, _, err := c.diskMgr.AllocShards(ctx, policy)
	if err != nil {
		return nil, errors.Info(err, "alloc shard failed")
	}

	newDiskInfo, err := c.diskMgr.GetDiskInfo(ctx, allocDiskID[0])
	if err != nil {
		return nil, errors.Info(err, "get disk info failed")
	}

	return &cmapi.AllocShardUnitRet{Suid: policy.Suids[0], DiskID: allocDiskID[0], Host: newDiskInfo.Host}, nil
}

func (c *CatalogMgr) PreUpdateShardUnit(ctx context.Context, args *cmapi.UpdateShardArgs) (err error) {
	span := trace.SpanFromContextSafe(ctx)
	shard := c.allShards.getShard(args.OldSuid.ShardID())
	if shard == nil {
		return ErrShardNotExist
	}

	err = shard.withRLocked(func() error {
		if int(args.OldSuid.Index()) >= len(shard.unitEpochs) {
			return ErrNewSuidNotMatch
		}
		unit := shard.info.Units[args.OldSuid.Index()]
		unitEpoch := shard.unitEpochs[args.OldSuid.Index()]
		if (unit.Suid != args.OldSuid && unit.Suid != args.NewSuid) || unitEpoch.nextEpoch < args.OldSuid.Epoch() {
			span.Errorf("shard's suid is %v", unit.Suid)
			return ErrOldSuidNotMatch
		}
		// idempotent retry update shard unit, return success
		if unit.Suid == args.NewSuid {
			return ErrRepeatUpdateShardUnit
		}
		if unit.Learner != args.OldIsLeaner {
			return ErrOldIsLeanerNotMatch
		}
		if proto.EncodeSuid(unit.Suid.ShardID(), unit.Suid.Index(), unitEpoch.nextEpoch) != args.NewSuid {
			span.Errorf("shard's suid is %v", proto.EncodeSuid(unit.Suid.ShardID(), unit.Suid.Index(), unitEpoch.nextEpoch))
			return ErrNewSuidNotMatch
		}
		return nil
	})
	if err != nil {
		return err
	}

	disk, err := c.diskMgr.GetDiskInfo(ctx, args.NewDiskID)
	if err != nil {
		span.Errorf("new diskID:%v not exist", args.NewDiskID)
		return apierrors.ErrCMDiskNotFound
	}
	getShardArgs := shardnode.GetShardArgs{
		DiskID: disk.DiskID,
		Suid:   args.NewSuid,
	}
	unitInfo, err := c.shardNodeClient.GetShardUintInfo(ctx, disk.Host, getShardArgs)
	if err != nil {
		span.Errorf("stat shardnode shard, disk id[%d], suid[%d] failed: %s", args.NewDiskID, args.NewSuid, err.Error())
		return apierrors.ErrGetShardFailed
	}
	if unitInfo.DiskID != args.NewDiskID {
		span.Errorf("new diskID:%v not match", args.NewDiskID)
		return ErrNewDiskIDNotMatch
	}

	return nil
}

func (c *CatalogMgr) UpdateShardUnit(ctx context.Context, args *cmapi.UpdateShardArgs) (err error) {
	span := trace.SpanFromContextSafe(ctx)

	data, err := json.Marshal(args)
	if err != nil {
		return errors.Info(err, "json marshal failed").Detail(err)
	}
	proposeInfo := base.EncodeProposeInfo(c.GetModuleName(), OperTypeUpdateShardUnit, data, base.ProposeContext{ReqID: span.TraceID()})
	err = c.raftServer.Propose(ctx, proposeInfo)
	if err != nil {
		return errors.Info(apierrors.ErrRaftPropose, "raft propose err: ", err)
	}

	return nil
}

func (c *CatalogMgr) ReportShard(ctx context.Context, args *cmapi.ShardReportArgs, rawData []byte) (ret *cmapi.ShardReportRet, err error) {
	span := trace.SpanFromContextSafe(ctx)

	proposeInfo := base.EncodeProposeInfo(c.GetModuleName(), OperTypeShardReport, rawData, base.ProposeContext{ReqID: span.TraceID()})
	err = c.raftServer.Propose(ctx, proposeInfo)
	if err != nil {
		span.Error(err)
		return nil, apierrors.ErrRaftPropose
	}

	ret = new(cmapi.ShardReportRet)
	for _, reportUnit := range args.Shards {
		shard := c.allShards.getShard(reportUnit.Suid.ShardID())
		if shard == nil {
			span.Warnf("shard not found, shardID: %d, suid: %d", reportUnit.Suid.ShardID(), reportUnit.Suid)
			continue
		}
		idx := reportUnit.Suid.Index()
		err = shard.withRLocked(func() error {
			unitInfo := shard.info.Units[idx]
			// in some case, the report suid epoch may bigger than epoch in cm, like repair, we should just ignore it
			if reportUnit.Suid.Epoch() > unitInfo.Suid.Epoch() {
				return errors.Newf("report suid: %d epoch is bigger than epoch in CM suid: %d", reportUnit.Suid, unitInfo.Suid)
			}
			if reportUnit.Suid.Epoch() < unitInfo.Suid.Epoch() {
				if isReplicateMember(reportUnit.DiskID, shard.info.Units) {
					reportSuidEpochNotConsistent(unitInfo.Suid, reportUnit.Suid, c.Region, c.ClusterID)
					return errors.Newf("report suid: %d epoch is not consistent with CM suid: %d", reportUnit.Suid, unitInfo.Suid)
				}
				ret.ShardTasks = append(ret.ShardTasks, cmapi.ShardTask{
					TaskType:     proto.ShardTaskTypeClearShard,
					DiskID:       reportUnit.DiskID,
					Suid:         reportUnit.Suid,
					RouteVersion: shard.info.RouteVersion,
				})
			}
			if reportUnit.RouteVersion != shard.info.RouteVersion {
				ret.ShardTasks = append(ret.ShardTasks, cmapi.ShardTask{
					TaskType:        proto.ShardTaskTypeSyncRouteVersion,
					DiskID:          reportUnit.DiskID,
					Suid:            reportUnit.Suid,
					OldRouteVersion: reportUnit.RouteVersion,
					RouteVersion:    shard.info.RouteVersion,
				})
			}
			return nil
		})
		if err != nil {
			span.Warnf("ReportShard err:%v", err)
		}
	}

	return ret, nil
}

func (c *CatalogMgr) UpdateShardUnitStatus(ctx context.Context, diskID proto.DiskID) (err error) {
	span := trace.SpanFromContextSafe(ctx)
	suidPrefixes, err := c.catalogTbl.ListShardUnit(diskID)
	if err != nil {
		return err
	}
	if len(suidPrefixes) == 0 {
		return nil
	}
	data, err := json.Marshal(suidPrefixes)
	if err != nil {
		return errors.Info(err, "json marshal failed").Detail(err)
	}

	proposeInfo := base.EncodeProposeInfo(c.GetModuleName(), OperTypeUpdateShardUnitStatus, data, base.ProposeContext{ReqID: span.TraceID()})
	err = c.raftServer.Propose(ctx, proposeInfo)
	if err != nil {
		return errors.Info(err, "raft propose err").Detail(err)
	}

	return
}

func (c *CatalogMgr) applyUpdateShardUnit(ctx context.Context, newSuid proto.Suid, newDiskID proto.DiskID, learner bool) (err error) {
	span := trace.SpanFromContextSafe(ctx)
	span.Debugf("start apply update shard unit, newSuid is %d, newDiskID is %d", newSuid, newDiskID)

	shard := c.allShards.getShard(newSuid.ShardID())
	index := newSuid.Index()
	err = shard.withRLocked(func() error {
		if shard.info.Units[index].Suid == newSuid {
			return ErrRepeatedApplyRequest
		}
		// when apply wal log happened, the next epoch of shard unit in db may larger than args new suid's epoch
		// just return nil in this situation
		if shard.unitEpochs[index].nextEpoch > newSuid.Epoch() {
			span.Debugf("shard nextEpoch: %d bigger than newSuid Epoch : %d", shard.unitEpochs[index].nextEpoch, newSuid.Epoch())
			return ErrRepeatedApplyRequest
		}
		return nil
	})
	if goerrors.Is(err, ErrRepeatedApplyRequest) {
		return nil
	}

	diskInfo, err := c.diskMgr.GetDiskInfo(ctx, newDiskID)
	if err != nil {
		span.Errorf("get diskInfo failed,diskID is %d", newDiskID)
		return err
	}

	newRouteVersion := c.routeMgr.genRouteVersion(ctx, 1)
	route := &routeItem{
		RouteVersion: proto.RouteVersion(newRouteVersion),
		Type:         proto.CatalogChangeItemUpdateShard,
		ItemDetail:   &routeItemShardUpdate{SuidPrefix: newSuid.SuidPrefix()},
	}

	err = shard.withLocked(func() error {
		shardRecord := shard.toShardRecord()
		shardRecord.RouteVersion = proto.RouteVersion(newRouteVersion)
		shardUnitRecord := shardUnitToShardUnitRecord(shard.info.Units[index], *shard.unitEpochs[index])
		shardUnitRecord.Epoch = newSuid.Epoch()
		shardUnitRecord.DiskID = newDiskID
		shardUnitRecord.Learner = learner

		shardRecords := []*catalogdb.ShardInfoRecord{shardRecord}
		unitRecords := []*catalogdb.ShardUnitInfoRecord{shardUnitRecord}
		routeRecords := []*catalogdb.RouteInfoRecord{routeItemToRouteRecord(route)}
		err := c.catalogTbl.UpdateUnitsAndPutShardsAndRouteItems(shardRecords, unitRecords, routeRecords)
		if err != nil {
			return err
		}

		shard.unitEpochs[index].epoch = newSuid.Epoch()

		shard.info.Units[index].Suid = newSuid
		shard.info.Units[index].DiskID = newDiskID
		shard.info.Units[index].Learner = learner
		shard.info.Units[index].Host = diskInfo.Host
		shard.info.RouteVersion = proto.RouteVersion(newRouteVersion)
		c.routeMgr.insertRouteItems(ctx, []*routeItem{route})
		return nil
	})
	if err != nil {
		return errors.Info(err, "catalog table update shard unit failed")
	}

	err = c.refreshShard(ctx, shard.shardID)
	if err != nil {
		span.Errorf("refresh shard failed, shardID is %d, error is %v", shard.shardID, err)
		return err
	}

	span.Debugf("finish apply update shard unit")
	return
}

func (c *CatalogMgr) applyUpdateShardUnitStatus(ctx context.Context, suidPrefixes []proto.SuidPrefix) (err error) {
	span := trace.SpanFromContextSafe(ctx)
	span.Debugf("start apply update unit status, suidPrefixes is %v", suidPrefixes)

	for _, suidPrefix := range suidPrefixes {
		err = c.refreshShard(ctx, suidPrefix.ShardID())
		if err != nil {
			span.Errorf("update unit status failed, err is %v", err)
			return err
		}
	}
	span.Debugf("finish apply update unit status")

	return
}

func (c *CatalogMgr) refreshShard(ctx context.Context, shardID proto.ShardID) error {
	span := trace.SpanFromContextSafe(ctx)
	shard := c.allShards.getShard(shardID)
	unitRecords := make([]*catalogdb.ShardUnitInfoRecord, 0, c.CodeMode.GetShardNum())
	return shard.withLocked(func() error {
		for i := range shard.info.Units {
			writable, err := c.diskMgr.IsDiskWritable(ctx, shard.info.Units[i].DiskID)
			if err != nil {
				return err
			}
			span.Debugf("disk writable is %v, disk_id: %d", writable, shard.info.Units[i].DiskID)
			if writable {
				shard.info.Units[i].Status = proto.ShardUnitStatusNormal
			} else {
				shard.info.Units[i].Status = proto.ShardUnitStatusOffline
			}
			unitRecords = append(unitRecords, shardUnitToShardUnitRecord(shard.info.Units[i], *shard.unitEpochs[i]))
		}
		return c.catalogTbl.PutShardsAndUnitsAndRouteItems(nil, unitRecords, nil)
	})
}

func (c *CatalogMgr) applyAllocShardUnit(ctx context.Context, args *allocShardUnitCtx) (err error) {
	span := trace.SpanFromContextSafe(ctx)
	span.Debugf("start apply alloc shard unit, args is %v", args)

	idx := args.Suid.Index()
	shard := c.allShards.getShard(args.Suid.ShardID())
	err = shard.withLocked(func() error {
		// concurrent alloc shard unit or wal log replay, do nothing and return
		if shard.unitEpochs[idx].nextEpoch >= args.NextEpoch {
			return ErrRepeatedApplyRequest
		}
		shard.unitEpochs[idx].nextEpoch = args.NextEpoch
		unitRecord := shardUnitToShardUnitRecord(shard.info.Units[idx], *shard.unitEpochs[idx])
		return c.catalogTbl.PutShardsAndUnitsAndRouteItems(nil, []*catalogdb.ShardUnitInfoRecord{unitRecord}, nil)
	})
	if err != nil {
		if goerrors.Is(err, ErrRepeatedApplyRequest) {
			return nil
		}
		return errors.Info(err, "put shard unit to catalogTbl failed")
	}

	// set pending entry in current process context
	newSuid := proto.EncodeSuid(args.Suid.ShardID(), idx, shard.unitEpochs[idx].nextEpoch)
	if _, ok := c.pendingEntries.Load(args.PendingSuidKey); ok {
		span.Debugf("new suid is %d", newSuid)
		c.pendingEntries.Store(args.PendingSuidKey, newSuid)
	}

	span.Debugf("finish apply alloc shard unit")
	return
}

// applyShardReport only change shardUnit leaner and shardItem leader info
func (c *CatalogMgr) applyShardReport(ctx context.Context, args *cmapi.ShardReportArgs) (err error) {
	span := trace.SpanFromContextSafe(ctx)

	for _, reportUnit := range args.Shards {
		shard := c.allShards.getShard(reportUnit.Suid.ShardID())
		if shard == nil {
			span.Warnf("shard not found, shardID: %d, suid: %d", reportUnit.Suid.ShardID(), reportUnit.Suid)
			continue
		}
		idx := reportUnit.Suid.Index()
		dirtyFlag := false
		shard.withLocked(func() error {
			unitInfo := shard.info.Units[idx]
			// The reported suid epoch is inconsistent with CM, and the reported info is ignored
			if reportUnit.Suid.Epoch() != unitInfo.Suid.Epoch() {
				return nil
			}
			if shard.info.LeaderDiskID != reportUnit.LeaderDiskID || unitInfo.Learner != reportUnit.Learner {
				shard.info.LeaderDiskID = reportUnit.LeaderDiskID
				shard.info.Units[idx].Learner = reportUnit.Learner
				dirtyFlag = true
			}
			return nil
		})
		// put on dirty shard and flush asynchronously
		if dirtyFlag {
			dirty := c.dirty.Load().(*concurrentShards)
			dirty.putShard(shard)
		}
	}

	return
}

func (c *CatalogMgr) applyAdminUpdateShard(ctx context.Context, shardInfo *cmapi.Shard) error {
	span := trace.SpanFromContextSafe(ctx)
	shard := c.allShards.getShard(shardInfo.ShardID)
	if shard == nil {
		span.Errorf("apply admin update shard, shardID %d not exist", shardInfo.ShardID)
		return ErrShardNotExist
	}

	return shard.withLocked(func() error {
		shard.info.RouteVersion = shardInfo.RouteVersion
		shard.info.Range = shardInfo.Range
		shardRecords := []*catalogdb.ShardInfoRecord{shard.toShardRecord()}
		unitRecords := make([]*catalogdb.ShardUnitInfoRecord, 0, len(shard.unitEpochs))
		for index := range shard.unitEpochs {
			unitRecords = append(unitRecords, shardUnitToShardUnitRecord(shard.info.Units[index], *shard.unitEpochs[index]))
		}
		return c.catalogTbl.PutShardsAndUnitsAndRouteItems(shardRecords, unitRecords, nil)
	})
}

func (c *CatalogMgr) applyAdminUpdateShardUnit(ctx context.Context, args *cmapi.AdminUpdateShardUnitArgs) error {
	span := trace.SpanFromContextSafe(ctx)
	shard := c.allShards.getShard(args.Suid.ShardID())
	if shard == nil {
		span.Errorf("apply admin update shard unit,suid %d not exist", args.Suid.ShardID())
		return ErrShardNotExist
	}
	index := args.Suid.Index()
	err := shard.withRLocked(func() error {
		if int(index) >= len(shard.unitEpochs) {
			span.Errorf("apply admin update shard unit,index:%d over suids length ", index)
			return ErrShardUnitNotExist
		}
		return nil
	})
	if err != nil {
		return err
	}

	return shard.withLocked(func() error {
		diskInfo, err := c.diskMgr.GetDiskInfo(ctx, args.DiskID)
		if err != nil {
			return err
		}
		shard.unitEpochs[index].epoch = args.Epoch
		shard.unitEpochs[index].nextEpoch = args.NextEpoch

		shard.info.Shard.Units[index].Suid = proto.EncodeSuid(args.Suid.ShardID(), index, args.Epoch)
		shard.info.Shard.Units[index].DiskID = diskInfo.DiskID
		shard.info.Shard.Units[index].Learner = args.Learner
		shard.info.Shard.Units[index].Host = diskInfo.Host
		shard.info.Shard.Units[index].Status = args.Status

		shardRecords := []*catalogdb.ShardInfoRecord{shard.toShardRecord()}
		unitRecords := []*catalogdb.ShardUnitInfoRecord{shardUnitToShardUnitRecord(shard.info.Units[index], *shard.unitEpochs[index])}
		return c.catalogTbl.UpdateUnitsAndPutShardsAndRouteItems(shardRecords, unitRecords, nil)
	})
}

func isReplicateMember(target proto.DiskID, units []cmapi.ShardUnit) bool {
	for i := range units {
		if units[i].DiskID == target {
			return true
		}
	}
	return false
}
