package client

import (
	"go.opentelemetry.io/otel/attribute"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/observability"
)

const (
	// TODO: Should this be the name of this module or the whole package?
	instrumentationName = "github.com/cloudevents/sdk-go/observability/opentelemetry/v2"
)

type OTelObservabilityServiceOption func(*OTelObservabilityService)

func WithSpanAttributesGetter(attrGetter func(cloudevents.Event) []attribute.KeyValue) OTelObservabilityServiceOption {
	return func(os *OTelObservabilityService) {
		if attrGetter != nil {
			os.spanAttributesGetter = attrGetter
		}
	}
}

func WithSpanNameFormatter(nameFormatter func(cloudevents.Event) string) OTelObservabilityServiceOption {
	return func(os *OTelObservabilityService) {
		if nameFormatter != nil {
			os.spanNameFormatter = nameFormatter
		}
	}
}

var defaultSpanNameFormatter func(cloudevents.Event) string = func(e cloudevents.Event) string {
	return observability.ClientSpanName + "." + e.Context.GetType()
}
