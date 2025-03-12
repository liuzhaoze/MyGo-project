package common

import (
	"github.com/gin-gonic/gin"
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
	c.JSON(http.StatusOK, response{
		ErrorCode: 0,
		Message:   "success",
		Data:      data,
		TraceID:   tracing.TraceID(c.Request.Context()),
	})
}

func (b *BaseResponse) error(c *gin.Context, err error) {
	c.JSON(http.StatusOK, response{
		ErrorCode: 2,
		Message:   err.Error(),
		Data:      nil,
		TraceID:   tracing.TraceID(c.Request.Context()),
	})
}

type response struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	TraceID   string `json:"trace_id"`
}
