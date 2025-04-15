package codegen

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func filterOutDocument(doc libopenapi.Document, cfg FilterConfig) (libopenapi.Document, error) {
	if cfg.Include.isEmpty() && cfg.Exclude.isEmpty() {
		return doc, nil
	}

	model, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	filterOperations(&model.Model, cfg)
	filterComponentSchemaProperties(&model.Model, cfg)

	_, doc, _, errs = doc.RenderAndReload()
	if errs != nil {
		return nil, fmt.Errorf("error reloading document: %w", errs[0])
	}

	return doc, nil
}

func filterOperations(model *v3high.Document, cfg FilterConfig) {
	for path, pathItem := range model.Paths.PathItems.FromOldest() {
		if cfg.Include.Paths != nil && !slices.Contains(cfg.Include.Paths, path) {
			model.Paths.PathItems.Delete(path)
			continue
		}

		if cfg.Exclude.Paths != nil && slices.Contains(cfg.Exclude.Paths, path) {
			model.Paths.PathItems.Delete(path)
			continue
		}

		for method, op := range pathItem.GetOperations().FromOldest() {
			remove := false

			// Tags
			for _, tag := range op.Tags {
				if slices.Contains(cfg.Exclude.Tags, tag) {
					remove = true
					break
				}
			}

			if !remove && len(cfg.Include.Tags) > 0 {
				// Only include if it matches Include.Tags
				includeMatch := false
				for _, tag := range op.Tags {
					if slices.Contains(cfg.Include.Tags, tag) {
						includeMatch = true
						break
					}
				}
				if !includeMatch {
					remove = true
				}
			}

			// OperationIDs
			if cfg.Exclude.OperationIDs != nil && slices.Contains(cfg.Exclude.OperationIDs, op.OperationId) {
				remove = true
			}
			if cfg.Include.OperationIDs != nil && !slices.Contains(cfg.Include.OperationIDs, op.OperationId) {
				remove = true
			}

			if remove {
				switch method {
				case "get":
					pathItem.Get = nil
				case "post":
					pathItem.Post = nil
				case "put":
					pathItem.Put = nil
				case "delete":
					pathItem.Delete = nil
				case "patch":
					pathItem.Patch = nil
				case "head":
					pathItem.Head = nil
				case "options":
					pathItem.Options = nil
				case "trace":
					pathItem.Trace = nil
				}
			}
		}
	}
}

func filterComponentSchemaProperties(model *v3high.Document, cfg FilterConfig) {
	if model.Components == nil || model.Components.Schemas == nil {
		return
	}

	includeMode := len(cfg.Include.SchemaProperties) > 0
	excludeMode := len(cfg.Exclude.SchemaProperties) > 0

	var propsFilter map[string][]string
	switch {
	case includeMode:
		propsFilter = cfg.Include.SchemaProperties
	case excludeMode:
		propsFilter = cfg.Exclude.SchemaProperties
	default:
		return
	}

	for schemaName, schemaProxy := range model.Components.Schemas.FromOldest() {
		filteredProps, ok := propsFilter[schemaName]
		if !ok {
			continue
		}

		schema := schemaProxy.Schema()
		if schema == nil || schema.Properties == nil {
			continue
		}

		var copiedKeys []string
		for prop := range schema.Properties.KeysFromOldest() {
			copiedKeys = append(copiedKeys, prop)
		}

		for _, propName := range copiedKeys {
			isRequired := slices.Contains(schema.Required, propName)
			if isRequired {
				continue
			}

			if includeMode && !slices.Contains(filteredProps, propName) {
				schema.Properties.Delete(propName)
			}

			if excludeMode && slices.Contains(filteredProps, propName) {
				schema.Properties.Delete(propName)
			}
		}
	}
}
