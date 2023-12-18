package cueflow

import "github.com/go-courier/logr"

type CanSuccess interface {
	Success() bool
}

type Result struct {
	//
	Ok bool `json:"ok"`
	// when not ok, should include error
	Reason string `json:"reason,omitempty"`
}

func (r *Result) Done(err error) {
	r.Ok = err == nil

	if !r.Ok {
		r.Reason = err.Error()
	}
}

func (r Result) Success() bool {
	return r.Ok
}

type WithResultValue interface {
	ResultValue() any
}

func logTaskResult(l logr.Logger, v any) {
	if x, ok := v.(WithResultValue); ok {
		ret := x.ResultValue()
		if r, ok := ret.(CanSuccess); ok {
			if r.Success() {
				l.WithValues("result", CueLogValue(ret)).Debug("success.")
				return
			}
			l.WithValues("result", CueLogValue(ret)).Debug("failed.")
			return
		}
		l.WithValues("result", CueLogValue(ret)).Debug("done.")
	}
}
