package file

import (
	"io"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-billy.v4"
)

// Copy will copy all files and directories over
func Copy(fs billy.Filesystem, src, dst string) error {
	info, err := fs.Lstat(src)
	if err != nil {
		return err
	}
	return copyAll(fs, src, dst, info)
}

func copyAll(fs billy.Filesystem, src, dest string, info os.FileInfo) error {
	if info.IsDir() {
		return dirCopy(fs, src, dest)
	}
	return fileCopy(fs, src, dest)
}

func fileCopy(fs billy.Filesystem, src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0744); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), 0644); err != nil {
		return err
	}

	s, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func dirCopy(fs billy.Filesystem, srcdir, destdir string) error {
	if err := os.MkdirAll(destdir, 0744); err != nil {
		return err
	}

	contents, err := fs.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := copyAll(fs, cs, cd, content); err != nil {
			return err
		}
	}
	return nil
}
