package tracetransform

import (
	"go.opentelemetry.io/otel/sdk/instrumentation"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

func InstrumentationLibrary(scope instrumentation.Scope) *commonpb.InstrumentationLibrary {
	if scope == (instrumentation.Scope{}) {
		return nil
	}

	return &commonpb.InstrumentationLibrary{
		Name:    scope.Name,
		Version: scope.Version,
	}
}
