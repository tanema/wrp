package config

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/tanema/wrp/src/file"
)

var errUnableToFindTag = fmt.Errorf("unable to find tag for repo")

type verType int

var (
	hashVersion   verType = 1 // iota is undefined for some reason?
	tagVersion    verType = 2
	branchVersion verType = 3
)

// Dependency describes a single dependency to fetch and files inside to keep
type Dependency struct {
	Pick []string `json:"pick,omitempty"`

	versionType verType
	version     string
	url         string
	fs          billy.Filesystem
}

func (dep *Dependency) fetch(url string) error {
	var err error
	dep.url, dep.versionType, dep.version, err = parseURL(url)
	if err != nil {
		return err
	}
	dep.fs, err = dep.getState()
	return err
}

// name@tag
// name#branch
// name!hash
func parseURL(repourl string) (string, verType, string, error) {
	var versionType verType
	var version string
	var splitURL string
	var parts []string
	if parts = strings.SplitN(repourl, "@", 2); len(parts) > 1 {
		versionType = tagVersion
	} else if parts = strings.SplitN(repourl, "#", 2); len(parts) > 1 {
		versionType = branchVersion
	} else if parts = strings.SplitN(repourl, "!", 2); len(parts) > 1 {
		versionType = hashVersion
	} else {
		versionType = hashVersion
	}
	splitURL = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}

	if strings.HasPrefix(splitURL, "http://") || strings.HasPrefix(splitURL, "git@") {
		return "", versionType, version, fmt.Errorf("invalid dependency url %v please only use domain", splitURL)
	}
	_, err := url.Parse("https://" + splitURL)
	if err != nil {
		return "", versionType, version, fmt.Errorf("cannot parse repo url %v, received err: %v", splitURL, err)
	}
	return splitURL, versionType, version, nil
}

func (dep *Dependency) formatURL() string {
	switch dep.versionType {
	case tagVersion:
		return dep.url + "@" + dep.version
	case branchVersion:
		return dep.url + "#" + dep.version
	case hashVersion:
		return dep.url + "!" + dep.version
	}
	return ""
}

func (dep *Dependency) write(dest string) error {
	if len(dep.Pick) > 0 {
		for _, pick := range dep.Pick {
			if err := file.Copy(dep.fs, pick, filepath.Join(dest, pick)); err != nil {
				return err
			}
		}
	} else {
		baseName := strings.TrimSuffix(path.Base(dep.url), ".git")
		if err := file.Copy(dep.fs, ".", filepath.Join(dest, baseName)); err != nil {
			return err
		}
	}
	return nil
}

func (dep *Dependency) remove(dest string) error {
	if len(dep.Pick) > 0 {
		for _, pick := range dep.Pick {
			if err := os.RemoveAll(filepath.Join(dest, pick)); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	} else {
		baseName := strings.TrimSuffix(path.Base(dep.url), ".git")
		if err := os.RemoveAll(filepath.Join(dest, baseName)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (dep *Dependency) getState() (billy.Filesystem, error) {
	fs := memfs.New()

	repo, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{URL: "https://" + dep.url})
	if err != nil {
		return fs, fmt.Errorf("problem cloning repo: %v", err)
	}

	fetchOptions := &git.FetchOptions{RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}}
	if err := repo.Fetch(fetchOptions); err != nil {
		return fs, err
	}

	tree, err := repo.Worktree()
	if err != nil {
		return fs, err
	}

	if dep.version == "" {
		if dep.version, err = getRepoTag(repo); err == errUnableToFindTag {
			if dep.version, err = getHeadHash(repo); err != nil {
				return fs, fmt.Errorf("unable to pin dependency: %v", err)
			}
		} else if err != nil {
			return fs, fmt.Errorf("problem fetching tags: %v", err)
		} else {
			dep.versionType = tagVersion
		}
	}

	switch dep.versionType {
	case tagVersion:
		return fs, tree.Checkout(&git.CheckoutOptions{Branch: plumbing.NewTagReferenceName(dep.version)})
	case branchVersion:
		return fs, tree.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(dep.version)})
	case hashVersion:
		return fs, tree.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(dep.version)})
	}

	return fs, nil
}

func getRepoTag(repo *git.Repository) (string, error) {
	iter, err := repo.Tags()
	if err != nil {
		return "", err
	}
	versions := []semver.Version{}
	versionMap := map[string]string{}
	iter.ForEach(func(ref *plumbing.Reference) error {
		tagName := strings.TrimPrefix(ref.Name().String(), "refs/tags/")
		version, err := semver.ParseTolerant(tagName)
		if err == nil {
			versionMap[version.String()] = tagName
			versions = append(versions, version)
		}
		return nil
	})

	if len(versions) > 0 {
		semver.Sort(versions)
		current := versions[len(versions)-1].String()
		if tagName, ok := versionMap[current]; ok {
			return tagName, nil
		}
	}

	return "", errUnableToFindTag
}

func getHeadHash(repo *git.Repository) (string, error) {
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}
	return ref.Hash().String(), nil
}
