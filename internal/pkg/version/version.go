package version

import "runtime/debug"

var Commit = func() string {
	vers := "no version"
	suffix := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				vers = setting.Value
			}
			if setting.Key == "vcs.modified" && setting.Value == "true" {
				suffix = "-dirty"
			}
		}
	}
	return vers + suffix
}()
