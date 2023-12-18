package cueflow

import "cuelang.org/go/cue"

type Scope interface {
	Value() Value
	Fill(path cue.Path, value Value) error
	Processed(path cue.Path) bool
}
