package decorator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"strings"

	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C any, R any] struct {
	logger *logrus.Logger
	base   QueryHandler[C, R]
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	fields := logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": string(body),
	}

	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "Query executed successfully")
		} else {
			logging.Errorf(ctx, fields, "Failed to execute query, err=%s", err)
		}
	}()
	result, err = q.base.Handle(ctx, cmd)
	return result, err
}

type commandLoggingDecorator[C any, R any] struct {
	logger *logrus.Logger
	base   CommandHandler[C, R]
}

func (c commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	fields := logrus.Fields{
		"command":      generateActionName(cmd),
		"command_body": string(body),
	}

	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "Command executed successfully")
		} else {
			logging.Errorf(ctx, fields, "Failed to execute command, err=%s", err)
		}
	}()
	result, err = c.base.Handle(ctx, cmd)
	return result, err
}

func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
