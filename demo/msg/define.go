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
package msg

//错误码：10000-20000
const (
	Err_LoginErr = uint32(10001) // 登录错误
)

// 消息码：20000+
const (
	//c2s#######################################################
	Msg_Login = uint32(20001) //登录
	Msg_Start = uint32(20002) // 开始

	//s2c#######################################################
	Msg_Login_rst = uint32(30001) //登录返回
	Msg_Start_rst = uint32(30002) // 开始
)
