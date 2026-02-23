package logger

import (
	"github.com/demotask/backend/utilities"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Entry
}

func NewLogger() *Logger {
	logger := logrus.New()
	newEntry := logrus.NewEntry(logger)
	return &Logger{newEntry}
}

func (l *Logger) Request(requestID string) *Logger {
	if requestID == "" {
		requestID = utilities.RandomString(20)
	}

	formatter := runtime.Formatter{
		File:         true,
		Package:      true,
		BaseNameOnly: true,
		Line:         true,
		ChildFormatter: &logrus.JSONFormatter{
			DataKey:     requestID,
			PrettyPrint: false,
		}}
	l.Logger.SetFormatter(&formatter)

	return l
}
