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

package ottl

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottltest"
)

// This is not in ottltest because it depends on a type that's a member of OTTL.
func Booleanp(b Boolean) *Boolean {
	return &b
}

func Test_parse(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		expected  *ParsedStatement
	}{
		{
			name:      "invocation with string",
			statement: `set("foo")`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							String: ottltest.Strp("foo"),
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "invocation with float",
			statement: `met(1.2)`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "met",
					Arguments: []Value{
						{
							Float: ottltest.Floatp(1.2),
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "invocation with int",
			statement: `fff(12)`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "fff",
					Arguments: []Value{
						{
							Int: ottltest.Intp(12),
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "complex invocation",
			statement: `set("foo", getSomething(bear.honey))`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							String: ottltest.Strp("foo"),
						},
						{
							Invocation: &Invocation{
								Function: "getSomething",
								Arguments: []Value{
									{
										Path: &Path{
											Fields: []Field{
												{
													Name: "bear",
												},
												{
													Name: "honey",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "complex path",
			statement: `set(foo.attributes["bar"].cat, "dog")`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name: "foo",
									},
									{
										Name:   "attributes",
										MapKey: ottltest.Strp("bar"),
									},
									{
										Name: "cat",
									},
								},
							},
						},
						{
							String: ottltest.Strp("dog"),
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "where == clause",
			statement: `set(foo.attributes["bar"].cat, "dog") where name == "fido"`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name: "foo",
									},
									{
										Name:   "attributes",
										MapKey: ottltest.Strp("bar"),
									},
									{
										Name: "cat",
									},
								},
							},
						},
						{
							String: ottltest.Strp("dog"),
						},
					},
				},
				WhereClause: &BooleanExpression{
					Left: &Term{
						Left: &BooleanValue{
							Comparison: &Comparison{
								Left: Value{
									Path: &Path{
										Fields: []Field{
											{
												Name: "name",
											},
										},
									},
								},
								Op: EQ,
								Right: Value{
									String: ottltest.Strp("fido"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:      "where != clause",
			statement: `set(foo.attributes["bar"].cat, "dog") where name != "fido"`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name: "foo",
									},
									{
										Name:   "attributes",
										MapKey: ottltest.Strp("bar"),
									},
									{
										Name: "cat",
									},
								},
							},
						},
						{
							String: ottltest.Strp("dog"),
						},
					},
				},
				WhereClause: &BooleanExpression{
					Left: &Term{
						Left: &BooleanValue{
							Comparison: &Comparison{
								Left: Value{
									Path: &Path{
										Fields: []Field{
											{
												Name: "name",
											},
										},
									},
								},
								Op: NE,
								Right: Value{
									String: ottltest.Strp("fido"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:      "ignore extra spaces",
			statement: `set  ( foo.attributes[ "bar"].cat,   "dog")   where name=="fido"`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name: "foo",
									},
									{
										Name:   "attributes",
										MapKey: ottltest.Strp("bar"),
									},
									{
										Name: "cat",
									},
								},
							},
						},
						{
							String: ottltest.Strp("dog"),
						},
					},
				},
				WhereClause: &BooleanExpression{
					Left: &Term{
						Left: &BooleanValue{
							Comparison: &Comparison{
								Left: Value{
									Path: &Path{
										Fields: []Field{
											{
												Name: "name",
											},
										},
									},
								},
								Op: EQ,
								Right: Value{
									String: ottltest.Strp("fido"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:      "handle quotes",
			statement: `set("fo\"o")`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							String: ottltest.Strp("fo\"o"),
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "Invocation with boolean false",
			statement: `convert_gauge_to_sum("cumulative", false)`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "convert_gauge_to_sum",
					Arguments: []Value{
						{
							String: ottltest.Strp("cumulative"),
						},
						{
							Bool: (*Boolean)(ottltest.Boolp(false)),
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "Invocation with boolean true",
			statement: `convert_gauge_to_sum("cumulative", true)`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "convert_gauge_to_sum",
					Arguments: []Value{
						{
							String: ottltest.Strp("cumulative"),
						},
						{
							Bool: (*Boolean)(ottltest.Boolp(true)),
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "Invocation with bytes",
			statement: `set(attributes["bytes"], 0x0102030405060708)`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name:   "attributes",
										MapKey: ottltest.Strp("bytes"),
									},
								},
							},
						},
						{
							Bytes: (*Bytes)(&[]byte{1, 2, 3, 4, 5, 6, 7, 8}),
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "Invocation with nil",
			statement: `set(attributes["test"], nil)`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name:   "attributes",
										MapKey: ottltest.Strp("test"),
									},
								},
							},
						},
						{
							IsNil: (*IsNil)(ottltest.Boolp(true)),
						},
					},
				},
				WhereClause: nil,
			},
		},
		{
			name:      "Invocation with Enum",
			statement: `set(attributes["test"], TEST_ENUM)`,
			expected: &ParsedStatement{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name:   "attributes",
										MapKey: ottltest.Strp("test"),
									},
								},
							},
						},
						{
							Enum: (*EnumSymbol)(ottltest.Strp("TEST_ENUM")),
						},
					},
				},
				WhereClause: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.statement, func(t *testing.T) {
			parsed, err := parseStatement(tt.statement)
			assert.NoError(t, err)
			assert.EqualValues(t, tt.expected, parsed)
		})
	}
}

