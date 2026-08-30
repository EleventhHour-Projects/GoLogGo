package fingerprint

// CurrentSchemaVersion defines the version of the fingerprint extraction and hashing algorithm.
const CurrentSchemaVersion = "fp-v1"

// FormatType represents the high-level syntax/format family of a log entry.
type FormatType string

const (
	FormatUnknown   FormatType = "unknown"
	FormatJSON      FormatType = "json"
	FormatXML       FormatType = "xml"
	FormatCSV       FormatType = "csv"
	FormatKeyValue  FormatType = "key_value"
	FormatSyslog    FormatType = "syslog"
	FormatCEF       FormatType = "cef"
	FormatLEEF      FormatType = "leef"
	FormatPlainText FormatType = "plain_text"
	FormatMultiline FormatType = "multiline"
)

// ValueType represents the semantic or primitive data type inferred for a field or token.
type ValueType string

const (
	TypeUnknown   ValueType = "unknown"
	TypeString    ValueType = "string"
	TypeNumber    ValueType = "number"
	TypeBool      ValueType = "bool"
	TypeNull      ValueType = "null"
	TypeObject    ValueType = "object"
	TypeArray     ValueType = "array"
	TypeIPv4      ValueType = "ipv4"
	TypeIPv6      ValueType = "ipv6"
	TypeTimestamp ValueType = "timestamp"
	TypeTime      ValueType = "time"
	TypeUUID      ValueType = "uuid"
	TypeMAC       ValueType = "mac"
	TypeURL       ValueType = "url"
	TypeEmail     ValueType = "email"
	TypeHexHash   ValueType = "hex_hash"
	TypePort      ValueType = "port"
)

// FieldSignature represents a key-type pair in structured logs.
type FieldSignature struct {
	Key  string    `json:"key"`
	Type ValueType `json:"type"`
}

// JSONSchemaNode represents a recursive schema tree for JSON structures.
type JSONSchemaNode struct {
	Type       ValueType                  `json:"type"`
	Properties map[string]*JSONSchemaNode `json:"properties,omitempty"` // sorted during canonicalization
	ItemType   *JSONSchemaNode            `json:"item_type,omitempty"`  // for arrays
}

// SyslogFeature encapsulates parsed Syslog structural characteristics.
type SyslogFeature struct {
	Version        int      `json:"version,omitempty"` // 0 for RFC3164, 1+ for RFC5424
	Facility       int      `json:"facility"`
	Severity       int      `json:"severity"`
	AppName        string   `json:"app_name,omitempty"`
	MsgID          string   `json:"msg_id,omitempty"`
	StructuredKeys []string `json:"structured_keys,omitempty"` // sorted [id.key]
	HasTimestamp   bool     `json:"has_timestamp"`
	HasHostname    bool     `json:"has_hostname"`
	HasPID         bool     `json:"has_pid"`
}

// CEFFeature encapsulates parsed Common Event Format (CEF) metadata and extension schema.
type CEFFeature struct {
	Version            int              `json:"version"`
	DeviceVendor       string           `json:"device_vendor"`
	DeviceProduct      string           `json:"device_product"`
	DeviceVersion      string           `json:"device_version,omitempty"`
	DeviceEventClassID string           `json:"device_event_class_id"`
	Severity           string           `json:"severity"`
	ExtensionFields    []FieldSignature `json:"extension_fields,omitempty"` // sorted by key
}

// LEEFFeature encapsulates parsed Log Event Extended Format (LEEF) metadata and extension schema.
type LEEFFeature struct {
	Version         string           `json:"version"`
	Vendor          string           `json:"vendor"`
	Product         string           `json:"product"`
	ProductVersion  string           `json:"product_version,omitempty"`
	EventID         string           `json:"event_id"`
	Delimiter       string           `json:"delimiter,omitempty"`
	ExtensionFields []FieldSignature `json:"extension_fields,omitempty"` // sorted by key
}

// CSVFeature encapsulates delimiter, column structure, and inferred column types for tabular logs.
type CSVFeature struct {
	Delimiter   string      `json:"delimiter"`
	ColumnCount int         `json:"column_count"`
	HasHeader   bool        `json:"has_header"`
	HeaderNames []string    `json:"header_names,omitempty"`
	ColumnTypes []ValueType `json:"column_types"`
	QuoteChar   string      `json:"quote_char,omitempty"`
}

// KVFeature encapsulates Key-Value (Logfmt / Splunk / Fortinet) structure.
type KVFeature struct {
	Separator     string           `json:"separator"` // e.g. "="
	Delimiter     string           `json:"delimiter"` // e.g. " " or ","
	Fields        []FieldSignature `json:"fields"`    // sorted by key
	QuotingStyle  string           `json:"quoting_style,omitempty"`
	HasDuplicates bool             `json:"has_duplicates"`
}

// MultilineFeature encapsulates characteristics of multi-line log events (e.g. stack traces).
type MultilineFeature struct {
	LineCount      int      `json:"line_count"`
	IsStackTrace   bool     `json:"is_stack_trace"`
	PrefixType     string   `json:"prefix_type,omitempty"` // e.g. "timestamp", "indent", "xml_tag"
	LineSignatures []string `json:"line_signatures"`       // normalized template for key lines
}

// FingerprintFeatures is the primary data structure capturing the deterministic structural features of a log.
type FingerprintFeatures struct {
	// Version is the fingerprint algorithm schema version (e.g. "fp-v1").
	Version string `json:"version"`

	// Format identifies the structural format of the log.
	Format FormatType `json:"format"`

	// Delimiter used in the log format (if applicable).
	Delimiter string `json:"delimiter,omitempty"`

	// FieldCount is the total number of distinct top-level fields extracted.
	FieldCount int `json:"field_count"`

	// Keys contains all top-level field names, always strictly sorted in ascending lexicographical order.
	Keys []string `json:"keys,omitempty"`

	// FieldSignatures contains sorted (key, semantic_type) pairs for structured formats.
	FieldSignatures []FieldSignature `json:"field_signatures,omitempty"`

	// JSONSchema holds the recursive schema AST for JSON logs.
	JSONSchema *JSONSchemaNode `json:"json_schema,omitempty"`

	// XMLHierarchy represents the normalized element hierarchy and attributes for XML logs.
	XMLHierarchy string `json:"xml_hierarchy,omitempty"`

	// Syslog holds Syslog-specific structural features.
	Syslog *SyslogFeature `json:"syslog,omitempty"`

	// CEF holds Common Event Format features.
	CEF *CEFFeature `json:"cef,omitempty"`

	// LEEF holds Log Event Extended Format features.
	LEEF *LEEFFeature `json:"leef,omitempty"`

	// CSV holds CSV/TSV structural features.
	CSV *CSVFeature `json:"csv,omitempty"`

	// KV holds Key-Value format features.
	KV *KVFeature `json:"kv,omitempty"`

	// Multiline holds multi-line specific characteristics.
	Multiline *MultilineFeature `json:"multiline,omitempty"`

	// VendorHint is a strong vendor hint (e.g. "cisco", "palo_alto", "fortinet", "aws") if detectable.
	VendorHint string `json:"vendor_hint,omitempty"`

	// ProductHint is a product hint (e.g. "asa", "pan_os", "fortigate", "vpc_flow") if detectable.
	ProductHint string `json:"product_hint,omitempty"`

	// Template is the normalized text template with volatile values abstracted (e.g. "<IPV4>", "<PORT>", "<TIMESTAMP>").
	Template string `json:"template,omitempty"`

	// StructuralTokens represents the ordered sequence of syntax/structural tokens.
	StructuralTokens []string `json:"structural_tokens,omitempty"`
}
