package fingerprint

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	// Precompiled regular expressions for high-performance normalization.
	uuidRegex = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\b`)
	macRegex  = regexp.MustCompile(`(?i)\b(?:[0-9a-f]{2}[:-]){5}[0-9a-f]{2}\b|\b[0-9a-f]{4}\.[0-9a-f]{4}\.[0-9a-f]{4}\b`)

	// Hashes: sha256 (64 hex), sha1 (40 hex), md5 (32 hex).
	sha256Regex = regexp.MustCompile(`\b[0-9a-fA-F]{64}\b`)
	sha1Regex   = regexp.MustCompile(`\b[0-9a-fA-F]{40}\b`)
	md5Regex    = regexp.MustCompile(`\b[0-9a-fA-F]{32}\b`)

	// Hex memory addresses / ObjectIDs / hex values (0x... or 24-char mongo id).
	hexAddrRegex = regexp.MustCompile(`\b0x[0-9a-fA-F]+\b`)
	mongoIDRegex = regexp.MustCompile(`\b[0-9a-fA-F]{24}\b`)

	// URLs & Emails
	urlRegex   = regexp.MustCompile(`(?i)\b(?:https?|ftp|file)://[^\s"'<>]+`)
	emailRegex = regexp.MustCompile(`(?i)\b[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}\b`)

	// ISO8601 / RFC3339 / RFC2822 / Apache combined timestamp.
	iso8601Regex  = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?\b`)
	apacheTsRegex = regexp.MustCompile(`\b\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}(?: [+-]\d{4})?\b`)
	syslogTsRegex = regexp.MustCompile(`\b(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\b`)
	dateOnlyRegex = regexp.MustCompile(`\b\d{4}[-/]\d{2}[-/]\d{2}\b`)
	timeOnlyRegex = regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}(?:\.\d+)?\b`)

	// IPv4 with optional port: e.g. 192.168.1.1:8080 or 10.0.0.1/52341
	ipv4WithPortRegex = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})[/:](\d{1,5})\b`)
	ipv4Regex         = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(?:/\d{1,2})?\b`)

	// IPv6 regex
	ipv6Regex = regexp.MustCompile(`(?i)\b(?:[0-9a-f]{1,4}:){7}[0-9a-f]{1,4}\b|\b(?:[0-9a-f]{1,4}:){1,7}:|\b:(?::[0-9a-f]{1,4}){1,7}\b|\b(?:[0-9a-f]{1,4}:){1,6}:[0-9a-f]{1,4}\b|\b(?:[0-9a-f]{1,4}:){1,5}(?::[0-9a-f]{1,4}){1,2}\b|\b(?:[0-9a-f]{1,4}:){1,4}(?::[0-9a-f]{1,4}){1,3}\b|\b(?:[0-9a-f]{1,4}:){1,3}(?::[0-9a-f]{1,4}){1,4}\b|\b(?:[0-9a-f]{1,4}:){1,2}(?::[0-9a-f]{1,4}){1,5}\b|\b[0-9a-f]{1,4}:(?:(?::[0-9a-f]{1,4}){1,6})\b`)

	// Dynamic hybrid identifiers with digits and letters like "session-abc123-98231" or "req_fa98b7c6"
	dynamicTokenRegex = regexp.MustCompile(`(?i)\b[a-z]+[-_][0-9a-z]*\d[0-9a-z]*(?:[-_][0-9a-z]+)*\b|\b[0-9a-z]*\d[0-9a-z]*[-_][a-z]+[-_][0-9a-z]+\b`)

	// Standalone numbers / counters / sequence numbers / epoch timestamps (10 to 13 digits).
	epochMsRegex  = regexp.MustCompile(`\b1[4-9]\d{11,12}\b`)       // 13-digit epoch ms
	epochSecRegex = regexp.MustCompile(`\b1[4-9]\d{8}(?:\.\d+)?\b`) // 10-digit epoch sec

	// Volatile entities: users, hosts, ports, ssh keys, web log user and user-agent in message context
	userAgentRegex   = regexp.MustCompile(`(?i)"(?:Mozilla/\d+\.\d+|curl/\S+|Postman\S+|Wget/\S+|python-requests/\S+|Go-http-client/\S+)[^"]*"`)
	sshKeyRegex      = regexp.MustCompile(`(?i)\b(SHA256|MD5|RSA):[A-Za-z0-9+/=_-]+`)
	webUserRegex     = regexp.MustCompile(`\s-\s+[^\s\[]+\s+\[`)
	userContextRegex = regexp.MustCompile(`(?i)\b(for\s+invalid\s+user|for\s+user|user(?:name)?)\s+([a-zA-Z0-9_.-]+)`)
	portContextRegex = regexp.MustCompile(`(?i)\b(port)\s+(\d+)`)
	hostContextRegex = regexp.MustCompile(`(?i)\b(from\s+host|host(?:name)?)\s+([a-zA-Z0-9_.-]+)`)
)

