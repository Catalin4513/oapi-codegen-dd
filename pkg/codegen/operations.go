// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codegen

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

// OperationDefinition describes an Operation
// ID The operation_id description from Swagger, used to generate function names
// PathParams Parameters in the path, eg, /path/:param
// HeaderParams Parameters in HTTP headers
// QueryParams Parameters in the query, /path?param
// TypeDefinitions These are all the types we need to define for this operation
// BodyRequired Whether the body is required for this operation
// Bodies The list of bodies for which to generate handlers.
// Responses The list of responses that can be accepted by handlers.
// Summary string from Swagger, used to generate a comment
// Method GET, POST, DELETE, etc.
// Path The Swagger path for the operation, like /resource/{id}
// Spec The OpenAPI3 Operation object
type OperationDefinition struct {
	ID          string
	Summary     string
	Description string
	Method      string
	Path        string

	PathParams      []ParameterDefinition
	HeaderParams    []ParameterDefinition
	QueryParams     []ParameterDefinition
	TypeDefinitions []TypeDefinition
	BodyRequired    bool

	Body     *RequestBodyDefinition
	Response ResponseDefinition

	Spec *openapi3.Operation
}

// Params returns the list of all parameters except Path parameters. Path parameters
// are handled differently from the rest, since they're mandatory.
func (o OperationDefinition) Params() []ParameterDefinition {
	result := append(o.QueryParams, o.HeaderParams...)
	return result
}

// AllParams returns all parameters
func (o OperationDefinition) AllParams() []ParameterDefinition {
	result := append(o.QueryParams, o.HeaderParams...)
	result = append(result, o.PathParams...)
	return result
}

// RequiresParamObject indicates If we have parameters other than path parameters, they're bundled into an
// object. Returns true if we have any of those. This is used from the template
// engine.
func (o OperationDefinition) RequiresParamObject() bool {
	return len(o.Params()) > 0
}

// SummaryAsComment returns the Operations summary as a multi line comment
func (o OperationDefinition) SummaryAsComment() string {
	if o.Summary == "" {
		return ""
	}
	trimmed := strings.TrimSuffix(o.Summary, "\n")
	parts := strings.Split(trimmed, "\n")
	for i, p := range parts {
		parts[i] = "// " + p
	}
	return strings.Join(parts, "\n")
}

// FilterParameterDefinitionByType returns the subset of the specified parameters which are of the
// specified type.
func FilterParameterDefinitionByType(params []ParameterDefinition, in string) []ParameterDefinition {
	var out []ParameterDefinition
	for _, p := range params {
		if p.In == in {
			out = append(out, p)
		}
	}
	return out
}

// OperationDefinitions returns all operations for a swagger definition.
func OperationDefinitions(swagger *openapi3.T) ([]OperationDefinition, error) {
	var operations []OperationDefinition

	if swagger == nil || swagger.Paths == nil {
		return operations, nil
	}

	for _, requestPath := range SortedMapKeys(swagger.Paths.Map()) {
		pathItem := swagger.Paths.Value(requestPath)
		// These are parameters defined for all methods on a given path. They
		// are shared by all methods.
		globalParams, err := DescribeParameters(pathItem.Parameters, nil)
		if err != nil {
			return nil, fmt.Errorf("error describing global parameters for %s: %s",
				requestPath, err)
		}

		// Each path can have a number of operations, POST, GET, OPTIONS, etc.
		pathOps := pathItem.Operations()
		for _, opName := range SortedMapKeys(pathOps) {
			op := pathOps[opName]
			if pathItem.Servers != nil {
				op.Servers = &pathItem.Servers
			}
			// We rely on OperationID to generate function names, it's required
			if op.OperationID == "" {
				op.OperationID, err = generateDefaultOperationID(opName, requestPath)
				if err != nil {
					return nil, fmt.Errorf("error generating default OperationID for %s/%s: %s",
						opName, requestPath, err)
				}
			} else {
				op.OperationID = nameNormalizer(op.OperationID)
			}
			op.OperationID = typeNamePrefix(op.OperationID) + op.OperationID

			var typeDefs []TypeDefinition

			// These are parameters defined for the specific path method that
			// we're iterating over.
			localParams, err := DescribeParameters(op.Parameters, []string{op.OperationID + "Params"})
			if err != nil {
				return nil, fmt.Errorf("error describing global parameters for %s/%s: %s",
					opName, requestPath, err)
			}
			// All the parameters required by a handler are the union of the
			// global parameters and the local parameters.
			allParams, err := CombineOperationParameters(globalParams, localParams)
			if err != nil {
				return nil, err
			}

			// Order the path parameters to match the order as specified in
			// the path, not in the swagger spec, and validate that the parameter
			// names match, as downstream code depends on that.
			pathParams := FilterParameterDefinitionByType(allParams, "path")
			pathParams, err = SortParamsByPath(requestPath, pathParams)
			if err != nil {
				return nil, err
			}

			bodyDefinition, bodyTypeDef, err := createBodyDefinition(op.OperationID, op.RequestBody)
			if err != nil {
				return nil, fmt.Errorf("error generating body definitions: %w", err)
			}

			if bodyTypeDef != nil {
				typeDefs = append(typeDefs, *bodyTypeDef)
			}
			if bodyDefinition != nil {
				typeDefs = append(typeDefs, bodyDefinition.Schema.AdditionalTypes...)
			}

			responseDef, responseTypes, err := getOperationResponses(op.OperationID, op.Responses.Map())
			if err != nil {
				return nil, fmt.Errorf("error getting operation responses: %w", err)
			}
			typeDefs = append(typeDefs, responseTypes...)

			opDef := OperationDefinition{
				ID: nameNormalizer(op.OperationID),
				// Replace newlines in summary.
				Summary:      op.Summary,
				Description:  op.Description,
				PathParams:   pathParams,
				HeaderParams: FilterParameterDefinitionByType(allParams, "header"),
				QueryParams:  FilterParameterDefinitionByType(allParams, "query"),
				Method:       opName,
				Path:         requestPath,
				Spec:         op,
				Response:     *responseDef,

				Body:            bodyDefinition,
				TypeDefinitions: typeDefs,
			}

			if op.RequestBody != nil {
				opDef.BodyRequired = op.RequestBody.Value.Required
			}

			// Generate all the type definitions needed for this operation
			opDef.TypeDefinitions = append(opDef.TypeDefinitions, GenerateTypeDefsForOperation(opDef)...)

			operations = append(operations, opDef)
		}
	}
	return operations, nil
}

