package opentelemetry

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"

	otelObs "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/extensions"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

var (
	event cloudevents.Event = createCloudEvent(extensions.DistributedTracingExtension{})
)

func TestRecordSendingEvent(t *testing.T) {

	tests := []struct {
		name             string
		expectedSpanName string
		expectedStatus   codes.Code
		expectedAttrs    []attribute.KeyValue
		nameFormatter    func(*cloudevents.Event) string
		attributesGetter func(*cloudevents.Event) []attribute.KeyValue
		expectedResult   protocol.Result
	}{

		{
			name:             "send with default options",
			expectedSpanName: "cloudevents.client.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&event, "RecordSendingEvent"),
			nameFormatter:    nil,
		},
		{
			name:             "send with custom span name",
			expectedSpanName: "test.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&event, "RecordSendingEvent"),
			nameFormatter: func(e *cloudevents.Event) string {
				return "test." + e.Context.GetType()
			},
		},
		{
			name:             "send with custom attributes",
			expectedSpanName: "test.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    append(otelObs.GetDefaultSpanAttributes(&event, "RecordSendingEvent"), attribute.String("my-attr", "some-value")),
			nameFormatter: func(e *cloudevents.Event) string {
				return "test." + e.Context.GetType()
			},
			attributesGetter: func(*cloudevents.Event) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("my-attr", "some-value"),
				}
			},
		},
		{
			name:             "send with error response",
			expectedSpanName: "cloudevents.client.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&event, "RecordSendingEvent"),
			expectedResult:   protocol.NewReceipt(false, "some error here"),
		},
		{
			name:             "send with http error response",
			expectedSpanName: "cloudevents.client.example.type send",
			expectedStatus:   codes.Error,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&event, "RecordSendingEvent"),
			expectedResult:   http.NewResult(500, "some server error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sr, _ := configureOtelTestSdk()
			ctx := context.Background()

			os := otelObs.NewOTelObservabilityService(
				otelObs.WithSpanNameFormatter(tc.nameFormatter),
				otelObs.WithSpanAttributesGetter(tc.attributesGetter))

			// act
			ctx, cb := os.RecordSendingEvent(ctx, event)
			cb(tc.expectedResult)

			spans := sr.Ended()

			// since the obs service started a span, the context should have the spancontext
			assert.NotNil(t, trace.SpanContextFromContext(ctx))
			assert.Equal(t, 1, len(spans))

			span := spans[0]
			assert.Equal(t, tc.expectedSpanName, span.Name())
			assert.Equal(t, tc.expectedStatus, span.Status().Code)

			if !reflect.DeepEqual(span.Attributes(), tc.expectedAttrs) {
				t.Errorf("p = %v, want %v", span.Attributes(), tc.expectedAttrs)
			}

			if tc.expectedResult != nil {
				assert.Equal(t, 1, len(span.Events()))
				assert.Equal(t, semconv.ExceptionEventName, span.Events()[0].Name)

				attrsMap := getSpanEventMap(span.Events()[0].Attributes)
				assert.Equal(t, tc.expectedResult.Error(), attrsMap[string(semconv.ExceptionMessageKey)])
			}
		})
	}

}

func getSpanEventMap(evtAttrs []attribute.KeyValue) map[string]string {
	attr := map[string]string{}
	for _, v := range evtAttrs {
		attr[string(v.Key)] = v.Value.AsString()
	}
	return attr
}