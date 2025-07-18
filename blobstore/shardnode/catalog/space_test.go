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

package catalog

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/cubefs/cubefs/blobstore/api/clustermgr"
	"github.com/cubefs/cubefs/blobstore/api/shardnode"
	"github.com/cubefs/cubefs/blobstore/common/codemode"
	apierr "github.com/cubefs/cubefs/blobstore/common/errors"
	"github.com/cubefs/cubefs/blobstore/common/proto"
	"github.com/cubefs/cubefs/blobstore/common/security"
	"github.com/cubefs/cubefs/blobstore/shardnode/mock"
	"github.com/cubefs/cubefs/blobstore/util/errors"
)

type mockSpace struct {
	space         *Space
	shardErrSpace *Space
	mockHandler   *mock.MockSpaceShardHandler
}

func newMockSpace(tb testing.TB) (*mockSpace, func()) {
	fixedFields := make(map[proto.FieldID]clustermgr.FieldMeta)
	fixedFields[1] = clustermgr.FieldMeta{
		Name:        "f1",
		FieldType:   proto.FieldTypeString,
		IndexOption: proto.IndexOptionNull,
	}
	fixedFields[2] = clustermgr.FieldMeta{
		Name:        "f2",
		FieldType:   proto.FieldTypeString,
		IndexOption: proto.IndexOptionNull,
	}
	handler := mock.NewMockSpaceShardHandler(C(tb))

	sg := mock.NewMockShardGetter(C(tb))
	sg.EXPECT().GetShard(A, A).Return(handler, nil).AnyTimes()

	sg2 := mock.NewMockShardGetter(C(tb))
	sg2.EXPECT().GetShard(A, A).Return(nil, apierr.ErrShardDoesNotExist).AnyTimes()

	space := &Space{
		sid:         1,
		name:        "space1",
		fieldMetas:  fixedFields,
		shardGetter: sg,
	}
	shardErrSpace := &Space{
		sid:         1,
		name:        "space1",
		fieldMetas:  fixedFields,
		shardGetter: sg2,
	}
	return &mockSpace{space: space, shardErrSpace: shardErrSpace, mockHandler: handler}, func() {
	}
}

func TestSpace_Item(t *testing.T) {
	mockSpace, cleanSpace := newMockSpace(t)
	defer cleanSpace()
	fields := []shardnode.Field{
		{ID: 1, Value: []byte("f1")},
		{ID: 2, Value: []byte("f2")},
	}
	oph := shardnode.ShardOpHeader{}
	gomock.InOrder(mockSpace.mockHandler.EXPECT().InsertItem(A, A, A, A).Return(nil))
	// insert
	err := mockSpace.space.InsertItem(ctx, oph, shardnode.Item{Fields: fields})
	require.Nil(t, err)
	gomock.InOrder(mockSpace.mockHandler.EXPECT().InsertItem(A, A, A, A).Return(errors.New("insert error")))
	err = mockSpace.space.InsertItem(ctx, oph, shardnode.Item{Fields: fields})
	require.Equal(t, errors.New("insert error"), err)
	err = mockSpace.space.InsertItem(ctx, oph, shardnode.Item{Fields: []shardnode.Field{
		{ID: 3, Value: []byte("string")},
	}})
	require.Equal(t, apierr.ErrUnknownField, err)
	err = mockSpace.shardErrSpace.InsertItem(ctx, oph, shardnode.Item{})
	require.Equal(t, apierr.ErrShardDoesNotExist, err)
	// get
	gomock.InOrder(mockSpace.mockHandler.EXPECT().GetItem(A, A, A).Return(shardnode.Item{Fields: fields}, nil))
	ret, err := mockSpace.space.GetItem(ctx, oph, []byte{1})
	require.Nil(t, err)
	require.Equal(t, shardnode.Item{Fields: fields}, ret)
	_, err = mockSpace.shardErrSpace.GetItem(ctx, oph, []byte{99})
	require.Equal(t, apierr.ErrShardDoesNotExist, err)
	// update
	gomock.InOrder(mockSpace.mockHandler.EXPECT().UpdateItem(A, A, A, A).Return(nil))
	err = mockSpace.space.UpdateItem(ctx, oph, shardnode.Item{Fields: fields})
	require.Nil(t, err)
	err = mockSpace.space.UpdateItem(ctx, oph, shardnode.Item{Fields: []shardnode.Field{
		{ID: 3, Value: []byte("string")},
	}})
	require.Equal(t, apierr.ErrUnknownField, err)
	err = mockSpace.shardErrSpace.UpdateItem(ctx, oph, shardnode.Item{})
	require.Equal(t, apierr.ErrShardDoesNotExist, err)
	// list
	gomock.InOrder(mockSpace.mockHandler.EXPECT().ListItem(A, A, A, A, A).Return([]shardnode.Item{
		{
			ID:     []byte("1"),
			Fields: fields,
		},
		{
			ID:     []byte("2"),
			Fields: fields,
		},
	}, mockSpace.space.generateSpaceKey([]byte("3")), nil))
	_, marker, err := mockSpace.space.ListItem(ctx, oph, nil, nil, 2)
	require.Nil(t, err)
	require.Equal(t, []byte("3"), marker)

	// delete
	gomock.InOrder(mockSpace.mockHandler.EXPECT().DeleteItem(A, A, A).Return(nil))
	err = mockSpace.space.DeleteItem(ctx, oph, []byte{1})
	require.Nil(t, err)
	err = mockSpace.shardErrSpace.DeleteItem(ctx, oph, []byte{1})
	require.Equal(t, apierr.ErrShardDoesNotExist, err)
}

