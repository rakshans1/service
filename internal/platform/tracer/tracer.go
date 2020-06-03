package tracer

import (
	"log"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init creates a new trace provider instance and registers it as global trace provider.
func Init(serviceName string, reporterURI string, probability float64, log *log.Logger) (func(),error) {
  tp, flush, err := jaeger.NewExportPipeline(
        jaeger.WithCollectorEndpoint(reporterURI),
          jaeger.WithProcess(jaeger.Process{
              ServiceName: serviceName,
          }),
        jaeger.RegisterAsGlobal(),
        jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
  )
	if err != nil {
		return nil, errors.Wrap(err, "creating new exporter")
	}
	global.SetTraceProvider(tp)
	return flush, nil
}
