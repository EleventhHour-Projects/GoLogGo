package fingerprint

import (
	"sort"
	"strconv"
	"strings"
)

// ExtractCEFFeatures parses logs formatted with ArcSight Common Event Format (CEF).
func ExtractCEFFeatures(raw string) (FingerprintFeatures, bool) {
	cefIdx := strings.Index(raw, "CEF:")
	if cefIdx == -1 {
		return FingerprintFeatures{}, false
	}

	cefPayload := raw[cefIdx+4:]
	// Header consists of 7 pipe-delimited fields: Version|Vendor|Product|DevVersion|EventClassID|Name|Severity|Extension
	parts := splitCEFHeader(cefPayload)
	if len(parts) < 7 {
		return FingerprintFeatures{}, false
	}

	version, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	vendor := strings.TrimSpace(parts[1])
	product := strings.TrimSpace(parts[2])
	devVersion := strings.TrimSpace(parts[3])
	eventClassID := strings.TrimSpace(parts[4])
	severity := strings.TrimSpace(parts[6])

	var extFields []FieldSignature
	var keys []string

	if len(parts) >= 8 {
		extPayload := parts[7]
		pairs, _, _, _, _ := parseKVPairs(extPayload)
		seen := make(map[string]bool)
		for _, p := range pairs {
			if !seen[p.key] {
				seen[p.key] = true
				keys = append(keys, p.key)
				extFields = append(extFields, FieldSignature{
					Key:  p.key,
					Type: p.valType,
				})
			}
		}
		sort.Strings(keys)
		sort.Slice(extFields, func(i, j int) bool {
			return extFields[i].Key < extFields[j].Key
		})
	}

	cefInfo := &CEFFeature{
		Version:            version,
		DeviceVendor:       vendor,
		DeviceProduct:      product,
		DeviceVersion:      devVersion,
		DeviceEventClassID: eventClassID,
		Severity:           severity,
		ExtensionFields:    extFields,
	}

	features := FingerprintFeatures{
		Version:         CurrentSchemaVersion,
		Format:          FormatCEF,
		Delimiter:       "|",
		FieldCount:      7 + len(keys),
		Keys:            keys,
		FieldSignatures: extFields,
		CEF:             cefInfo,
		VendorHint:      strings.ToLower(vendor),
		ProductHint:     strings.ToLower(product),
	}

	return features, true
}

func splitCEFHeader(s string) []string {
	var parts []string
	var current strings.Builder
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if c == '|' {
			parts = append(parts, current.String())
			current.Reset()
			if len(parts) == 7 {
				// The rest is the extension field
				if i+1 < len(s) {
					parts = append(parts, s[i+1:])
				}
				return parts
			}
			continue
		}

		current.WriteByte(c)
	}

	if current.Len() > 0 || len(parts) > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