func TestSpace_CreateBlob(t *testing.T) {
	ctx := context.Background()
	mockSpace, cleanSpace := newMockSpace(t)
	defer cleanSpace()
	space := mockSpace.space

	alc := mock.NewMockAllocator(C(t))
	space.allocator = alc

	slices := []proto.Slice{
		{
			MinSliceID: proto.BlobID(1),
			Vid:        proto.Vid(100),
			Count:      160,
			ValidSize:  uint64(10240),
		},
	}

	name := []byte("blob")
	oph := shardnode.ShardOpHeader{}
	args := &shardnode.CreateBlobArgs{
		Header:    oph,
		Name:      name,
		CodeMode:  codemode.EC6P6,
		Size_:     1024 * 10,
		SliceSize: 64,
	}

	b := proto.Blob{
		Name: name,
		Location: proto.Location{
			CodeMode:  codemode.EC6P6,
			Size_:     1024 * 10,
			SliceSize: 64,
			Slices:    slices,
		},
	}
	security.LocationCrcFill(&b.Location)

	gomock.InOrder(alc.EXPECT().AllocSlices(A, A, A, A, A).Return(slices, nil))

	mockSpace.mockHandler.EXPECT().GetBlob(A, A, A).Return(proto.Blob{}, apierr.ErrKeyNotFound)
	mockSpace.mockHandler.EXPECT().CreateBlob(A, A, A, A).Return(b, nil)

	ret, err := mockSpace.space.CreateBlob(ctx, args)
	require.Nil(t, err)
	b = ret.Blob
	require.NotNil(t, b.Location.Slices)
	require.True(t, security.LocationCrcVerify(&b.Location))

	// repeat create
	mockSpace.mockHandler.EXPECT().GetBlob(A, A, A).Return(b, nil)

	ret, err = mockSpace.space.CreateBlob(ctx, args)
	require.Equal(t, apierr.ErrBlobAlreadyExists, err)

	// create with alloc size 0
	args.Size_ = 0

	mockSpace.mockHandler.EXPECT().GetBlob(A, A, A).Return(proto.Blob{}, apierr.ErrKeyNotFound)
	b.Location.Slices = nil
	mockSpace.mockHandler.EXPECT().CreateBlob(A, A, A, A).Return(b, nil)
	ret, err = mockSpace.space.CreateBlob(ctx, args)
	require.Nil(t, err)
	b = ret.Blob
	require.Nil(t, b.Location.Slices)
}

