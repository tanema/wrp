package config

import (
	"crypto/md5"
	"encoding/json"
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

// Dependency describes a single dependency to fetch and files inside to keep
type Dependency struct {
	Pick   []string `json:"pick,omitempty"`
	Tag    string   `json:"tag,omitempty"`
	Branch string   `json:"branch,omitempty"`
	Hash   string   `json:"hash,omitempty"`
	Check  string   `json:"check,omitempty"`

	url string
	fs  billy.Filesystem
}

func (dep *Dependency) MarshalJSON() ([]byte, error) {
	return json.Marshal(dep)
}

func (dep *Dependency) fetch(dest string, url string, lock Dependency) error {
	var err error
	dep.url, err = parseURL(url)
	if err != nil {
		return err
	}
	dep.fs, err = dep.getState()
	return err
}

func (dep *Dependency) requiresUpdate(dest string, lock Dependency) bool {
	sum, err := dep.calcChecksum(dest)
	if err != nil {
		return true
	}
	if lock.Check != sum {
		return true
	}
	return lock.Tag != dep.Tag && lock.Branch != dep.Branch && lock.Hash != dep.Hash
}

func parseURL(repourl string) (string, error) {
	if strings.HasPrefix(repourl, "http://") || strings.HasPrefix(repourl, "git@") {
		return "", fmt.Errorf("invalid dependency url %v please only use domain", repourl)
	}
	_, err := url.Parse("https://" + repourl)
	if err != nil {
		return "", fmt.Errorf("cannot parse repo url %v, received err: %v", repourl, err)
	}
	return repourl, nil
}

func (dep *Dependency) calcChecksum(dest string) (string, error) {
	hasher := md5.New()
	if len(dep.Pick) > 0 {
		for _, pick := range dep.Pick {
			if err := file.Sum(filepath.Join(dest, pick), hasher); err != nil {
				return "", err
			}
		}
	} else {
		baseName := strings.TrimSuffix(path.Base(dep.url), ".git")
		if err := file.Sum(filepath.Join(dest, baseName), hasher); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func (dep *Dependency) write(dest string) error {
	hasher := md5.New()
	if len(dep.Pick) > 0 {
		for _, pick := range dep.Pick {
			if err := file.Copy(dep.fs, pick, filepath.Join(dest, pick), hasher); err != nil {
				return err
			}
		}
	} else {
		baseName := strings.TrimSuffix(path.Base(dep.url), ".git")
		if err := file.Copy(dep.fs, ".", filepath.Join(dest, baseName), hasher); err != nil {
			return err
		}
	}
	dep.Check = fmt.Sprintf("%x", hasher.Sum(nil))
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

	if dep.Tag == "" && dep.Branch == "" && dep.Hash == "" {
		if dep.Tag, err = getRepoTag(repo); err == errUnableToFindTag {
			if dep.Hash, err = getHeadHash(repo); err != nil {
				return fs, fmt.Errorf("unable to pin dependency: %v", err)
			}
		} else if err != nil {
			return fs, fmt.Errorf("problem fetching tags: %v", err)
		}
	}

	if dep.Tag != "" {
		// TODO: version range
		return fs, tree.Checkout(&git.CheckoutOptions{Branch: plumbing.NewTagReferenceName(dep.Tag)})
	} else if dep.Branch != "" {
		return fs, tree.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(dep.Branch)})
	} else if dep.Hash != "" {
		return fs, tree.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(dep.Hash)})
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
