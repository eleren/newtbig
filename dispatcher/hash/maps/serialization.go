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
package maps

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func ToString(value interface{}) string {
	switch value.(type) {
	case string:
		return value.(string)
	case int8:
		return strconv.FormatInt(int64(value.(int8)), 10)
	case int16:
		return strconv.FormatInt(int64(value.(int16)), 10)
	case int32:
		return strconv.FormatInt(int64(value.(int32)), 10)
	case int64:
		return strconv.FormatInt(int64(value.(int64)), 10)
	case uint8:
		return strconv.FormatUint(uint64(value.(uint8)), 10)
	case uint16:
		return strconv.FormatUint(uint64(value.(uint16)), 10)
	case uint32:
		return strconv.FormatUint(uint64(value.(uint32)), 10)
	case uint64:
		return strconv.FormatUint(uint64(value.(uint64)), 10)
	case float32:
		return strconv.FormatFloat(float64(value.(float32)), 'g', -1, 64)
	case float64:
		return strconv.FormatFloat(float64(value.(float64)), 'g', -1, 64)
	case bool:
		return strconv.FormatBool(value.(bool))
	default:
		return fmt.Sprintf("%+v", value)
	}
}

func (tree *Map) ToJSON() ([]byte, error) {
	elements := make(map[string]interface{})
	it := tree.Iterator()
	for it.Next() {
		elements[ToString(it.Key())] = it.Value()
	}
	return json.Marshal(&elements)
}

func (tree *Map) FromJSON(data []byte) error {
	elements := make(map[string]interface{})
	err := json.Unmarshal(data, &elements)
	if err == nil {
		tree.Clear()
		for key, value := range elements {
			tree.Put(key, value)
		}
	}
	return err
}
