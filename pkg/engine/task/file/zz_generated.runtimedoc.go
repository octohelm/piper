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

func (v Exists) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "Cwd":
			return []string{
				"current workdir",
			}, true
		case "Path":
			return []string{
				"path",
			}, true
		case "Info":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Exists to check path exists",
	}, true
}

func (v File) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "WorkDir":
			return []string{
				"current work dir",
			}, true
		case "Filename":
			return []string{
				"filename related from current work dir",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Info) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "IsDir":
			return []string{}, true
		case "Mode":
			return []string{}, true
		case "Size":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v ReadAsString) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "File":
			return []string{
				"file",
			}, true
		case "Contents":
			return []string{
				"text contents",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"ReadAsString read file as string",
	}, true
}

func (v ReadAsTable) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "File":
			return []string{
				"file",
			}, true
		case "With":
			return []string{
				"options",
			}, true
		case "Data":
			return []string{
				"file contents",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"ReadAsTable file read as table",
	}, true
}

func (v ReadAsTableOption) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "StrictColNum":
			return []string{
				"strict column num",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v ReadFromJSON) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "File":
			return []string{
				"file",
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
		"ReadFromJSON read and parse json",
	}, true
}

func (v ReadFromTOML) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "File":
			return []string{
				"file",
			}, true
		case "With":
			return []string{
				"options",
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
		"ReadFromTOML read and parse yaml",
	}, true
}

func (v ReadFromTOMLOption) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "AsList":
			return []string{
				"read as list",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v ReadFromYAML) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "File":
			return []string{
				"file",
			}, true
		case "With":
			return []string{
				"options",
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
		"ReadFromYAML read and parse yaml",
	}, true
}

func (v ReadFromYAMLOption) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "AsList":
			return []string{
				"read as list",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v StringOrFile) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "File":
			return []string{}, true
		case "String":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Sync) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "Source":
			return []string{
				"source file",
			}, true
		case "With":
			return []string{
				"sync option",
			}, true
		case "Dest":
			return []string{
				"dest fie",
			}, true
		case "File":
			return []string{
				"synced file same as dest",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Sync file to contents",
	}, true
}

func (v SyncOption) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "MaxConcurrent":
			return []string{
				"once maxConcurrent larger than 1,",
				"file will split to chunk for partially read and write when syncing",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Write) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "OutFile":
			return []string{
				"output file",
			}, true
		case "Contents":
			return []string{
				"file contents",
			}, true
		case "File":
			return []string{
				"writen file",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Write file with contents",
	}, true
}

func (v WriteAsJSON) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "OutFile":
			return []string{
				"output file",
			}, true
		case "Data":
			return []string{
				"data could convert to json",
			}, true
		case "File":
			return []string{
				"writen file",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"WriteAsJSON read and parse json",
	}, true
}

func (v WriteAsTOML) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "OutFile":
			return []string{
				"filename",
			}, true
		case "Data":
			return []string{
				"data could convert to yaml",
			}, true
		case "File":
			return []string{
				"writen file",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"WriteAsYAML read and parse yaml",
	}, true
}

func (v WriteAsYAML) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "OutFile":
			return []string{
				"output file",
			}, true
		case "Data":
			return []string{
				"data could convert to yaml",
			}, true
		case "File":
			return []string{
				"writen file",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"WriteAsYAML read and parse yaml",
	}, true
}
