package decorator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C any, R any] struct {
	logger *logrus.Entry
	base   QueryHandler[C, R]
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	logger := q.logger.WithFields(logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": string(body),
	})
	logger.Debug("Executing query")

	defer func() {
		if err == nil {
			logger.Info("Query succeeded")
		} else {
			logger.Error("Failed to execute query", err)
		}
	}()
	return q.base.Handle(ctx, cmd)
}

type commandLoggingDecorator[C any, R any] struct {
	logger *logrus.Entry
	base   CommandHandler[C, R]
}

func (c commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	logger := c.logger.WithFields(logrus.Fields{
		"command":      generateActionName(cmd),
		"command_body": string(body),
	})
	logger.Debug("Executing command")

	defer func() {
		if err == nil {
			logger.Info("Command succeeded")
		} else {
			logger.Error("Failed to execute command", err)
		}
	}()
	return c.base.Handle(ctx, cmd)
}

func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
