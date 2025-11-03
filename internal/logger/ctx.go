package logger

import (
	"go.opentelemetry.io/otel/trace"

	contextx "github.com/octohelm/x/context"
)

var TracerContext = contextx.New[trace.Tracer]()
