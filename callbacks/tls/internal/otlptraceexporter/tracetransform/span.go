package tracetransform

import (
	"math"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func Spans(spans []tracesdk.ReadOnlySpan) []*tracepb.ResourceSpans {
	if len(spans) == 0 {
		return nil
	}

	type key struct {
		resource attribute.Distinct
		scope    instrumentation.Scope
	}

	rsm := make(map[attribute.Distinct]*tracepb.ResourceSpans)
	ilsm := make(map[key]*tracepb.InstrumentationLibrarySpans)

	var resourceCount int
	for _, sd := range spans {
		if sd == nil {
			continue
		}

		rKey := sd.Resource().Equivalent()
		k := key{
			resource: rKey,
			scope:    sd.InstrumentationScope(),
		}

		ils, scopeSeen := ilsm[k]
		if !scopeSeen {
			ils = &tracepb.InstrumentationLibrarySpans{
				InstrumentationLibrary: InstrumentationLibrary(sd.InstrumentationScope()),
				Spans:                  []*tracepb.Span{},
			}
		}
		ils.Spans = append(ils.Spans, span(sd))
		ilsm[k] = ils

		rs, resourceSeen := rsm[rKey]
		if !resourceSeen {
			resourceCount++
			rs = &tracepb.ResourceSpans{
				Resource:                    Resource(sd.Resource()),
				InstrumentationLibrarySpans: []*tracepb.InstrumentationLibrarySpans{ils},
			}
			rsm[rKey] = rs
			continue
		}

		if !scopeSeen {
			rs.InstrumentationLibrarySpans = append(rs.InstrumentationLibrarySpans, ils)
		}
	}

	out := make([]*tracepb.ResourceSpans, 0, resourceCount)
	for _, rs := range rsm {
		out = append(out, rs)
	}
	return out
}

func span(sd tracesdk.ReadOnlySpan) *tracepb.Span {
	tid := sd.SpanContext().TraceID()
	sid := sd.SpanContext().SpanID()

	s := &tracepb.Span{
		TraceId:                tid[:],
		SpanId:                 sid[:],
		TraceState:             sd.SpanContext().TraceState().String(),
		Status:                 status(sd.Status().Code, sd.Status().Description),
		StartTimeUnixNano:      uint64(maxInt64(0, sd.StartTime().UnixNano())),
		EndTimeUnixNano:        uint64(maxInt64(0, sd.EndTime().UnixNano())),
		Links:                  links(sd.Links()),
		Kind:                   spanKind(sd.SpanKind()),
		Name:                   sd.Name(),
		Attributes:             KeyValues(sd.Attributes()),
		Events:                 spanEvents(sd.Events()),
		DroppedAttributesCount: clampUint32(sd.DroppedAttributes()),
		DroppedEventsCount:     clampUint32(sd.DroppedEvents()),
		DroppedLinksCount:      clampUint32(sd.DroppedLinks()),
	}

	if psid := sd.Parent().SpanID(); psid.IsValid() {
		s.ParentSpanId = psid[:]
	}

	return s
}

func status(code codes.Code, message string) *tracepb.Status {
	var statusCode tracepb.Status_StatusCode
	switch code {
	case codes.Ok:
		statusCode = tracepb.Status_STATUS_CODE_OK
	case codes.Error:
		statusCode = tracepb.Status_STATUS_CODE_ERROR
	default:
		statusCode = tracepb.Status_STATUS_CODE_UNSET
	}

	return &tracepb.Status{
		Code:    statusCode,
		Message: message,
	}
}

func links(links []tracesdk.Link) []*tracepb.Span_Link {
	if len(links) == 0 {
		return nil
	}

	out := make([]*tracepb.Span_Link, 0, len(links))
	for _, link := range links {
		link := link
		tid := link.SpanContext.TraceID()
		sid := link.SpanContext.SpanID()

		out = append(out, &tracepb.Span_Link{
			TraceId:    tid[:],
			SpanId:     sid[:],
			Attributes: KeyValues(link.Attributes),
		})
	}

	return out
}

func spanEvents(events []tracesdk.Event) []*tracepb.Span_Event {
	if len(events) == 0 {
		return nil
	}

	out := make([]*tracepb.Span_Event, len(events))
	for i := range events {
		out[i] = &tracepb.Span_Event{
			Name:         events[i].Name,
			TimeUnixNano: uint64(maxInt64(0, events[i].Time.UnixNano())),
			Attributes:   KeyValues(events[i].Attributes),
		}
	}

	return out
}

func spanKind(kind trace.SpanKind) tracepb.Span_SpanKind {
	switch kind {
	case trace.SpanKindInternal:
		return tracepb.Span_SPAN_KIND_INTERNAL
	case trace.SpanKindClient:
		return tracepb.Span_SPAN_KIND_CLIENT
	case trace.SpanKindServer:
		return tracepb.Span_SPAN_KIND_SERVER
	case trace.SpanKindProducer:
		return tracepb.Span_SPAN_KIND_PRODUCER
	case trace.SpanKindConsumer:
		return tracepb.Span_SPAN_KIND_CONSUMER
	default:
		return tracepb.Span_SPAN_KIND_UNSPECIFIED
	}
}

func clampUint32(v int) uint32 {
	if v < 0 {
		return 0
	}
	if int64(v) > math.MaxUint32 {
		return math.MaxUint32
	}

	return uint32(v)
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
