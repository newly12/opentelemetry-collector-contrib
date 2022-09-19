// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package traces

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

var (
	TestSpanStartTime      = time.Date(2020, 2, 11, 20, 26, 12, 321, time.UTC)
	TestSpanStartTimestamp = pcommon.NewTimestampFromTime(TestSpanStartTime)

	TestSpanEndTime      = time.Date(2020, 2, 11, 20, 26, 13, 789, time.UTC)
	TestSpanEndTimestamp = pcommon.NewTimestampFromTime(TestSpanEndTime)

	traceID = [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID  = [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	spanID2 = [8]byte{8, 7, 6, 5, 4, 3, 2, 1}
)

func TestProcess(t *testing.T) {
	tests := []struct {
		query string
		want  func(td ptrace.Traces)
	}{
		{
			query: `set(attributes["test"], "pass") where name == "operationA"`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where resource.attributes["host.name"] == "localhost"`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `keep_keys(attributes, "http.method") where name == "operationA"`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().Clear()
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("http.method", "get")
			},
		},
		{
			query: `set(status.code, 1) where attributes["http.path"] == "/health"`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Status().SetCode(ptrace.StatusCodeOk)
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).Status().SetCode(ptrace.StatusCodeOk)
			},
		},
		{
			query: `set(attributes["test"], "pass") where dropped_attributes_count == 1`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where dropped_events_count == 1`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where dropped_links_count == 1`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where span_id == SpanID(0x0102030405060708)`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where parent_span_id == SpanID(0x0807060504030201)`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where trace_id == TraceID(0x0102030405060708090a0b0c0d0e0f10)`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where trace_state == "new"`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `replace_pattern(attributes["http.method"], "get", "post")`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("http.method", "post")
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).Attributes().PutString("http.method", "post")
			},
		},
		{
			query: `replace_all_patterns(attributes, "get", "post")`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("http.method", "post")
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).Attributes().PutString("http.method", "post")
			},
		},
		{
			query: `set(attributes["test"], "pass") where IsMatch(name, "operation[AC]") == true`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where attributes["doesnt exist"] == nil`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `delete_key(attributes, "http.url") where name == "operationA"`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().Clear()
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("http.method", "get")
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("http.path", "/health")
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("flags", "A|B|C")
			},
		},
		{
			query: `delete_matching_keys(attributes, "http.*t.*") where name == "operationA"`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().Clear()
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("http.url", "http://localhost/health")
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("flags", "A|B|C")
			},
		},
		{
			query: `set(attributes["test"], "pass") where kind == SPAN_KIND_INTERNAL`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "pass")
			},
		},
		{
			query: `set(kind, SPAN_KIND_SERVER) where kind == 1`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetKind(2)
			},
		},
		{
			query: `set(attributes["test"], Concat(": ", attributes["http.method"], attributes["http.url"]))`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "get: http://localhost/health")
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).Attributes().PutString("test", "get: http://localhost/health")
			},
		},
		{
			query: `set(attributes["test"], Concat("", attributes["http.method"], ": ", attributes["http.url"]))`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "get: http://localhost/health")
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).Attributes().PutString("test", "get: http://localhost/health")
			},
		},
		{
			query: `set(attributes["test"], Concat(": ", attributes["http.method"], attributes["http.url"])) where name == Concat("", "operation", "A")`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("test", "get: http://localhost/health")
			},
		},
		{
			query: `set(attributes["kind"], Concat("", "kind", ": ", kind)) where kind == 1`,
			want: func(td ptrace.Traces) {
				td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutString("kind", "kind: 1")
			},
		},
		{
			query: `set(attributes["test"], Split(attributes["flags"], "|"))`,
			want: func(td ptrace.Traces) {
				v1 := td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutEmptySlice("test")
				v1.AppendEmpty().SetStringVal("A")
				v1.AppendEmpty().SetStringVal("B")
				v1.AppendEmpty().SetStringVal("C")
				v2 := td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).Attributes().PutEmptySlice("test")
				v2.AppendEmpty().SetStringVal("C")
				v2.AppendEmpty().SetStringVal("D")
			},
		},
		{
			query: `set(attributes["test"], Split(attributes["flags"], "|")) where name == "operationA"`,
			want: func(td ptrace.Traces) {
				v1 := td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().PutEmptySlice("test")
				v1.AppendEmpty().SetStringVal("A")
				v1.AppendEmpty().SetStringVal("B")
				v1.AppendEmpty().SetStringVal("C")
			},
		},
		{
			query: `set(attributes["test"], Split(attributes["not_exist"], "|"))`,
			want:  func(td ptrace.Traces) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			td := constructTraces()
			processor, err := NewProcessor([]string{tt.query}, Functions(), component.ProcessorCreateSettings{})
			assert.NoError(t, err)

			_, err = processor.ProcessTraces(context.Background(), td)
			assert.NoError(t, err)

			exTd := constructTraces()
			tt.want(exTd)

			assert.Equal(t, exTd, td)
		})
	}
}

