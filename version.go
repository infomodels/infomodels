package main

import "github.com/blang/semver"

// SemVer components.
const (
	progMajor        = 0
	progMinor        = 1
	progPatch        = 0
	progReleaseLevel = "alpha"
)

var (
	// Populated at build time. See the Makefile for details.
	// Note, in environments where the git information is not
	// available, progBuild will not be populated. In
	// environments where the BUILD_NUM environment variable
	// is not available, progReleaseNum will not be populated.
	progBuild      string
	progReleaseNum string

	// Main semantic version for the CLI tool. If the release
	// level is not final, it will be added along with the
	// sha and build number at build time.
	progVersion = semver.Version{
		Major: progMajor,
		Minor: progMinor,
		Patch: progPatch,
	}
)

func init() {
	if progReleaseLevel != "final" {

		// Add the release number.
		progVersion.Pre = []semver.PRVersion{{
			VersionStr: progReleaseLevel,
		}}

		// Add the build git sha if available.
		if progBuild != "" {
			progVersion.Build = []string{progBuild}
		}

		// Add the build number if available.
		if progReleaseNum != "" {
			releaseNumPRVersion, _ := semver.NewPRVersion(progReleaseNum)
			progVersion.Pre = append(progVersion.Pre, releaseNumPRVersion)
		}
	}
}
