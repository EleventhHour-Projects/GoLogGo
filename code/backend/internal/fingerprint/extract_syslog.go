package fingerprint

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	// RFC5424 header: <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID
	rfc5424Regex = regexp.MustCompile(`^<(\d{1,3})>(\d+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)`)
	// RFC3164 header: <PRI>TIMESTAMP HOSTNAME ...
	rfc3164Regex = regexp.MustCompile(`^<(\d{1,3})>([A-Za-z]{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+([^:\s\[]+)(?:\[(\d+)\])?:\s*(.*)`)
	// General PRI prefix
	priPrefixRegex = regexp.MustCompile(`^<(\d{1,3})>`)
	// Structured data element: [id key="val" ...]
	sdElementRegex = regexp.MustCompile(`\[([^\s\]]+)(?:\s+([^\]]+))?\]`)
)

// ExtractSyslogFeatures parses RFC 3164 and RFC 5424 Syslog logs.
func ExtractSyslogFeatures(raw string) (FingerprintFeatures, bool) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "<") {
		// Could still be non-PRI syslog format, e.g. "Aug 30 13:02:11 hostname sshd[1234]: Accepted publickey..."
		if isBSDSyslogWithoutPRI(trimmed) {
			return parseNonPRISyslog(trimmed)
		}
		return FingerprintFeatures{}, false
	}

	priMatches := priPrefixRegex.FindStringSubmatch(trimmed)
	if len(priMatches) < 2 {
		return FingerprintFeatures{}, false
	}

	priVal, err := strconv.Atoi(priMatches[1])
	if err != nil || priVal < 0 || priVal > 191 {
		return FingerprintFeatures{}, false
	}

	facility := priVal / 8
	severity := priVal % 8

	// Try RFC 5424
	if matches := rfc5424Regex.FindStringSubmatch(trimmed); len(matches) >= 8 {
		version, _ := strconv.Atoi(matches[2])
		hostname := matches[4]
		appName := matches[5]
		procID := matches[6]
		msgID := matches[7]

		rest := strings.TrimSpace(trimmed[len(matches[0]):])
		sdKeys, body := parseStructuredData(rest)

		normTemplate := NormalizeTemplate(body)
		vVendor, vProduct := DetectVendorProduct(raw)

		syslogInfo := &SyslogFeature{
			Version:        version,
			Facility:       facility,
			Severity:       severity,
			AppName:        appName,
			MsgID:          msgID,
			StructuredKeys: sdKeys,
			HasTimestamp:   matches[3] != "-",
			HasHostname:    hostname != "-",
			HasPID:         procID != "-",
		}

		features := FingerprintFeatures{
			Version:          CurrentSchemaVersion,
			Format:           FormatSyslog,
			FieldCount:       len(sdKeys) + 5,
			Keys:             sdKeys,
			Syslog:           syslogInfo,
			VendorHint:       vVendor,
			ProductHint:      vProduct,
			Template:         normTemplate,
			StructuralTokens: []string{"syslog", "rfc5424", appName, msgID},
		}

		return features, true
	}

	// Try RFC 3164
	if matches := rfc3164Regex.FindStringSubmatch(trimmed); len(matches) >= 7 {
		tag := matches[4]
		hasPID := matches[5] != ""
		msgBody := matches[6]
		normTemplate := NormalizeTemplate(msgBody)
		vVendor, vProduct := DetectVendorProduct(raw)

		syslogInfo := &SyslogFeature{
			Version:      0, // RFC3164
			Facility:     facility,
			Severity:     severity,
			AppName:      tag,
			HasTimestamp: true,
			HasHostname:  true,
			HasPID:       hasPID,
		}

		features := FingerprintFeatures{
			Version:          CurrentSchemaVersion,
			Format:           FormatSyslog,
			FieldCount:       4,
			Syslog:           syslogInfo,
			VendorHint:       vVendor,
			ProductHint:      vProduct,
			Template:         normTemplate,
			StructuralTokens: []string{"syslog", "rfc3164", tag},
		}

		return features, true
	}

	// Generic PRI-prefixed log
	afterPRI := trimmed[len(priMatches[0]):]
	normTemplate := NormalizeTemplate(afterPRI)
	vVendor, vProduct := DetectVendorProduct(raw)

	syslogInfo := &SyslogFeature{
		Version:  0,
		Facility: facility,
		Severity: severity,
	}

	features := FingerprintFeatures{
		Version:     CurrentSchemaVersion,
		Format:      FormatSyslog,
		Syslog:      syslogInfo,
		VendorHint:  vVendor,
		ProductHint: vProduct,
		Template:    normTemplate,
	}

	return features, true
}

func isBSDSyslogWithoutPRI(s string) bool {
	// Checks if begins with "MMM d HH:mm:ss " or "MMM  d HH:mm:ss "
	if len(s) < 16 {
		return false
	}
	return syslogTsRegex.MatchString(s[:16]) || (len(s) >= 15 && syslogTsRegex.MatchString(s[:15]))
}

var bsdSyslogNoPRIRegex = regexp.MustCompile(`^([A-Za-z]{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+([^:\s\[]+)(?:\[(\d+)\])?:\s*(.*)`)

func parseNonPRISyslog(raw string) (FingerprintFeatures, bool) {
	matches := bsdSyslogNoPRIRegex.FindStringSubmatch(raw)
	if len(matches) < 6 {
		return FingerprintFeatures{}, false
	}

	tag := matches[3]
	hasPID := matches[4] != ""
	body := matches[5]
	normTemplate := NormalizeTemplate(body)
	vVendor, vProduct := DetectVendorProduct(raw)

	syslogInfo := &SyslogFeature{
		Version:      0,
		AppName:      tag,
		HasTimestamp: true,
		HasHostname:  true,
		HasPID:       hasPID,
	}

	features := FingerprintFeatures{
		Version:          CurrentSchemaVersion,
		Format:           FormatSyslog,
		FieldCount:       3,
		Syslog:           syslogInfo,
		VendorHint:       vVendor,
		ProductHint:      vProduct,
		Template:         normTemplate,
		StructuralTokens: []string{"syslog", "bsd", tag},
	}

	return features, true
}

func parseStructuredData(rest string) ([]string, string) {
	var keys []string
	idx := 0

	for idx < len(rest) && rest[idx] == '[' {
		endBracket := strings.IndexByte(rest[idx:], ']')
		if endBracket == -1 {
			break
		}
		sdContent := rest[idx+1 : idx+endBracket]
		idx += endBracket + 1

		// Parse SD-ID and SD-PARAMS
		parts := strings.Fields(sdContent)
		if len(parts) > 0 {
			sdID := parts[0]
			for _, param := range parts[1:] {
				if eqIdx := strings.IndexByte(param, '='); eqIdx != -1 {
					paramName := param[:eqIdx]
					keys = append(keys, sdID+"."+paramName)
				}
			}
		}

		// Skip trailing space after bracket
		for idx < len(rest) && rest[idx] == ' ' {
			idx++
		}
	}

	sort.Strings(keys)
	body := ""
	if idx < len(rest) {
		body = strings.TrimSpace(rest[idx:])
	}

	return keys, body
}
