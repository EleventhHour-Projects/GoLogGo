package fingerprint

import (
	"sort"
	"strings"
	"unicode"
)

// ExtractKVFeatures parses Key-Value formatted logs (e.g. Fortinet, Splunk, Logfmt).
func ExtractKVFeatures(raw string) (FingerprintFeatures, bool) {
	pairs, sep, delim, quoteStyle, hasDups := parseKVPairs(raw)
	if len(pairs) < 2 { // At least 2 key=value pairs required to confidently classify as KV format
		return FingerprintFeatures{}, false
	}

	keysMap := make(map[string]ValueType)
	var keys []string
	var signatures []FieldSignature

	for _, p := range pairs {
		if _, exists := keysMap[p.key]; !exists {
			keysMap[p.key] = p.valType
			keys = append(keys, p.key)
			signatures = append(signatures, FieldSignature{
				Key:  p.key,
				Type: p.valType,
			})
		}
	}

	sort.Strings(keys)
	sort.Slice(signatures, func(i, j int) bool {
		return signatures[i].Key < signatures[j].Key
	})

	vVendor, vProduct := DetectVendorProduct(raw)

	kvFeature := &KVFeature{
		Separator:     sep,
		Delimiter:     delim,
		Fields:        signatures,
		QuotingStyle:  quoteStyle,
		HasDuplicates: hasDups,
	}

	features := FingerprintFeatures{
		Version:         CurrentSchemaVersion,
		Format:          FormatKeyValue,
		Delimiter:       delim,
		FieldCount:      len(keys),
		Keys:            keys,
		FieldSignatures: signatures,
		KV:              kvFeature,
		VendorHint:      vVendor,
		ProductHint:     vProduct,
	}

	return features, true
}

type kvPair struct {
	key     string
	val     string
	valType ValueType
}

func parseKVPairs(raw string) ([]kvPair, string, string, string, bool) {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, "", "", "", false
	}

	var pairs []kvPair
	seenKeys := make(map[string]bool)
	hasDups := false
	quoteStyle := ""
	separator := "="
	delimiter := " "

	// Detect if commas are used between pairs
	commaCount := strings.Count(trimmed, ",")
	spaceCount := strings.Count(trimmed, " ")
	if commaCount > spaceCount && commaCount > 1 {
		delimiter = ","
	}

	n := len(trimmed)
	i := 0

	for i < n {
		// Skip leading delimiter/whitespace
		for i < n && (unicode.IsSpace(rune(trimmed[i])) || trimmed[i] == ',') {
			i++
		}
		if i >= n {
			break
		}

		// Read key
		keyStart := i
		for i < n && trimmed[i] != '=' && trimmed[i] != ':' && !unicode.IsSpace(rune(trimmed[i])) && trimmed[i] != ',' {
			i++
		}
		if i >= n || (trimmed[i] != '=' && trimmed[i] != ':') {
			// Not a valid key=value pair start, skip word
			for i < n && !unicode.IsSpace(rune(trimmed[i])) && trimmed[i] != ',' {
				i++
			}
			continue
		}

		key := trimmed[keyStart:i]
		if len(key) == 0 {
			i++
			continue
		}

		sepChar := trimmed[i]
		separator = string(sepChar)
		i++ // skip '=' or ':'

		// Read value
		if i >= n {
			// Empty value at end
			pairs = append(pairs, kvPair{key: key, val: "", valType: TypeString})
			if seenKeys[key] {
				hasDups = true
			}
			seenKeys[key] = true
			break
		}

		var val string
		if trimmed[i] == '"' {
			quoteStyle = "double"
			i++
			valStart := i
			for i < n && trimmed[i] != '"' {
				if trimmed[i] == '\\' && i+1 < n {
					i += 2
					continue
				}
				i++
			}
			val = trimmed[valStart:i]
			if i < n && trimmed[i] == '"' {
				i++
			}
		} else if trimmed[i] == '\'' {
			quoteStyle = "single"
			i++
			valStart := i
			for i < n && trimmed[i] != '\'' {
				if trimmed[i] == '\\' && i+1 < n {
					i += 2
					continue
				}
				i++
			}
			val = trimmed[valStart:i]
			if i < n && trimmed[i] == '\'' {
				i++
			}
		} else {
			// Unquoted value
			valStart := i
			for i < n && !unicode.IsSpace(rune(trimmed[i])) && trimmed[i] != ',' {
				i++
			}
			val = trimmed[valStart:i]
		}

		if seenKeys[key] {
			hasDups = true
		}
		seenKeys[key] = true

		valType := DetectValueType(val)
		pairs = append(pairs, kvPair{
			key:     key,
			val:     val,
			valType: valType,
		})
	}

	return pairs, separator, delimiter, quoteStyle, hasDups
}
