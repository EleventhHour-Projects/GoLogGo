package fingerprint

import (
	"strings"
)

// ExtractPlainTextFeatures generates normalized message templates and token sequences for plain text logs.
func ExtractPlainTextFeatures(raw string) FingerprintFeatures {
	normTemplate := NormalizeTemplate(raw)
	tokens := extractStructuralTokens(normTemplate)
	vVendor, vProduct := DetectVendorProduct(raw)

	return FingerprintFeatures{
		Version:          CurrentSchemaVersion,
		Format:           FormatPlainText,
		FieldCount:       len(tokens),
		Template:         normTemplate,
		StructuralTokens: tokens,
		VendorHint:       vVendor,
		ProductHint:      vProduct,
	}
}

func extractStructuralTokens(template string) []string {
	words := strings.Fields(template)
	var tokens []string
	for _, w := range words {
		// Clean punctuation
		cleaned := strings.Trim(w, `()[]{}:,"'`)
		if cleaned != "" {
			tokens = append(tokens, cleaned)
		}
	}
	return tokens
}
