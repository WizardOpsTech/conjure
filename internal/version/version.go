package version

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// Do I make these config params later?
	// MaxVersionLength limits the total length of a version string
	MaxVersionLength = 50
	// MaxVersionComponent limits the size of each version component
	MaxVersionComponent = 1000
)

type Semver struct {
	Major int
	Minor int
	Patch int
}

func ParseSemver(version string) (*Semver, error) {
	if len(version) > MaxVersionLength {
		return nil, fmt.Errorf("version string exceeds maximum length of %d characters", MaxVersionLength)
	}

	if version == "" {
		return nil, fmt.Errorf("version string cannot be empty")
	}

	if strings.HasPrefix(version, "v") || strings.HasPrefix(version, "V") {
		return nil, fmt.Errorf("version string must not include 'v' prefix, use format 'x.x.x'")
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid semver format '%s', expected 'x.x.x'", version)
	}

	major, err := parseVersionComponent(parts[0], "major")
	if err != nil {
		return nil, err
	}

	minor, err := parseVersionComponent(parts[1], "minor")
	if err != nil {
		return nil, err
	}

	patch, err := parseVersionComponent(parts[2], "patch")
	if err != nil {
		return nil, err
	}

	return &Semver{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

func parseVersionComponent(component, name string) (int, error) {
	if component == "" {
		return 0, fmt.Errorf("%s version component cannot be empty", name)
	}

	if len(component) > 1 && component[0] == '0' {
		return 0, fmt.Errorf("%s version component '%s' has invalid leading zero", name, component)
	}

	value, err := strconv.Atoi(component)
	if err != nil {
		return 0, fmt.Errorf("%s version component '%s' is not a valid integer: %w", name, component, err)
	}

	if value < 0 {
		return 0, fmt.Errorf("%s version component cannot be negative", name)
	}

	if value > MaxVersionComponent {
		return 0, fmt.Errorf("%s version component %d exceeds maximum allowed value of %d", name, value, MaxVersionComponent)
	}

	return value, nil
}

func (v *Semver) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Semver) Compare(other *Semver) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	return 0
}

func (v *Semver) IsGreaterThan(other *Semver) bool {
	return v.Compare(other) > 0
}

func (v *Semver) IsLessThan(other *Semver) bool {
	return v.Compare(other) < 0
}

func (v *Semver) Equal(other *Semver) bool {
	return v.Compare(other) == 0
}

func FindLatest(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions provided")
	}

	parsedVersions := make([]*Semver, 0, len(versions))
	for _, v := range versions {
		parsed, err := ParseSemver(v)
		if err != nil {
			return "", fmt.Errorf("invalid version '%s': %w", v, err)
		}
		parsedVersions = append(parsedVersions, parsed)
	}

	latest := parsedVersions[0]
	latestStr := versions[0]

	for i := 1; i < len(parsedVersions); i++ {
		if parsedVersions[i].IsGreaterThan(latest) {
			latest = parsedVersions[i]
			latestStr = versions[i]
		}
	}

	return latestStr, nil
}

func ValidateVersion(version string) error {
	_, err := ParseSemver(version)
	return err
}
