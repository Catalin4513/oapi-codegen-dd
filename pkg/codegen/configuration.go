package codegen

import (
	"errors"
	"fmt"
)

type AdditionalImport struct {
	Alias   string `yaml:"alias,omitempty"`
	Package string `yaml:"package"`
}

// Configuration defines code generation customizations
type Configuration struct {
	// PackageName to generate the code under
	PackageName string `yaml:"package"`
	// OutputOptions are used to modify the output code in some way.
	OutputOptions OutputOptions `yaml:"output-options,omitempty"`
	// ImportMapping specifies the golang package path for each external reference
	ImportMapping map[string]string `yaml:"import-mapping,omitempty"`
	// AdditionalImports defines any additional Go imports to add to the generated code
	AdditionalImports []AdditionalImport `yaml:"additional-imports,omitempty"`
}

// Validate checks whether Configuration represent a valid configuration
func (o Configuration) Validate() error {
	if o.PackageName == "" {
		return errors.New("package name must be specified")
	}

	var errs []error
	if problems := o.OutputOptions.Validate(); problems != nil {
		for k, v := range problems {
			errs = append(errs, fmt.Errorf("`output-options` configuration for %v was incorrect: %v", k, v))
		}
	}

	err := errors.Join(errs...)
	if err != nil {
		return fmt.Errorf("failed to validate configuration: %w", err)
	}

	return nil
}

// OutputOptions are used to modify the output code in some way.
type OutputOptions struct {
	// Whether to skip go imports on the generated code
	SkipFmt bool `yaml:"skip-fmt,omitempty"`
	// Whether to skip pruning unused components on the generated code
	SkipPrune bool `yaml:"skip-prune,omitempty"`
	// Only include operations that have one of these tags. Ignored when empty.
	IncludeTags []string `yaml:"include-tags,omitempty"`
	// Exclude operations that have one of these tags. Ignored when empty.
	ExcludeTags []string `yaml:"exclude-tags,omitempty"`
	// Only include operations that have one of these operation-ids. Ignored when empty.
	IncludeOperationIDs []string `yaml:"include-operation-ids,omitempty"`
	// Exclude operations that have one of these operation-ids. Ignored when empty.
	ExcludeOperationIDs []string `yaml:"exclude-operation-ids,omitempty"`
	// Override built-in templates from user-provided files
	UserTemplates map[string]string `yaml:"user-templates,omitempty"`

	// Exclude from generation schemas with given names. Ignored when empty.
	ExcludeSchemas []string `yaml:"exclude-schemas,omitempty"`
	// The suffix used for responses types
	ResponseTypeSuffix string `yaml:"response-type-suffix,omitempty"`
	// Override the default generated client type with the value
	ClientTypeName string `yaml:"client-type-name,omitempty"`
	// Whether to use the initialism overrides
	InitialismOverrides bool `yaml:"initialism-overrides,omitempty"`
	// AdditionalInitialisms is a list of additional initialisms to use when generating names.
	// NOTE that this has no effect unless the `name-normalizer` is set to `ToCamelCaseWithInitialisms`
	AdditionalInitialisms []string `yaml:"additional-initialisms,omitempty"`
	// Whether to generate nullable type for nullable fields
	NullableType bool `yaml:"nullable-type,omitempty"`

	// DisableTypeAliasesForType allows defining which OpenAPI `type`s will explicitly not use type aliases
	// Currently supports:
	//   "array"
	DisableTypeAliasesForType []string `yaml:"disable-type-aliases-for-type"`

	// NameNormalizer is the method used to normalize Go names and types, for instance converting the text `MyApi` to `MyAPI`. Corresponds with the constants defined for `codegen.NameNormalizerFunction`
	NameNormalizer string `yaml:"name-normalizer,omitempty"`

	// EnableYamlTags adds YAML tags to generated structs, in addition to default JSON ones
	EnableYamlTags bool `yaml:"yaml-tags,omitempty"`

	// ClientResponseBytesFunction decides whether to enable the generation of a `Bytes()` method on response objects for `ClientWithResponses`
	ClientResponseBytesFunction bool `yaml:"client-response-bytes-function,omitempty"`
}

func (oo OutputOptions) Validate() map[string]string {
	if NameNormalizerFunction(oo.NameNormalizer) != NameNormalizerFunctionToCamelCaseWithInitialisms && len(oo.AdditionalInitialisms) > 0 {
		return map[string]string{
			"additional-initialisms": "You have specified `additional-initialisms`, but the `name-normalizer` is not set to `ToCamelCaseWithInitialisms`. Please specify `name-normalizer: ToCamelCaseWithInitialisms` or remove the `additional-initialisms` configuration",
		}
	}

	return nil
}
