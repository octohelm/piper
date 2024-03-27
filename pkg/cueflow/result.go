package cueflow

type CanSuccess interface {
	Success() bool
}

type ResultValuer interface {
	ResultValue() map[string]any
}