func TestSpace_AllocSlice(t *testing.T) {
	ctx := context.Background()
	mockSpace, cleanSpace := newMockSpace(t)
	defer cleanSpace()
	space := mockSpace.space

	alc := mock.NewMockAllocator(C(t))
	space.allocator = alc

	locSlices := []proto.Slice{
		{Vid: 1, MinSliceID: 1, Count: 10, ValidSize: 0},
		{Vid: 1, MinSliceID: 2, Count: 20, ValidSize: 0},
		{Vid: 1, MinSliceID: 3, Count: 30, ValidSize: 0},
	}
	name := []byte("blob")
	mode := codemode.EC6P6
	b := proto.Blob{
		Name: name,
		Location: proto.Location{
			CodeMode:  mode,
			Slices:    locSlices,
			SliceSize: 10,
		},
	}
	mockSpace.mockHandler.EXPECT().GetBlob(A, A, A).Return(b, nil).Times(5)

	newSlices := []proto.Slice{
		{Vid: 1, MinSliceID: 4, Count: 10, ValidSize: 100},
	}
	alc.EXPECT().AllocSlices(A, A, A, A, A).Return(newSlices, nil).Times(3)
	mockSpace.mockHandler.EXPECT().UpdateBlob(A, A, A, A).Return(nil).Times(3)
	args := &shardnode.AllocSliceArgs{
		Header: shardnode.ShardOpHeader{},
		Name:   name,
		Size_:  64,
	}
	// failedSlice == nil
	ret, err := space.AllocSlice(ctx, args)
	require.Nil(t, err)
	require.Equal(t, newSlices, ret.Slices)

	// illegal failedSlice
	args.FailedSlice = proto.Slice{Vid: 1, MinSliceID: 4, Count: 30, ValidSize: 300}
	_, err = space.AllocSlice(ctx, args)
	require.Equal(t, apierr.ErrIllegalSlices, err)

	// failedSlice: part
	args.FailedSlice = proto.Slice{Vid: 1, MinSliceID: 2, Count: 20, ValidSize: 100}
	ret, err = space.AllocSlice(ctx, args)
	require.Nil(t, err)
	require.Equal(t, newSlices, ret.Slices)

	// failedSlice: all
	args.FailedSlice.Count = 20
	args.FailedSlice.ValidSize = 0
	ret, err = space.AllocSlice(ctx, args)
	require.Nil(t, err)
	require.Equal(t, newSlices, ret.Slices)

	// alloc failed
	alc.EXPECT().AllocSlices(A, A, A, A, A).Return(nil, apierr.ErrNoCodemodeVolume)
	ret, err = space.AllocSlice(ctx, args)
	require.Equal(t, apierr.ErrNoCodemodeVolume, errors.Cause(err))

	// blob already sealed
	b.Sealed = true
	mockSpace.mockHandler.EXPECT().GetBlob(A, A, A).Return(b, nil).Times(1)
	ret, err = space.AllocSlice(ctx, args)
	require.Equal(t, err, apierr.ErrBlobAlreadySealed)
}

func TestSpace_SealBlob(t *testing.T) {
	ctx := context.Background()
	mockSpace, cleanSpace := newMockSpace(t)
	defer cleanSpace()
	space := mockSpace.space

	alc := mock.NewMockAllocator(C(t))
	space.allocator = alc

	// seal size should <= 100+100+100+200=500, and > 300
	locSlices := []proto.Slice{
		{Vid: 1, MinSliceID: 1, Count: 10, ValidSize: 100},
		{Vid: 1, MinSliceID: 2, Count: 10, ValidSize: 100},
		{Vid: 1, MinSliceID: 3, Count: 10, ValidSize: 0},
		{Vid: 1, MinSliceID: 4, Count: 20, ValidSize: 0},
	}
	name := []byte("blob")
	mode := codemode.EC6P6
	b := proto.Blob{
		Name: name,
		Location: proto.Location{
			CodeMode:  mode,
			Slices:    locSlices,
			SliceSize: 10,
		},
	}
	reqSlice := make([]proto.Slice, len(locSlices))
	copy(reqSlice, locSlices)

	mockSpace.mockHandler.EXPECT().GetBlob(A, A, A).Return(b, nil).Times(14)
	mockSpace.mockHandler.EXPECT().UpdateBlob(A, A, A, A).Return(nil).Times(2)

	// length not equal
	args := &shardnode.SealBlobArgs{
		Header: shardnode.ShardOpHeader{},
		Name:   name,
		Slices: reqSlice[:2],
	}
	err := space.SealBlob(ctx, args)
	require.NotNil(t, err)

	// seal size too small, can't fill slice in middle
	reqSlice[2].ValidSize = 100
	reqSlice[3].ValidSize = 110
	args = &shardnode.SealBlobArgs{
		Header: shardnode.ShardOpHeader{},
		Name:   name,
		Size_:  290,
		Slices: reqSlice,
	}
	err = space.SealBlob(ctx, args)
	require.Equal(t, apierr.ErrIllegalLocationSize, err)

	// not full write
	args.Size_ = 410
	err = space.SealBlob(ctx, args)
	require.Nil(t, err)

	// full write
	reqSlice[3].ValidSize = 200
	args.Size_ = 500
	err = space.SealBlob(ctx, args)
	require.Nil(t, err)
	/*local blob slice is updated to:
	[]proto.Slice{
		{Vid: 1, MinSliceID: 1, Count: 10, ValidSize: 100},
		{Vid: 1, MinSliceID: 2, Count: 10, ValidSize: 100},
		{Vid: 1, MinSliceID: 3, Count: 10, ValidSize: 100},
		{Vid: 1, MinSliceID: 4, Count: 20, ValidSize: 200},
	}*/

	// seal size too small than local
	args.Size_ = 200
	err = space.SealBlob(ctx, args)
	require.Equal(t, apierr.ErrIllegalLocationSize, err)

	// seal size too small than local, quit at last slice
	args.Size_ = 410
	err = space.SealBlob(ctx, args)
	require.Equal(t, apierr.ErrIllegalLocationSize, err)

	// seal size too large
	args.Size_ = 700
	err = space.SealBlob(ctx, args)
	require.Equal(t, apierr.ErrIllegalLocationSize, err)

	args.Slices[0].MinSliceID = 3
	err = space.SealBlob(ctx, args)
	require.Equal(t, apierr.ErrIllegalSlices, err)
}

