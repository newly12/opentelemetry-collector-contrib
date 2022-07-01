// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

var (
	once sync.Once

	jobsMapGcTotal = stats.Int64("jobs_map_gc_total", "total gc done by jobs map", stats.UnitDimensionless)
	jobsMapTsTotal = stats.Int64("jobs_map_timeseries_total", "total timeseries held by jobs map", stats.UnitDimensionless)

	serviceIdKey  = tag.MustNewKey("service_instance_id")
	tsLocationKey = tag.MustNewKey("timeseries_location")

	jobsMapTimeSeries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jobs_map_gc_total",
			Help: "total gc done by jobs map",
		},
		[]string{"receiver", "timeseries_location"},
	)
)

func RegisterView() {
	views := []*view.View{
		{
			Name:        jobsMapGcTotal.Name(),
			Description: jobsMapGcTotal.Description(),
			TagKeys:     []tag.Key{serviceIdKey},
			Measure:     jobsMapGcTotal,
			Aggregation: view.Sum(),
		},
		{
			Name:        jobsMapTsTotal.Name(),
			Description: jobsMapTsTotal.Description(),
			TagKeys:     []tag.Key{serviceIdKey, tsLocationKey},
			Measure:     jobsMapTsTotal,
			Aggregation: view.Sum(),
		},
	}

	once.Do(func() {
		view.Register(views...)
		prometheus.MustRegister(
			jobsMapTimeSeries,
		)
	})
}

func Len(m *pmetric.Metric) int {
	var total int
	switch m.DataType() {
	case pmetric.MetricDataTypeGauge:
		total += m.Gauge().DataPoints().Len()
	case pmetric.MetricDataTypeSummary:
		total += m.Summary().DataPoints().Len()
	case pmetric.MetricDataTypeSum:
		total += m.Sum().DataPoints().Len()
	case pmetric.MetricDataTypeHistogram:
		total += m.Histogram().DataPoints().Len()
	case pmetric.MetricDataTypeExponentialHistogram:
		total += m.ExponentialHistogram().DataPoints().Len()
	}
	return total
}
