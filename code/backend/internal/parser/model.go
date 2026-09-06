package parser

type NormalizedLog struct {
	Timestamp string `json:"timestamp" bson:"timestamp"`
	Severity  string `json:"severity" bson:"severity"` // INFO, ERROR etc
	Service   string `json:"service" bson:"service"`   // nginx, redis etc
	Host      string `json:"host" bson:"host"`         // server1
	Source    string `json:"source" bson:"source"`     // nginx, redis etc
	Message   string `json:"message" bson:"message"`
	EventType string `json:"event_type" bson:"event_type"` // http_request

	TraceID   string `json:"trace_id,omitempty" bson:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty" bson:"request_id,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

type Parser struct {
	Pattern         string                    `json:"pattern" bson:"pattern"`
	Mapping         map[string]string         `json:"mapping" bson:"mapping"`
	Transformations map[string]Transformation `json:"transformations,omitempty" bson:"transformations,omitempty"`
}

type Transformation struct {
	Type   string `json:"type" bson:"type"`
	Format string `json:"format,omitempty" bson:"format,omitempty"`
}
