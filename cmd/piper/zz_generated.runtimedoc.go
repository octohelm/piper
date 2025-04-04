/*
Package main GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package main

func (v *Do) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		}
		if doc, ok := runtimeDoc(&v.Logger, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(&v.Pipeline, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Init) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		}
		if doc, ok := runtimeDoc(&v.InitRun, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *InitRun) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Tidy) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
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
