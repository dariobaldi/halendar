package vcs

import (
	"fmt"
	"runtime/debug"
)

func Version() string {
	var time string
	var revision string
	var modified bool

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.time":
				time = setting.Value
			case "vcs.revision":
				revision = setting.Value
			case "vcs.modified":
				modified = (setting.Value == "true")
			}
		}
	}
	version := fmt.Sprintf("%s-%s", time, revision)

	if modified {
		return version + "-dirty"
	}
	return version
}
