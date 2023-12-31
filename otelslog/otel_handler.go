package otelslog

import (
	"context"
	otel "github.com/agoda-com/opentelemetry-logs-go/logs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"io"
	"log/slog"
	"sync"
)

const (
	instrumentationName = "github.com/chameleon82/go-modules-test/otelslog"
)

// OtelHandler is a Handler that writes Records to OTLP
type OtelHandler struct {
	otelHandler
}

type otelHandler struct {
	logger otel.Logger
	opts   HandlerOptions
	mu     *sync.Mutex
	w      io.Writer
}

// compilation time verification handler implement interface
var _ slog.Handler = &otelHandler{}

var instrumentationScope = instrumentation.Scope{
	Name:      instrumentationName,
	Version:   Version(),
	SchemaURL: semconv.SchemaURL,
}

func (o otelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (o otelHandler) Handle(ctx context.Context, record slog.Record) error {

	spanContext := trace.SpanFromContext(ctx).SpanContext()
	var traceID *trace.TraceID = nil
	var spanID *trace.SpanID = nil
	var traceFlags *trace.TraceFlags = nil
	if spanContext.IsValid() {
		tid := spanContext.TraceID()
		sid := spanContext.SpanID()
		tf := spanContext.TraceFlags()
		traceID = &tid
		spanID = &sid
		traceFlags = &tf
	}
	levelString := record.Level.String()
	severity := otel.SeverityNumber(int(record.Level.Level()) + 9)

	var attributes []attribute.KeyValue

	record.Attrs(func(attr slog.Attr) bool {
		attributes = append(attributes, otelAttribute(attr)...)
		return true
	})

	lrc := otel.LogRecordConfig{
		Timestamp:            &record.Time,
		ObservedTimestamp:    record.Time,
		TraceId:              traceID,
		SpanId:               spanID,
		TraceFlags:           traceFlags,
		SeverityText:         &levelString,
		SeverityNumber:       &severity,
		Body:                 &record.Message,
		Resource:             nil,
		InstrumentationScope: &instrumentationScope,
		Attributes:           &attributes,
	}

	r := otel.NewLogRecord(lrc)
	o.logger.Emit(r)
	return nil
}

func (o otelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	//TODO implement me
	panic("not implemented yet")
}

func (o otelHandler) WithGroup(name string) slog.Handler {
	//TODO implement me
	panic("not implemented yet")
}

// HandlerOptions are options for a OtelHandler.
// A zero HandlerOptions consists entirely of default values.
type HandlerOptions struct {
}

// NewOtelHandler creates a OtelHandler that writes to otlp,
// using the given options.
// If opts is nil, the default options are used.
func NewOtelHandler(loggerProvider otel.LoggerProvider, opts *HandlerOptions) *OtelHandler {
	logger := loggerProvider.Logger(
		instrumentationScope.Name,
		otel.WithInstrumentationVersion(instrumentationScope.Version),
	)
	return &OtelHandler{
		otelHandler{
			logger: logger,
		},
	}
}
