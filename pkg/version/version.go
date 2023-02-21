package version

import (
	"fmt"
	"runtime"
)

var (
	gitCommit string
	version   string
	buildDate string
)

// Version holds version data
type Version struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
	Arch      string `json:"arch"`
}

// Get returns the Version object
func Get() *Version {
	return &Version{
		GitCommit: gitCommit,
		Version:   version,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

func (v *Version) String() string {
	return fmt.Sprintf("{version:%s, gitCommit:%s, buildDate:%s, goVersion:%s, compiler:%s, platform:%s, arch:%s}",
		v.Version, v.GitCommit, v.BuildDate, v.GoVersion, v.Compiler, v.Platform, v.Arch,
	)
}