func generateDefaultOperationID(opName string, requestPath string) (string, error) {
	var operationId = strings.ToLower(opName)

	if opName == "" {
		return "", ErrOperationNameEmpty
	}

	if requestPath == "" {
		return "", ErrRequestPathEmpty
	}

	for _, part := range strings.Split(requestPath, "/") {
		if part != "" {
			operationId = operationId + "-" + part
		}
	}

	return nameNormalizer(operationId), nil
}

func GenerateTypeDefsForOperation(op OperationDefinition) []TypeDefinition {
	var typeDefs []TypeDefinition
	// Start with the params object itself
	if len(op.Params()) != 0 {
		typeDefs = append(typeDefs, GenerateParamsTypes(op)...)
	}

	// Now, go through all the additional types we need to declare.
	for _, param := range op.AllParams() {
		typeDefs = append(typeDefs, param.Schema.AdditionalTypes...)
	}

	return typeDefs
}

// GenerateParamsTypes defines the schema for a parameters definition object
// which encapsulates all the query, header and cookie parameters for an operation.
func GenerateParamsTypes(op OperationDefinition) []TypeDefinition {
	var typeDefs []TypeDefinition

	objectParams := op.QueryParams
	objectParams = append(objectParams, op.HeaderParams...)

	typeName := op.ID + "Params"

	s := Schema{}
	for _, param := range objectParams {
		pSchema := param.Schema
		param.Style()
		if pSchema.HasAdditionalProperties {
			propRefName := strings.Join([]string{typeName, param.GoName()}, "_")
			pSchema.RefType = propRefName
			typeDefs = append(typeDefs, TypeDefinition{
				TypeName: propRefName,
				Schema:   param.Schema,
			})
		}
		prop := Property{
			Description:   param.Spec.Description,
			JsonFieldName: param.ParamName,
			Required:      param.Required,
			Schema:        pSchema,
			NeedsFormTag:  param.Style() == "form",
			Extensions:    param.Spec.Extensions,
		}
		s.Properties = append(s.Properties, prop)
	}

	s.Description = op.Description
	s.GoType = GenStructFromSchema(s)

	td := TypeDefinition{
		TypeName: typeName,
		Schema:   s,
	}
	return append(typeDefs, td)
}

// GenerateTypesForOperations generates code for all types produced within operations
func GenerateTypesForOperations(t *template.Template, ops []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	addTypes, err := GenerateTemplates([]string{"param-types.tmpl", "request-bodies.tmpl"}, t, ops)
	if err != nil {
		return "", fmt.Errorf("error generating type boilerplate for operations: %w", err)
	}
	if _, err := w.WriteString(addTypes); err != nil {
		return "", fmt.Errorf("error writing boilerplate to buffer: %w", err)
	}

	// Generate boiler plate for all additional types.
	var td []TypeDefinition
	for _, op := range ops {
		td = append(td, op.TypeDefinitions...)
	}

	addProps, err := GenerateAdditionalPropertyBoilerplate(t, td)
	if err != nil {
		return "", fmt.Errorf("error generating additional properties boilerplate for operations: %w", err)
	}

	if _, err := w.WriteString("\n"); err != nil {
		return "", fmt.Errorf("error generating additional properties boilerplate for operations: %w", err)
	}

	if _, err := w.WriteString(addProps); err != nil {
		return "", fmt.Errorf("error generating additional properties boilerplate for operations: %w", err)
	}

	if err = w.Flush(); err != nil {
		return "", fmt.Errorf("error flushing output buffer for server interface: %w", err)
	}

	return buf.String(), nil
}

// GenerateClient uses the template engine to generate the function which registers our wrappers
// as Echo path handlers.
func GenerateClient(t *template.Template, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"client.tmpl"}, t, ops)
}

// GenerateClientWithResponses generates a client which extends the basic client which does response
// unmarshaling.
func GenerateClientWithResponses(t *template.Template, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"client-with-responses.tmpl"}, t, ops)
}

// GenerateTemplates used to generate templates
func GenerateTemplates(templates []string, t *template.Template, ops interface{}) (string, error) {
	var generatedTemplates []string
	for _, tmpl := range templates {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)

		if err := t.ExecuteTemplate(w, tmpl, ops); err != nil {
			return "", fmt.Errorf("error generating %s: %s", tmpl, err)
		}
		if err := w.Flush(); err != nil {
			return "", fmt.Errorf("error flushing output buffer for %s: %s", tmpl, err)
		}
		generatedTemplates = append(generatedTemplates, buf.String())
	}

	return strings.Join(generatedTemplates, "\n"), nil
}
