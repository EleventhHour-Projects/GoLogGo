package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// GenerateFingerprintHash converts a FingerprintFeatures structure into a deterministic SHA-256 hex string.
//
// This is Function 2 of the two-stage fingerprinting pipeline.
func GenerateFingerprintHash(features FingerprintFeatures) string {
	canonical := CanonicalRepresentation(features)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

// CanonicalRepresentation builds a deterministic, strictly-ordered canonical string representation of FingerprintFeatures.
func CanonicalRepresentation(feat FingerprintFeatures) string {
	version := feat.Version
	if version == "" {
		version = CurrentSchemaVersion
	}

	var sb strings.Builder
	sb.WriteString(version)
	sb.WriteString(":format=")
	sb.WriteString(string(feat.Format))

	if feat.VendorHint != "" {
		sb.WriteString("|vendor=")
		sb.WriteString(feat.VendorHint)
	}
	if feat.ProductHint != "" {
		sb.WriteString("|product=")
		sb.WriteString(feat.ProductHint)
	}

	switch feat.Format {
	case FormatJSON:
		if feat.JSONSchema != nil {
			sb.WriteString("|schema=")
			sb.WriteString(canonicalJSONSchema(feat.JSONSchema))
		}

	case FormatKeyValue:
		if feat.KV != nil {
			sb.WriteString("|sep=")
			sb.WriteString(feat.KV.Separator)
			sb.WriteString("|delim=")
			sb.WriteString(feat.KV.Delimiter)
			sb.WriteString("|fields=")
			sb.WriteString(canonicalFieldSignatures(feat.KV.Fields))
		} else {
			sb.WriteString("|fields=")
			sb.WriteString(canonicalFieldSignatures(feat.FieldSignatures))
		}

	case FormatCEF:
		if feat.CEF != nil {
			sb.WriteString(fmt.Sprintf("|ver=%d|dev_vendor=%s|dev_product=%s|dev_ver=%s|class_id=%s|sev=%s|ext=",
				feat.CEF.Version, feat.CEF.DeviceVendor, feat.CEF.DeviceProduct, feat.CEF.DeviceVersion, feat.CEF.DeviceEventClassID, feat.CEF.Severity))
			sb.WriteString(canonicalFieldSignatures(feat.CEF.ExtensionFields))
		}

	case FormatLEEF:
		if feat.LEEF != nil {
			sb.WriteString(fmt.Sprintf("|ver=%s|vendor=%s|product=%s|prod_ver=%s|event_id=%s|delim=%s|ext=",
				feat.LEEF.Version, feat.LEEF.Vendor, feat.LEEF.Product, feat.LEEF.ProductVersion, feat.LEEF.EventID, feat.LEEF.Delimiter))
			sb.WriteString(canonicalFieldSignatures(feat.LEEF.ExtensionFields))
		}

	case FormatSyslog:
		if feat.Syslog != nil {
			sb.WriteString(fmt.Sprintf("|ver=%d|fac=%d|sev=%d|app=%s|msg_id=%s",
				feat.Syslog.Version, feat.Syslog.Facility, feat.Syslog.Severity, feat.Syslog.AppName, feat.Syslog.MsgID))
			if len(feat.Syslog.StructuredKeys) > 0 {
				sb.WriteString("|sd=")
				sb.WriteString(strings.Join(feat.Syslog.StructuredKeys, ","))
			}
		}
		if feat.Template != "" {
			sb.WriteString("|template=")
			sb.WriteString(feat.Template)
		}

	case FormatCSV:
		if feat.CSV != nil {
			sb.WriteString(fmt.Sprintf("|delim=%s|cols=%d|header=%t", feat.CSV.Delimiter, feat.CSV.ColumnCount, feat.CSV.HasHeader))
			if len(feat.CSV.HeaderNames) > 0 {
				sb.WriteString("|headers=")
				sb.WriteString(strings.Join(feat.CSV.HeaderNames, ","))
			}
			if len(feat.CSV.ColumnTypes) > 0 {
				sb.WriteString("|types=")
				for i, t := range feat.CSV.ColumnTypes {
					if i > 0 {
						sb.WriteByte(',')
					}
					sb.WriteString(string(t))
				}
			}
		}

	case FormatXML:
		if feat.XMLHierarchy != "" {
			sb.WriteString("|hierarchy=")
			sb.WriteString(feat.XMLHierarchy)
		}

	case FormatMultiline:
		if feat.Multiline != nil {
			sb.WriteString(fmt.Sprintf("|stack=%t|prefix=%s|lines=", feat.Multiline.IsStackTrace, feat.Multiline.PrefixType))
			sb.WriteString(strings.Join(feat.Multiline.LineSignatures, "||"))
		}

	case FormatPlainText, FormatUnknown:
		fallthrough
	default:
		if feat.Template != "" {
			sb.WriteString("|template=")
			sb.WriteString(feat.Template)
		}
	}

	return sb.String()
}

func canonicalJSONSchema(node *JSONSchemaNode) string {
	if node == nil {
		return string(TypeUnknown)
	}

	switch node.Type {
	case TypeObject:
		var sb strings.Builder
		sb.WriteString("object{")
		if len(node.Properties) > 0 {
			keys := make([]string, 0, len(node.Properties))
			for k := range node.Properties {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for i, k := range keys {
				if i > 0 {
					sb.WriteByte(',')
				}
				sb.WriteString(k)
				sb.WriteByte(':')
				sb.WriteString(canonicalJSONSchema(node.Properties[k]))
			}
		}
		sb.WriteString("}")
		return sb.String()

	case TypeArray:
		var sb strings.Builder
		sb.WriteString("array[")
		if node.ItemType != nil {
			sb.WriteString(canonicalJSONSchema(node.ItemType))
		} else {
			sb.WriteString(string(TypeUnknown))
		}
		sb.WriteString("]")
		return sb.String()

	default:
		return string(node.Type)
	}
}

func canonicalFieldSignatures(signatures []FieldSignature) string {
	if len(signatures) == 0 {
		return ""
	}

	// Copy and sort deterministically by Key
	sortedSigs := make([]FieldSignature, len(signatures))
	copy(sortedSigs, signatures)
	sort.Slice(sortedSigs, func(i, j int) bool {
		return sortedSigs[i].Key < sortedSigs[j].Key
	})

	var sb strings.Builder
	for i, sig := range sortedSigs {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(sig.Key)
		sb.WriteByte(':')
		sb.WriteString(string(sig.Type))
	}
	return sb.String()
}