// DetectValueType evaluates a string value and infers its semantic ValueType.
func DetectValueType(val string) ValueType {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return TypeString
	}

	// Boolean
	lower := strings.ToLower(trimmed)
	if lower == "true" || lower == "false" || lower == "yes" || lower == "no" {
		return TypeBool
	}
	if lower == "null" || lower == "nil" || lower == "none" {
		return TypeNull
	}

	// UUID
	if uuidRegex.MatchString(trimmed) && len(trimmed) == 36 {
		return TypeUUID
	}

	// MAC
	if macRegex.MatchString(trimmed) && (len(trimmed) == 17 || len(trimmed) == 14) {
		return TypeMAC
	}

	// IP addresses
	if ip := net.ParseIP(trimmed); ip != nil {
		if strings.Contains(trimmed, ":") {
			return TypeIPv6
		}
		return TypeIPv4
	}

	// Hashes
	if (len(trimmed) == 64 || len(trimmed) == 40 || len(trimmed) == 32) && isPureHex(trimmed) {
		return TypeHexHash
	}

	// URL
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "ftp://") {
		return TypeURL
	}

	// Email
	if emailRegex.MatchString(trimmed) {
		return TypeEmail
	}

	// Timestamps
	if iso8601Regex.MatchString(trimmed) || apacheTsRegex.MatchString(trimmed) || syslogTsRegex.MatchString(trimmed) {
		return TypeTimestamp
	}
	if timeOnlyRegex.MatchString(trimmed) && len(trimmed) <= 12 {
		return TypeTime
	}

	// Epoch timestamps
	if epochMsRegex.MatchString(trimmed) || epochSecRegex.MatchString(trimmed) {
		return TypeTimestamp
	}

	// Number / Port
	if num, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		if num >= 1 && num <= 65535 && isPortContext(trimmed) {
			return TypePort
		}
		return TypeNumber
	}
	if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return TypeNumber
	}

	return TypeString
}

func isPortContext(s string) bool {
	// Utility for specific port ranges if needed
	return false
}

func isPureHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}