func Test_parse_failure(t *testing.T) {
	tests := []string{
		`set(`,
		`set("foo)`,
		`set(name.)`,
		`("foo")`,
		`set("foo") where name =||= "fido"`,
		`set(span_id, SpanIDWrapper{not a hex string})`,
		`set(span_id, SpanIDWrapper{01})`,
		`set(span_id, SpanIDWrapper{010203040506070809})`,
		`set(trace_id, TraceIDWrapper{not a hex string})`,
		`set(trace_id, TraceIDWrapper{0102030405060708090a0b0c0d0e0f})`,
		`set(trace_id, TraceIDWrapper{0102030405060708090a0b0c0d0e0f1011})`,
		`set("foo") where name = "fido"`,
		`set("foo") where name or "fido"`,
		`set("foo") where name and "fido"`,
		`set("foo") where name and`,
		`set("foo") where name or`,
		`set("foo") where (`,
		`set("foo") where )`,
		`set("foo") where (name == "fido"))`,
		`set("foo") where ((name == "fido")`,
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			_, err := parseStatement(tt)
			assert.Error(t, err)
		})
	}
}

func testParsePath(val *Path) (GetSetter, error) {
	if val != nil && len(val.Fields) > 0 && val.Fields[0].Name == "name" {
		return &testGetSetter{
			getter: func(ctx TransformContext) interface{} {
				return ctx.GetItem()
			},
			setter: func(ctx TransformContext, val interface{}) {
				ctx.GetItem()
			},
		}, nil
	}
	return nil, fmt.Errorf("bad path %v", val)
}

// Helper for test cases where the WHERE clause is all that matters.
// Parse string should start with `set(name, "test") where`...
func setNameTest(b *BooleanExpression) *ParsedStatement {
	return &ParsedStatement{
		Invocation: Invocation{
			Function: "set",
			Arguments: []Value{
				{
					Path: &Path{
						Fields: []Field{
							{
								Name: "name",
							},
						},
					},
				},
				{
					String: ottltest.Strp("test"),
				},
			},
		},
		WhereClause: b,
	}
}

