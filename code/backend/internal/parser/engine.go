package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseLog parses a raw log string using the provided Parser configuration
// and returns a NormalizedLog.
func ParseLog(rawLog string, p *Parser) (*NormalizedLog, error) {
	if p == nil {
		return nil, fmt.Errorf("parser cannot be nil")
	}

	if p.Pattern == "" {
		return nil, fmt.Errorf("parser pattern cannot be empty")
	}

	re, err := regexp.Compile(p.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid parser pattern: %w", err)
	}

	match := re.FindStringSubmatch(rawLog)
	if match == nil {
		return nil, fmt.Errorf("log does not match parser pattern")
	}
	if match[0] != rawLog {
		return nil, fmt.Errorf("log does not fully match parser pattern")
	}

	fields := extractFields(re, match)

	normalized := &NormalizedLog{
		Metadata: make(map[string]interface{}),
	}

	// If explicit mapping is provided, map extracted fields according to p.Mapping
	if len(p.Mapping) > 0 {
		for extractedField, targetField := range p.Mapping {
			value, ok := fields[extractedField]
			if !ok {
				return nil, fmt.Errorf("mapped field %q not found in extracted fields", extractedField)
			}

			// Apply transformation on the extracted field if defined
			if p.Transformations != nil {
				if transformation, exists := p.Transformations[extractedField]; exists {
					var err error
					value, err = applyTransformation(value, transformation)
					if err != nil {
						return nil, fmt.Errorf("transformation failed for %q: %w", extractedField, err)
					}
				} else if transformation, exists := p.Transformations[targetField]; exists {
					// Fallback for compatibility if transformation key matches targetField
					var err error
					value, err = applyTransformation(value, transformation)
					if err != nil {
						return nil, fmt.Errorf("transformation failed for %q: %w", targetField, err)
					}
				}
			}

			setNormalizedField(normalized, targetField, value)
		}
	} else {
		// If no explicit mapping is provided, map extracted fields directly
		for extractedField, value := range fields {
			if p.Transformations != nil {
				if transformation, exists := p.Transformations[extractedField]; exists {
					var err error
					value, err = applyTransformation(value, transformation)
					if err != nil {
						return nil, fmt.Errorf("transformation failed for %q: %w", extractedField, err)
					}
				}
			}

			setNormalizedField(normalized, extractedField, value)
		}
	}

	return normalized, nil
}

// Parse is a convenience method on Parser to parse a raw log.
func (p *Parser) Parse(rawLog string) (*NormalizedLog, error) {
	return ParseLog(rawLog, p)
}

// extractFields maps named regex capture groups to their matched string values.
func extractFields(re *regexp.Regexp, match []string) map[string]string {
	fields := make(map[string]string)
	names := re.SubexpNames()

	for i, name := range names {
		if i == 0 || name == "" {
			continue
		}

		if i < len(match) {
			fields[name] = match[i]
		}
	}

	return fields
}

// setNormalizedField sets the corresponding field on NormalizedLog or stores it in Metadata.
func setNormalizedField(log *NormalizedLog, field string, value string) {
	if log.Metadata == nil {
		log.Metadata = make(map[string]interface{})
	}

	switch field {
	case "timestamp":
		log.Timestamp = value

	case "severity":
		log.Severity = value

	case "service":
		log.Service = value

	case "host":
		log.Host = value

	case "source":
		log.Source = value

	case "message":
		log.Message = value

	case "event_type":
		log.EventType = value

	case "trace_id":
		log.TraceID = value

	case "request_id":
		log.RequestID = value

	default:
		if strings.HasPrefix(field, "metadata.") {
			key := strings.TrimPrefix(field, "metadata.")
			log.Metadata[key] = value
		} else {
			log.Metadata[field] = value
		}
	}
}

// applyTransformation applies the specified transformation on the extracted value.
func applyTransformation(value string, t Transformation) (string, error) {
	switch strings.ToLower(t.Type) {
	case "uppercase", "upper":
		return strings.ToUpper(value), nil

	case "lowercase", "lower":
		return strings.ToLower(value), nil

	case "trim", "trim_space":
		return strings.TrimSpace(value), nil

	case "integer", "int":
		number, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid integer value %q: %w", value, err)
		}
		return strconv.FormatInt(number, 10), nil

	case "float", "decimal", "number":
		number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return "", fmt.Errorf("invalid float value %q: %w", value, err)
		}
		return strconv.FormatFloat(number, 'f', -1, 64), nil

	case "datetime", "date", "timestamp", "time":
		return parseAndFormatDateTime(value, t.Format)

	case "epoch_s", "epoch_sec", "unix":
		sec, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid unix timestamp seconds %q: %w", value, err)
		}
		return time.Unix(sec, 0).UTC().Format(time.RFC3339), nil

	case "epoch_ms", "epoch_milli":
		ms, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid unix timestamp milliseconds %q: %w", value, err)
		}
		return time.UnixMilli(ms).UTC().Format(time.RFC3339), nil

	default:
		return "", fmt.Errorf("unknown transformation type: %s", t.Type)
	}
}

// parseAndFormatDateTime parses a date/time string and formats it to RFC3339 UTC.
func parseAndFormatDateTime(value string, format string) (string, error) {
	trimmed := strings.TrimSpace(value)

	// If a format layout is explicitly provided
	if format != "" {
		// First try standard Go layout format directly
		parsed, err := time.Parse(format, trimmed)
		if err == nil {
			return parsed.UTC().Format(time.RFC3339), nil
		}

		// Try converting common strftime or ISO format tokens (e.g. YYYY-MM-DD HH:mm:ss) to Go layout
		convertedLayout := convertToGoDateFormat(format)
		if convertedLayout != format {
			parsed, err = time.Parse(convertedLayout, trimmed)
			if err == nil {
				return parsed.UTC().Format(time.RFC3339), nil
			}
		}

		return "", fmt.Errorf("failed to parse datetime %q with format %q: %w", value, format, err)
	}

	// If no format is provided, try common date/time formats
	commonFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05.000000",
		"2006-01-02 15:04:05,000",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000",
		"02/Jan/2006:15:04:05 -0700",
		"02/Jan/2006:15:04:05",
		"Jan 02 15:04:05",
		"Jan  2 15:04:05",
		"2006/01/02 15:04:05",
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.DateTime,
		time.DateOnly,
	}

	for _, layout := range commonFormats {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed.UTC().Format(time.RFC3339), nil
		}
	}

	return "", fmt.Errorf("unable to auto-detect datetime format for %q", value)
}

// convertToGoDateFormat maps popular date format specifiers (e.g. YYYY-MM-DD HH:mm:ss) to Go's reference format.
func convertToGoDateFormat(format string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"yyyy", "2006",
		"YY", "06",
		"yy", "06",
		"MM", "01",
		"DD", "02",
		"dd", "02",
		"HH", "15",
		"hh", "03",
		"mm", "04",
		"ss", "05",
		"SS", "05",
		"SSS", "000",
		"%Y", "2006",
		"%y", "06",
		"%m", "01",
		"%d", "02",
		"%H", "15",
		"%I", "03",
		"%M", "04",
		"%S", "05",
		"%f", "000000",
		"%z", "-0700",
		"%Z", "MST",
	)
	return replacer.Replace(format)
}
