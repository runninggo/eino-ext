package otlptraceexporter

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/sdk/trace"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/cloudwego/eino-ext/callbacks/tls/internal/otlptraceexporter/tracetransform"
)

type Exporter struct {
	client   coltracepb.TraceServiceClient
	conn     *grpc.ClientConn
	metadata metadata.MD
}

func New(endpoint string, headers map[string]string, tlsEnabled bool) (*Exporter, error) {
	dialOpts := make([]grpc.DialOption, 0, 1)
	if tlsEnabled {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(endpoint, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial otlp trace endpoint: %w", err)
	}

	return &Exporter{
		client:   coltracepb.NewTraceServiceClient(conn),
		conn:     conn,
		metadata: metadata.New(headers),
	}, nil
}

func (e *Exporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	if len(spans) == 0 {
		return nil
	}
	if len(e.metadata) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, e.metadata)
	}

	_, err := e.client.Export(ctx, &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: tracetransform.Spans(spans),
	})
	if err != nil {
		return fmt.Errorf("export spans: %w", err)
	}

	return nil
}

func (e *Exporter) Shutdown(context.Context) error {
	if e == nil || e.conn == nil {
		return nil
	}

	return e.conn.Close()
}
