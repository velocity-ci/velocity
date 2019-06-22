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
		var config zap.Config
		if strings.ToLower(os.Getenv("DEBUG")) == "true" {
			config = zap.NewDevelopmentConfig()
		} else {
			config = zap.NewProductionConfig()
		}
		// config.OutputPaths = []string{
		// 	// "vcli.log",
		// }
		logger, _ = config.Build()
	})
	return logger
}
