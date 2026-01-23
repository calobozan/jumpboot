package jumpboot

import (
	"fmt"
	"strings"
)

// Version represents a semantic version with major, minor, and patch components.
// Minor and Patch may be -1 if not specified (e.g., "3" parses as {3, -1, -1}).
type Version struct {
	// Major is the major version number (required).
	Major int

	// Minor is the minor version number (-1 if not specified).
	Minor int

	// Patch is the patch version number (-1 if not specified).
	Patch int
}

// ParseVersion parses a version string into a Version struct.
// Accepts formats: "X.Y.Z", "X.Y", or "X". Any trailing text is ignored.
//
// Examples:
//   - "3.10.5" -> {3, 10, 5}
//   - "3.10" -> {3, 10, -1}
//   - "3" -> {3, -1, -1}
//   - "2.1.0-beta" -> {2, 1, 0}
func ParseVersion(versionStr string) (Version, error) {
	version := Version{
		Minor: -1,
		Patch: -1,
	}
	_, err := fmt.Sscanf(versionStr, "%d.%d.%d", &version.Major, &version.Minor, &version.Patch)
	if err != nil {
		// If the version string is not in the format "X.Y.Z", try parsing it as "X.Y"
		_, err = fmt.Sscanf(versionStr, "%d.%d", &version.Major, &version.Minor)
		if err != nil {
			// If the version string is not in the format "X.Y", try parsing it as "X"
			_, err = fmt.Sscanf(versionStr, "%d", &version.Major)
			if err != nil {
				return Version{}, fmt.Errorf("error parsing version: %v", err)
			}
		}
	}
	if version.Major < 0 || version.Minor < -1 || version.Patch < -1 {
		return Version{}, fmt.Errorf("invalid version: %s", versionStr)
	}
	return version, nil
}

// ParsePythonVersion parses output from "python --version" (e.g., "Python 3.10.5").
func ParsePythonVersion(versionStr string) (Version, error) {
	parts := strings.Split(versionStr, " ")
	if len(parts) != 2 {
		return Version{}, fmt.Errorf("invalid version string: %s", versionStr)
	}
	if parts[0] != "Python" {
		return Version{}, fmt.Errorf("invalid version string: %s", versionStr)
	}
	return ParseVersion(parts[1])
}

// ParsePipVersion parses output from "pip --version" (e.g., "pip 23.0 from ...").
func ParsePipVersion(versionStr string) (Version, error) {
	parts := strings.Split(versionStr, " ")
	if len(parts) < 2 {
		return Version{}, fmt.Errorf("invalid version string: %s", versionStr)
	}
	if !strings.HasPrefix(parts[0], "pip") {
		return Version{}, fmt.Errorf("invalid version string: %s", versionStr)
	}
	return ParseVersion(parts[1])
}

// Compare returns -1 if v < other, 0 if v == other, or 1 if v > other.
// Comparison is done component by component (major, then minor, then patch).
func (v *Version) Compare(other Version) int {
	if v.Major > other.Major {
		return 1
	}
	if v.Major < other.Major {
		return -1
	}
	if v.Minor > other.Minor {
		return 1
	}
	if v.Minor < other.Minor {
		return -1
	}
	if v.Patch > other.Patch {
		return 1
	}
	if v.Patch < other.Patch {
		return -1
	}
	return 0
}

// String returns the version as a string, omitting unspecified components.
// Examples: "3.10.5", "3.10", "3"
func (v *Version) String() string {
	if v.Patch != -1 {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
	if v.Minor != -1 {
		return fmt.Sprintf("%d.%d", v.Major, v.Minor)
	}
	return fmt.Sprintf("%d", v.Major)
}

// MinorString returns the version as "major.minor" (e.g., "3.10").
// Used for paths like "python3.10" or "site-packages/python3.10".
func (v *Version) MinorString() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// MinorStringCompact returns the version without separator (e.g., "310").
// Used for Windows paths like "python310.dll".
func (v *Version) MinorStringCompact() string {
	return fmt.Sprintf("%d%d", v.Major, v.Minor)
}
