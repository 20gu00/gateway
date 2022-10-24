package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/20gu00/gateway/common/lib"
	"github.com/gin-gonic/gin"
	"strings"
)

type ResponseCode int //响应码,方便开发过程快速定位错误,注意这个跟http请求的响应码不同,这是用来程序内部处理错误的

//公用的中间件,统一处理错误
const (
	SuccessCode ResponseCode = iota //0
	UndefErrorCode
	ValidErrorCode
	InternalErrorCode                    //3
	InvalidRequestErrorCode ResponseCode = 401
	CustomizeCode           ResponseCode = 1000
	GROUPALL_SAVE_FLOWERROR ResponseCode = 2001
)

type Response struct {
	ErrorCode ResponseCode `json:"errno"`  //响应码
	ErrorMsg  string       `json:"errmsg"` //响应消息
	Data      interface{}  `json:"data"`
	TraceId   interface{}  `json:"trace_id"` //追踪id
	Stack     interface{}  `json:"stack"`    //堆栈信息
}

func ResponseError(ctx *gin.Context, code ResponseCode, err error) {
	trace, _ := ctx.Get("trace")                 //获取链路追踪的value
	traceContext, _ := trace.(*lib.TraceContext) //断言得到追踪上下文
	traceId := ""
	if traceContext != nil {
		traceId = traceContext.TraceId //获取链路追踪的id
	}

	stack := ""
	//是不是debug模式或者开发模式
	if ctx.Query("is_debug") == "1" || lib.GetConfEnv() == "dev" {
		stack = strings.Replace(fmt.Sprintf("%+v", err), err.Error()+"\n", "", -1) //全部替换,%+v将字段的值一块打印
	}

	//包装请求处理的响应的信息
	resp := &Response{ErrorCode: code, ErrorMsg: err.Error(), Data: "", TraceId: traceId, Stack: stack}
	ctx.JSON(200, resp)                   //json
	response, _ := json.Marshal(resp)     //json编码
	ctx.Set("response", string(response)) //context设置key value
	ctx.AbortWithError(200, err)          //有错误就停止context链,不再向子context传递
}

func ResponseSuccess(ctx *gin.Context, data interface{}) {
	trace, _ := ctx.Get("trace")
	traceContext, _ := trace.(*lib.TraceContext)
	traceId := ""
	if traceContext != nil {
		traceId = traceContext.TraceId
	}

	//响应成功的编码0
	resp := &Response{ErrorCode: SuccessCode, ErrorMsg: "", Data: data, TraceId: traceId}
	ctx.JSON(200, resp)
	response, _ := json.Marshal(resp)
	ctx.Set("response", string(response)) //继续传递
}
