/*
Package wd GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package wd

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v CurrentWorkDir) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Cwd":
			return []string{
				"current word dir",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Local) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "SetupTask":
			return []string{}, true
		case "Dir":
			return []string{
				"related dir on the root of project",
			}, true
		case "WorkDir":
			return []string{
				"the local work dir",
			}, true

		}
		if doc, ok := runtimeDoc(v.SetupTask, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Local",
		"create a local work dir",
	}, true
}

func (v Platform) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "OS":
			return []string{}, true
		case "Architecture":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Release) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{
				"OS name",
			}, true
		case "Version":
			return []string{
				"OS Version",
			}, true
		case "ID":
			return []string{
				"OS id, like `ubuntu` `windows`",
			}, true
		case "IDLike":
			return []string{
				"if os is based on some upstream",
				"like debian when id is `ubuntu`",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v SSH) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "SetupTask":
			return []string{}, true
		case "Config":
			return []string{
				"path to ssh config",
			}, true
		case "HostKey":
			return []string{
				"host key of ssh config",
			}, true
		case "Address":
			return []string{
				"custom setting",
				"ssh address",
			}, true
		case "Port":
			return []string{
				"ssh port",
			}, true
		case "IdentityFile":
			return []string{
				"ssh identity file",
			}, true
		case "User":
			return []string{
				"ssh user",
			}, true
		case "WorkDir":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.SetupTask, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"SSH",
		"create ssh work dir for remote executing",
	}, true
}

func (v Su) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "CurrentWorkDir":
			return []string{}, true
		case "User":
			return []string{
				"new user",
			}, true
		case "WorkDir":
			return []string{
				"new work dir",
			}, true

		}
		if doc, ok := runtimeDoc(v.CurrentWorkDir, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Su",
		"switch user",
	}, true
}

func (v Sub) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "CurrentWorkDir":
			return []string{}, true
		case "Dir":
			return []string{
				"related path from current work dir",
			}, true
		case "WorkDir":
			return []string{
				"new work dir",
			}, true

		}
		if doc, ok := runtimeDoc(v.CurrentWorkDir, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Sub",
		"create new work dir base on current work dir",
	}, true
}

func (v SysInfo) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "CurrentWorkDir":
			return []string{}, true
		case "Release":
			return []string{
				"os release info",
			}, true
		case "Platform":
			return []string{
				"os platform",
			}, true

		}
		if doc, ok := runtimeDoc(v.CurrentWorkDir, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"SysInfo",
		"get sys info of current work dir",
	}, true
}

func (v WorkDir) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {

		}

		return nil, false
	}
	return []string{}, true
}
