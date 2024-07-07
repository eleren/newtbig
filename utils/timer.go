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
package utils

import (
	"time"
)

type Timer struct {
	Time     time.Time
	Interval int
}

func NewTimer(interval int) *Timer {
	t := &Timer{}
	t.Time = time.Now()
	t.Interval = interval
	return t
}

func (c *Timer) Run() bool {
	if time.Now().After(c.Time.Add(time.Duration(1e6 * c.Interval))) {
		c.Time = time.Now()
		return true
	} else {
		return false
	}
}

func (c *Timer) AddRandomMillisec(maxms int) {
	c.Time = c.Time.Add(time.Duration(RandBetween(0, maxms) * 1e6))
}