func Test_parseWhere(t *testing.T) {
	tests := []struct {
		statement string
		expected  *ParsedStatement
	}{
		{
			statement: `true`,
			expected: setNameTest(&BooleanExpression{
				Left: &Term{
					Left: &BooleanValue{
						ConstExpr: Booleanp(true),
					},
				},
			}),
		},
		{
			statement: `true and false`,
			expected: setNameTest(&BooleanExpression{
				Left: &Term{
					Left: &BooleanValue{
						ConstExpr: Booleanp(true),
					},
					Right: []*OpAndBooleanValue{
						{
							Operator: "and",
							Value: &BooleanValue{
								ConstExpr: Booleanp(false),
							},
						},
					},
				},
			}),
		},
		{
			statement: `true and true and false`,
			expected: setNameTest(&BooleanExpression{
				Left: &Term{
					Left: &BooleanValue{
						ConstExpr: Booleanp(true),
					},
					Right: []*OpAndBooleanValue{
						{
							Operator: "and",
							Value: &BooleanValue{
								ConstExpr: Booleanp(true),
							},
						},
						{
							Operator: "and",
							Value: &BooleanValue{
								ConstExpr: Booleanp(false),
							},
						},
					},
				},
			}),
		},
		{
			statement: `true or false`,
			expected: setNameTest(&BooleanExpression{
				Left: &Term{
					Left: &BooleanValue{
						ConstExpr: Booleanp(true),
					},
				},
				Right: []*OpOrTerm{
					{
						Operator: "or",
						Term: &Term{
							Left: &BooleanValue{
								ConstExpr: Booleanp(false),
							},
						},
					},
				},
			}),
		},
		{
			statement: `false and true or false`,
			expected: setNameTest(&BooleanExpression{
				Left: &Term{
					Left: &BooleanValue{
						ConstExpr: Booleanp(false),
					},
					Right: []*OpAndBooleanValue{
						{
							Operator: "and",
							Value: &BooleanValue{
								ConstExpr: Booleanp(true),
							},
						},
					},
				},
				Right: []*OpOrTerm{
					{
						Operator: "or",
						Term: &Term{
							Left: &BooleanValue{
								ConstExpr: Booleanp(false),
							},
						},
					},
				},
			}),
		},
		{
			statement: `(false and true) or false`,
			expected: setNameTest(&BooleanExpression{
				Left: &Term{
					Left: &BooleanValue{
						SubExpr: &BooleanExpression{
							Left: &Term{
								Left: &BooleanValue{
									ConstExpr: Booleanp(false),
								},
								Right: []*OpAndBooleanValue{
									{
										Operator: "and",
										Value: &BooleanValue{
											ConstExpr: Booleanp(true),
										},
									},
								},
							},
						},
					},
				},
				Right: []*OpOrTerm{
					{
						Operator: "or",
						Term: &Term{
							Left: &BooleanValue{
								ConstExpr: Booleanp(false),
							},
						},
					},
				},
			}),
		},
		{
			statement: `false and (true or false)`,
			expected: setNameTest(&BooleanExpression{
				Left: &Term{
					Left: &BooleanValue{
						ConstExpr: Booleanp(false),
					},
					Right: []*OpAndBooleanValue{
						{
							Operator: "and",
							Value: &BooleanValue{
								SubExpr: &BooleanExpression{
									Left: &Term{
										Left: &BooleanValue{
											ConstExpr: Booleanp(true),
										},
									},
									Right: []*OpOrTerm{
										{
											Operator: "or",
											Term: &Term{
												Left: &BooleanValue{
													ConstExpr: Booleanp(false),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}),
		},
		{
			statement: `name != "foo" and name != "bar"`,
			expected: setNameTest(&BooleanExpression{
				Left: &Term{
					Left: &BooleanValue{
						Comparison: &Comparison{
							Left: Value{
								Path: &Path{
									Fields: []Field{
										{
											Name: "name",
										},
									},
								},
							},
							Op: NE,
							Right: Value{
								String: ottltest.Strp("foo"),
							},
						},
					},
					Right: []*OpAndBooleanValue{
						{
							Operator: "and",
							Value: &BooleanValue{
								Comparison: &Comparison{
									Left: Value{
										Path: &Path{
											Fields: []Field{
												{
													Name: "name",
												},
											},
										},
									},
									Op: NE,
									Right: Value{
										String: ottltest.Strp("bar"),
									},
								},
							},
						},
					},
				},
			}),
		},
		{
			statement: `name == "foo" or name == "bar"`,
			expected: setNameTest(&BooleanExpression{
				Left: &Term{
					Left: &BooleanValue{
						Comparison: &Comparison{
							Left: Value{
								Path: &Path{
									Fields: []Field{
										{
											Name: "name",
										},
									},
								},
							},
							Op: EQ,
							Right: Value{
								String: ottltest.Strp("foo"),
							},
						},
					},
				},
				Right: []*OpOrTerm{
					{
						Operator: "or",
						Term: &Term{
							Left: &BooleanValue{
								Comparison: &Comparison{
									Left: Value{
										Path: &Path{
											Fields: []Field{
												{
													Name: "name",
												},
											},
										},
									},
									Op: EQ,
									Right: Value{
										String: ottltest.Strp("bar"),
									},
								},
							},
						},
					},
				},
			}),
		},
	}

	// create a test name that doesn't confuse vscode so we can rerun tests with one click
	pat := regexp.MustCompile("[^a-zA-Z0-9]+")
	for _, tt := range tests {
		name := pat.ReplaceAllString(tt.statement, "_")
		t.Run(name, func(t *testing.T) {
			statement := `set(name, "test") where ` + tt.statement
			parsed, err := parseStatement(statement)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}

var testSymbolTable = map[EnumSymbol]Enum{
	"TEST_ENUM":     0,
	"TEST_ENUM_ONE": 1,
	"TEST_ENUM_TWO": 2,
}

func testParseEnum(val *EnumSymbol) (*Enum, error) {
	if val != nil {
		if enum, ok := testSymbolTable[*val]; ok {
			return &enum, nil
		}
		return nil, fmt.Errorf("enum symbol not found")
	}
	return nil, fmt.Errorf("enum symbol not provided")
}

// This test doesn't validate parser results, simply checks whether the parse succeeds or not.
// It's a fast way to check a large range of possible syntaxes.
func Test_parseStatement(t *testing.T) {
	tests := []struct {
		statement string
		wantErr   bool
	}{
		{`set(foo.attributes["bar"].cat, "dog")`, false},
		{`set(foo.attributes["animal"], "dog") where animal == "cat"`, false},
		{`drop() where service == "pinger" or foo.attributes["endpoint"] == "/x/alive"`, false},
		{`drop() where service == "pinger" or foo.attributes["verb"] == "GET" and foo.attributes["endpoint"] == "/x/alive"`, false},
		{`drop() where animal > "cat"`, false},
		{`drop() where animal >= "cat"`, false},
		{`drop() where animal <= "cat"`, false},
		{`drop() where animal < "cat"`, false},
		{`drop() where animal =< "dog"`, true},
		{`drop() where animal => "dog"`, true},
		{`drop() where animal <> "dog"`, true},
		{`drop() where animal = "dog"`, true},
		{`drop() where animal`, true},
		{`drop() where animal ==`, true},
		{`drop() where ==`, true},
		{`drop() where == animal`, true},
		{`drop() where attributes["path"] == "/healthcheck"`, false},
	}
	pat := regexp.MustCompile("[^a-zA-Z0-9]+")
	for _, tt := range tests {
		name := pat.ReplaceAllString(tt.statement, "_")
		t.Run(name, func(t *testing.T) {
			_, err := parseStatement(tt.statement)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStatement(%s) error = %v, wantErr %v", tt.statement, err, tt.wantErr)
				return
			}
		})
	}
}
