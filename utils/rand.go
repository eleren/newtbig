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
	"math/rand"
	"time"
)

func RandBetween(min, max int) int {
	if min < 0 || max < 0 {
		return 0
	}
	if min > max {
		min, max = max, min
	}
	return rand.Intn(max-min+1) + min
}

func Uint32MaxBetween(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func Uint32MinBetween(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func Int64MaxBetween(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

var (
	x0  uint32 = uint32(time.Now().UnixNano())
	a   uint32 = 1664525
	c   uint32 = 1013904223
	LCG chan uint32
)

const (
	PRERNG = 1024
)

func init() {
	LCG = make(chan uint32, PRERNG)
	go func() {
		for {
			x0 = a*x0 + c
			LCG <- x0
		}
	}()
}
