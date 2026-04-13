//  Copyright 2026 Walter Schulze
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package parse

import (
	"io"
	"reflect"
	"testing"

	"github.com/katydid/parser-go/expect"
	"github.com/katydid/parser-go/parse"
)

func TestSkipUnknownObjectOpen(t *testing.T) {
	input := struct{}{} // {}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.NoErr(t, p.Skip)
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestSkipUnknownObjectAfterOpen(t *testing.T) {
	input := struct{}{} // {}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.NoErr(t, p.Skip) // skip object close
	expect.EOF(t, p)
}

func TestSkipUnknownArrayOpen(t *testing.T) {
	input := []struct{}{} // []
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.NoErr(t, p.Skip)
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestSkipUnknownArrayAfterOpen(t *testing.T) {
	input := []struct{}{} // []
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.NoErr(t, p.Skip)
	expect.EOF(t, p)
}

func TestSkipSingletonArrayAfterOpen(t *testing.T) {
	input := []int{1} // [0:1]
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.NoErr(t, p.Skip)
	expect.EOF(t, p)
}

func TestSkipUnknownString(t *testing.T) {
	input := `"abc"` // "abc"
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.NoErr(t, p.Skip)
	expect.EOF(t, p)
}

// If the kind '[' was returned by Next, then the whole array is skipped.
func TestSkipArrayOpen(t *testing.T) {
	input := []int{1, 2} // [0:1, 1:2]
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.NoErr(t, p.Skip)
	// skipped over 0:1,1:2]
	expect.EOF(t, p)
}

func TestSkipArrayFirst(t *testing.T) {
	input := []int{1, 2} // [0:1, 1:2]
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.NoErr(t, p.Skip)
	// skipped over 1
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 1)
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 2)
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestSkipArrayNestedOpen(t *testing.T) {
	input := [][]int{[]int{1, 2}} // [0:[0:1,1:2]]
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)

	expect.Hint(t, p, parse.EnterHint)

	expect.NoErr(t, p.Skip)
	// skipped over 0:1,1:2]
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

// If an array element was parsed, then the rest of the array is skipped.
func TestSkipArrayElement(t *testing.T) {
	input := []int{1, 2, 3} // [0:1,1:2,2:3]
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 1)
	expect.NoErr(t, p.Skip)
	// skipped over 1:2,2:3]
	expect.EOF(t, p)
}

func TestSkipArrayNestedElement(t *testing.T) {
	input := []any{1, []int{2, 3, 4}, 5} // [0:1, 1:[0:2,1:3,2:4], 2:5]
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 1)

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 1)
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 2)

	expect.NoErr(t, p.Skip)
	// skipped over 1:3,2:4]

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 2)
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 5)

	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestSkipArrayRecursiveElement1(t *testing.T) {
	input := []any{1, []int{2, 3}, []any{[]int{4, 5, 6}}} // [0:1, 1:[0:2,1:3], 2:[0:[0:4,1:5,2:6]]]
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 1)
	expect.NoErr(t, p.Skip)
	// skipped over 1:[0:2,1:3], 2:[0:[0:4,1:5,2:6]]]
	expect.EOF(t, p)
}

func TestSkipArrayRecursiveElement2(t *testing.T) {
	input := []any{"a", []string{"b", "c"}, []any{[]string{"d", "e", "f"}}} // [0:1, 1:[0:2,1:3], 2:[0:[0:4,1:5,2:6]]]
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "a")
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 1)
	expect.NoErr(t, p.Skip) // skip over [0:"b",1:"c"]
	expect.NoErr(t, p.Skip) // skip over 2:[0:[0:"d",1:"e",2:"f"]]]
	expect.EOF(t, p)
}

func TestSkipArrayRecursiveElement3(t *testing.T) {
	input := []any{"a", []string{"b", "c"}, [][]string{[]string{"d", "e", "f"}}} // [0:"a", 1:[0:"b",1:"c"], 2:[0:[0:"d",1:"e",2:"f"]]]
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "a")
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 1)
	expect.NoErr(t, p.Skip) // skip over [0:"b",1:"c"]
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 2)
	expect.NoErr(t, p.Skip) // skip over [0:[0:"d",1:"e",2:"f"]]
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

