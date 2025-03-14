/*
Package task GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package task

func (v *Checkpoint) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		}
		if doc, ok := runtimeDoc(&v.Task, "no need the ok", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Group) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		}
		if doc, ok := runtimeDoc(&v.Task, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Secret) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Key":
			return []string{}, true
		case "Value":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v *SetupTask) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		}
		if doc, ok := runtimeDoc(&v.Task, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Task) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Dep":
			return []string{
				"task hook to make task could run after some others",
			}, true
		case "Ok":
			return []string{
				"task result",
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
