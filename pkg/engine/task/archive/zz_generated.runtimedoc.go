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
		case "CurrentWorkDir":
			return []string{}, true
		case "Filename":
			return []string{
				"final tar filename base on the current work dir",
			}, true
		case "Dir":
			return []string{
				"specified dir for tar",
			}, true
		case "Output":
			return []string{
				"output tar file when created",
				"just group cwd and filename",
			}, true

		}
		if doc, ok := runtimeDoc(v.CurrentWorkDir, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Tar",
		"make a tar archive file of specified dir",
	}, true
}
