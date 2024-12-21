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

package common

const BROATCAST = "broatcast"
const REQUEST = "req"
const PUBLISH = "pub"
const SEND = "send"

type RouteType int8

const (
	Default RouteType = iota // 0
	Hash
	Direct
)

var errCodeMap = map[uint32]string{
	Err_OK:                 "Success",
	Err_ParameterIllegal:   "参数不合法",
	Err_UnauthorizedUserId: "非法的用户Id",
	Err_Unauthorized:       "未授权",
	Err_ServerError:        "没有数据",
	Err_NotData:            "系统错误",
	Err_OperationFailure:   "操作失败",
	Err_RoutingNotExist:    "路由不存在",
	Err_HandlerFuncRetErr:  "路由方法返回错误",
	Err_HandlerFuncErr:     "路由方法错误",
}

// 保留错误码：小于 600
const (
	Default_Frame_MAX      = uint32(1000) // 框架保留消息码最大值
	Error_MAX              = uint32(600)  // 框架错误保留最大值
	Err_OK                 = uint32(200)  // Success
	Err_ParameterIllegal   = uint32(201)  // 参数不合法
	Err_UnauthorizedUserId = uint32(202)  // 非法的用户Id
	Err_Unauthorized       = uint32(203)  // 未授权
	Err_ServerError        = uint32(204)  // 系统错误
	Err_NotData            = uint32(205)  // 没有数据
	Err_OperationFailure   = uint32(206)  // 操作失败
	Err_RoutingNotExist    = uint32(207)  // 路由不存在
	Err_HandlerFuncRetErr  = uint32(208)  // 路由方法返回错误
	Err_HandlerFuncErr     = uint32(209)  // 路由方法错误
	Err_SeedEncoderErr     = uint32(210)  // 加密秘钥生成错误
	Err_SeedDecoderErr     = uint32(211)  // 解密秘钥生成错误
	Err_RpcRequestErr      = uint32(212)  // rpc请求错误
)

// 保留消息码：大于600  小于1000
const (
	//c2s#######################################################
	Msg_Heartbeat       = uint32(701) //心跳
	Msg_Verify          = uint32(702) //身份校验
	Msg_Kick            = uint32(704) //踢出用户
	Msg_RegBroatCast    = uint32(706) //注册广播
	Msg_DisregBroatCast = uint32(707) //取消广播
	Msg_ReRoute         = uint32(708) //重置
	//s2c#######################################################
	Msg_Heartbeat_Rst   = uint32(801) //心跳返回
	Msg_Verify_Rst_Suc  = uint32(802) //身份校验成功
	Msg_Verify_Rst_Fail = uint32(803) //身份校验失败
	Msg_Kick_Rst        = uint32(804) //踢出用户
	Msg_BroatCast_Rst   = uint32(805) //系统广播

)

// 根据错误码 获取错误信息
func GetErrorMessage(code uint32, message string) string {
	var codeMessage string
	if message == "" {
		if value, ok := errCodeMap[code]; ok {
			codeMessage = value
		} else {
			codeMessage = "未定义错误类型!"
		}
	} else {
		codeMessage = message
	}

	return codeMessage
}
