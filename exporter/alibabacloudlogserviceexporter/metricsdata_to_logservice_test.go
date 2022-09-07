// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package alibabacloudlogserviceexporter

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestMetricDataToLogService(t *testing.T) {
	logger := zap.NewNop()

	md := pmetric.NewMetrics()
	md.ResourceMetrics().AppendEmpty() // Add an empty ResourceMetrics
	rm := md.ResourceMetrics().AppendEmpty()

	rm.Resource().Attributes().UpsertString("labelB", "valueB")
	rm.Resource().Attributes().UpsertString("labelA", "valueA")
	rm.Resource().Attributes().UpsertString("service.name", "unknown-service")
	rm.Resource().Attributes().UpsertString("a", "b")
	sms := rm.ScopeMetrics()
	sms.AppendEmpty() // Add an empty ScopeMetrics
	sm := sms.AppendEmpty()

	metrics := sm.Metrics()

	badNameMetric := metrics.AppendEmpty()
	badNameMetric.SetName("")

	noneMetric := metrics.AppendEmpty()
	noneMetric.SetName("none")

	intGaugeMetric := metrics.AppendEmpty()
	intGaugeMetric.SetDataType(pmetric.MetricDataTypeGauge)
	intGaugeMetric.SetName("int_gauge")
	intGauge := intGaugeMetric.Gauge()
	intGaugeDataPoints := intGauge.DataPoints()
	intGaugeDataPoint := intGaugeDataPoints.AppendEmpty()
	intGaugeDataPoint.Attributes().UpsertString("innerLabel", "innerValue")
	intGaugeDataPoint.SetIntVal(10)
	intGaugeDataPoint.SetTimestamp(pcommon.Timestamp(100_000_000))

	doubleGaugeMetric := metrics.AppendEmpty()
	doubleGaugeMetric.SetDataType(pmetric.MetricDataTypeGauge)
	doubleGaugeMetric.SetName("double_gauge")
	doubleGauge := doubleGaugeMetric.Gauge()
	doubleGaugeDataPoints := doubleGauge.DataPoints()
	doubleGaugeDataPoint := doubleGaugeDataPoints.AppendEmpty()
	doubleGaugeDataPoint.Attributes().UpsertString("innerLabel", "innerValue")
	doubleGaugeDataPoint.SetDoubleVal(10.1)
	doubleGaugeDataPoint.SetTimestamp(pcommon.Timestamp(100_000_000))

	intSumMetric := metrics.AppendEmpty()
	intSumMetric.SetDataType(pmetric.MetricDataTypeSum)
	intSumMetric.SetName("int_sum")
	intSum := intSumMetric.Sum()
	intSumDataPoints := intSum.DataPoints()
	intSumDataPoint := intSumDataPoints.AppendEmpty()
	intSumDataPoint.Attributes().UpsertString("innerLabel", "innerValue")
	intSumDataPoint.SetIntVal(11)
	intSumDataPoint.SetTimestamp(pcommon.Timestamp(100_000_000))

	doubleSumMetric := metrics.AppendEmpty()
	doubleSumMetric.SetDataType(pmetric.MetricDataTypeSum)
	doubleSumMetric.SetName("double_sum")
	doubleSum := doubleSumMetric.Sum()
	doubleSumDataPoints := doubleSum.DataPoints()
	doubleSumDataPoint := doubleSumDataPoints.AppendEmpty()
	doubleSumDataPoint.Attributes().UpsertString("innerLabel", "innerValue")
	doubleSumDataPoint.SetDoubleVal(10.1)
	doubleSumDataPoint.SetTimestamp(pcommon.Timestamp(100_000_000))

	doubleHistogramMetric := metrics.AppendEmpty()
	doubleHistogramMetric.SetDataType(pmetric.MetricDataTypeHistogram)
	doubleHistogramMetric.SetName("double_$histogram")
	doubleHistogram := doubleHistogramMetric.Histogram()
	doubleHistogramDataPoints := doubleHistogram.DataPoints()
	doubleHistogramDataPoint := doubleHistogramDataPoints.AppendEmpty()
	doubleHistogramDataPoint.Attributes().UpsertString("innerLabel", "innerValue")
	doubleHistogramDataPoint.SetCount(2)
	doubleHistogramDataPoint.SetSum(10.1)
	doubleHistogramDataPoint.SetTimestamp(pcommon.Timestamp(100_000_000))
	doubleHistogramDataPoint.SetBucketCounts(pcommon.NewImmutableUInt64Slice([]uint64{1, 2, 3}))
	doubleHistogramDataPoint.SetExplicitBounds(pcommon.NewImmutableFloat64Slice([]float64{1, 2}))

	doubleSummaryMetric := metrics.AppendEmpty()
	doubleSummaryMetric.SetDataType(pmetric.MetricDataTypeSummary)
	doubleSummaryMetric.SetName("double-summary")
	doubleSummary := doubleSummaryMetric.Summary()
	doubleSummaryDataPoints := doubleSummary.DataPoints()
	doubleSummaryDataPoint := doubleSummaryDataPoints.AppendEmpty()
	doubleSummaryDataPoint.SetCount(2)
	doubleSummaryDataPoint.SetSum(10.1)
	doubleSummaryDataPoint.SetTimestamp(pcommon.Timestamp(100_000_000))
	doubleSummaryDataPoint.Attributes().UpsertString("innerLabel", "innerValue")
	quantileVal := doubleSummaryDataPoint.QuantileValues().AppendEmpty()
	quantileVal.SetValue(10.2)
	quantileVal.SetQuantile(0.9)
	quantileVal2 := doubleSummaryDataPoint.QuantileValues().AppendEmpty()
	quantileVal2.SetValue(10.5)
	quantileVal2.SetQuantile(0.95)

	gotLogs := metricsDataToLogServiceData(logger, md)
	gotLogPairs := make([][]logKeyValuePair, 0, len(gotLogs))

	for _, log := range gotLogs {
		pairs := make([]logKeyValuePair, 0, len(log.Contents))
		for _, content := range log.Contents {
			pairs = append(pairs, logKeyValuePair{
				Key:   content.GetKey(),
				Value: content.GetValue(),
			})
		}
		gotLogPairs = append(gotLogPairs, pairs)

	}

	wantLogs := make([][]logKeyValuePair, 0, len(gotLogs))

	if err := loadFromJSON("./testdata/logservice_metric_data.json", &wantLogs); err != nil {
		t.Errorf("Failed load log key value pairs from file, error: %v", err)
		return
	}
	assert.Equal(t, len(wantLogs), len(gotLogs))
	for j := 0; j < len(gotLogs); j++ {
		sort.Sort(logKeyValuePairs(gotLogPairs[j]))
		sort.Sort(logKeyValuePairs(wantLogs[j]))
		if !reflect.DeepEqual(gotLogPairs[j], wantLogs[j]) {
			t.Errorf("Unsuccessful conversion \nGot:\n\t%v\nWant:\n\t%v", gotLogPairs, wantLogs)
		}
	}
}

func TestMetricCornerCases(t *testing.T) {
	assert.Equal(t, min(1, 2), 1)
	assert.Equal(t, min(2, 1), 1)
	assert.Equal(t, min(1, 1), 1)
	var label KeyValues
	label.Append("a", "b")
	assert.Equal(t, label.String(), "a#$#b")
}

func TestMetricLabelSanitize(t *testing.T) {
	var label KeyValues
	label.Append("_test", "key_test")
	label.Append("0test", "key_0test")
	label.Append("test_normal", "test_normal")
	label.Append("0test", "key_0test")
	assert.Equal(t, label.String(), "key_test#$#key_test|key_0test#$#key_0test|test_normal#$#test_normal|key_0test#$#key_0test")
	label.Sort()
	assert.Equal(t, label.String(), "key_0test#$#key_0test|key_0test#$#key_0test|key_test#$#key_test|test_normal#$#test_normal")
}
