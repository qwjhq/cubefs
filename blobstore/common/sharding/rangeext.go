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

package sharding

import "fmt"

// Belong check if compare item inside the Range
func (m *Range) Belong(ci *CompareItem) bool {
	switch m.Type {
	case RangeType_RangeTypeHash:
		return (*hashRange)(m).Belong(ci)
	default:
		panic(fmt.Sprintf(ErrUnsupportedRangeType.Error(), m.Type))
	}
}

// Contain return true when the target range contain in source range
func (m *Range) Contain(sub *Range) bool {
	switch m.Type {
	case RangeType_RangeTypeHash:
		return (*hashRange)(m).Contain(sub)
	default:
		panic(fmt.Sprintf(ErrUnsupportedRangeType.Error(), m.Type))
	}
}

// Split the source range into two children range with specified sub range index
// An ErrUnsupportedSplit should be return if the shard can not be split
func (m *Range) Split(idx int) ([2]Range, error) {
	switch m.Type {
	case RangeType_RangeTypeHash:
		return (*hashRange)(m).Split(idx)
	default:
		panic(fmt.Sprintf(ErrUnsupportedRangeType.Error(), m.Type))
	}
}

func (m *Range) IsEmpty() bool {
	switch m.Type {
	case RangeType_RangeTypeHash:
		return (*hashRange)(m).IsEmpty()
	default:
		panic(fmt.Sprintf(ErrUnsupportedRangeType.Error(), m.Type))
	}
}

// MaxBoundary return the max boundary of this Range
func (m *Range) MaxBoundary() Boundary {
	switch m.Type {
	case RangeType_RangeTypeHash:
		return (*hashRange)(m).MaxBoundary()
	default:
		panic(fmt.Sprintf(ErrUnsupportedRangeType.Error(), m.Type))
	}
}
