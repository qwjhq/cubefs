// Copyright 2022 The CubeFS Authors.
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

package storage

import (
	"encoding/binary"

	"github.com/cubefs/cubefs/blobstore/api/clustermgr"
	"github.com/cubefs/cubefs/blobstore/common/proto"
	shardnodeproto "github.com/cubefs/cubefs/blobstore/shardnode/proto"
)

const (
	MaxKeySize   = 1 << 15
	MaxValueSize = 1 << 24

	dataCF  = "data"
	lockCF  = "lock"
	writeCF = "write"
)

var (
	// top level prefix
	shardDataPrefix = []byte{'d'}
	shardInfoPrefix = []byte{'s'}

	// shard's internal suffix
	itemSuffix = []byte{'a'}
	blobSuffix = []byte{'b'}
	maxSuffix  = []byte{'z'}
)

type Timestamp struct{}

// proto for storage encoding/decoding and function return value

type (
	item      = shardnodeproto.Item
	shardInfo = clustermgr.Shard
)

// todo: merge these encode and decode function into shard?

func shardDataPrefixSize() int {
	return len(shardDataPrefix) + 4
}

func shardInfoPrefixSize() int {
	return len(shardInfoPrefix) + 8
}

func shardItemPrefixSize() int {
	return shardDataPrefixSize() + len(itemSuffix)
}

func shardBlobPrefixSize() int {
	return shardDataPrefixSize() + len(blobSuffix)
}

func shardMaxPrefixSize() int {
	return shardDataPrefixSize() + len(maxSuffix)
}

func encodeShardInfoListPrefix(raw []byte) {
	if raw == nil || cap(raw) == 0 {
		panic("invalid raw input")
	}
	copy(raw, shardInfoPrefix)
}

func encodeShardInfoPrefix(suid proto.Suid, raw []byte) {
	if raw == nil || cap(raw) == 0 {
		panic("invalid raw input")
	}
	prefixSize := len(shardInfoPrefix)
	copy(raw, shardInfoPrefix)
	binary.BigEndian.PutUint64(raw[prefixSize:], uint64(suid))
}

func decodeShardInfoPrefix(raw []byte) proto.Suid {
	if raw == nil || cap(raw) == 0 {
		panic("invalid raw input")
	}
	prefixSize := len(shardInfoPrefix)
	return proto.Suid(binary.BigEndian.Uint64(raw[prefixSize:]))
}

func encodeShardDataPrefix(shardID proto.ShardID, raw []byte) {
	copy(raw, shardDataPrefix)
	binary.BigEndian.PutUint32(raw[len(shardDataPrefix):], uint32(shardID))
}

func encodeShardItemPrefix(shardID proto.ShardID, raw []byte) {
	shardPrefixSize := shardDataPrefixSize()
	encodeShardDataPrefix(shardID, raw)
	copy(raw[shardPrefixSize:], itemSuffix)
}

func encodeShardBlobPrefix(shardID proto.ShardID, raw []byte) {
	shardPrefixSize := shardDataPrefixSize()
	encodeShardDataPrefix(shardID, raw)
	copy(raw[shardPrefixSize:], blobSuffix)
}

func encodeShardDataMaxPrefix(shardID proto.ShardID, raw []byte) {
	shardPrefixSize := shardDataPrefixSize()
	encodeShardDataPrefix(shardID, raw)
	copy(raw[shardPrefixSize:], maxSuffix)
}
