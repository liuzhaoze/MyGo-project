package logging

import (
	"context"
	"encoding/json"
	"github.com/liuzhaoze/MyGo-project/common/util"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	Method   = "method"
	Args     = "args"
	Cost     = "cost_ms"
	Response = "response"
	Error    = "err"
)

type ArgFormatter interface {
	FormatArg() (string, error)
}

func WhenMySQL(ctx context.Context, method string, args ...any) (logrus.Fields, func(any, *error)) {
	fields := logrus.Fields{
		Method: method,
		Args:   formatArgs(args),
	}
	start := time.Now()
	return fields, func(resp any, err *error) {
		level, msg := logrus.InfoLevel, "mysql_success"
		fields[Cost] = time.Since(start).Milliseconds()
		fields[Response] = resp

		if err != nil && (*err != nil) {
			level, msg = logrus.ErrorLevel, "mysql_error"
			fields[Error] = (*err).Error()
		}

		logf(ctx, level, fields, "%s", msg)
	}
}

func formatArgs(args []any) string {
	var item []string
	for _, arg := range args {
		item = append(item, formatArg(arg))
	}
	return strings.Join(item, " || ")
}

func formatArg(arg any) string {
	var (
		str string
		err error
	)

	defer func() {
		if err != nil {
			str = "unsupported type in formatArg || err=" + err.Error()
		}
	}()

	switch v := arg.(type) {
	case ArgFormatter:
		str, err = util.MarshalString(v)
	default:
		bytes, e := json.Marshal(v)
		str, err = string(bytes), e
	}

	return str
}
