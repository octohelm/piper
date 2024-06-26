/*
Package archive GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package archive

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v Tar) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "SrcDir":
			return []string{
				"specified dir for tar",
			}, true
		case "OutFile":
			return []string{
				"tar out filename base on the current work dir",
			}, true
		case "File":
			return []string{
				"created tarfile",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Tar",
		"make a tar archive file of specified dir",
	}, true
}

func (v UnTar) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Task":
			return []string{}, true
		case "SrcFile":
			return []string{
				"tar filename base on the current work outDir",
			}, true
		case "ContentEncoding":
			return []string{
				"tar file content encoding",
			}, true
		case "OutDir":
			return []string{
				"output outDir for tar",
			}, true
		case "Dir":
			return []string{
				"final dir contains tar files",
			}, true

		}
		if doc, ok := runtimeDoc(v.Task, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"UnTar",
		"un tar files into specified outDir",
	}, true
}
