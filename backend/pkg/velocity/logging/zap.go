package logging

import (
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
)

var logger *zap.Logger
var once sync.Once

func GetLogger() *zap.Logger {
	once.Do(func() {
		if strings.ToLower(os.Getenv("DEBUG")) == "true" {
			logger, _ = zap.NewDevelopment()
		} else {
			logger, _ = zap.NewProduction()
		}
	})
	return logger
}
