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
package hash

import (
	"errors"
	"hash/crc32"
	"strconv"
	"sync"

	"github.com/eleren/newtbig/dispatcher/hash/maps"
)

var ErrEmptyRing = errors.New("empty ring")

type (
	HashRing struct {
		m_RingMap    map[uint32]string
		m_MemberMap  map[string]bool
		m_SortedKeys *maps.Map
		rRep         uint32
		rLen         uint32
		sync.RWMutex
	}

	IHashRing interface {
		Add(elt string) []uint32
		Remove(elt string)
		HasMember(elt string) bool
		Members() []string
		Get(name string) (error, string)
		Get64(val int64) (error, uint32)
	}
)

func NewHashRing(rep uint32, rLen uint32) *HashRing {
	pRing := new(HashRing)
	pRing.m_RingMap = make(map[uint32]string)
	pRing.m_MemberMap = make(map[string]bool)
	pRing.m_SortedKeys = maps.NewWithUInt32Comparator()
	pRing.rLen = rLen
	pRing.rRep = rep
	return pRing
}
func (hr *HashRing) addInt(id uint32, elt string) {
	hr.m_RingMap[id] = elt
	hr.m_SortedKeys.Put(id, true)
	hr.m_MemberMap[elt] = true
	return
}

func (hr *HashRing) removeInt(id uint32) {
	elt := hr.m_RingMap[id]
	delete(hr.m_RingMap, id)
	hr.m_SortedKeys.Remove(id)
	delete(hr.m_MemberMap, elt)
}

func (hr *HashRing) eltKey(elt string, idx int) string {
	return strconv.Itoa(idx) + elt
}

func (hr *HashRing) add(elt string) []uint32 {
	rst := make([]uint32, 0)
	for i := 0; i < int(hr.rRep); i++ {
		Id := hr.hashKey(hr.eltKey(elt, i))
		hr.m_RingMap[Id] = elt
		hr.m_SortedKeys.Put(Id, true)
		rst = append(rst, Id)
	}
	hr.m_MemberMap[elt] = true
	return rst
}

func (hr *HashRing) remove(elt string) {
	for i := 0; i < int(hr.rRep); i++ {
		Id := hr.hashKey(hr.eltKey(elt, i))
		delete(hr.m_RingMap, Id)
		hr.m_SortedKeys.Remove(Id)
	}
	delete(hr.m_MemberMap, elt)
}

func (hr *HashRing) hashKey(key string) uint32 {
	if len(key) < 64 {
		var scratch [64]byte
		copy(scratch[:], key)
		return crc32.ChecksumIEEE(scratch[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

func (hr *HashRing) Add(elt string) []uint32 {
	hr.Lock()
	defer hr.Unlock()
	return hr.add(elt)
}

func (hr *HashRing) Remove(elt string) {
	hr.Lock()
	defer hr.Unlock()
	hr.remove(elt)
}

func (hr *HashRing) HasMember(elt string) bool {
	hr.RLock()
	defer hr.RUnlock()
	_, bEx := hr.m_MemberMap[elt]
	return bEx
}

func (hr *HashRing) Members() []string {
	hr.RLock()
	defer hr.RUnlock()
	var m []string
	for k := range hr.m_MemberMap {
		m = append(m, k)
	}
	return m
}

func (hr *HashRing) Get(name string) (error, string) {
	hr.RLock()
	defer hr.RUnlock()
	if len(hr.m_RingMap) == 0 {
		return ErrEmptyRing, ""
	}
	key := hr.hashKey(name)
	node, bOk := hr.m_SortedKeys.Ceiling(key)
	if !bOk {
		itr := hr.m_SortedKeys.Iterator()
		if itr.First() {
			return nil, hr.m_RingMap[itr.Key().(uint32)]
		}
		return ErrEmptyRing, ""
	}
	return nil, hr.m_RingMap[node.Key.(uint32)]
}

func (hr *HashRing) Get64(val int64) (error, uint32) {
	hr.RLock()
	defer hr.RUnlock()
	if len(hr.m_RingMap) == 0 {
		return ErrEmptyRing, 0
	}
	key := hr.hashKey(strconv.FormatInt(val, 10))
	node, bOk := hr.m_SortedKeys.Ceiling(key)
	if !bOk {
		itr := hr.m_SortedKeys.Iterator()
		if itr.First() {
			return nil, hr.hashKey(hr.m_RingMap[itr.Key().(uint32)])
		}
		return ErrEmptyRing, 0
	}
	return nil, hr.hashKey(hr.m_RingMap[node.Key.(uint32)])
}
