package otel

import (
	telemetry "github.com/dagger/otel-go"
)

const (
	LogAttrScope = "$scope"

	LogAttrProgressCurrent = telemetry.ProgressCurrentAttr
	LogAttrProgressTotal   = telemetry.ProgressTotalAttr
)
