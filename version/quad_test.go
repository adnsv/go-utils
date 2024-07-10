package version

import (
	"reflect"
	"testing"

	"github.com/blang/semver/v4"
)

func TestMakeVersionQuad(t *testing.T) {
	type args struct {
		semantic          string
		additionalCommits int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{"release", args{"1.0.0", 0}, "1.0.0.50000", nil},
		{"release", args{"1.0.0", 1}, "1.0.0.50001", nil},
		{"release", args{"1.0.0", 99}, "1.0.0.50099", nil},
		{"release", args{"1.0.0", 100}, "1.0.0.50099", nil},
		{"release-1", args{"1.0.0-1", 0}, "1.0.0.50100", nil},
		{"release-1.ignored", args{"1.0.0-1.ignored", 0}, "1.0.0.50100", nil},
		{"release-99", args{"1.0.0-99", 0}, "1.0.0.59900", nil},
		{"release-100", args{"1.0.0-100", 0}, "", ErrReleaseNumberIsTooLarge},
		{"alpha", args{"1.0.0-alpha", 0}, "1.0.0.10000", nil},
		{"alpha.1", args{"1.0.0-alpha.1", 0}, "1.0.0.10100", nil},
		{"alpha.2", args{"1.0.0-alpha.2", 0}, "1.0.0.10200", nil},
		{"alpha.2", args{"1.0.0-alpha.2", 99}, "1.0.0.10299", nil},
		{"alpha.ignored", args{"1.0.0-alpha.ignored", 99}, "1.0.0.10099", nil},
		{"alpha.99", args{"1.0.0-alpha.99", 0}, "1.0.0.19900", nil},
		{"alpha.100", args{"1.0.0-alpha.100", 0}, "", ErrReleaseNumberIsTooLarge},
		{"a", args{"1.0.0-a", 0}, "1.0.0.10000", nil},
		{"beta", args{"1.0.0-beta", 0}, "1.0.0.20000", nil},
		{"b", args{"1.0.0-b", 0}, "1.0.0.20000", nil},
		{"rc", args{"1.0.0-rc", 0}, "1.0.0.30000", nil},
		{"rc.0", args{"1.0.0-rc.0", 0}, "1.0.0.30000", nil},
		{"rc.1", args{"1.0.0-rc.1", 0}, "1.0.0.30100", nil},
		{"rc.1", args{"1.0.0-rc.1", 1}, "1.0.0.30101", nil},
		{"rc.99", args{"1.0.0-rc.99", 0}, "1.0.0.39900", nil},
		{"rc.99", args{"1.0.0-rc.99", 99}, "1.0.0.39999", nil},
		{"rc.2", args{"1.0.0-rc.2", 0}, "1.0.0.30200", nil},
		{"rc.2+meta", args{"1.0.0-rc.2+meta", 0}, "1.0.0.30200", nil},
		{"gamma", args{"1.0.0-gamma", 0}, "", ErrUnsupportedPR},
		{"delta", args{"1.0.0-delta", 0}, "", ErrUnsupportedPR},
		{"large", args{"1000000.0.0", 0}, "", ErrVersionNumberIsTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv := semver.MustParse(tt.args.semantic)
			q, err := MakeQuad(sv, tt.args.additionalCommits)
			if err != tt.wantErr {
				t.Errorf("MakeVersionQuad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if !reflect.DeepEqual(q.String(), tt.want) {
				t.Errorf("MakeVersionQuad() = %v, want %v", q.String(), tt.want)
			}
		})
	}
}
