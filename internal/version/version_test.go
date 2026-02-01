package version

import (
	"strings"
	"testing"
)

func TestParseSemver(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    *Semver
		wantErr bool
	}{
		{
			name:    "valid version",
			version: "1.2.3",
			want:    &Semver{Major: 1, Minor: 2, Patch: 3},
			wantErr: false,
		},
		{
			name:    "valid version with zeros",
			version: "0.0.0",
			want:    &Semver{Major: 0, Minor: 0, Patch: 0},
			wantErr: false,
		},
		{
			name:    "valid version with large numbers",
			version: "999.888.777",
			want:    &Semver{Major: 999, Minor: 888, Patch: 777},
			wantErr: false,
		},
		{
			name:    "empty string",
			version: "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "version with v prefix",
			version: "v1.2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "version with V prefix",
			version: "V1.2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - two parts",
			version: "1.2",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - four parts",
			version: "1.2.3.4",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric major",
			version: "a.2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric minor",
			version: "1.b.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric patch",
			version: "1.2.c",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - negative major",
			version: "-1.2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - negative minor",
			version: "1.-2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - negative patch",
			version: "1.2.-3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - leading zero in major",
			version: "01.2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - leading zero in minor",
			version: "1.02.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - leading zero in patch",
			version: "1.2.03",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "version exceeds max component",
			version: "1001.2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "version exceeds max length",
			version: strings.Repeat("1", 51),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty major component",
			version: ".2.3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty minor component",
			version: "1..3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty patch component",
			version: "1.2.",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSemver(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSemver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch {
					t.Errorf("ParseSemver() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSemverString(t *testing.T) {
	tests := []struct {
		name    string
		version *Semver
		want    string
	}{
		{
			name:    "standard version",
			version: &Semver{Major: 1, Minor: 2, Patch: 3},
			want:    "1.2.3",
		},
		{
			name:    "version with zeros",
			version: &Semver{Major: 0, Minor: 0, Patch: 0},
			want:    "0.0.0",
		},
		{
			name:    "large version numbers",
			version: &Semver{Major: 999, Minor: 888, Patch: 777},
			want:    "999.888.777",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("Semver.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemverCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   *Semver
		v2   *Semver
		want int
	}{
		{
			name: "equal versions",
			v1:   &Semver{Major: 1, Minor: 2, Patch: 3},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 3},
			want: 0,
		},
		{
			name: "v1 major greater",
			v1:   &Semver{Major: 2, Minor: 0, Patch: 0},
			v2:   &Semver{Major: 1, Minor: 9, Patch: 9},
			want: 1,
		},
		{
			name: "v1 major less",
			v1:   &Semver{Major: 1, Minor: 9, Patch: 9},
			v2:   &Semver{Major: 2, Minor: 0, Patch: 0},
			want: -1,
		},
		{
			name: "v1 minor greater",
			v1:   &Semver{Major: 1, Minor: 3, Patch: 0},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 9},
			want: 1,
		},
		{
			name: "v1 minor less",
			v1:   &Semver{Major: 1, Minor: 2, Patch: 9},
			v2:   &Semver{Major: 1, Minor: 3, Patch: 0},
			want: -1,
		},
		{
			name: "v1 patch greater",
			v1:   &Semver{Major: 1, Minor: 2, Patch: 4},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 3},
			want: 1,
		},
		{
			name: "v1 patch less",
			v1:   &Semver{Major: 1, Minor: 2, Patch: 3},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 4},
			want: -1,
		},
		{
			name: "zeros equal",
			v1:   &Semver{Major: 0, Minor: 0, Patch: 0},
			v2:   &Semver{Major: 0, Minor: 0, Patch: 0},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.Compare(tt.v2); got != tt.want {
				t.Errorf("Semver.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemverIsGreaterThan(t *testing.T) {
	tests := []struct {
		name string
		v1   *Semver
		v2   *Semver
		want bool
	}{
		{
			name: "v1 greater",
			v1:   &Semver{Major: 2, Minor: 0, Patch: 0},
			v2:   &Semver{Major: 1, Minor: 9, Patch: 9},
			want: true,
		},
		{
			name: "v1 less",
			v1:   &Semver{Major: 1, Minor: 0, Patch: 0},
			v2:   &Semver{Major: 2, Minor: 0, Patch: 0},
			want: false,
		},
		{
			name: "v1 equal",
			v1:   &Semver{Major: 1, Minor: 2, Patch: 3},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 3},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.IsGreaterThan(tt.v2); got != tt.want {
				t.Errorf("Semver.IsGreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemverIsLessThan(t *testing.T) {
	tests := []struct {
		name string
		v1   *Semver
		v2   *Semver
		want bool
	}{
		{
			name: "v1 less",
			v1:   &Semver{Major: 1, Minor: 0, Patch: 0},
			v2:   &Semver{Major: 2, Minor: 0, Patch: 0},
			want: true,
		},
		{
			name: "v1 greater",
			v1:   &Semver{Major: 2, Minor: 0, Patch: 0},
			v2:   &Semver{Major: 1, Minor: 9, Patch: 9},
			want: false,
		},
		{
			name: "v1 equal",
			v1:   &Semver{Major: 1, Minor: 2, Patch: 3},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 3},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.IsLessThan(tt.v2); got != tt.want {
				t.Errorf("Semver.IsLessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemverEqual(t *testing.T) {
	tests := []struct {
		name string
		v1   *Semver
		v2   *Semver
		want bool
	}{
		{
			name: "equal versions",
			v1:   &Semver{Major: 1, Minor: 2, Patch: 3},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 3},
			want: true,
		},
		{
			name: "different major",
			v1:   &Semver{Major: 2, Minor: 2, Patch: 3},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 3},
			want: false,
		},
		{
			name: "different minor",
			v1:   &Semver{Major: 1, Minor: 3, Patch: 3},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 3},
			want: false,
		},
		{
			name: "different patch",
			v1:   &Semver{Major: 1, Minor: 2, Patch: 4},
			v2:   &Semver{Major: 1, Minor: 2, Patch: 3},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.Equal(tt.v2); got != tt.want {
				t.Errorf("Semver.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindLatest(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		want     string
		wantErr  bool
	}{
		{
			name:     "single version",
			versions: []string{"1.2.3"},
			want:     "1.2.3",
			wantErr:  false,
		},
		{
			name:     "multiple versions in order",
			versions: []string{"1.0.0", "1.1.0", "1.2.0"},
			want:     "1.2.0",
			wantErr:  false,
		},
		{
			name:     "multiple versions out of order",
			versions: []string{"1.2.0", "1.0.0", "1.1.0"},
			want:     "1.2.0",
			wantErr:  false,
		},
		{
			name:     "versions with different majors",
			versions: []string{"1.9.9", "2.0.0", "1.10.0"},
			want:     "2.0.0",
			wantErr:  false,
		},
		{
			name:     "versions with different minors",
			versions: []string{"1.2.9", "1.3.0", "1.2.10"},
			want:     "1.3.0",
			wantErr:  false,
		},
		{
			name:     "versions with different patches",
			versions: []string{"1.2.3", "1.2.5", "1.2.4"},
			want:     "1.2.5",
			wantErr:  false,
		},
		{
			name:     "empty list",
			versions: []string{},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "invalid version in list",
			versions: []string{"1.2.3", "v1.2.4", "1.2.5"},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "all zeros",
			versions: []string{"0.0.0"},
			want:     "0.0.0",
			wantErr:  false,
		},
		{
			name:     "mix of zeros and non-zeros",
			versions: []string{"0.0.1", "0.0.0", "0.1.0"},
			want:     "0.1.0",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindLatest(tt.versions)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindLatest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindLatest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "valid version",
			version: "1.2.3",
			wantErr: false,
		},
		{
			name:    "invalid version with v prefix",
			version: "v1.2.3",
			wantErr: true,
		},
		{
			name:    "invalid version format",
			version: "1.2",
			wantErr: true,
		},
		{
			name:    "empty version",
			version: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateVersion(tt.version); (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
