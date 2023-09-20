package dependencies

import (
	"fmt"
	"strings"

	g "github.com/SAP/configuration-tools-for-gitops/v2/cmd/coco/graph"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/git"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/log"
)

type gitRepo interface {
	Checkout(branch string, force bool) (res *git.Tree, err error)
}

var (
	repo func(
		string, string, string, string, string, int, log.Level,
	) (gitRepo, error) = newRepo
	mergeBase func(*git.Tree, *git.Tree) (*git.Tree, error) = mb
	diffPaths func(*git.Tree, *git.Tree) ([]string, error)  = diff
	graphh    func(path, depFileName string) (
		g.ComponentDependencies, map[string]string, error,
	) = Graph
)

func ChangeAffectedComponents(
	giturl, remote, gitToken, path, depFileName, sourceBranch, targetBranch string,
	graphDepth, gitDepth int, logLvl log.Level,
) (g.ComponentDependencies, error) {
	c := log.Context{
		"git.URL":         giturl,
		"git.path":        path,
		"git.remote":      remote,
		"git.depth":       gitDepth,
		"dependency-file": depFileName,
		"source-branch":   sourceBranch,
		"target-branch":   targetBranch,
	}
	c.Log("setup local git repository", log.Debug())
	r, err := repo(path, giturl, gitToken, remote, sourceBranch, gitDepth, logLvl)
	if logErr(c, err) {
		return g.ComponentDependencies{}, err
	}

	_, err = r.Checkout(sourceBranch, false)
	if logErr(c, err) {
		return g.ComponentDependencies{}, err
	}

	allDependencies, components, err := graphh(path, depFileName)
	if logErr(c, err) {
		return g.ComponentDependencies{}, err
	}
	c["dependency-graph"] = fmt.Sprintf("%+v", allDependencies)
	c.Log("dependency graph constructed", log.Debug())

	changed, sl, err := changedComponents(
		r.Checkout, components, sourceBranch, targetBranch,
	)
	c["changed-components"] = sl
	c.Log("changed components calculated", log.Debug())
	if logErr(c, err) {
		return g.ComponentDependencies{}, err
	}

	res := make(g.ComponentDependencies, len(allDependencies))
	for cc := range changed {
		var ok bool
		res[cc], ok = allDependencies[cc]
		if !ok {
			c["component"] = cc
			err := fmt.Errorf("changed component does not appear in dependency graph")
			logErr(c, err)
			return g.ComponentDependencies{}, err
		}
	}
	return res.MaxDepth(graphDepth), nil
}

func newRepo(
	path, url, token, remote, defaultBranch string, maxDepth int, logLvl log.Level,
) (gitRepo, error) {
	r, err := git.New(path, url, token, remote, defaultBranch, maxDepth, logLvl)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func changedComponents(
	checkout func(string, bool) (*git.Tree, error),
	components map[string]string, sourceBranch, targetBranch string,
) (changeMap map[string]bool, changeSlice []string, err error) {
	target, err := checkout(targetBranch, false)
	if err != nil {
		return
	}
	source, err := checkout(sourceBranch, false)
	if err != nil {
		return
	}
	base, err := mergeBase(source, target)
	if err != nil {
		return
	}
	diffFiles, err := diffPaths(source, base)
	if err != nil {
		return
	}
	changeMap, changeSlice = siftForChangedComponents(diffFiles, components)
	return
}

func siftForChangedComponents(
	diffFiles []string, components map[string]string,
) (changeMap map[string]bool, changeSlice []string) {
	changeMap = map[string]bool{}
	changeSlice = []string{}
	for _, df := range diffFiles {
		for name, path := range components {
			if strings.HasPrefix(df, path) {
				changeMap[name] = true
				changeSlice = append(changeSlice, name)
				break
			}
		}
	}
	return changeMap, changeSlice
}

func mb(source, target *git.Tree) (*git.Tree, error) {
	return source.MergeBase(target)
}
func diff(source, target *git.Tree) ([]string, error) {
	return source.DiffPaths(target)
}
