package adapter

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestTargetInfosExposeAllCanonicalCategoryStatuses(t *testing.T) {
	valid := map[string]bool{
		categoryStatusSupported:   true,
		categoryStatusUnsupported: true,
		categoryStatusMapped:      true,
		categoryStatusPartial:     true,
	}
	for _, info := range TargetInfos() {
		if len(info.CategoryStatuses) != len(categoryOrder) {
			t.Errorf("%s category status count = %d, want %d", info.ID, len(info.CategoryStatuses), len(categoryOrder))
		}
		unsupported := make(map[string]bool, len(info.UnsupportedCategories))
		for _, category := range info.UnsupportedCategories {
			unsupported[category] = true
		}
		for _, category := range categoryOrder {
			status := info.CategoryStatuses[category]
			if !valid[status] {
				t.Errorf("%s %s status = %q", info.ID, category, status)
			}
			if (status == categoryStatusUnsupported) != unsupported[category] {
				t.Errorf("%s %s status %q disagrees with renderer unsupported categories", info.ID, category, status)
			}
		}
	}
}

func TestTargetInfosMatchCanonicalMappings(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", ".."))
	for _, info := range TargetInfos() {
		mapped, err := targetCategoryStatus(root, info.ID, true)
		if err != nil {
			t.Fatalf("targetCategoryStatus(%s) error = %v", info.ID, err)
		}
		if !reflect.DeepEqual(info.CategoryStatuses, mapped) {
			t.Errorf("%s category statuses = %#v, mappings = %#v", info.ID, info.CategoryStatuses, mapped)
		}
	}
}
