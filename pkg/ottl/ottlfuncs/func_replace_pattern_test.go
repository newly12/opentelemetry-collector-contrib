// Copyright The OpenTelemetry Authors
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

package ottlfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottltest"
)

func Test_replacePattern(t *testing.T) {
	input := pcommon.NewValueString("application passwd=sensitivedtata otherarg=notsensitive key1 key2")

	target := &ottl.StandardGetSetter{
		Getter: func(ctx ottl.TransformContext) interface{} {
			return ctx.GetItem().(pcommon.Value).StringVal()
		},
		Setter: func(ctx ottl.TransformContext, val interface{}) {
			ctx.GetItem().(pcommon.Value).SetStringVal(val.(string))
		},
	}

	tests := []struct {
		name        string
		target      ottl.GetSetter
		pattern     string
		replacement string
		want        func(pcommon.Value)
	}{
		{
			name:        "replace regex match",
			target:      target,
			pattern:     `passwd\=[^\s]*(\s?)`,
			replacement: "passwd=*** ",
			want: func(expectedValue pcommon.Value) {
				expectedValue.SetStringVal("application passwd=*** otherarg=notsensitive key1 key2")
			},
		},
		{
			name:        "no regex match",
			target:      target,
			pattern:     `nomatch\=[^\s]*(\s?)`,
			replacement: "shouldnotbeinoutput",
			want: func(expectedValue pcommon.Value) {
				expectedValue.SetStringVal("application passwd=sensitivedtata otherarg=notsensitive key1 key2")
			},
		},
		{
			name:        "multiple regex match",
			target:      target,
			pattern:     `key[^\s]*(\s?)`,
			replacement: "**** ",
			want: func(expectedValue pcommon.Value) {
				expectedValue.SetStringVal("application passwd=sensitivedtata otherarg=notsensitive **** **** ")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenarioValue := pcommon.NewValueString(input.StringVal())

			ctx := ottltest.TestTransformContext{
				Item: scenarioValue,
			}

			exprFunc, _ := ReplacePattern(tt.target, tt.pattern, tt.replacement)
			exprFunc(ctx)

			expected := pcommon.NewValueString("")
			tt.want(expected)

			assert.Equal(t, expected, scenarioValue)
		})
	}
}

func Test_replacePattern_bad_input(t *testing.T) {
	input := pcommon.NewValueInt(1)
	ctx := ottltest.TestTransformContext{
		Item: input,
	}

	target := &ottl.StandardGetSetter{
		Getter: func(ctx ottl.TransformContext) interface{} {
			return ctx.GetItem()
		},
		Setter: func(ctx ottl.TransformContext, val interface{}) {
			t.Errorf("nothing should be set in this scenario")
		},
	}

	exprFunc, err := ReplacePattern(target, "regexp", "{replacement}")
	assert.Nil(t, err)
	exprFunc(ctx)

	assert.Equal(t, pcommon.NewValueInt(1), input)
}

func Test_replacePattern_get_nil(t *testing.T) {
	ctx := ottltest.TestTransformContext{
		Item: nil,
	}

	target := &ottl.StandardGetSetter{
		Getter: func(ctx ottl.TransformContext) interface{} {
			return ctx.GetItem()
		},
		Setter: func(ctx ottl.TransformContext, val interface{}) {
			t.Errorf("nothing should be set in this scenario")
		},
	}

	exprFunc, _ := ReplacePattern(target, `nomatch\=[^\s]*(\s?)`, "{anything}")
	exprFunc(ctx)
}

func Test_replacePatterns_invalid_pattern(t *testing.T) {
	target := &ottl.StandardGetSetter{
		Getter: func(ctx ottl.TransformContext) interface{} {
			t.Errorf("nothing should be received in this scenario")
			return nil
		},
		Setter: func(ctx ottl.TransformContext, val interface{}) {
			t.Errorf("nothing should be set in this scenario")
		},
	}

	invalidRegexPattern := "*"
	exprFunc, err := ReplacePattern(target, invalidRegexPattern, "{anything}")
	assert.Nil(t, exprFunc)
	assert.Contains(t, err.Error(), "error parsing regexp:")
}
