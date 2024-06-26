/*
Package client GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package client

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v Any) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Value":
			return []string{}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v EnvInterface) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{
				"to avoid added ok",
			}, true
		case "RequiredEnv":
			return []string{}, true
		case "OptionalEnv":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"EnvInterface of client",
	}, true
}

func (v GroupInterface) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Group":
			return []string{}, true
		}
		if doc, ok := runtimeDoc(v.Group, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Group",
	}, true
}

func (v Merge) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "Inputs":
			return []string{}, true
		case "Output":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Merge",
		"read secret value for the secret",
	}, true
}

func (v Module) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "Module":
			return []string{
				"root module",
			}, true
		case "Deps":
			return []string{
				"{ dep: version }",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v ReadSecret) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "Secret":
			return []string{
				"secret ref",
			}, true
		case "Value":
			return []string{
				"secret value",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"ReadSecret",
		"read secret value for the secret",
	}, true
}

func (v RevInfo) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "Version":
			return []string{
				"get pseudo version same of go mod",
				"like",
				"v0.0.0-20231222030512-c093d5e89975",
				"v0.0.0-dirty.0.20231222022414-5f9d1d44dacc",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v Secret) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		}

		return nil, false
	}
	return []string{}, true
}

func (v SecretOrString) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Secret":
			return []string{}, true
		case "Value":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Skip) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "When":
			return []string{}, true
		}

		return nil, false
	}
	return []string{
		"Skip will skip task when matched",
	}, true
}

func (v StringOrBool) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "String":
			return []string{}, true
		case "Bool":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (StringOrBytes) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}

func (StringOrSlice) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}

func (v WaitInterface) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Checkpoint":
			return []string{}, true
		case "Ok":
			return []string{
				"as assertion, one $ok is false",
				"all task should break",
			}, true

		}
		if doc, ok := runtimeDoc(v.Checkpoint, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"WaitInterface for wait task ready",
	}, true
}
