package plan

import contextx "github.com/octohelm/x/context"

var ContextContext = contextx.New[Context]()

type Context interface {
	PlanRoot() string
}
