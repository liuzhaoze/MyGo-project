package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

func RequestLog(e *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestIn(c, e)
		defer requestOut(c, e)
		c.Next()
	}
}

func requestIn(c *gin.Context, e *logrus.Entry) {
	now := time.Now()
	body := c.Request.Body
	bodyBytes, _ := io.ReadAll(body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	var compactJson bytes.Buffer
	_ = json.Compact(&compactJson, bodyBytes)
	c.Set("request_start", now)
	e.WithContext(c.Request.Context()).WithFields(logrus.Fields{
		"start":      now.Unix(),
		logging.Args: compactJson.String(),
		"from":       c.RemoteIP(),
		"uri":        c.Request.RequestURI,
	}).Info("_request_in")
}

func requestOut(c *gin.Context, e *logrus.Entry) {
	resp, _ := c.Get("response")
	start, _ := c.Get("request_start")
	startTime := start.(time.Time)
	e.WithContext(c.Request.Context()).WithFields(logrus.Fields{
		logging.Cost:     time.Since(startTime).Milliseconds(),
		logging.Response: resp,
	}).Info("_request_out")
}
