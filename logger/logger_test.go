package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestFields(t *testing.T) {
	m := map[string]interface{}{
		"a": "b",
		"b": 1,
		"c": true,
	}

	f := Fields(m)

	if len(f) != len(m) {
		t.Errorf("Fields %q have length %q, expected %q.", f, len(f), len(m))
	}

	switch v := interface{}(f).(type) {
	case logrus.Fields:
	default:
		t.Errorf("Fields %q have type %q, expected %q.", f, v, logrus.Fields{})
	}
}
