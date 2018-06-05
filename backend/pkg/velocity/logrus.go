package velocity

import (
	"os"
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
