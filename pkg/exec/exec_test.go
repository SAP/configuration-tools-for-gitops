package exec_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/SAP/configuration-tools-for-gitops/pkg/exec"
	"github.com/SAP/configuration-tools-for-gitops/pkg/testfuncs"
)

const (
	scriptName = "script.sh"
)

type scenario struct {
	title      string
	timeout    time.Duration
	cmd        string
	env        []exec.Input
	args       []exec.Input
	want       string
	wantErr    error
	wantLogs   string
	workingDir string
	script     string
}

var scenarios = []scenario{
	{
		title:      "timeout",
		timeout:    100 * time.Millisecond,
		cmd:        "sleep",
		args:       exec.Public("5"),
		want:       "",
		wantErr:    errors.New("signal: killed"),
		workingDir: "",
		wantLogs:   "workdir: %q -- exec: \"   sleep 5\"",
	},
	{
		title:      "not empty",
		cmd:        "ls",
		args:       []exec.Input{},
		want:       "",
		wantErr:    nil,
		workingDir: "./",
		wantLogs:   "workdir: %q -- exec: \"   ls \"",
	},
	{
		title:      "env var used",
		cmd:        "/bin/sh",
		env:        exec.Public("TEST=hello"),
		args:       exec.Public(scriptName),
		want:       "hello",
		wantErr:    nil,
		workingDir: "./",
		script: `
/bin/bash
printf "%s" $TEST
`,
		wantLogs: "workdir: %q -- exec: \"TEST=hello   /bin/sh script.sh\"",
	},
	{
		title: "private env vars",
		cmd:   "/bin/sh",
		env: append(
			exec.Public("TEST=hello"),
			exec.Private("PRIVATE=bye", "PRIVATE=xxx"),
			exec.Private("PRIVATE_ALL=secret"),
		),
		args:       exec.Public(scriptName),
		want:       "hello bye secret",
		wantErr:    nil,
		workingDir: "./",
		script: `
/bin/bash
printf "%s %s %s" $TEST $PRIVATE $PRIVATE_ALL
`,
		wantLogs: "workdir: %q -- exec: \"TEST=hello PRIVATE=xxx xxx   sh -c '/bin/sh script.sh'\"",
	},
	{
		title: "private args",
		cmd:   "/bin/sh",
		args: append(
			exec.Public(scriptName, "input1"),
			exec.Private("this_is=secret", "this_is=xxx"),
			exec.Private("really_secret"),
		),
		want:       "input1 this_is=secret really_secret",
		wantErr:    nil,
		workingDir: "./",
		script: `
/bin/bash
printf "$1 $2 $3"
`,
		wantLogs: "workdir: %q -- exec: \"   /bin/sh script.sh input1 this_is=xxx xxx\"",
	},
	{
		title:      "use $HOME",
		cmd:        "/bin/sh",
		args:       exec.Public(scriptName, "input1"),
		want:       "not empty",
		wantErr:    nil,
		workingDir: "./",
		script: `
/bin/bash
printf "%s" ${HOME}
`,
		wantLogs: "not empty",
	},
}

func TestOutput(t *testing.T) {
	for _, s := range scenarios {
		t.Logf("test scenario: %s\n", s.title)
		s.test(t)
	}
}

func (s *scenario) test(t *testing.T) {
	tmpDir, err := s.setup()
	if err != nil {
		t.Logf("setup failed: %v\n", err)
		t.FailNow()
	}
	defer tmpDir.Cleanup(t)

	l := logger{}

	ctx := context.TODO()
	if s.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.timeout)
		defer cancel()
	}

	if s.title == "use $HOME" {
		ex, err := exec.NewHome(ctx, &l, filepath.Join(tmpDir.Path(), s.workingDir), s.env...)
		testfuncs.CheckErrs(t, nil, err)
		res, err := ex.Command(s.cmd, s.args...).Output()
		testfuncs.CheckErrs(t, s.wantErr, err)
		if len(res) == 0 {
			t.Error("passing HOME variable failed")
			t.Fail()
		}
		if l.output == "" {
			t.Error("log output when passing HOME variable failed")
			t.Fail()
		}
	} else {
		res, err := exec.New(ctx, &l, filepath.Join(tmpDir.Path(), s.workingDir), s.env...).
			Command(s.cmd, s.args...).Output()
		testfuncs.CheckErrs(t, s.wantErr, err)
		s.CheckRes(t, res)
		s.CheckLogs(t, l.output, tmpDir.Path())
	}
}

type logger struct {
	output string
}

func (l *logger) Debugf(template string, args ...interface{}) {
	l.output = fmt.Sprintf(template, args...)
}
func (l *logger) Debug(msg ...interface{}) {
	l.output = fmt.Sprint(msg...)
}

func (s *scenario) CheckRes(t *testing.T, got []byte) {
	gotString := string(got)
	if s.want != gotString {
		t.Errorf(
			"results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"",
			s.want,
			gotString,
		)
		t.Fail()
	}
}

func (s *scenario) CheckLogs(t *testing.T, got, workdir string) {
	want := fmt.Sprintf(s.wantLogs, workdir)
	if want != got {
		t.Errorf(
			"logs do not match: \nwant = \"%+v\"\ngot =  \"%+v\"",
			want,
			got,
		)
		t.Fail()
	}
}

func (s *scenario) setup() (testfuncs.TestDir, error) {
	allFiles := map[string][]byte{}
	if s.script != "" {
		allFiles[scriptName] = []byte(s.script)
	}

	return testfuncs.PrepareTestDirTree(allFiles)
}
