package common

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/liuzhaoze/MyGo-project/common/handler/errors"
	"github.com/liuzhaoze/MyGo-project/common/tracing"
	"net/http"
)

type BaseResponse struct {
}

func (b *BaseResponse) Response(c *gin.Context, err error, data interface{}) {
	if err != nil {
		b.error(c, err)
	} else {
		b.success(c, data)
	}
}

func (b *BaseResponse) success(c *gin.Context, data interface{}) {
	errno, errmsg := errors.Output(nil)
	resp := response{
		ErrorCode: errno,
		Message:   errmsg,
		Data:      data,
		TraceID:   tracing.TraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, resp)

	respJson, _ := json.Marshal(resp)
	c.Set("response", respJson)
}

func (b *BaseResponse) error(c *gin.Context, err error) {
	errno, errmsg := errors.Output(err)
	resp := response{
		ErrorCode: errno,
		Message:   errmsg,
		Data:      nil,
		TraceID:   tracing.TraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, resp)

	respJson, _ := json.Marshal(resp)
	c.Set("response", respJson)
}

type response struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	TraceID   string `json:"trace_id"`
}
