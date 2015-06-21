package logs

import (
	"testing"
)

func TestLoggers(t *testing.T) {
	Trace.Println("Test Trace")
	Info.Println("Test Info")
	Error.Println("Test Error")
}
