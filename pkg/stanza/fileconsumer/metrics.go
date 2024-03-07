package fileconsumer

import (
	"go.opentelemetry.io/otel/metric"
)

type internalTelemetry struct {
	offsetGauge  metric.Int64ObservableGauge
	registration metric.Registration
}
