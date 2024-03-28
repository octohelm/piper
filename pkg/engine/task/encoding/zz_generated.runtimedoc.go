/*
Package file GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package file

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v MarshalTOML) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "Data":
			return []string{
				"data",
			}, true
		case "Contents":
			return []string{
				"raw",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"MarshalTOML",
	}, true
}

func (v UnmarshalTOML) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "Contents":
			return []string{
				"toml raw",
			}, true
		case "Data":
			return []string{
				"data",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"UnmarshalTOML",
	}, true
}