package tracetransform

import (
	"go.opentelemetry.io/otel/sdk/resource"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

func Resource(r *resource.Resource) *resourcepb.Resource {
	if r == nil {
		return nil
	}

	return &resourcepb.Resource{Attributes: ResourceAttributes(r)}
}
