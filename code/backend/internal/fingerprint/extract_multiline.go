package fingerprint

import (
	"regexp"
	"strings"
)

var (
	stackTracePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?m)^\s+at\s+([a-zA-Z0-9_$.]+)\(`),
		regexp.MustCompile(`(?m)^Caused by:\s+([a-zA-Z0-9_$.]+)`),
		regexp.MustCompile(`(?m)^\s+File\s+"[^"]+",\s+line\s+\d+`),
		regexp.MustCompile(`(?m)^Traceback \(most recent call last\):`),
		regexp.MustCompile(`(?m)^goroutine \d+ \[[^\]]+\]:`),
	}
)

// ExtractMultilineFeatures parses multi-line log events, such as application stack traces and multiline errors.
func ExtractMultilineFeatures(raw string) (FingerprintFeatures, bool) {
	lines := strings.Split(raw, "\n")
	var nonEmptyLines []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if len(trimmed) > 0 {
			nonEmptyLines = append(nonEmptyLines, trimmed)
		}
	}

	if len(nonEmptyLines) <= 1 {
		return FingerprintFeatures{}, false
	}

	isStackTrace := false
	for _, pat := range stackTracePatterns {
		if pat.MatchString(raw) {
			isStackTrace = true
			break
		}
	}

	// Extract normalized signature for significant lines
	var lineSignatures []string
	prefixType := "text"

	// First line is typically the main error message or trigger
	firstLineTemplate := NormalizeTemplate(nonEmptyLines[0])
	lineSignatures = append(lineSignatures, firstLineTemplate)

	// Check if lines are indented or stack-trace calls
	for i := 1; i < len(nonEmptyLines); i++ {
		line := nonEmptyLines[i]
		if strings.HasPrefix(line, "at ") {
			prefixType = "java_stack"
			// Normalize java line e.g. "at com.example.Service.process(Service.java:123)" -> "at com.example.Service.process"
			if idx := strings.IndexByte(line, '('); idx != -1 {
				lineSignatures = append(lineSignatures, line[:idx])
			} else {
				lineSignatures = append(lineSignatures, line)
			}
		} else if strings.HasPrefix(line, "Caused by: ") {
			lineSignatures = append(lineSignatures, NormalizeTemplate(line))
		} else if strings.HasPrefix(line, "File ") {
			prefixType = "python_stack"
			lineSignatures = append(lineSignatures, NormalizeTemplate(line))
		} else if i < 5 { // Include up to first 5 lines for general multiline logs
			lineSignatures = append(lineSignatures, NormalizeTemplate(line))
		}
	}

	vVendor, vProduct := DetectVendorProduct(raw)

	multilineInfo := &MultilineFeature{
		LineCount:      len(nonEmptyLines),
		IsStackTrace:   isStackTrace,
		PrefixType:     prefixType,
		LineSignatures: lineSignatures,
	}

	features := FingerprintFeatures{
		Version:     CurrentSchemaVersion,
		Format:      FormatMultiline,
		FieldCount:  len(lineSignatures),
		Multiline:   multilineInfo,
		Template:    firstLineTemplate,
		VendorHint:  vVendor,
		ProductHint: vProduct,
	}

	return features, true
}
