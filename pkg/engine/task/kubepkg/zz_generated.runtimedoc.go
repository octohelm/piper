/*
Package kubepkg GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package kubepkg

func (v *Apply) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Kubeconfig":
			return []string{
				"path",
			}, true
		case "Manifests":
			return []string{
				"of k8s resources",
			}, true

		}
		if doc, ok := runtimeDoc(&v.Task, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"to kubernetes",
	}, true
}

func (v *Cueify) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "KubePkg":
			return []string{
				"spec",
			}, true
		case "PkgName":
			return []string{
				"pkg name",
			}, true
		case "OutDir":
			return []string{
				"for cue files",
			}, true

		}
		if doc, ok := runtimeDoc(&v.Task, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *KubePkg) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Spec":
			return []string{}, true
		case "Status":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(&v.TypeMeta, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(&v.ObjectMeta, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v *Manifests) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "KubePkg":
			return []string{
				"spec",
			}, true
		case "Rename":
			return []string{
				"for image repo name",
				"go template rule",
				"`{{ .registry }}/{{ .namespace }}/{{ .name }}`",
			}, true
		case "Recursive":
			return []string{
				"recursively extract KubePkg in sub manifests",
			}, true
		case "Manifests":
			return []string{
				"of k8s resources",
			}, true

		}
		if doc, ok := runtimeDoc(&v.Task, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"extract manifests from KubePkg",
	}, true
}

func (v *OciTar) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "KubePkg":
			return []string{
				"spec",
			}, true
		case "Platforms":
			return []string{
				"of oci tar, if empty it will based on KubePkg",
			}, true
		case "WithAnnotations":
			return []string{
				"pick annotations of KubePkg as image annotations",
			}, true
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

func (v *PushOciTar) RuntimeDoc(names ...string) ([]string, bool) {
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
