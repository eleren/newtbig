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
package service

import (
	"context"

	"github.com/eleren/newtbig/module"
)

type serverKey struct{}

func wait(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	wait, ok := ctx.Value("wait").(bool)
	if !ok {
		return false
	}
	return wait
}

func FromContext(ctx context.Context) (module.Server, bool) {
	c, ok := ctx.Value(serverKey{}).(module.Server)
	return c, ok
}

func NewContext(ctx context.Context, s module.Server) context.Context {
	return context.WithValue(ctx, serverKey{}, s)
}
