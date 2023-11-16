package pjsekaioverlay

import (
	"slices"
	"strings"
)

const Version = "0.0.0"

func ConvertENVersion(version string) string {
	version_numbers := strings.Split(version, ".")
	jp_version_numbers := slices.Delete(version_numbers, 3, len(version_numbers))
	jp_version := strings.Join(jp_version_numbers, ".")
	return jp_version
}
