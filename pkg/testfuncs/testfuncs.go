package testfuncs

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const (
	dirPermissions     = 0755
	msgSkipIntegration = "skipping integration test - to run set env variable \"export INTEGRATION_TESTS=true\""
)

func MustBeNil(t *testing.T, err error) {
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

func CheckErrs(t *testing.T, want, got error) {
	checkErrs(t, want, got,
		func(want, got error) bool { return want.Error() != got.Error() },
	)
}

func CheckSimilarErrs(t *testing.T, want, got error) {
	checkErrs(t, want, got,
		func(want, got error) bool { return strings.HasPrefix(got.Error(), want.Error()) },
	)
}

func checkErrs(t *testing.T, want, got error, compare func(want, got error) bool) {
	if got == nil && want == nil {
		return
	} else if got == nil && want != nil {
		t.Errorf("errors do not match: \nwant = \"%+v\"\ngot =  \"%+v\"", want, "nil")
		t.Fail()
	} else if got != nil && want == nil {
		t.Errorf("errors do not match: \nwant = \"%+v\"\ngot =  \"%+v\"", "nil", got)
		t.Fail()
	} else if compare(want, got) {
		t.Errorf("errors do not match: \nwant = \"%+v\"\ngot =  \"%+v\"", want, got)
		t.Fail()
	}
}

func CheckEqualityInterface(t *testing.T, want, got interface{}) {
	if !reflect.DeepEqual(want, got) {
		t.Errorf(
			"want and got do not match: \nwant = \"%s\"\ngot =  \"%s\"",
			want, got,
		)
		t.Fail()
	}
}

func Error(t *testing.T, msgPrefix, want, got interface{}) {
	t.Errorf(
		"%s results do not match: \nwant = \"%v\"\ngot =  \"%v\"",
		msgPrefix, want, got,
	)
	t.Fail()
}

func FromEnv(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Errorf("environment variable \"%s\" must not be empty", key)
	}
	return value
}

type TestDir interface {
	Cleanup(t *testing.T)
	Path() string
}

type testDir struct {
	path string
}

func (td testDir) Cleanup(t *testing.T) {
	err := os.RemoveAll(td.path)
	CheckErrs(t, nil, err)
	t.Logf("clean up \"%s\" successful", td.path)
}

func (td testDir) Path() string {
	return td.path
}

func PrepareTestDirTree(files map[string][]byte) (TestDir, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return testDir{}, fmt.Errorf("error creating temp directory: %v", err)
	}

	for name, content := range files {
		fSlice := strings.Split(name, "/")
		fileName := fSlice[len(fSlice)-1]
		filePath := strings.Join(fSlice[:len(fSlice)-1], "/")

		if err := os.MkdirAll(filepath.Join(tmpDir, filePath), dirPermissions); err != nil {
			os.RemoveAll(tmpDir)
			return testDir{}, fmt.Errorf("failed to create dir %s: %s", filePath, err)
		}

		if err := writeFile(filepath.Join(tmpDir, filePath, fileName), content); err != nil {
			os.RemoveAll(tmpDir)
			return testDir{}, err
		}
	}

	return testDir{tmpDir}, nil
}

func writeFile(path string, content []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %s", path, err)
	}
	defer file.Close()
	if content != nil {
		if _, err := file.Write(content); err != nil {
			return fmt.Errorf("failed to write content: %s", err)
		}
	}
	return nil
}

func RunIntegrationTests(t *testing.T) {
	env := os.Getenv("INTEGRATION_TESTS")
	if env == "" {
		t.Skip(msgSkipIntegration)
	}
	integration, err := strconv.ParseBool(env)
	if err != nil {
		t.Skip(msgSkipIntegration)
	}
	if !integration {
		t.Skip(msgSkipIntegration)
	}
}
