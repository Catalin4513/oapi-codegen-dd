// Copyright 2025 DoorDash, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package codegen

import (
	"fmt"
	"iter"
	"log/slog"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func pruneSchema(doc libopenapi.Document) error {
	// Build model once - we'll mutate it directly and skip RenderAndReload
	model, err := doc.BuildV3Model()
	if err != nil {
		return fmt.Errorf("error building model: %w", err)
	}

	// Aggressively remove everything we don't generate code for
	slog.Debug("Pruning: removing webhooks, security schemes, callbacks, component examples, links")
	model.Model.Webhooks = nil
	if model.Model.Components != nil {
		// Set to nil - we don't generate code for these
		model.Model.Components.SecuritySchemes = nil
		model.Model.Components.Callbacks = nil
		model.Model.Components.Examples = nil
		model.Model.Components.Links = nil
	}

	// Iteratively prune unreferenced components
	iteration := 0
	maxIterations := 1000

	// Iteratively prune unreferenced components until nothing left to remove
	for {
		iteration++
		if iteration > maxIterations {
			return fmt.Errorf("pruning exceeded maximum iterations (%d), possible infinite loop", maxIterations)
		}

		refs := findOperationRefs(&model.Model)
		slog.Debug("Found operation refs", "count", len(refs), "iteration", iteration)

		countRemoved := removeOrphanedComponents(&model.Model, refs)
		slog.Debug("Removed orphaned components", "count", countRemoved, "iteration", iteration)

		// No RenderAndReload! We just keep mutating the model directly
		// The code generator will use this mutated model via BuildV3Model() which returns the cached model

		if countRemoved < 1 {
			// Done pruning
			return nil
		}
	}
}

func removeOrphanedComponents(model *v3high.Document, refs []string) int {
	if model.Components == nil {
		return 0
	}

	countRemoved := 0

	if model.Components.Schemas != nil {
		for _, key := range getComponentKeys(model.Components.Schemas.KeysFromOldest()) {
			ref := fmt.Sprintf("#/components/schemas/%s", key)
			if !slices.Contains(refs, ref) {
				countRemoved++
				model.Components.Schemas.Delete(key)
			}
		}
	}

	if model.Components.Parameters != nil {
		for _, key := range getComponentKeys(model.Components.Parameters.KeysFromOldest()) {
			ref := fmt.Sprintf("#/components/parameters/%s", key)
			if !slices.Contains(refs, ref) {
				countRemoved++
				model.Components.Parameters.Delete(key)
			}
		}
	}

	if model.Components.RequestBodies != nil {
		for _, key := range getComponentKeys(model.Components.RequestBodies.KeysFromOldest()) {
			ref := fmt.Sprintf("#/components/requestBodies/%s", key)
			if !slices.Contains(refs, ref) {
				countRemoved++
				model.Components.RequestBodies.Delete(key)
			}
		}
	}

	if model.Components.Responses != nil {
		for _, key := range getComponentKeys(model.Components.Responses.KeysFromOldest()) {
			ref := fmt.Sprintf("#/components/responses/%s", key)
			if !slices.Contains(refs, ref) {
				countRemoved++
				model.Components.Responses.Delete(key)
			}
		}
	}

	if model.Components.Headers != nil {
		for _, key := range getComponentKeys(model.Components.Headers.KeysFromOldest()) {
			ref := fmt.Sprintf("#/components/headers/%s", key)
			if !slices.Contains(refs, ref) {
				countRemoved++
				model.Components.Headers.Delete(key)
			}
		}
	}

	// Note: Links, Callbacks, Examples are set to nil in pruneSchema, so we don't need to prune them here

	return countRemoved
}

func findOperationRefs(model *v3high.Document) []string {
	refSet := make(map[string]bool)

	if model.Paths == nil || model.Paths.PathItems == nil {
		return []string{}
	}

	// Walk all operations and collect refs
	for _, pathItem := range model.Paths.PathItems.FromOldest() {
		// Collect path-level parameters
		for _, param := range pathItem.Parameters {
			collectRefFromProxy(param, refSet)
		}

		// Collect operation-level refs
		for _, op := range pathItem.GetOperations().FromOldest() {
			// Request body
			if op.RequestBody != nil {
				collectRefFromProxy(op.RequestBody, refSet)
			}

			// Parameters
			for _, param := range op.Parameters {
				collectRefFromProxy(param, refSet)
			}

			// Responses
			if op.Responses != nil {
				if op.Responses.Default != nil {
					collectRefFromProxy(op.Responses.Default, refSet)
				}
				for _, resp := range op.Responses.Codes.FromOldest() {
					collectRefFromProxy(resp, refSet)
				}
			}
		}
	}

	// Walk component parameters/requestBodies/responses/headers to collect schema refs
	// These components are only kept if referenced by operations (collected above)
	// But we need to walk them to find schemas they reference
	if model.Components != nil {
		if model.Components.Parameters != nil {
			for _, param := range model.Components.Parameters.FromOldest() {
				collectRefFromProxy(param, refSet)
			}
		}
		if model.Components.RequestBodies != nil {
			for _, reqBody := range model.Components.RequestBodies.FromOldest() {
				collectRefFromProxy(reqBody, refSet)
			}
		}
		if model.Components.Responses != nil {
			for _, resp := range model.Components.Responses.FromOldest() {
				collectRefFromProxy(resp, refSet)
			}
		}
		if model.Components.Headers != nil {
			for _, header := range model.Components.Headers.FromOldest() {
				collectRefFromProxy(header, refSet)
			}
		}
		// Walk component schemas to collect refs to other schemas (for composition)
		if model.Components.Schemas != nil {
			for _, schemaProxy := range model.Components.Schemas.FromOldest() {
				if schemaProxy == nil {
					continue
				}
				// Check if the schema proxy itself is a $ref
				if schemaRef := schemaProxy.GoLow().GetReference(); schemaRef != "" {
					refSet[schemaRef] = true
				}
				// Collect refs from the schema's content (allOf, oneOf, anyOf, properties, etc.)
				collectSchemaRefs(schemaProxy.Schema(), refSet)
			}
		}
	}

	refs := make([]string, 0, len(refSet))
	for r := range refSet {
		refs = append(refs, r)
	}
	slices.Sort(refs)
	slog.Debug("All collected refs", "refs", refs)
	return refs
}

// addParentSchemaRef adds the parent schema reference if the given ref is a property reference
// e.g., if ref is "#/components/schemas/Foo/properties/bar", also add "#/components/schemas/Foo"
func addParentSchemaRef(ref string, refSet map[string]bool) {
	// Check if this is a property reference pattern: #/components/schemas/{name}/properties/{prop}
	if strings.Contains(ref, "/properties/") {
		// Extract the parent schema reference
		parts := strings.Split(ref, "/properties/")
		if len(parts) > 0 {
			parentRef := parts[0]
			if !refSet[parentRef] {
				refSet[parentRef] = true
			}
		}
	}
}

func collectSchemaRefs(schema *base.Schema, refSet map[string]bool) {
	if schema == nil {
		return
	}

	// Check if this schema is a $ref
	if ref := schema.GoLow().GetReference(); ref != "" {
		if refSet[ref] {
			return
		}
		refSet[ref] = true
		// Also add parent schema if this is a property reference
		// e.g., if ref is "#/components/schemas/Foo/properties/bar", also add "#/components/schemas/Foo"
		addParentSchemaRef(ref, refSet)
	}

	// Traverse object properties
	if schema.Properties != nil {
		for _, prop := range schema.Properties.FromOldest() {
			if prop == nil {
				continue
			}
			// Check for $ref on the property itself
			if pRef := prop.GoLow().GetReference(); pRef != "" {
				if !refSet[pRef] {
					refSet[pRef] = true
					// Also add parent schema if this is a property reference
					addParentSchemaRef(pRef, refSet)
					// Don't return — we still want to walk its schema if possible
				} else {
					continue
				}
			}
			collectSchemaRefs(prop.Schema(), refSet)
		}
	}

	// Traverse array items
	if items := schema.Items; items != nil && items.IsA() && items.A != nil {
		if iRef := items.A.GoLow().GetReference(); iRef != "" {
			if !refSet[iRef] {
				refSet[iRef] = true
				// keep walking
			} else {
				return
			}
		}
		collectSchemaRefs(items.A.Schema(), refSet)
	}

	// Traverse additionalProperties
	if ap := schema.AdditionalProperties; ap != nil && ap.IsA() && ap.A != nil {
		if aRef := ap.A.GoLow().GetReference(); aRef != "" {
			if !refSet[aRef] {
				refSet[aRef] = true
			} else {
				return
			}
		}
		collectSchemaRefs(ap.A.Schema(), refSet)
	}

	// allOf / oneOf / anyOf / not
	for _, group := range [][]*base.SchemaProxy{schema.AllOf, schema.OneOf, schema.AnyOf, {schema.Not}} {
		for _, sp := range group {
			if sp == nil {
				continue
			}
			if sRef := sp.GoLow().GetReference(); sRef != "" {
				if !refSet[sRef] {
					refSet[sRef] = true
				} else {
					continue
				}
			}
			collectSchemaRefs(sp.Schema(), refSet)
		}
	}

	// Note: We don't collect refs from x-extensions because we don't parse them for code generation
	// Extensions can contain $refs, but we leave them as-is and don't follow them
}

func getComponentKeys(component iter.Seq[string]) []string {
	keys := make([]string, 0)
	for k := range component {
		keys = append(keys, k)
	}
	return keys
}

// collectRefFromProxy collects the $ref from a proxy object and all schema refs from its content
// This works for Parameters, RequestBodies, Responses, Headers - anything with GoLow().GetReference()
func collectRefFromProxy(proxy interface{}, refSet map[string]bool) {
	if proxy == nil {
		return
	}

	// Collect the proxy's own $ref (e.g., $ref to component parameter)
	// Try different types that have GoLow().GetReference()
	switch p := proxy.(type) {
	case *v3high.Parameter:
		if ref := p.GoLow().GetReference(); ref != "" {
			slog.Debug("Collected parameter ref", "ref", ref)
			refSet[ref] = true
		}
	case *v3high.RequestBody:
		if ref := p.GoLow().GetReference(); ref != "" {
			refSet[ref] = true
		}
	case *v3high.Response:
		if ref := p.GoLow().GetReference(); ref != "" {
			refSet[ref] = true
		}
	case *v3high.Header:
		if ref := p.GoLow().GetReference(); ref != "" {
			refSet[ref] = true
		}
	}

	// Collect schema refs from the proxy's content
	switch v := proxy.(type) {
	case *v3high.Parameter:
		collectSchemaProxy(v.Schema, refSet)
		if v.Content != nil {
			for _, mediaType := range v.Content.FromOldest() {
				collectSchemaProxy(mediaType.Schema, refSet)
			}
		}

	case *v3high.RequestBody:
		if v.Content != nil {
			for _, mediaType := range v.Content.FromOldest() {
				collectSchemaProxy(mediaType.Schema, refSet)
			}
		}

	case *v3high.Response:
		if v.Content != nil {
			for _, mediaType := range v.Content.FromOldest() {
				collectSchemaProxy(mediaType.Schema, refSet)
			}
		}
		if v.Headers != nil {
			for _, header := range v.Headers.FromOldest() {
				collectRefFromProxy(header, refSet)
			}
		}

	case *v3high.Header:
		collectSchemaProxy(v.Schema, refSet)
		if v.Content != nil {
			for _, mediaType := range v.Content.FromOldest() {
				collectSchemaProxy(mediaType.Schema, refSet)
			}
		}
	}
}

// collectSchemaProxy collects refs from a SchemaProxy (both the proxy's own $ref and its schema content)
func collectSchemaProxy(schemaProxy *base.SchemaProxy, refSet map[string]bool) {
	if schemaProxy == nil {
		return
	}
	// Collect the proxy's own $ref
	if ref := schemaProxy.GoLow().GetReference(); ref != "" {
		refSet[ref] = true
	}
	// Collect refs from the schema's content
	collectSchemaRefs(schemaProxy.Schema(), refSet)
}
