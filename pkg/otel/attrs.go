package otel

import (
	"dagger.io/dagger/telemetry"
)

const (
	LogAttrScope = "$scope"

	LogAttrProgressCurrent = telemetry.ProgressCurrentAttr
	LogAttrProgressTotal   = telemetry.ProgressTotalAttr
)
