package ocitar

import (
	"archive/tar"
	"io"
	"os"
)

func Replace(filename string, replace func(hdr *tar.Header, r io.Reader) (io.Reader, error)) (err error) {
	tmpFilename := filename + ".tmp"

	defer func() {
		if err == nil {
			if e := os.RemoveAll(filename); e != nil {
				err = e
			}
			if e := os.Rename(tmpFilename, filename); e != nil {
				err = e
			}
		}
	}()

	in, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(tmpFilename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer out.Close()

	tw := tar.NewWriter(out)
	defer tw.Close()

	tr := tar.NewReader(in)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		replaced, err := replace(hdr, tr)
		if err != nil {
			return err
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := io.Copy(tw, replaced); err != nil {
			return err
		}
	}
	
	return nil
}
