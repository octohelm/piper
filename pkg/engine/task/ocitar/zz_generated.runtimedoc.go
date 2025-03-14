/*
Package ocitar GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package ocitar

func (v *PackExecutable) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Dest":
			return []string{}, true
		case "Files":
			return []string{
				"[Platform]: _",
			}, true
		case "Annotations":
			return []string{}, true
		case "OutFile":
			return []string{
				"of OciTar",
			}, true
		case "File":
			return []string{
				"of tar created",
			}, true

		}
		if doc, ok := runtimeDoc(&v.Task, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Pull) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Source":
			return []string{
				"image from",
			}, true
		case "Platforms":
			return []string{
				"of oci tar, if empty it will based on KubePkg",
			}, true
		case "Annotations":
			return []string{}, true
		case "Rename":
			return []string{
				"for image repo name",
				"go template rule",
				"`{{ .registry }}/{{ .namespace }}/{{ .name }}`",
			}, true
		case "OutFile":
			return []string{
				"of OciTar",
			}, true
		case "File":
			return []string{
				"of tar created",
			}, true

		}
		if doc, ok := runtimeDoc(&v.Task, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Push) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "SrcFile":
			return []string{
				"of oci tar",
			}, true
		case "RemoteURL":
			return []string{
				"of container registry",
			}, true
		case "Rename":
			return []string{
				"for image repo name",
				"go template rule",
				"`{{ .registry }}/{{ .namespace }}/{{ .name }}`",
			}, true

		}
		if doc, ok := runtimeDoc(&v.Task, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Rename) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		}
		if doc, ok := runtimeDoc(&v.Renamer, "", names...); ok {
			return doc, ok
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
