package git

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const (
	notUsed = "notUsed"
)

var (
	newRefName = plumbing.NewBranchReferenceName
	newCommit  = plumbing.NewHash
)

type Repository struct {
	Client   *git.Repository
	token    string
	Path     string
	Remote   string
	MaxDepth int
}

// New registers a new git repository. If there is a git repository already present
// in "path", the url and remote will be validated and then the repository will be used.
func New(
	path, url, token, remote, defaultBranch string, maxDepth int, logLvl log.Level,
) (repo Repository, err error) {
	repo = Repository{}

	client, err := git.PlainOpen(path)
	if err != nil {
		var output io.Writer
		if logLvl.Is(log.Debug()) {
			output = os.Stdout
		}
		client, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:               url,
			Auth:              &http.BasicAuth{Username: notUsed, Password: token},
			RemoteName:        remote,
			ReferenceName:     newRefName(defaultBranch),
			SingleBranch:      true,
			NoCheckout:        true,
			Depth:             maxDepth,
			RecurseSubmodules: 1,
			Progress:          output,
			Tags:              0,
			InsecureSkipTLS:   false,
			CABundle:          []byte{},
		})
		if err != nil {
			return
		}
	} else {
		var remotes []*git.Remote
		remotes, err = client.Remotes()
		if err != nil {
			return
		}

		remoteFound := false
		for _, r := range remotes {
			if r.Config().URLs[0] == url {
				remoteFound = true
				break
			}
		}
		if !remoteFound {
			if _, err = client.CreateRemote(&config.RemoteConfig{
				Name: remote,
				URLs: []string{url},
				Fetch: []config.RefSpec{config.RefSpec(
					fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", remote),
				)},
			}); err != nil {
				err = fmt.Errorf("cannot create remote \"%s\": %s", remote, err)
				return
			}
		}
	}

	return Repository{client, token, path, remote, maxDepth}, nil
}

func (r *Repository) Checkout(branch string, force bool) (res *Tree, err error) {
	if _, err = r.Client.Branch(branch); err != nil {
		if err = r.Client.Fetch(&git.FetchOptions{
			RemoteName: r.Remote,
			RefSpecs:   []config.RefSpec{config.RefSpec("refs/*:refs/*")},
			Auth:       &http.BasicAuth{Username: notUsed, Password: r.token},
			Depth:      r.MaxDepth,
			Force:      true,
			Progress:   nil,
		}); err != nil && err != git.NoErrAlreadyUpToDate {
			return
		}
	}
	return checkoutTree(r.Client, branch, branchT, force)
}

type Tree struct {
	RepoClient *git.Repository
	T          *object.Tree
	HeadCommit *object.Commit
}

func (t *Tree) MergeBase(other *Tree) (base *Tree, err error) {
	mb, err := t.HeadCommit.MergeBase(other.HeadCommit)
	if err != nil {
		return nil, err
	}
	if len(mb) != 1 {
		return nil, fmt.Errorf("merge-base is not unique - found: \n%+v", mb)
	}
	return checkoutTree(t.RepoClient, mb[0].Hash.String(), commitT, true)
}

// Diff compares the provided GitTree c with g and returns a slice of changed
// file paths.
func (t *Tree) DiffPaths(c *Tree) ([]string, error) {
	changes, err := t.T.Diff(c.T)
	if err != nil {
		return []string{}, err
	}

	allChanges := []string{}
	for _, c := range changes {
		allChanges = append(allChanges, c.To.Name, c.From.Name)
	}
	return unique(allChanges), nil
}

func unique(slice []string) []string {
	uniqMap := make(map[string]interface{})
	for _, v := range slice {
		uniqMap[v] = nil
	}

	uniqSlice := make([]string, 0, len(uniqMap))
	for v := range uniqMap {
		if v != "" {
			uniqSlice = append(uniqSlice, v)
		}
	}
	sort.Strings(uniqSlice)
	return uniqSlice
}

// FindFiles searches t for all instances of the provided pattern in the file
// paths.
func (t *Tree) FindFiles(pattern string) (files []string, err error) {
	err = t.T.Files().ForEach(
		func(f *object.File) error {
			if strings.Contains(f.Name, pattern) {
				files = append(files, f.Name)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return files, err
}

// ReadFile reads the file at location filepath and returns its content as a string.
func (t *Tree) ReadFile(filepath string) (string, error) {
	file, err := t.T.File(filepath)
	if err != nil {
		return "", err
	}
	return file.Contents()
}

type refType string

const (
	branchT refType = "branch"
	commitT refType = "commit"
)

func checkoutTree(
	c *git.Repository, ref string, rt refType, force bool,
) (t *Tree, err error) {
	options := git.CheckoutOptions{Force: force}
	switch rt {
	case branchT:
		options.Branch = newRefName(ref)
	case commitT:
		options.Hash = newCommit(ref)
	default:
		err = fmt.Errorf("illegal refType \"%s\" found", rt)
		return
	}
	w, err := c.Worktree()
	if err != nil {
		return
	}

	if err = w.Checkout(&options); err != nil {
		return
	}

	reference, err := c.Head()
	if err != nil {
		return
	}

	commit, err := c.CommitObject(reference.Hash())
	if err != nil {
		return
	}

	tree, err := commit.Tree()
	if err != nil {
		return
	}
	return &Tree{c, tree, commit}, nil
}
