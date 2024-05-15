package otel

import "github.com/dagger/dagger/telemetry"

const (
	LogAttrScope = "$scope"

	LogAttrProgressCurrent = telemetry.ProgressCurrentAttr
	LogAttrProgressTotal   = telemetry.ProgressTotalAttr
)
