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
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

var (
	once sync.Once

	// jobsMapGcTotal = stats.Int64("jobs_map_gc_total", "total gc done by jobs map", stats.UnitDimensionless)
	// jobsMapTsTotal = stats.Int64("jobs_map_timeseries_total", "total timeseries held by jobs map", stats.UnitDimensionless)
	//
	// serviceIdKey  = tag.MustNewKey("service_instance_id")
	// tsLocationKey = tag.MustNewKey("timeseries_location")

	jobsMapTimeSeries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jobs_map_timeseries",
			Help: "total timeseries held by jobs map",
		},
		[]string{"receiver", "timeseries_location"},
	)
	familiesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "metrics_builder_families",
			Help: "total timeseries held by jobs map",
		},
		[]string{"receiver"},
	)

	jobsMapGcTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_map_gc_total",
			Help: "total gc done by jobs map",
		},
		[]string{"receiver"},
	)
	jobsMapGcDeletedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_map_gc_deleted_total",
			Help: "",
		},
		[]string{"receiver"},
	)
	TsiMapGcDeletedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tsi_map_gc_deleted_total",
			Help: "",
		},
		[]string{"receiver"},
	)
	MetricsAdjusterResetTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "metrics_adjust_reset_total",
			Help: "",
		},
		[]string{"receiver"},
	)

	metricsGroupCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "metrics_group_created_total",
			Help: "total created metrics group",
		},
		[]string{"receiver"},
	)
	metricsFamilyAddedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "metrics_family_added_total",
			Help: "",
		},
		[]string{"receiver"},
	)
	timeseriesInfoCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "timeseries_info_created_total",
			Help: "",
		},
		[]string{"receiver"},
	)

	toMetricTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "to_metric_created_total",
			Help: "total created metrics group",
		},
		[]string{"type", "status"},
	)
)

func RegisterView() {
	views := []*view.View{
		// {
		// 	Name:        jobsMapGcTotal.Name(),
		// 	Description: jobsMapGcTotal.Description(),
		// 	TagKeys:     []tag.Key{serviceIdKey},
		// 	Measure:     jobsMapGcTotal,
		// 	Aggregation: view.Sum(),
		// },
		// {
		// 	Name:        jobsMapTsTotal.Name(),
		// 	Description: jobsMapTsTotal.Description(),
		// 	TagKeys:     []tag.Key{serviceIdKey, tsLocationKey},
		// 	Measure:     jobsMapTsTotal,
		// 	Aggregation: view.Sum(),
		// },
	}

	once.Do(func() {
		view.Register(views...)
		prometheus.MustRegister(
			jobsMapTimeSeries,
			familiesTotal,
			jobsMapGcTotal,
			metricsFamilyAddedTotal,
			jobsMapGcDeletedTotal,
			TsiMapGcDeletedTotal,
			metricsGroupCreatedTotal,
			toMetricTotal,
			timeseriesInfoCreatedTotal,
			MetricsAdjusterResetTotal,
		)
	})
}

func Len(m *pmetric.Metric) int {
	// TODO count attrs?
	// metrics := pmetric.NewMetrics()
	// for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
	// 	rm := metrics.ResourceMetrics().At(i)
	// 	for j := 0; j < rm.ScopeMetrics().Len(); j++ {
	// 		ilm := rm.ScopeMetrics().At(j)
	// 		for k := 0; k < ilm.Metrics().Len(); k++ {
	// 			m := ilm.Metrics().At(k)
	// 		}
	// 	}
	// }

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
