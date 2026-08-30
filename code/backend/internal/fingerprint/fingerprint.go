package fingerprint

import (
	"strings"
)

// ExtractFingerprintFeatures analyzes a raw log string and extracts a deterministic,
// semantic representation of the log's structure while normalizing volatile values.
//
// This is Function 1 of the two-stage fingerprinting pipeline.
func ExtractFingerprintFeatures(log string) (features FingerprintFeatures) {
	// Guard against panics on any malformed input
	defer func() {
		if r := recover(); r != nil {
			features = FingerprintFeatures{
				Version:    CurrentSchemaVersion,
				Format:     FormatUnknown,
				Template:   NormalizeTemplate(log),
				FieldCount: 0,
			}
		}
	}()

	trimmed := strings.TrimSpace(log)
	if len(trimmed) == 0 {
		return FingerprintFeatures{
			Version: CurrentSchemaVersion,
			Format:  FormatUnknown,
		}
	}

	// 1. Check CEF (Common Event Format) - may appear standalone or encapsulated in Syslog
	if strings.Contains(trimmed, "CEF:") {
		if feat, ok := ExtractCEFFeatures(trimmed); ok {
			return feat
		}
	}

	// 2. Check LEEF (Log Event Extended Format)
	if strings.Contains(trimmed, "LEEF:") {
		if feat, ok := ExtractLEEFFeatures(trimmed); ok {
			return feat
		}
	}

	// 3. Check JSON format (single or multiline)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		if feat, ok := ExtractJSONFeatures(trimmed); ok {
			return feat
		}
	}

	// 4. Check Syslog format (RFC 5424 or RFC 3164)
	if feat, ok := ExtractSyslogFeatures(trimmed); ok {
		return feat
	}

	// 5. Check XML format
	if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") && !strings.HasPrefix(trimmed, "<0>") {
		if feat, ok := ExtractXMLFeatures(trimmed); ok {
			return feat
		}
	}

	// 6. Check Multiline (stack traces, multiline exceptions)
	if strings.Contains(trimmed, "\n") {
		if feat, ok := ExtractMultilineFeatures(trimmed); ok {
			return feat
		}
	}

	// 7. Check Key-Value format (Logfmt, Fortinet, Splunk)
	if strings.Contains(trimmed, "=") {
		if feat, ok := ExtractKVFeatures(trimmed); ok {
			return feat
		}
	}

	// 8. Check CSV / Delimited tabular logs
	if feat, ok := ExtractCSVFeatures(trimmed); ok {
		return feat
	}

	// 9. Fallback: Plain text / CLI event template extraction
	return ExtractPlainTextFeatures(trimmed)
}
