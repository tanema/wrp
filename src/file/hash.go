package file

import (
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// Sum will add to a hash
func Sum(src string, sum hash.Hash) error {
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return sumAll(src, info, sum)
}

func sumAll(src string, info os.FileInfo, sum hash.Hash) error {
	if info.IsDir() {
		return dirSum(src, sum)
	}
	return fileSum(src, sum)
}

func fileSum(src string, sum hash.Hash) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	_, err = io.Copy(sum, s)
	return err
}

func dirSum(srcdir string, sum hash.Hash) error {
	paths := []string{}
	filepath.Walk(srcdir, func(path string, info os.FileInfo, err error) error {
		if srcdir != path && info.Mode().IsRegular() {
			paths = append(paths, path)
		}
		return nil
	})
	sort.Strings(paths)
	for _, path := range paths {
		if err := fileSum(path, sum); err != nil {
			return err
		}
	}
	return nil
}
