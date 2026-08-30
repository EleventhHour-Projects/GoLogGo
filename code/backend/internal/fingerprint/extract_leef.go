package fingerprint

import (
	"sort"
	"strings"
)

// ExtractLEEFFeatures parses logs formatted with IBM QRadar Log Event Extended Format (LEEF).
func ExtractLEEFFeatures(raw string) (FingerprintFeatures, bool) {
	leefIdx := strings.Index(raw, "LEEF:")
	if leefIdx == -1 {
		return FingerprintFeatures{}, false
	}

	leefPayload := raw[leefIdx+5:]
	parts := strings.Split(leefPayload, "|")
	if len(parts) < 5 {
		return FingerprintFeatures{}, false
	}

	version := strings.TrimSpace(parts[0])
	vendor := strings.TrimSpace(parts[1])
	product := strings.TrimSpace(parts[2])
	prodVersion := strings.TrimSpace(parts[3])
	eventID := strings.TrimSpace(parts[4])

	delimiter := "\t" // Default LEEF delimiter
	var extPayload string

	if strings.HasPrefix(version, "2") && len(parts) >= 7 {
		delimiter = parts[5]
		if delimiter == "x09" || delimiter == "\\t" {
			delimiter = "\t"
		}
		extPayload = strings.Join(parts[6:], "|")
	} else if len(parts) >= 6 {
		extPayload = strings.Join(parts[5:], "|")
	}

	var extFields []FieldSignature
	var keys []string

	if extPayload != "" {
		extFields, keys = parseLEEFExtensions(extPayload, delimiter)
	}

	leefInfo := &LEEFFeature{
		Version:         version,
		Vendor:          vendor,
		Product:         product,
		ProductVersion:  prodVersion,
		EventID:         eventID,
		Delimiter:       delimiter,
		ExtensionFields: extFields,
	}

	features := FingerprintFeatures{
		Version:         CurrentSchemaVersion,
		Format:          FormatLEEF,
		Delimiter:       "|",
		FieldCount:      5 + len(keys),
		Keys:            keys,
		FieldSignatures: extFields,
		LEEF:            leefInfo,
		VendorHint:      strings.ToLower(vendor),
		ProductHint:     strings.ToLower(product),
	}

	return features, true
}

func parseLEEFExtensions(payload, delimiter string) ([]FieldSignature, []string) {
	var pairs []string
	if delimiter == "\t" {
		pairs = strings.Split(payload, "\t")
	} else if len(delimiter) > 0 {
		pairs = strings.Split(payload, delimiter)
	} else {
		pairs = strings.Fields(payload)
	}

	seen := make(map[string]bool)
	var keys []string
	var signatures []FieldSignature

	for _, pair := range pairs {
		trimmed := strings.TrimSpace(pair)
		if eqIdx := strings.IndexByte(trimmed, '='); eqIdx != -1 {
			k := strings.TrimSpace(trimmed[:eqIdx])
			v := strings.TrimSpace(trimmed[eqIdx+1:])
			if k != "" && !seen[k] {
				seen[k] = true
				keys = append(keys, k)
				signatures = append(signatures, FieldSignature{
					Key:  k,
					Type: DetectValueType(v),
				})
			}
		}
	}

	sort.Strings(keys)
	sort.Slice(signatures, func(i, j int) bool {
		return signatures[i].Key < signatures[j].Key
	})

	return signatures, keys
}
