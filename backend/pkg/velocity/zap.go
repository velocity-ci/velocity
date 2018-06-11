package velocity

import (
	"os"
	"sync"

	"go.uber.org/zap"
)

func SetLogLevel() {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		// glog.SetLevel(glog.DebugLevel)
		break
	case "info":
		// glog.SetLevel(glog.InfoLevel)
		break
	case "warn":
		// glog.SetLevel(glog.WarnLevel)
		break
	case "error":
		// glog.SetLevel(glog.ErrorLevel)
		break
	default:
		// glog.SetLevel(glog.InfoLevel)
	}
}

var logger *zap.Logger
var once sync.Once

func GetLogger() *zap.Logger {
	once.Do(func() {
		logger, _ = zap.NewProduction()
	})
	return logger
}
