package tracer

import (
	"math/rand"
)

// Span is the common struct we use to represent a dapper-like span.
// More information about the structure of the Span can be found
// here: http://research.google.com/pubs/pub36356.html
type Span struct {
	Name     string             `json:"name"`      // the name of what we're monitoring (e.g. redis.command)
	Service  string             `json:"service"`   // the service related to this trace (e.g. redis)
	Resource string             `json:"resource"`  // the natural key of what we measure (e.g. GET)
	Type     string             `json:"type"`      // protocol associated with the span
	Start    int64              `json:"start"`     // span start time expressed in nanoseconds since epoch
	Duration int64              `json:"duration"`  // duration of the span expressed in nanoseconds
	Error    int32              `json:"error"`     // error status of the span; 0 means no errors
	Meta     map[string]string  `json:"meta"`      // arbitrary map of metadata
	Metrics  map[string]float64 `json:"metrics"`   // arbitrary map of numeric metrics
	SpanID   uint64             `json:"span_id"`   // identifier of this span
	TraceID  uint64             `json:"trace_id"`  // identifier of the root span
	ParentID uint64             `json:"parent_id"` // identifier of the span's direct parent
}

// NewSpan creates a new Span with the given arguments, and sets
// the internal Start field.
func NewSpan(spanID, traceID, parentID uint64, service, name, resource string) *Span {
	return &Span{
		SpanID:   spanID,
		TraceID:  traceID,
		ParentID: parentID,
		Service:  service,
		Name:     name,
		Resource: resource,
		Start:    Now(),
	}
}

// Trace is an helper that creates a new root Span with a
// random identifier. This high-level API is commonly used to
// start a new tracing session.
func Trace(service, name, resource string) *Span {
	spanID := nextSpanID()
	// the parentID must be set to 0 otherwise this is not
	// a root span
	return NewSpan(spanID, spanID, 0, service, name, resource)
}

// Nest returns a new span that is child of the current Span instance
// This high-level API is commonly used to create a nested span in the
// current tracing session.
func (s *Span) Nest(service, name, resource string) *Span {
	spanID := nextSpanID()
	return NewSpan(spanID, s.TraceID, s.ParentID, service, name, resource)
}

// SetMeta adds an arbitrary meta field to the current Span.
// This method is not thread-safe and the Span should not be modified
// by multiple go routine.
func (s *Span) SetMeta(key, value string) {
	// TODO: should we make the Span thread-safe? this means adding a
	// sync.Mutex to the Span struct
	if s.Meta == nil {
		s.Meta = make(map[string]string)
	}
	s.Meta[key] = value
}

// IsFinished returns true if the span.Finish() method has been called.
// Under the hood, any Span with a Duration has to be considered closed.
func (s *Span) IsFinished() bool {
	return s.Duration > 0
}

// Finish closes this Span (but not its children) providing the duration
// of this part of the tracing session. This method is idempotent so
// calling this method multiple times is safe and doesn't update the
// current Span.
func (s *Span) Finish() {
	if !s.IsFinished() {
		s.Duration = Now() - s.Start

		// TODO: a go routine should start and this Span must be sent
		// in the outgoingPacket channel for more processing.
	}
}

// nextSpanID returns a new random identifier. It is meant to be used as a
// SpanID for the Span struct. Changing this function impacts the whole
// package.
func nextSpanID() uint64 {
	return uint64(rand.Int63())
}
