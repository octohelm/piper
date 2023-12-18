package wd

import (
	"context"
	"io"
	"io/fs"
	"os"

	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/webdav"
)

func WrapSFTP(c *ssh.Client) (filesystem.FileSystem, error) {
	cc, err := sftp.NewClient(c)
	if err != nil {
		return nil, err
	}
	return &sftpFs{c: cc}, nil
}

type sftpFs struct {
	c *sftp.Client
}

func (fs *sftpFs) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return fs.c.Mkdir(name)
}

func (fs *sftpFs) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	f, err := fs.c.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}

	return &sftpFile{
		File: f,
		c:    fs.c,
		path: name,
	}, nil
}

type sftpFile struct {
	*sftp.File
	c    *sftp.Client
	path string
}

func (f *sftpFile) Truncate(size int64) error {
	return f.File.Truncate(size)
}

func (f *sftpFile) ReadFrom(r io.Reader) (int64, error) {
	return f.File.ReadFrom(r)
}

func (f *sftpFile) Readdir(count int) ([]fs.FileInfo, error) {
	list, err := f.c.ReadDir(f.path)
	if err != nil {
		return nil, err
	}

	if count > 0 && count < len(list) {
		final := make([]fs.FileInfo, 0, count)
		for i := 0; i < count; i++ {
			final = append(final, list[i])
		}
		return final, nil
	}

	return list, nil
}

func (fs *sftpFs) RemoveAll(ctx context.Context, name string) error {
	return fs.c.RemoveAll(name)
}

func (fs *sftpFs) Rename(ctx context.Context, oldName, newName string) error {
	return fs.c.Rename(oldName, newName)
}

func (fs *sftpFs) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return fs.c.Stat(name)
}
