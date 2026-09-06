package parsergen

import (
	"context"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parser"
)

// ParserGenMessage is the payload published to RabbitMQ to request parser generation.
type ParserGenMessage struct {
	Hash     string                          `json:"hash"`
	RawLog   string                          `json:"raw_log"`
	Features fingerprint.FingerprintFeatures `json:"features"`
}

// MessageConsumer abstracts consuming messages from a message broker (e.g. RabbitMQ).
type MessageConsumer interface {
	ConsumeMessages(ctx context.Context, handler func(message []byte) error) error
}

// MLParserClient abstracts requesting parsers from an ML service.
type MLParserClient interface {
	RequestParser(ctx context.Context, rawLog string, features fingerprint.FingerprintFeatures) (*parser.Parser, error)
}
