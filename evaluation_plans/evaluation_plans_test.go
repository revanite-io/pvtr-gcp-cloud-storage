package evaluation_plans

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// catalogFile mirrors just enough of the Gemara Layer 2 catalog schema to
// collect assessment requirement ids.
type catalogFile struct {
	Controls []struct {
		ID                     string `yaml:"id"`
		AssessmentRequirements []struct {
			ID string `yaml:"id"`
		} `yaml:"assessment-requirements"`
	} `yaml:"controls"`
}

// TestPlanCoversCatalog asserts a bidirectional match between the assessment
// requirement ids in the vendored catalog(s) and the keys of the CCC_ObjStor plan
// map, so catalog upgrades cannot silently orphan either side. Keys outside
// the vendored catalogs' namespaces (resolved from reference catalogs
// at runtime) are not checked here.
func TestPlanCoversCatalog(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "data", "catalogs", "*.yaml"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no catalog files found: %v", err)
	}

	catalogIDs := map[string]bool{}
	prefixes := map[string]bool{}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		var catalog catalogFile
		if err := yaml.Unmarshal(raw, &catalog); err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, control := range catalog.Controls {
			// Namespace prefix, e.g. "CCC.ObjStor." from "CCC.ObjStor.CN01"
			if i := strings.LastIndex(control.ID, "."); i > 0 {
				prefixes[control.ID[:i+1]] = true
			}
			for _, requirement := range control.AssessmentRequirements {
				catalogIDs[requirement.ID] = true
			}
		}
	}
	if len(catalogIDs) == 0 {
		t.Fatal("no assessment requirements parsed from catalogs")
	}

	inScope := func(id string) bool {
		for prefix := range prefixes {
			if strings.HasPrefix(id, prefix) {
				return true
			}
		}
		return false
	}

	for id := range catalogIDs {
		if _, ok := CCC_ObjStor[id]; !ok {
			t.Errorf("catalog requirement %s has no entry in the CCC_ObjStor plan map", id)
		}
	}
	for id := range CCC_ObjStor {
		if inScope(id) && !catalogIDs[id] {
			t.Errorf("plan map key %s does not exist in any vendored catalog", id)
		}
	}
}