// If the kind '{' was returned by Next, then the whole object is skipped.
func TestSkipObjectOpen(t *testing.T) {
	input := struct {
		a int
		b int
	}{a: 1, b: 2} // {"a": 1, "b": 2}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.NoErr(t, p.Skip)
	// skipped over "a":1,"b":2}
	expect.EOF(t, p)
}

func TestSkipObjectNestedOpen(t *testing.T) {
	input := struct {
		a struct {
			b int
			c int
		}
	}{a: struct {
		b int
		c int
	}{b: 1, c: 2}} // {"a":{"b":1,"c":2}}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "a")
	expect.Hint(t, p, parse.EnterHint)
	expect.NoErr(t, p.Skip)
	// skipped over "b":1,"c":2}
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

// If an object value was just parsed, then the rest of the object is skipped.
func TestSkipObjectKey(t *testing.T) {
	input := struct {
		a int
		b int
		c int
	}{a: 1, b: 2, c: 3} // {"a":1,"b":2,"c":3}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "a")
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 1)
	expect.NoErr(t, p.Skip)
	// skipped over "b":2,"c":3}
	expect.EOF(t, p)
}

func TestSkipObjectNestedKey(t *testing.T) {
	input := struct {
		a struct {
			b int
			c int
			d int
		}
	}{a: struct {
		b int
		c int
		d int
	}{b: 1, c: 2, d: 3}} // {"a": {"b":1,"c":2,"d":3}}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "a")
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "b")
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 1)
	expect.NoErr(t, p.Skip)
	// skipped over "c":2,"d":3}
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

// If a object key was just parsed, then that key's value is skipped.
func TestSkipObjectValue(t *testing.T) {
	input := struct {
		a int
		b int
	}{a: 1, b: 2} // {"a":1,"b":2}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "a")
	expect.NoErr(t, p.Skip)
	// skipped over 1
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "b")
	expect.NoErr(t, p.Skip)
	// skipped over 2
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestSkipObjectRecursiveValue(t *testing.T) {
	// {"a":1, "b": {"c": {"d": {"e": "f"}, "g": [0:1, 1:2]}}
	input := struct {
		a int
		b struct {
			c struct {
				d struct{ e string }
				g []int
			}
		}
	}{
		a: 1,
		b: struct {
			c struct {
				d struct{ e string }
				g []int
			}
		}{
			c: struct {
				d struct{ e string }
				g []int
			}{
				d: struct{ e string }{e: "f"},
				g: []int{1, 2},
			},
		},
	}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "a")
	expect.NoErr(t, p.Skip)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "b")
	expect.NoErr(t, p.Skip)
	// skipped over {"c": {"d": {"e": "f"}, "g": [0:1, 1:2]}
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestSkipObjectDeepRecursiveValue(t *testing.T) {
	// {"a":1, "b": {"c": {"d": {"e": "f"}, "g": [0:1, 1:2]}}
	input := struct {
		a int
		b struct {
			c struct {
				d struct{ e string }
				g []int
			}
		}
	}{
		a: 1,
		b: struct {
			c struct {
				d struct{ e string }
				g []int
			}
		}{
			c: struct {
				d struct{ e string }
				g []int
			}{
				d: struct{ e string }{e: "f"},
				g: []int{1, 2},
			},
		},
	}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "a")
	expect.NoErr(t, p.Skip)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "b")
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "c")
	expect.NoErr(t, p.Skip)
	// skipped over {"d":{"e":"f"},"g":[1,2]}
	expect.Hint(t, p, parse.LeaveHint)
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestSkipTagMixTwoObjectsWithIndexes(t *testing.T) {
	input := []map[string]bool{map[string]bool{"mykey1": true}, map[string]bool{"mykey2": false}} // [0: {"mykey1":true}, 1: {"mykey2":false}]
	// will be parsed the same as : [0: {"mykey1":true}, 1: {"mykey2":false}]
	p := NewParser()
	p.Init(reflect.ValueOf(input))

	// 1: first array
	expect.Hint(t, p, parse.EnterHint)

	// 2: first object
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	if err := p.Skip(); err != nil {
		t.Fatal(err)
	}

	// 2: second object
	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 1)
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "mykey2")
	expect.Hint(t, p, parse.ValueHint)
	expect.False(t, p)
	expect.Hint(t, p, parse.LeaveHint)

	// 1: first array
	expect.Hint(t, p, parse.LeaveHint)
	// in endState, at top of stack return EOF
	if _, err := p.Next(); err != io.EOF {
		t.Fatalf("expected EOF, but got %v", err)
	}
}

