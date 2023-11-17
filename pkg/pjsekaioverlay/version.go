package pjsekaioverlay

import (
	"strings"
)

const Version = "0.0.0"

func GetVersion(version string, index int) string {
	versionGroup := []string{Version, "unknown version"}
	if splitVersion := strings.Split(Version, "en"); len(splitVersion) == 2 {
		versionGroup = []string{splitVersion[0], "v" + splitVersion[1]}
	}
	return versionGroup[index]
}
