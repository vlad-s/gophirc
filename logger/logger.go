package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Log is the Logger instance used to log everything here & in the clients.
var Log = logrus.New()

func init() {
	Log.SetLevel(logrus.DebugLevel)
	Log.Out = os.Stdout
}

// Fields is a wrapper to logrus' `Fields{}` in order to return a `logrus.Fields`
// structure without importing the package.
func Fields(args map[string]interface{}) logrus.Fields {
	fields := make(map[string]interface{})
	for k, v := range args {
		fields[k] = v
	}
	return fields
}