func TestSpace_DeleteBlob(t *testing.T) {
	ctx := context.Background()
	mockSpace, cleanSpace := newMockSpace(t)
	defer cleanSpace()
	space := mockSpace.space

	mockSpace.mockHandler.EXPECT().DeleteBlob(A, A, A).Return(nil)

	err := space.DeleteBlob(ctx, &shardnode.DeleteBlobArgs{
		Header: shardnode.ShardOpHeader{},
		Name:   []byte("blob"),
	})
	require.Nil(t, err)
}

func TestSpace_FindAndDeleteBlob(t *testing.T) {
	ctx := context.Background()
	mockSpace, cleanSpace := newMockSpace(t)
	defer cleanSpace()
	space := mockSpace.space

	blob := proto.Blob{Name: []byte("blob"), Location: proto.Location{CodeMode: codemode.EC6P6}}

	mockSpace.mockHandler.EXPECT().DeleteBlob(A, A, A).Return(nil)
	mockSpace.mockHandler.EXPECT().GetBlob(A, A, A).Return(blob, nil)

	ret, err := space.FindAndDeleteBlob(ctx, &shardnode.DeleteBlobArgs{
		Header: shardnode.ShardOpHeader{},
		Name:   []byte("blob"),
	})
	require.Nil(t, err)
	require.Equal(t, ret.Blob, blob)
}

func TestSpace_ListBlob(t *testing.T) {
	mockSpace, cleanSpace := newMockSpace(t)
	defer cleanSpace()
	space := mockSpace.space

	blobs := make([]proto.Blob, 0)
	for i := 0; i < 10; i++ {
		blob := proto.Blob{Name: []byte(fmt.Sprintf("b%d", i)), Location: proto.Location{CodeMode: codemode.EC6P6}}
		blobs = append(blobs, blob)
	}

	nextMarker := []byte("next")
	mockSpace.mockHandler.EXPECT().ListBlob(A, A, A, A, A).Return(
		blobs, space.generateSpaceKey(nextMarker), nil,
	)
	blobs, m, err := space.ListBlob(ctx, shardnode.ShardOpHeader{}, nil, []byte("b1"), 10)
	require.Nil(t, err)
	require.Equal(t, nextMarker, m)
	require.Equal(t, 10, len(blobs))
}

func Test_SpaceKey(t *testing.T) {
	space := Space{sid: 1000, spaceVersion: 0}

	key1 := []byte("blob0")
	key2 := []byte("blob00")
	key3 := []byte("blob1")
	key4 := []byte("blob10")
	key5 := []byte("blob10000")
	key6 := []byte("blob2")

	// test decode and compare
	keys := [][]byte{key1, key2, key3, key4, key5, key6}
	for i := range keys {
		spaceKey := space.generateSpaceKey(keys[i])
		require.True(t, bytes.Equal(keys[i], space.decodeSpaceKey(spaceKey)))

		var frontSpaceKey []byte
		if i > 0 {
			require.Equal(t, 1, bytes.Compare(keys[i], keys[i-1]))
			frontSpaceKey = space.generateSpaceKey(keys[i-1])
			require.Equal(t, 1, bytes.Compare(spaceKey, frontSpaceKey))
		}

		space.spaceVersion = 123
		newSpaceKey := space.generateSpaceKey(keys[i])
		require.True(t, bytes.Equal(keys[i], space.decodeSpaceKey(newSpaceKey)))

		// new version in the front
		// newFrontSpaceKey < frontSpaceKey < newSpaceKey < spaceKey
		require.Equal(t, 1, bytes.Compare(spaceKey, newSpaceKey))
		if i > 0 {
			require.Equal(t, 1, bytes.Compare(newSpaceKey, frontSpaceKey))

			newFrontSpaceKey := space.generateSpaceKey(keys[i-1])
			require.Equal(t, 1, bytes.Compare(frontSpaceKey, newFrontSpaceKey))
		}
		space.spaceVersion = 0
	}

	// test prefix
	_prefix := space.generateSpacePrefix(nil)
	require.True(t, bytes.HasPrefix(space.generateSpaceKey(keys[0]), _prefix))

	_prefix = space.generateSpacePrefix([]byte("b"))
	require.True(t, bytes.HasPrefix(space.generateSpaceKey(keys[0]), _prefix))
}
