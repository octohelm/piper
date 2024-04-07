package logger

import (
	contextx "github.com/octohelm/x/context"
	"go.opentelemetry.io/otel/trace"
)

var TracerContext = contextx.New[trace.Tracer]()
