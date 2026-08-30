package fingerprint

import (
	"bytes"
	"encoding/xml"
	"io"
	"sort"
	"strings"
)

// ExtractXMLFeatures parses XML documents and extracts structural element tag hierarchies and attributes.
func ExtractXMLFeatures(raw string) (FingerprintFeatures, bool) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "<") {
		return FingerprintFeatures{}, false
	}

	decoder := xml.NewDecoder(strings.NewReader(trimmed))
	var tagStack []string
	var paths []string
	tagCounts := make(map[string]int)
	attributeMap := make(map[string][]string)

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			// If error happened but we already extracted a valid XML structure, continue
			break
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			tagName := elem.Name.Local
			tagStack = append(tagStack, tagName)
			fullPath := strings.Join(tagStack, "/")
			paths = append(paths, fullPath)
			tagCounts[tagName]++

			if len(elem.Attr) > 0 {
				var attrNames []string
				for _, attr := range elem.Attr {
					attrNames = append(attrNames, attr.Name.Local)
				}
				sort.Strings(attrNames)
				attributeMap[fullPath] = attrNames
			}

		case xml.EndElement:
			if len(tagStack) > 0 {
				tagStack = tagStack[:len(tagStack)-1]
			}
		}
	}

	if len(paths) == 0 {
		return FingerprintFeatures{}, false
	}

	// Deduplicate and canonicalize XML hierarchy representation
	var uniquePaths []string
	seenPaths := make(map[string]bool)
	for _, p := range paths {
		if !seenPaths[p] {
			seenPaths[p] = true
			uniquePaths = append(uniquePaths, p)
		}
	}
	sort.Strings(uniquePaths)

	var hierarchyBuf bytes.Buffer
	for _, p := range uniquePaths {
		hierarchyBuf.WriteString(p)
		if attrs, ok := attributeMap[p]; ok && len(attrs) > 0 {
			hierarchyBuf.WriteString("[")
			hierarchyBuf.WriteString(strings.Join(attrs, ","))
			hierarchyBuf.WriteString("]")
		}
		hierarchyBuf.WriteByte(';')
	}

	vVendor, vProduct := DetectVendorProduct(raw)

	features := FingerprintFeatures{
		Version:      CurrentSchemaVersion,
		Format:       FormatXML,
		FieldCount:   len(uniquePaths),
		Keys:         uniquePaths,
		XMLHierarchy: hierarchyBuf.String(),
		VendorHint:   vVendor,
		ProductHint:  vProduct,
	}

	return features, true
}
