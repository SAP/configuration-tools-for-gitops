package version

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"

	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
)

const (
	maxCommitLength = 7
	versionElements = 4
)

// Version information set by link flags during build. We fall back to these sane
// default values when we build outside the Makefile context (e.g. go run, go build, or go test).
var (
	version      = "99.99.99"             // inserted from goreleaser
	buildDate    = "1970-01-01T00:00:00Z" // inserted from goreleaser
	gitCommit    = ""                     // inserted from goreleaser
	gitTag       = ""                     // inserted from goreleaser
	gitTreeState = ""                     // inserted from goreleaser

	re            = regexp.MustCompile(`(\d*)\.(\d*)\.(\d*)`)
	semver SemVer = SemVer{}
)

//nolint:gochecknoinits // The init sets the binaries version once and for all. This comes from build time inputs.
func init() {
	matches := re.FindStringSubmatch(version)
	if len(matches) != versionElements {
		log.Log.Sugar().DPanicf("illegal version \"%v\" provided", version)
	}
	var s SemVer
	var err error
	s.Major, err = strconv.Atoi(matches[1])
	if err != nil {
		log.Log.Sugar().DPanicf("major version in \"%v\" not an int: %v", semver, err)
	}
	s.Minor, err = strconv.Atoi(matches[2])
	if err != nil {
		log.Log.Sugar().DPanicf("minor version in \"%v\" not an int: %v", semver, err)
	}
	s.Patch, err = strconv.Atoi(matches[3])
	if err != nil {
		log.Log.Sugar().DPanicf("patch version in \"%v\" not an int: %v", semver, err)
	}
	semver = s
}

// Version contains version information
type Version struct {
	Version      string
	SemVer       SemVer
	BuildDate    string
	GitCommit    string
	GitTag       string
	GitTreeState string
	GoVersion    string
	Compiler     string
	Platform     string
}

type SemVer struct {
	Major int
	Minor int
	Patch int
}

// Read returns the version string. This is either the gitTag for released version
// or a combination of version number, commit hash and gitTreeState.
func Read() string {
	if gitCommit != "" && gitTag != "" && gitTreeState == "clean" {
		// if we have a clean tree state and the current commit is tagged,
		// this is an official release.
		return gitTag
	}
	// otherwise formulate a version string based on as much metadata
	// information we have available.
	versionString := "v" + version
	if len(gitCommit) >= maxCommitLength {
		versionString += "+" + gitCommit[0:7]
		if gitTreeState != "clean" {
			versionString += ".dirty"
		}
	} else {
		versionString += "+unknown"
	}
	return versionString
}

func SemanticVersion() SemVer {
	return semver
}

// ReadAll returns the complete version information
func ReadAll() *Version {
	return &Version{
		Version:      Read(),
		SemVer:       semver,
		BuildDate:    buildDate,
		GitCommit:    gitCommit,
		GitTag:       gitTag,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
