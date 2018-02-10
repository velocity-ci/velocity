package velocity

import (
	"os"

	"github.com/Sirupsen/logrus"
)

func SetLogLevel() {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
		break
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
