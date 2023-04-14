package git

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/configuration-tools-for-gitops/pkg/testfuncs"
)

const (
	giturl      = "https://github.wdf.sap.corp/AI/test-repository.git"
	remoteName  = "origin"
	mainBranch  = "main"
	otherBranch = "otherBranch"
)

func TestIntegration(t *testing.T) {
	testfuncs.RunIntegrationTests(t)
	gitToken := os.Getenv("GITHUB_WDF_TOKEN")
	if gitToken == "" {
		t.Error("no environment variable GITHUB_WDF_TOKEN provided")
		t.FailNow()
	}

	tmpDir, err := os.MkdirTemp("", "")
	testfuncs.CheckErrs(t, nil, err)
	fmt.Println(tmpDir)
	defer os.RemoveAll(tmpDir)

	repo, err := New(tmpDir, giturl, gitToken, remoteName, mainBranch, 10, log.New("Debug"))
	testfuncs.CheckErrs(t, nil, err)
	t.Logf("repository path: %s\n", repo.Path)

	t.Log("test repeated initialization does not lead to double download")
	repo, err = New(tmpDir, giturl, gitToken, remoteName, mainBranch, 10, log.New("Debug"))
	testfuncs.CheckErrs(t, nil, err)

	t.Log("test checkout")
	mainTree, err := repo.Checkout(mainBranch, false)
	testfuncs.CheckErrs(t, nil, err)

	otherTree, err := repo.Checkout(otherBranch, false)
	testfuncs.CheckErrs(t, nil, err)

	t.Log("test diffPaths functionality")
	changesOtherMain, err := otherTree.DiffPaths(mainTree)
	testfuncs.CheckErrs(t, nil, err)

	changesMainOther, err := mainTree.DiffPaths(otherTree)
	testfuncs.CheckErrs(t, nil, err)

	if !reflect.DeepEqual(changesMainOther, changesOtherMain) {
		t.Errorf("DiffPaths between branch %s and %s depend on the comparisons order: \nmain-other: %+v\nother-main: %+v\n",
			mainBranch, otherBranch,
			changesMainOther, changesOtherMain,
		)
	}

	t.Log("test FindFiles")

	files, err := otherTree.FindFiles("dependencies.yaml")
	testfuncs.CheckErrs(t, nil, err)
	expected := []string{
		"services/s1/dependencies.yaml",
		"services/s2/dependencies.yaml",
	}
	if !reflect.DeepEqual(expected, files) {
		t.Errorf(
			"dependency files for branch %s do not match expectations:\n found: %+v\nexpected: %+v\n",
			otherBranch, files, expected,
		)
	}

	files, err = mainTree.FindFiles("dependencies.yaml")
	testfuncs.CheckErrs(t, nil, err)
	expectedMain := []string{
		"dependencies.yaml",
		"services/s1/dependencies.yaml",
	}
	if !reflect.DeepEqual(expectedMain, files) {
		t.Errorf(
			"dependency files for branch %s do not match expectations:\n found: %+v\nexpected: %+v\n",
			mainBranch, files, expectedMain,
		)
	}

	t.Log("test ReadFile")
	fileToRead := "dependencies.yaml"
	expectedContent := "global: dependencies"
	content, err := mainTree.ReadFile(fileToRead)
	testfuncs.CheckErrs(t, nil, err)
	if content != expectedContent {
		t.Errorf(
			"unexpected file content in branch %s for file %s:\n found: %+v\nexpected: %+v\n",
			mainBranch, fileToRead, content, expectedContent,
		)
	}
}
