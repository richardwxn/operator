package mesh

import (
	"path/filepath"
	"testing"

	"istio.io/operator/pkg/util"
)

func TestManifestMigrate(t *testing.T) {
	testDataDir = filepath.Join(repoRootDir, "cmd/mesh/testdata/manifest-migrate")
	tests := []struct {
		desc       string
		diffSelect string
		diffIgnore string
	}{
		{
			desc: "values",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			inPath := filepath.Join(testDataDir, "input", tt.desc+".yaml")
			outPath := filepath.Join(testDataDir, "output", tt.desc+".yaml")

			got, err := runManifestMigrate(inPath)
			if err != nil {
				t.Fatal(err)
			}
			want, err := readFile(outPath)
			if err != nil {
				t.Fatal(err)
			}
			if !util.IsYAMLEqual(got, want) {
				t.Errorf("manifest-migrate command(%s): got:\n%s\n\nwant:\n%s\nDiff:\n%s\n", tt.desc, got, want, util.YAMLDiff(got, want))
			}
		})
	}
}

func runManifestMigrate(path string) (string, error) {
	return runCommand("manifest migrate " + path)
}