func BenchmarkTwoSpans(b *testing.B) {
	tests := []struct {
		name    string
		queries []string
	}{
		{
			name:    "no processing",
			queries: []string{},
		},
		{
			name:    "set attribute",
			queries: []string{`set(attributes["test"], "pass") where name == "operationA"`},
		},
		{
			name:    "keep_keys attribute",
			queries: []string{`keep_keys(attributes, "http.method") where name == "operationA"`},
		},
		{
			name:    "no match",
			queries: []string{`keep_keys(attributes, "http.method") where name == "unknownOperation"`},
		},
		{
			name:    "inner field",
			queries: []string{`set(status.code, 1) where attributes["http.path"] == "/health"`},
		},
		{
			name: "inner field both spans",
			queries: []string{
				`set(status.code, 1) where name == "operationA"`,
				`set(status.code, 2) where name == "operationB"`,
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			processor, err := NewProcessor(tt.queries, Functions(), component.ProcessorCreateSettings{})
			assert.NoError(b, err)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				td := constructTraces()
				_, err = processor.ProcessTraces(context.Background(), td)
				assert.NoError(b, err)
			}
		})
	}
}

func BenchmarkHundredSpans(b *testing.B) {
	tests := []struct {
		name    string
		queries []string
	}{
		{
			name:    "no processing",
			queries: []string{},
		},
		{
			name: "set status code",
			queries: []string{
				`set(status.code, 1) where name == "operationA"`,
				`set(status.code, 2) where name == "operationB"`,
			},
		},
		{
			name: "hundred queries",
			queries: func() []string {
				var queries []string
				queries = append(queries, `set(status.code, 1) where name == "operationA"`)
				for i := 0; i < 99; i++ {
					queries = append(queries, `keep_keys(attributes, "http.method") where name == "unknownOperation"`)
				}
				return queries
			}(),
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			processor, err := NewProcessor(tt.queries, Functions(), component.ProcessorCreateSettings{})
			assert.NoError(b, err)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				td := constructTracesNum(100)
				_, err = processor.ProcessTraces(context.Background(), td)
				assert.NoError(b, err)
			}
		})
	}
}

func constructTraces() ptrace.Traces {
	td := ptrace.NewTraces()
	rs0 := td.ResourceSpans().AppendEmpty()
	rs0.Resource().Attributes().PutString("host.name", "localhost")
	rs0ils0 := rs0.ScopeSpans().AppendEmpty()
	fillSpanOne(rs0ils0.Spans().AppendEmpty())
	fillSpanTwo(rs0ils0.Spans().AppendEmpty())
	return td
}

func constructTracesNum(num int) ptrace.Traces {
	td := ptrace.NewTraces()
	rs0 := td.ResourceSpans().AppendEmpty()
	rs0ils0 := rs0.ScopeSpans().AppendEmpty()
	for i := 0; i < num; i++ {
		fillSpanOne(rs0ils0.Spans().AppendEmpty())
	}
	return td
}

func fillSpanOne(span ptrace.Span) {
	span.SetName("operationA")
	span.SetSpanID(spanID)
	span.SetParentSpanID(spanID2)
	span.SetTraceID(traceID)
	span.SetStartTimestamp(TestSpanStartTimestamp)
	span.SetEndTimestamp(TestSpanEndTimestamp)
	span.SetDroppedAttributesCount(1)
	span.SetDroppedLinksCount(1)
	span.SetDroppedEventsCount(1)
	span.SetKind(1)
	span.TraceStateStruct().FromRaw("new")
	span.Attributes().PutString("http.method", "get")
	span.Attributes().PutString("http.path", "/health")
	span.Attributes().PutString("http.url", "http://localhost/health")
	span.Attributes().PutString("flags", "A|B|C")
	status := span.Status()
	status.SetCode(ptrace.StatusCodeError)
	status.SetMessage("status-cancelled")
}

func fillSpanTwo(span ptrace.Span) {
	span.SetName("operationB")
	span.SetStartTimestamp(TestSpanStartTimestamp)
	span.SetEndTimestamp(TestSpanEndTimestamp)
	span.Attributes().PutString("http.method", "get")
	span.Attributes().PutString("http.path", "/health")
	span.Attributes().PutString("http.url", "http://localhost/health")
	span.Attributes().PutString("flags", "C|D")
	link0 := span.Links().AppendEmpty()
	link0.SetDroppedAttributesCount(4)
	link1 := span.Links().AppendEmpty()
	link1.SetDroppedAttributesCount(4)
	span.SetDroppedLinksCount(3)
	status := span.Status()
	status.SetCode(ptrace.StatusCodeError)
	status.SetMessage("status-cancelled")
}
