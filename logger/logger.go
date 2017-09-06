package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func init() {
	Log.SetLevel(logrus.DebugLevel)
	Log.Out = os.Stdout
}

func Fields(args map[string]interface{}) logrus.Fields {
	fields := make(map[string]interface{})
	for k, v := range args {
		fields[k] = v
	}
	return fields
}
