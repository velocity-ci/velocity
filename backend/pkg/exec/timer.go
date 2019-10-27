package exec

import (
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
)

// Time logs out how long has elapsed since the given start time. Useful for debugging function timing information
func Time(start time.Time, name string) {
	elapsed := time.Since(start)
	logging.GetLogger().Debug("timed", zap.String("name", name), zap.Duration("duration", elapsed))
}
