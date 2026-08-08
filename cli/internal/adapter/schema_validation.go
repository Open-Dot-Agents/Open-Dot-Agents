package adapter

import (
	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/schema"
)

const (
	categoryStatusSupported   = schema.CategoryStatusSupported
	categoryStatusUnsupported = schema.CategoryStatusUnsupported
	categoryStatusMapped      = schema.CategoryStatusMapped
	categoryStatusPartial     = schema.CategoryStatusPartial
)

func validateContracts(root, target string, strict bool) error {
	return schema.ValidateContracts(root, target, strict)
}

func targetCategoryStatus(root, target string, strict bool) (map[string]string, error) {
	return schema.TargetCategoryStatus(root, target, strict)
}

func validateCategoryStatusAlignment(target string, statusByCategory map[string]string, unsupportedCategories []string) error {
	return schema.ValidateCategoryStatusAlignment(target, statusByCategory, unsupportedCategories)
}

func validateAgentsManifestCompatibility(root string) error {
	return schema.ValidateAgentsManifestCompatibility(root)
}
