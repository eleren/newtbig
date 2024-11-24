// Copyright 2020 newtbig Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package maps

import "github.com/eleren/newtbig/dispatcher/hash/containers"

func assertIteratorImplementation() {
	var _ containers.ReverseIteratorWithKey = (*Iterator)(nil)
}

type Iterator struct {
	maps     *Map
	node     *Node
	position position
}

type position byte

const (
	begin, between, end position = 0, 1, 2
)

func (m *Map) Iterator() Iterator {
	return Iterator{maps: m, node: nil, position: begin}
}

func (ite *Iterator) Next() bool {
	if ite.position == end {
		goto end
	}
	if ite.position == begin {
		left := ite.maps.Left()
		if left == nil {
			goto end
		}
		ite.node = left
		goto between
	}
	if ite.node.Right != nil {
		ite.node = ite.node.Right
		for ite.node.Left != nil {
			ite.node = ite.node.Left
		}
		goto between
	}
	if ite.node.Parent != nil {
		node := ite.node
		for ite.node.Parent != nil {
			ite.node = ite.node.Parent
			if ite.maps.Comparator(node.Key, ite.node.Key) <= 0 {
				goto between
			}
		}
	}

end:
	ite.node = nil
	ite.position = end
	return false

between:
	ite.position = between
	return true
}

func (ite *Iterator) Prev() bool {
	if ite.position == begin {
		goto begin
	}
	if ite.position == end {
		right := ite.maps.Right()
		if right == nil {
			goto begin
		}
		ite.node = right
		goto between
	}
	if ite.node.Left != nil {
		ite.node = ite.node.Left
		for ite.node.Right != nil {
			ite.node = ite.node.Right
		}
		goto between
	}
	if ite.node.Parent != nil {
		node := ite.node
		for ite.node.Parent != nil {
			ite.node = ite.node.Parent
			if ite.maps.Comparator(node.Key, ite.node.Key) >= 0 {
				goto between
			}
		}
	}

begin:
	ite.node = nil
	ite.position = begin
	return false

between:
	ite.position = between
	return true
}

func (ite *Iterator) Value() interface{} {
	return ite.node.Value
}

func (ite *Iterator) Key() interface{} {
	return ite.node.Key
}

func (ite *Iterator) Begin() {
	ite.node = nil
	ite.position = begin
}

func (ite *Iterator) End() {
	ite.node = nil
	ite.position = end
}

func (ite *Iterator) First() bool {
	ite.Begin()
	return ite.Next()
}

func (ite *Iterator) Last() bool {
	ite.End()
	return ite.Prev()
}