func TestSkipUpUpUpTwoFields(t *testing.T) {
	// {
	// 	"A": [
	// 		{"a": "b", "c": "d"},
	// 		"b",
	// 		"c"
	// 	],
	// 	"B": 1
	// }
	input := struct {
		A any
		B int
	}{A: []any{
		struct {
			a string
			c string
		}{a: "b", c: "d"},
		"b",
		"c",
	},
		B: 1,
	}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	// assertNoErr(t, p.Next)
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	// expect(t, p.String, "A")
	expect.String(t, p, "A")
	// p.Down()
	// assertNoErr(t, p.Next)
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	// expect(t, p.Int, 0)
	expect.Int(t, p, 0)

	// p.Down()
	// assertNoErr(t, p.Next)
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)

	// expect(t, p.String, "a")
	expect.String(t, p, "a")

	// p.Down()
	// assertNoErr(t, p.Next)
	expect.Hint(t, p, parse.ValueHint)

	// expect(t, p.String, "b")
	expect.String(t, p, "b")

	// p.Up()                 // back up to object
	// assertNoErr(t, p.Next) // go to "c"
	expect.Hint(t, p, parse.FieldHint)

	// p.Up()                 // skip over , `"c": "d"` and `}`,
	// assertNoErr(t, p.Next) // go to `1:b`
	expect.NoErr(t, p.Skip)
	expect.NoErr(t, p.Skip)
	expect.Hint(t, p, parse.FieldHint)

	// p.Up()                 // skip over `1:"b", 2:"c"` and `]`
	// assertNoErr(t, p.Next) // go to `"B":1`
	expect.NoErr(t, p.Skip) // skip over `1:"b"`
	expect.NoErr(t, p.Skip) // skip over `2:"c"]`
	expect.Hint(t, p, parse.FieldHint)

	// expect(t, p.String, "B")
	expect.String(t, p, "B")

	// expectEOF(t, p.Next)
	expect.NoErr(t, p.Skip) // skip over 1
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestSkipUpUpUpEndOfObject(t *testing.T) {
	// {
	// 	"A": [
	// 		{"a": "b"},
	// 		"b",
	// 		"c"
	// 	],
	// 	"B": 1
	// }
	input := struct {
		A any
		B int
	}{A: []any{
		struct{ a string }{a: "b"},
		"b",
		"c",
	},
		B: 1,
	}
	p := NewParser()
	p.Init(reflect.ValueOf(input))
	// assertNoErr(t, p.Next)
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	// expect(t, p.String, "A")
	expect.String(t, p, "A")
	// p.Down()
	// assertNoErr(t, p.Next)
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	// expect(t, p.Int, 0)
	expect.Int(t, p, 0)

	// p.Down()
	// assertNoErr(t, p.Next)
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)

	// expect(t, p.String, "a")
	expect.String(t, p, "a")

	// p.Down()
	// assertNoErr(t, p.Next)
	expect.Hint(t, p, parse.ValueHint)

	// expect(t, p.String, "b")
	expect.String(t, p, "b")

	// p.Up()                 // back up to object
	// expectEOF(t, p.Next)   // go to "}"
	expect.Hint(t, p, parse.LeaveHint)

	// p.Up()                 // skip over `}`,
	// assertNoErr(t, p.Next) // go to `1:b`
	expect.Hint(t, p, parse.FieldHint)

	// p.Up()                 // skip over `1:b, 2:c` and `]`
	// assertNoErr(t, p.Next) // go to `"B":1`
	expect.NoErr(t, p.Skip) // skip over `1:"b"`
	expect.NoErr(t, p.Skip) // skip over `2:"c"]`
	expect.Hint(t, p, parse.FieldHint)

	// expect(t, p.String, "B")
	expect.String(t, p, "B")

	// expectEOF(t, p.Next)
	expect.NoErr(t, p.Skip) // skip over 1
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}
