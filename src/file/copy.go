package file

import (
	"hash"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-billy.v4"
)

// Copy will copy all files and directories over
func Copy(fs billy.Filesystem, src, dst string, sum hash.Hash) error {
	info, err := fs.Lstat(src)
	if err != nil {
		return err
	}
	return copyAll(fs, src, dst, info, sum)
}

func copyAll(fs billy.Filesystem, src, dest string, info os.FileInfo, sum hash.Hash) error {
	if info.IsDir() {
		return dirCopy(fs, src, dest, sum)
	}
	return fileCopy(fs, src, dest, sum)
}

func fileCopy(fs billy.Filesystem, src, dest string, sum hash.Hash) error {
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

	if _, err = io.Copy(f, s); err != nil {
		return err
	}

	s.Seek(0, io.SeekStart)
	_, err = io.Copy(sum, s)
	return err
}

func dirCopy(fs billy.Filesystem, srcdir, destdir string, sum hash.Hash) error {
	if err := os.MkdirAll(destdir, 0744); err != nil {
		return err
	}

	contents, err := fs.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := copyAll(fs, cs, cd, content, sum); err != nil {
			return err
		}
	}
	return nil
}
