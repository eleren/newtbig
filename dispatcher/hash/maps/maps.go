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

import (
	"fmt"

	"github.com/eleren/newtbig/dispatcher/hash/containers"
)

type IMap interface {
	containers.Container
}

func assertMapImplementation() {
	var _ IMap = (*Map)(nil)
}

type color bool

const (
	black, red color = true, false
)

type Map struct {
	Root       *Node
	size       int
	Comparator containers.Comparator
}

type Node struct {
	Key    interface{}
	Value  interface{}
	color  color
	Left   *Node
	Right  *Node
	Parent *Node
}

func NewWith(comparator containers.Comparator) *Map {
	return &Map{Comparator: comparator}
}

func NewWithIntComparator() *Map {
	return &Map{Comparator: containers.IntComparator}
}

func NewWithUInt32Comparator() *Map {
	return &Map{Comparator: containers.UInt32Comparator}
}

func NewWithStringComparator() *Map {
	return &Map{Comparator: containers.StringComparator}
}

func (this *Map) Put(key interface{}, value interface{}) {
	var insertedNode *Node
	if this.Root == nil {
		this.Comparator(key, key)
		this.Root = &Node{Key: key, Value: value, color: red}
		insertedNode = this.Root
	} else {
		node := this.Root
		loop := true
		for loop {
			compare := this.Comparator(key, node.Key)
			switch {
			case compare == 0:
				node.Key = key
				node.Value = value
				return
			case compare < 0:
				if node.Left == nil {
					node.Left = &Node{Key: key, Value: value, color: red}
					insertedNode = node.Left
					loop = false
				} else {
					node = node.Left
				}
			case compare > 0:
				if node.Right == nil {
					node.Right = &Node{Key: key, Value: value, color: red}
					insertedNode = node.Right
					loop = false
				} else {
					node = node.Right
				}
			}
		}
		insertedNode.Parent = node
	}
	this.insertCase1(insertedNode)
	this.size++
}

func (this *Map) Get(key interface{}) (value interface{}, found bool) {
	node := this.lookup(key)
	if node != nil {
		return node.Value, true
	}
	return nil, false
}

func (this *Map) Remove(key interface{}) {
	var child *Node
	node := this.lookup(key)
	if node == nil {
		return
	}
	if node.Left != nil && node.Right != nil {
		pred := node.Left.maximumNode()
		node.Key = pred.Key
		node.Value = pred.Value
		node = pred
	}
	if node.Left == nil || node.Right == nil {
		if node.Right == nil {
			child = node.Left
		} else {
			child = node.Right
		}
		if node.color == black {
			node.color = nodeColor(child)
			this.deleteCase1(node)
		}
		this.replaceNode(node, child)
		if node.Parent == nil && child != nil {
			child.color = black
		}
	}
	this.size--
}

func (this *Map) Empty() bool {
	return this.size == 0
}

func (this *Map) Size() int {
	return this.size
}

func (this *Map) Keys() []interface{} {
	keys := make([]interface{}, this.size)
	it := this.Iterator()
	for i := 0; it.Next(); i++ {
		keys[i] = it.Key()
	}
	return keys
}

func (this *Map) Values() []interface{} {
	values := make([]interface{}, this.size)
	it := this.Iterator()
	for i := 0; it.Next(); i++ {
		values[i] = it.Value()
	}
	return values
}

func (this *Map) Left() *Node {
	var parent *Node
	current := this.Root
	for current != nil {
		parent = current
		current = current.Left
	}
	return parent
}

func (this *Map) Right() *Node {
	var parent *Node
	current := this.Root
	for current != nil {
		parent = current
		current = current.Right
	}
	return parent
}

func (this *Map) Floor(key interface{}) (floor *Node, found bool) {
	found = false
	node := this.Root
	for node != nil {
		compare := this.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			node = node.Left
		case compare > 0:
			floor, found = node, true
			node = node.Right
		}
	}
	if found {
		return floor, true
	}
	return nil, false
}

func (this *Map) Ceiling(key interface{}) (ceiling *Node, found bool) {
	found = false
	node := this.Root
	for node != nil {
		compare := this.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			ceiling, found = node, true
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}
	if found {
		return ceiling, true
	}
	return nil, false
}

func (this *Map) Clear() {
	this.Root = nil
	this.size = 0
}

func (this *Map) String() string {
	str := "RedBlackTree\n"
	if !this.Empty() {
		output(this.Root, "", true, &str)
	}
	return str
}

func (node *Node) String() string {
	return fmt.Sprintf("%v", node.Key)
}

func output(node *Node, prefix string, isTail bool, str *string) {
	if node.Right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		output(node.Right, newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += node.String() + "\n"
	if node.Left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		output(node.Left, newPrefix, true, str)
	}
}

func (this *Map) lookup(key interface{}) *Node {
	node := this.Root
	for node != nil {
		compare := this.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node
		case compare < 0:
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}
	return nil
}

