/*
Package internal GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package internal

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v Controller) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Prefix":
			return []string{}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (OptionFunc) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
