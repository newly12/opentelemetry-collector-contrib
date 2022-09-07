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

package tqlotel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/telemetryquerylanguage/tql"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/telemetryquerylanguage/tql/tqltest"
)

func Test_truncateAll(t *testing.T) {
	input := pcommon.NewMap()
	input.UpsertString("test", "hello world")
	input.UpsertInt("test2", 3)
	input.UpsertBool("test3", true)

	target := &tql.StandardGetSetter{
		Getter: func(ctx tql.TransformContext) interface{} {
			return ctx.GetItem()
		},
		Setter: func(ctx tql.TransformContext, val interface{}) {
			ctx.GetItem().(pcommon.Map).Clear()
			val.(pcommon.Map).CopyTo(ctx.GetItem().(pcommon.Map))
		},
	}

	tests := []struct {
		name   string
		target tql.GetSetter
		limit  int64
		want   func(pcommon.Map)
	}{
		{
			name:   "truncate map",
			target: target,
			limit:  1,
			want: func(expectedMap pcommon.Map) {
				expectedMap.Clear()
				expectedMap.UpsertString("test", "h")
				expectedMap.UpsertInt("test2", 3)
				expectedMap.UpsertBool("test3", true)
			},
		},
		{
			name:   "truncate map to zero",
			target: target,
			limit:  0,
			want: func(expectedMap pcommon.Map) {
				expectedMap.Clear()
				expectedMap.UpsertString("test", "")
				expectedMap.UpsertInt("test2", 3)
				expectedMap.UpsertBool("test3", true)
			},
		},
		{
			name:   "truncate nothing",
			target: target,
			limit:  100,
			want: func(expectedMap pcommon.Map) {
				expectedMap.Clear()
				expectedMap.UpsertString("test", "hello world")
				expectedMap.UpsertInt("test2", 3)
				expectedMap.UpsertBool("test3", true)
			},
		},
		{
			name:   "truncate exact",
			target: target,
			limit:  11,
			want: func(expectedMap pcommon.Map) {
				expectedMap.Clear()
				expectedMap.UpsertString("test", "hello world")
				expectedMap.UpsertInt("test2", 3)
				expectedMap.UpsertBool("test3", true)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenarioMap := pcommon.NewMap()
			input.CopyTo(scenarioMap)

			ctx := tqltest.TestTransformContext{
				Item: scenarioMap,
			}

			exprFunc, _ := TruncateAll(tt.target, tt.limit)
			exprFunc(ctx)

			expected := pcommon.NewMap()
			tt.want(expected)

			assert.Equal(t, expected, scenarioMap)
		})
	}
}

func Test_truncateAll_validation(t *testing.T) {
	tests := []struct {
		name   string
		target tql.GetSetter
		limit  int64
	}{
		{
			name:   "limit less than zero",
			target: &tql.StandardGetSetter{},
			limit:  int64(-1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := TruncateAll(tt.target, tt.limit)
			assert.Error(t, err, "invalid limit for truncate_all function, -1 cannot be negative")
		})
	}
}

func Test_truncateAll_bad_input(t *testing.T) {
	input := pcommon.NewValueString("not a map")
	ctx := tqltest.TestTransformContext{
		Item: input,
	}

	target := &tql.StandardGetSetter{
		Getter: func(ctx tql.TransformContext) interface{} {
			return ctx.GetItem()
		},
		Setter: func(ctx tql.TransformContext, val interface{}) {
			t.Errorf("nothing should be set in this scenario")
		},
	}

	exprFunc, _ := TruncateAll(target, 1)
	exprFunc(ctx)

	assert.Equal(t, pcommon.NewValueString("not a map"), input)
}

func Test_truncateAll_get_nil(t *testing.T) {
	ctx := tqltest.TestTransformContext{
		Item: nil,
	}

	target := &tql.StandardGetSetter{
		Getter: func(ctx tql.TransformContext) interface{} {
			return ctx.GetItem()
		},
		Setter: func(ctx tql.TransformContext, val interface{}) {
			t.Errorf("nothing should be set in this scenario")
		},
	}

	exprFunc, _ := TruncateAll(target, 1)
	exprFunc(ctx)
}
