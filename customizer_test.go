package raven

import "testing"

func Test_CustomizerLoggerInterfaces(t *testing.T) {
	var _ Logger = &ContextLogger{}
	var _ ChildLogger = &ContextLogger{}
}
