/*
Package engine GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package engine

func (*LogLevel) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}

func (v *Logger) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "LogLevel":
			return []string{
				"Log level",
			}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Pipeline) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Action":
			return []string{}, true
		case "Project":
			return []string{
				"plan root file",
				"and the dir of the root file will be the cwd for all cue files",
			}, true
		case "CacheDir":
			return []string{
				"cache dir root",
				"for cache files",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

// nolint:deadcode,unused
func runtimeDoc(v any, prefix string, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		doc, ok := c.RuntimeDoc(names...)
		if ok {
			if prefix != "" && len(doc) > 0 {
				doc[0] = prefix + doc[0]
				return doc, true
			}

			return doc, true
		}
	}
	return nil, false
}
