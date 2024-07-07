// Copyright 2020 newtbig Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package process

import "sort"

type perfRecorder struct {
	msgID     uint32
	name      string
	totalTime int64
	count     int32
	maxTime   int64
}

type perfRecorderSortBy func(p1, p2 *perfRecorder) bool

func (by perfRecorderSortBy) Sort(records []*perfRecorder) {
	ps := &perfRecorderSorter{
		records: records,
		by:      by,
	}
	sort.Sort(ps)
}

type perfRecorderSorter struct {
	records []*perfRecorder
	by      perfRecorderSortBy
}

func (s *perfRecorderSorter) Len() int {
	return len(s.records)
}

func (s *perfRecorderSorter) Swap(i, j int) {
	s.records[i], s.records[j] = s.records[j], s.records[i]
}

func (s *perfRecorderSorter) Less(i, j int) bool {
	return s.by(s.records[i], s.records[j])
}
