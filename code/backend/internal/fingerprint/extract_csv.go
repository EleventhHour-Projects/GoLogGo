package fingerprint

import (
	"encoding/csv"
	"strings"
)

// ExtractCSVFeatures analyzes delimiter-separated values (CSV, TSV, pipe-delimited).
func ExtractCSVFeatures(raw string) (FingerprintFeatures, bool) {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) == 0 {
		return FingerprintFeatures{}, false
	}

	// Web access log guard
	if strings.Contains(trimmed, "HTTP/") && strings.Contains(trimmed, " \"") {
		return FingerprintFeatures{}, false
	}

	// Delimiter candidates
	delimiters := []rune{',', '\t', ';', '|'}
	var bestDelim rune
	var bestRecords [][]string
	maxCols := 0

	for _, delim := range delimiters {
		// Semicolon or pipe on a single line with many spaces is likely text/headers, require multiple rows or strong comma/tab structure
		if delim == ';' && !strings.Contains(trimmed, "\n") {
			continue
		}

		r := csv.NewReader(strings.NewReader(trimmed))
		r.Comma = delim
		r.LazyQuotes = true
		r.FieldsPerRecord = -1 // Allow variable to detect consistency

		records, err := r.ReadAll()
		if err == nil && len(records) > 0 {
			firstRowLen := len(records[0])
			if firstRowLen >= 4 && firstRowLen > maxCols {
				// Check consistency across rows if multiline
				consistent := true
				for _, row := range records {
					if len(row) != firstRowLen {
						consistent = false
						break
					}
				}
				if consistent {
					maxCols = firstRowLen
					bestDelim = delim
					bestRecords = records
				}
			}
		}
	}

	if maxCols < 4 || len(bestRecords) == 0 {
		return FingerprintFeatures{}, false
	}

	firstRow := bestRecords[0]
	hasHeader := isHeaderRow(firstRow)
	var headerNames []string
	var colTypes []ValueType

	if hasHeader {
		headerNames = make([]string, len(firstRow))
		for i, h := range firstRow {
			headerNames[i] = strings.ToLower(strings.TrimSpace(h))
		}
		if len(bestRecords) > 1 {
			for _, colVal := range bestRecords[1] {
				colTypes = append(colTypes, DetectValueType(colVal))
			}
		} else {
			for range firstRow {
				colTypes = append(colTypes, TypeString)
			}
		}
	} else {
		for _, colVal := range firstRow {
			colTypes = append(colTypes, DetectValueType(colVal))
		}
	}

	vVendor, vProduct := DetectVendorProduct(raw)

	csvFeature := &CSVFeature{
		Delimiter:   string(bestDelim),
		ColumnCount: maxCols,
		HasHeader:   hasHeader,
		HeaderNames: headerNames,
		ColumnTypes: colTypes,
	}

	features := FingerprintFeatures{
		Version:     CurrentSchemaVersion,
		Format:      FormatCSV,
		Delimiter:   string(bestDelim),
		FieldCount:  maxCols,
		Keys:        headerNames,
		CSV:         csvFeature,
		VendorHint:  vVendor,
		ProductHint: vProduct,
	}

	return features, true
}

func isHeaderRow(row []string) bool {
	if len(row) == 0 {
		return false
	}
	allAlpha := true
	for _, col := range row {
		trimmed := strings.TrimSpace(col)
		if len(trimmed) == 0 {
			continue
		}
		// If column contains IPs, timestamps, numbers, it's not a header
		valType := DetectValueType(trimmed)
		if valType == TypeIPv4 || valType == TypeIPv6 || valType == TypeTimestamp || valType == TypeNumber {
			return false
		}
		for _, r := range trimmed {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != '_' && r != '-' && r != ' ' {
				allAlpha = false
				break
			}
		}
	}
	return allAlpha
}
