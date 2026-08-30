package fingerprint

import (
	"encoding/json"
	"sort"
	"strings"
)

// ExtractJSONFeatures parses a JSON string and extracts its deterministic recursive schema AST.
func ExtractJSONFeatures(raw string) (FingerprintFeatures, bool) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return FingerprintFeatures{}, false
	}

	var root interface{}
	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return FingerprintFeatures{}, false
	}

	schema := buildJSONSchema(root)
	if schema == nil {
		return FingerprintFeatures{}, false
	}

	// Extract top-level keys and field signatures
	var keys []string
	var signatures []FieldSignature

	if obj, ok := root.(map[string]interface{}); ok {
		for k, v := range obj {
			keys = append(keys, k)
			signatures = append(signatures, FieldSignature{
				Key:  k,
				Type: inferJSONValueType(v),
			})
		}
		sort.Strings(keys)
		sort.Slice(signatures, func(i, j int) bool {
			return signatures[i].Key < signatures[j].Key
		})
	}

	vVendor, vProduct := DetectVendorProduct(raw)

	features := FingerprintFeatures{
		Version:         CurrentSchemaVersion,
		Format:          FormatJSON,
		FieldCount:      len(keys),
		Keys:            keys,
		FieldSignatures: signatures,
		JSONSchema:      schema,
		VendorHint:      vVendor,
		ProductHint:     vProduct,
	}

	return features, true
}

func buildJSONSchema(val interface{}) *JSONSchemaNode {
	if val == nil {
		return &JSONSchemaNode{Type: TypeNull}
	}

	switch v := val.(type) {
	case map[string]interface{}:
		node := &JSONSchemaNode{
			Type:       TypeObject,
			Properties: make(map[string]*JSONSchemaNode, len(v)),
		}
		for k, child := range v {
			node.Properties[k] = buildJSONSchema(child)
		}
		return node

	case []interface{}:
		node := &JSONSchemaNode{
			Type: TypeArray,
		}
		if len(v) > 0 {
			// Infer unified item type across elements
			node.ItemType = buildArrayItemSchema(v)
		} else {
			node.ItemType = &JSONSchemaNode{Type: TypeUnknown}
		}
		return node

	case bool:
		return &JSONSchemaNode{Type: TypeBool}

	case json.Number:
		return &JSONSchemaNode{Type: TypeNumber}

	case string:
		semanticType := DetectValueType(v)
		return &JSONSchemaNode{Type: semanticType}

	default:
		return &JSONSchemaNode{Type: TypeString}
	}
}

func buildArrayItemSchema(items []interface{}) *JSONSchemaNode {
	if len(items) == 0 {
		return &JSONSchemaNode{Type: TypeUnknown}
	}

	first := buildJSONSchema(items[0])
	if first.Type != TypeObject {
		return first
	}

	// For array of objects, merge all observed keys across objects for schema completeness
	mergedProps := make(map[string]*JSONSchemaNode)
	for _, item := range items {
		if obj, ok := item.(map[string]interface{}); ok {
			for k, v := range obj {
				if _, exists := mergedProps[k]; !exists {
					mergedProps[k] = buildJSONSchema(v)
				}
			}
		}
	}

	return &JSONSchemaNode{
		Type:       TypeObject,
		Properties: mergedProps,
	}
}

func inferJSONValueType(val interface{}) ValueType {
	if val == nil {
		return TypeNull
	}
	switch v := val.(type) {
	case map[string]interface{}:
		return TypeObject
	case []interface{}:
		return TypeArray
	case bool:
		return TypeBool
	case json.Number:
		return TypeNumber
	case string:
		return DetectValueType(v)
	default:
		return TypeString
	}
}
