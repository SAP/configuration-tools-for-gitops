package git

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/configuration-tools-for-gitops/pkg/testfuncs"
	"github.com/go-git/go-git/v5"
)

const (
	irrelevant = "irrelevant"
)

func TestFailCheckout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	testfuncs.CheckErrs(t, nil, err)
	defer os.RemoveAll(tmpDir)

	_, err = git.PlainInit(tmpDir, true)
	testfuncs.CheckErrs(t, nil, err)

	repo, err := New(tmpDir, irrelevant, irrelevant, remoteName, mainBranch,
		0, log.New("Debug"),
	)
	testfuncs.CheckErrs(t, nil, err)

	t.Log("fail checkout")
	_, err = repo.Checkout("doesNotExist", false)
	testfuncs.CheckErrs(t, fmt.Errorf("repository not found"), err)
}

func TestFailNew(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	testfuncs.CheckErrs(t, nil, err)
	defer os.RemoveAll(tmpDir)

	_, err = New(tmpDir, irrelevant, irrelevant, remoteName, mainBranch, 0, log.New("Debug"))
	testfuncs.CheckErrs(t, fmt.Errorf("repository not found"), err)
}

func TestRemoteNotFound(t *testing.T) {
	fetchRules := "+refs/heads/*:refs/remotes/origin/*"

	tmpDir, err := os.MkdirTemp("", "")
	testfuncs.CheckErrs(t, nil, err)
	defer os.RemoveAll(tmpDir)

	_, err = git.PlainInit(tmpDir, true)
	testfuncs.CheckErrs(t, nil, err)

	repo, err := New(tmpDir, irrelevant, irrelevant, remoteName, mainBranch, 10, log.New("Debug"))
	testfuncs.CheckErrs(t, nil, err)
	remotes, err := repo.Client.Remotes()
	testfuncs.CheckErrs(t, nil, err)

	if len(remotes) != 1 {
		t.Errorf(
			"expected only 1 remote but got %v: %+v \n",
			len(remotes), remotes,
		)
		t.FailNow()
	}
	r := remotes[0].Config()
	if r.Name != remoteName {
		t.Errorf(
			"wrong remote name: \nwant = \"%+v\"\ngot =  \"%+v\"\n",
			remoteName, r.Name,
		)
	}
	if !reflect.DeepEqual(r.URLs, []string{irrelevant}) {
		t.Errorf(
			"wrong remote urls: \nwant = \"%+v\"\ngot =  \"%+v\"\n",
			[]string{irrelevant}, r.URLs,
		)
	}
	if len(r.Fetch) != 1 {
		t.Errorf(
			"expected only 1 fetch rule %v: %+v \n",
			len(r.Fetch), r.Fetch,
		)
		t.FailNow()
	}
	if string(r.Fetch[0]) != fetchRules {
		t.Errorf(
			"wrong fetch rules: \nwant = \"%+v\"\ngot =  \"%+v\"\n",
			fetchRules, r.Fetch[0],
		)
	}
}