func (node *Node) grandparent() *Node {
	if node != nil && node.Parent != nil {
		return node.Parent.Parent
	}
	return nil
}

func (node *Node) uncle() *Node {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	return node.Parent.sibling()
}

func (node *Node) sibling() *Node {
	if node == nil || node.Parent == nil {
		return nil
	}
	if node == node.Parent.Left {
		return node.Parent.Right
	}
	return node.Parent.Left
}

func (this *Map) rotateLeft(node *Node) {
	right := node.Right
	this.replaceNode(node, right)
	node.Right = right.Left
	if right.Left != nil {
		right.Left.Parent = node
	}
	right.Left = node
	node.Parent = right
}

func (this *Map) rotateRight(node *Node) {
	left := node.Left
	this.replaceNode(node, left)
	node.Left = left.Right
	if left.Right != nil {
		left.Right.Parent = node
	}
	left.Right = node
	node.Parent = left
}

func (this *Map) replaceNode(old *Node, new *Node) {
	if old.Parent == nil {
		this.Root = new
	} else {
		if old == old.Parent.Left {
			old.Parent.Left = new
		} else {
			old.Parent.Right = new
		}
	}
	if new != nil {
		new.Parent = old.Parent
	}
}

func (this *Map) insertCase1(node *Node) {
	if node.Parent == nil {
		node.color = black
	} else {
		this.insertCase2(node)
	}
}

func (this *Map) insertCase2(node *Node) {
	if nodeColor(node.Parent) == black {
		return
	}
	this.insertCase3(node)
}

func (this *Map) insertCase3(node *Node) {
	uncle := node.uncle()
	if nodeColor(uncle) == red {
		node.Parent.color = black
		uncle.color = black
		node.grandparent().color = red
		this.insertCase1(node.grandparent())
	} else {
		this.insertCase4(node)
	}
}

func (this *Map) insertCase4(node *Node) {
	grandparent := node.grandparent()
	if node == node.Parent.Right && node.Parent == grandparent.Left {
		this.rotateLeft(node.Parent)
		node = node.Left
	} else if node == node.Parent.Left && node.Parent == grandparent.Right {
		this.rotateRight(node.Parent)
		node = node.Right
	}
	this.insertCase5(node)
}

func (this *Map) insertCase5(node *Node) {
	node.Parent.color = black
	grandparent := node.grandparent()
	grandparent.color = red
	if node == node.Parent.Left && node.Parent == grandparent.Left {
		this.rotateRight(grandparent)
	} else if node == node.Parent.Right && node.Parent == grandparent.Right {
		this.rotateLeft(grandparent)
	}
}

func (node *Node) maximumNode() *Node {
	if node == nil {
		return nil
	}
	for node.Right != nil {
		node = node.Right
	}
	return node
}

func (this *Map) deleteCase1(node *Node) {
	if node.Parent == nil {
		return
	}
	this.deleteCase2(node)
}

func (this *Map) deleteCase2(node *Node) {
	sibling := node.sibling()
	if nodeColor(sibling) == red {
		node.Parent.color = red
		sibling.color = black
		if node == node.Parent.Left {
			this.rotateLeft(node.Parent)
		} else {
			this.rotateRight(node.Parent)
		}
	}
	this.deleteCase3(node)
}

func (this *Map) deleteCase3(node *Node) {
	sibling := node.sibling()
	if nodeColor(node.Parent) == black &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left) == black &&
		nodeColor(sibling.Right) == black {
		sibling.color = red
		this.deleteCase1(node.Parent)
	} else {
		this.deleteCase4(node)
	}
}

func (this *Map) deleteCase4(node *Node) {
	sibling := node.sibling()
	if nodeColor(node.Parent) == red &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left) == black &&
		nodeColor(sibling.Right) == black {
		sibling.color = red
		node.Parent.color = black
	} else {
		this.deleteCase5(node)
	}
}

func (this *Map) deleteCase5(node *Node) {
	sibling := node.sibling()
	if node == node.Parent.Left &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left) == red &&
		nodeColor(sibling.Right) == black {
		sibling.color = red
		sibling.Left.color = black
		this.rotateRight(sibling)
	} else if node == node.Parent.Right &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Right) == red &&
		nodeColor(sibling.Left) == black {
		sibling.color = red
		sibling.Right.color = black
		this.rotateLeft(sibling)
	}
	this.deleteCase6(node)
}

func (this *Map) deleteCase6(node *Node) {
	sibling := node.sibling()
	sibling.color = nodeColor(node.Parent)
	node.Parent.color = black
	if node == node.Parent.Left && nodeColor(sibling.Right) == red {
		sibling.Right.color = black
		this.rotateLeft(node.Parent)
	} else if nodeColor(sibling.Left) == red {
		sibling.Left.color = black
		this.rotateRight(node.Parent)
	}
}

func nodeColor(node *Node) color {
	if node == nil {
		return black
	}
	return node.color
}