// NormalizeValue takes a string value and normalizes volatile content while preserving enum keywords.
func NormalizeValue(key, val string) string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return ""
	}

	// Strip surrounding quotes if present
	if (strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`)) ||
		(strings.HasPrefix(trimmed, `'`) && strings.HasSuffix(trimmed, `'`)) {
		if len(trimmed) >= 2 {
			trimmed = trimmed[1 : len(trimmed)-1]
		}
	}

	// Check key context: preserved schema enum fields
	kLower := strings.ToLower(key)
	if isEnumKey(kLower) {
		// Preserve small static keywords/status codes
		if len(trimmed) <= 15 && isIdentifierOrSmallNumber(trimmed) {
			return strings.ToLower(trimmed)
		}
	}

	// Infer type
	t := DetectValueType(trimmed)
	switch t {
	case TypeIPv4:
		return "<IPV4>"
	case TypeIPv6:
		return "<IPV6>"
	case TypeMAC:
		return "<MAC>"
	case TypeUUID:
		return "<UUID>"
	case TypeTimestamp:
		return "<TIMESTAMP>"
	case TypeTime:
		return "<TIME>"
	case TypeHexHash:
		return "<HASH>"
	case TypeURL:
		return "<URL>"
	case TypeEmail:
		return "<EMAIL>"
	case TypeNumber:
		if isSmallEnumNumber(trimmed, kLower) {
			return trimmed
		}
		return "<NUMBER>"
	case TypeBool:
		return strings.ToLower(trimmed)
	case TypeNull:
		return "<NULL>"
	}

	// Check for dynamic token
	if dynamicTokenRegex.MatchString(trimmed) {
		return "<DYNAMIC_ID>"
	}

	// Text normalization for compound string values
	return NormalizeTemplate(trimmed)
}

// isEnumKey returns true if the key typically holds schema-defining enums rather than volatile values.
func isEnumKey(k string) bool {
	switch k {
	case "action", "act", "status", "level", "severity", "sev", "pri", "priority",
		"proto", "protocol", "method", "http_method", "type", "event_type",
		"version", "v", "ver", "op", "operation", "result", "disposition",
		"category", "cat", "facility", "fac", "state", "direction", "dir":
		return true
	}
	return false
}

// isSmallEnumNumber checks if a number should be preserved verbatim (e.g. HTTP status 200, Syslog severity 6).
func isSmallEnumNumber(numStr, key string) bool {
	if key == "" {
		return false
	}
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return false
	}

	switch key {
	case "version", "v", "ver":
		return n >= 0 && n <= 10
	case "severity", "sev", "pri", "facility", "fac", "level":
		return n >= 0 && n <= 24
	case "status", "http_status", "code", "response_code", "status_code":
		return (n >= 100 && n <= 599) || (n >= 0 && n <= 10)
	case "proto", "protocol":
		// common IP protocol numbers: 1 (ICMP), 6 (TCP), 17 (UDP), 47 (GRE), 50 (ESP), 58 (ICMPv6)
		return n == 1 || n == 6 || n == 17 || n == 47 || n == 50 || n == 58
	}
	return false
}

func isIdentifierOrSmallNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

// NormalizeTemplate performs broad, robust normalization of volatile tokens in arbitrary log strings.
func NormalizeTemplate(text string) string {
	if text == "" {
		return ""
	}

	// Bound max length to avoid ReDoS or extreme overhead on corrupted 100KB+ blobs
	maxLen := 4096
	truncated := false
	res := text
	if len(res) > maxLen {
		res = res[:maxLen]
		truncated = true
	}

	// Web access log user and user-agent normalization
	res = webUserRegex.ReplaceAllString(res, " - <USER> [")
	res = userAgentRegex.ReplaceAllString(res, "\"<USER_AGENT>\"")

	// SSH key fingerprints
	res = sshKeyRegex.ReplaceAllString(res, "$1:<HASH>")

	// Contextual User and Host normalization
	res = userContextRegex.ReplaceAllString(res, "$1 <USER>")
	res = hostContextRegex.ReplaceAllString(res, "$1 <HOST>")
	res = portContextRegex.ReplaceAllString(res, "port <PORT>")

	// 1. URLs
	res = urlRegex.ReplaceAllString(res, "<URL>")

	// 2. Email addresses
	res = emailRegex.ReplaceAllString(res, "<EMAIL>")

	// 3. Timestamps & Dates
	res = iso8601Regex.ReplaceAllString(res, "<TIMESTAMP>")
	res = apacheTsRegex.ReplaceAllString(res, "<TIMESTAMP>")
	res = syslogTsRegex.ReplaceAllString(res, "<TIMESTAMP>")
	res = dateOnlyRegex.ReplaceAllString(res, "<DATE>")

	// 4. UUIDs
	res = uuidRegex.ReplaceAllString(res, "<UUID>")

	// 5. MAC addresses
	res = macRegex.ReplaceAllString(res, "<MAC>")

	// 6. Hashes & Hex Addresses
	res = sha256Regex.ReplaceAllString(res, "<HASH>")
	res = sha1Regex.ReplaceAllString(res, "<HASH>")
	res = md5Regex.ReplaceAllString(res, "<HASH>")
	res = hexAddrRegex.ReplaceAllString(res, "<HEX_ADDR>")
	res = mongoIDRegex.ReplaceAllString(res, "<HEX_ID>")

	// 7. IPv4 with Port (e.g. 10.0.0.5:52341 or 10.0.0.1/52341 -> <IPV4>:<PORT>)
	res = ipv4WithPortRegex.ReplaceAllString(res, "<IPV4>:<PORT>")

	// 8. IPv4 and IPv6 standalone
	res = ipv4Regex.ReplaceAllString(res, "<IPV4>")
	res = ipv6Regex.ReplaceAllString(res, "<IPV6>")

	// 9. Time standalone
	res = timeOnlyRegex.ReplaceAllString(res, "<TIME>")

	// 10. Epoch timestamps
	res = epochMsRegex.ReplaceAllString(res, "<TIMESTAMP>")
	res = epochSecRegex.ReplaceAllString(res, "<TIMESTAMP>")

	// 11. Dynamic identifiers (e.g. abc123-session-98231 -> <DYNAMIC_ID>)
	res = dynamicTokenRegex.ReplaceAllString(res, "<DYNAMIC_ID>")

	// 12. Remaining numbers (context-aware token replacement)
	res = normalizeRemainingNumbers(res)

	// 13. Collapse redundant whitespace
	res = collapseWhitespace(res)

	if truncated {
		res += " <TRUNCATED>"
	}

	return res
}

func normalizeRemainingNumbers(s string) string {
	// Walk through tokens and replace numbers that are not schema keywords
	var sb strings.Builder
	sb.Grow(len(s))

	tokens := strings.Fields(s)
	for i, tok := range tokens {
		if i > 0 {
			sb.WriteByte(' ')
		}

		// Handle key=value token like "src_port=52341" or "count=123" or "version=2"
		if eqIdx := strings.IndexByte(tok, '='); eqIdx != -1 {
			k := tok[:eqIdx]
			v := tok[eqIdx+1:]
			kLower := strings.ToLower(k)
			if isSmallEnumNumber(v, kLower) {
				sb.WriteString(tok)
				continue
			}
			if _, err := strconv.ParseFloat(v, 64); err == nil {
				sb.WriteString(k)
				sb.WriteString("=<NUMBER>")
				continue
			}
			sb.WriteString(tok)
			continue
		}

		// Handle standalone numbers
		if _, err := strconv.ParseFloat(tok, 64); err == nil {
			sb.WriteString("<NUMBER>")
			continue
		}

		// Check if token has trailing punctuation like "accepted." or "port: 443"
		trimmed := strings.TrimRight(tok, ",.;:)]}")
		punct := tok[len(trimmed):]
		if _, err := strconv.ParseFloat(trimmed, 64); err == nil && len(trimmed) > 0 {
			sb.WriteString("<NUMBER>")
			sb.WriteString(punct)
			continue
		}

		sb.WriteString(tok)
	}

	return sb.String()
}

func collapseWhitespace(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	inSpace := false

	for _, r := range s {
		if unicode.IsSpace(r) {
			if !inSpace {
				sb.WriteByte(' ')
				inSpace = true
			}
		} else {
			sb.WriteRune(r)
			inSpace = false
		}
	}

	return strings.TrimSpace(sb.String())
}

// ExtractTimestampFromLayout attempts to parse various time layouts.
func ExtractTimestampFromLayout(val string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z0700",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.000",
		"02/Jan/2006:15:04:05 -0700",
		"Jan _2 15:04:05",
		"Jan 02 15:04:05",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, val); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
