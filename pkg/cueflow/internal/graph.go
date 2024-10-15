package internal

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
)

func printGraph(scope string, nodes []Node) {
	buffer := bytes.NewBuffer(nil)

	w, err := zlib.NewWriterLevel(buffer, 9)
	if err != nil {
		panic(fmt.Errorf("fail to write: %w", err))
	}

	wrap := func(name string) string {
		return fmt.Sprintf("%q", name)
	}

	_, _ = fmt.Fprintf(w, `direction: right
`)

	for _, n := range nodes {
		for _, d := range n.Deps() {
			_, _ = fmt.Fprintf(w, `
%s -> %s
`, wrap(d.String()), wrap(n.String()))
		}
	}
	_ = w.Close()

	url := fmt.Sprintf("https://kroki.io/d2/svg/%s?theme=101", base64.URLEncoding.EncodeToString(buffer.Bytes()))

	fmt.Println(scope, url)
}
