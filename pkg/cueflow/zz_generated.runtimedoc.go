/*
Package cueflow GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package cueflow

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v Progress) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Current":
			return []string{}, true
		case "Total":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (TaskOptionFunc) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
