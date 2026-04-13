// Copyright 2026 Walter Schulze
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parse

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/katydid/parser-go/expect"
	"github.com/katydid/parser-go/parse"
)

type TestStruct struct {
	A string
	B *int64
	C []string
	M map[string]int64
}

func RandomTestStruct(r *rand.Rand) *TestStruct {
	s := &TestStruct{}
	s = random(r, s).(*TestStruct)
	return s
}

func checkTestStruct(t *testing.T, s *TestStruct, p Parser) {
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "A")
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, s.A)

	if s.B != nil {
		expect.Hint(t, p, parse.FieldHint)
		expect.String(t, p, "B")
		expect.Hint(t, p, parse.ValueHint)
		expect.Int(t, p, *s.B)
	}

	if len(s.C) > 0 {
		expect.Hint(t, p, parse.FieldHint)
		expect.String(t, p, "C")
		expect.Hint(t, p, parse.EnterHint)
		for i, c := range s.C {
			expect.Hint(t, p, parse.FieldHint)
			expect.Int(t, p, int64(i))
			expect.Hint(t, p, parse.ValueHint)
			expect.String(t, p, c)
		}
		expect.Hint(t, p, parse.LeaveHint)
	}

	if len(s.M) > 0 {
		expect.Hint(t, p, parse.FieldHint)
		expect.String(t, p, "M")
		checkMap(t, reflect.ValueOf(s.M), p)
	}
	expect.Hint(t, p, parse.LeaveHint)
}

func checkMap(t *testing.T, m reflect.Value, p Parser) {
	keys := m.MapKeys()
	ks := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		ks[i] = keys[i].String()
	}
	expect.Hint(t, p, parse.EnterHint)
	for i := 0; i < m.Len(); i++ {
		expect.Hint(t, p, parse.FieldHint)
		kind, value, err := p.Token()
		if err != nil {
			t.Fatal(err)
		}
		if kind != parse.StringKind {
			t.Fatalf("expected string kind, but got %v", kind)
		}
		k := string(value)
		if ok := contains(ks, k); !ok {
			t.Fatalf("expected key in map keys %v, but got %s, %s", ks, k, err)
		}
		if m.Type().Elem().Kind() == reflect.Int64 {
			expect.Hint(t, p, parse.ValueHint)
			mapvalue := m.MapIndex(reflect.ValueOf(k)).Interface().(int64)
			expect.Int(t, p, mapvalue)
		} else {
			checkTestStruct(t, m.MapIndex(reflect.ValueOf(k)).Interface().(*TestStruct), p)
		}
	}
	expect.Hint(t, p, parse.LeaveHint)
}

func contains(xs []string, x string) bool {
	for i := 0; i < len(xs); i++ {
		if xs[i] == x {
			return true
		}
	}
	return false
}

func testWithRandomStruct(t *testing.T, r *rand.Rand) {
	s := RandomTestStruct(r)
	p := NewParser()
	p.Init(reflect.ValueOf(s))
	checkTestStruct(t, s, p)
	expect.EOF(t, p)
}

func TestWithRandomStruct(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100; i++ {
		testWithRandomStruct(t, r)
	}
}

func testWithRandomMapInt64(t *testing.T, r *rand.Rand) {
	s := randMap(r, reflect.TypeOf(make(map[string]int64)))
	p := NewParser()
	p.Init(reflect.ValueOf(s))
	checkMap(t, reflect.ValueOf(s), p)
}

func TestWithRandomMapInt64(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100; i++ {
		testWithRandomMapInt64(t, r)
	}
}

func testWithRandomMapTestStruct(t *testing.T, r *rand.Rand) {
	s := randMap(r, reflect.TypeOf(make(map[string]*TestStruct)))
	p := NewParser()
	p.Init(reflect.ValueOf(s))
	checkMap(t, reflect.ValueOf(s), p)
}

func TestWithRandomMapTestStruct(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100; i++ {
		testWithRandomMapTestStruct(t, r)
	}
}

func TestSliceOfString2(t *testing.T) {
	m := []string{"a", "b"}
	p := NewParser()
	p.Init(reflect.ValueOf(m))
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "a")

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 1)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "b")

	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestSliceOfString3(t *testing.T) {
	m := []string{"a", "b", "c"}
	p := NewParser()
	p.Init(reflect.ValueOf(m))
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "a")

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 1)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "b")

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 2)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "c")

	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestStructOfStrings(t *testing.T) {
	m := struct {
		A string
		B string
	}{A: "a", B: "b"}
	p := NewParser()
	p.Init(reflect.ValueOf(m))
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "A")
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "a")

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "B")
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "b")

	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestStructOfStructOfStrings(t *testing.T) {
	m := struct {
		Z struct {
			A string
			B string
		}
	}{Z: struct {
		A string
		B string
	}{A: "a", B: "b"}}
	p := NewParser()
	p.Init(reflect.ValueOf(m))
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "Z")
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "A")
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "a")

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "B")
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "b")

	expect.Hint(t, p, parse.LeaveHint)

	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestStructOfSliceOfString2(t *testing.T) {
	m := struct{ A []string }{A: []string{"a", "b"}}
	p := NewParser()
	p.Init(reflect.ValueOf(m))
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "A")
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "a")

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 1)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "b")

	expect.Hint(t, p, parse.LeaveHint)

	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestMapOfInterface(t *testing.T) {
	m := map[string]any{"MenuPaperclip": []any{"a", "b", "c"}}
	p := NewParser()
	p.Init(reflect.ValueOf(m))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "MenuPaperclip")

	expect.Hint(t, p, parse.EnterHint)
	for i := 0; i < len(m["MenuPaperclip"].([]any)); i++ {
		expect.Hint(t, p, parse.FieldHint)
		expect.Int(t, p, int64(i))
		expect.Hint(t, p, parse.ValueHint)
		expect.String(t, p, m["MenuPaperclip"].([]any)[int64(i)].(string))
	}
	expect.Hint(t, p, parse.LeaveHint)

	expect.Hint(t, p, parse.LeaveHint)
}

func TestWasJSON(t *testing.T) {
	jsonStr := `{
		"Number": 456
	}`
	v := make(map[string]any)
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		t.Fatalf("err <%v> unmarshaling json from <%s>", err, jsonStr)
	}
	p := NewParser(WithJsonNumber)
	p.Init(reflect.ValueOf(v))
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "Number")
	expect.Hint(t, p, parse.ValueHint)
	expect.Float(t, p, 456)
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}

func TestListOfStructsWasJSON(t *testing.T) {
	jsonStr := `{
		"Name": "Robert",
		"Addresses": [
			{
				"Number": 456,
				"Street": "TheStreet"
			}
		],
		"Telephone": "0127897897"
	}`
	v := make(map[string]any)
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		t.Fatalf("err <%v> unmarshaling json from <%s>", err, jsonStr)
	}
	p := NewParser()
	p.Init(reflect.ValueOf(v))
	missing := map[string]struct{}{"Name": {}, "Addresses": {}, "Telephone": {}}
	expect.Hint(t, p, parse.EnterHint)
	h, looperr := p.Next()
	for looperr == nil && h != parse.LeaveHint {
		if h != parse.FieldHint {
			t.Fatalf("expected field, but got %v", h)
		}
		fieldKind, fieldName, err := p.Token()
		if err != nil {
			t.Fatal(err)
		}
		if fieldKind != parse.StringKind {
			t.Fatalf("expected string, but got %v", fieldKind)
		}
		switch string(fieldName) {
		case "Name":
			delete(missing, "Name")
			expect.Hint(t, p, parse.ValueHint)
			expect.String(t, p, "Robert")
		case "Addresses":
			delete(missing, "Addresses")
			expect.Hint(t, p, parse.EnterHint)
			expect.Hint(t, p, parse.FieldHint)
			expect.Int(t, p, 0)
			expect.Hint(t, p, parse.EnterHint)
			missing2 := map[string]struct{}{"Number": {}, "Street": {}}
			h2, looperr2 := p.Next()
			for looperr2 == nil && h2 != parse.LeaveHint {
				if h2 != parse.FieldHint {
					t.Fatalf("expected field, but got %v", h2)
				}
				fieldKind2, fieldName2, err := p.Token()
				if err != nil {
					t.Fatal(err)
				}
				if fieldKind2 != parse.StringKind {
					t.Fatalf("expected string, but got %v", fieldKind2)
				}
				switch string(fieldName2) {
				case "Number":
					delete(missing2, "Number")
					expect.Hint(t, p, parse.ValueHint)
					expect.Float(t, p, 456)
				case "Street":
					delete(missing2, "Street")
					expect.Hint(t, p, parse.ValueHint)
					expect.String(t, p, "TheStreet")
				}
				h2, looperr2 = p.Next()
			}
			if looperr2 != nil {
				t.Fatal(looperr2)
			}
			if len(missing2) > 0 {
				t.Fatalf("missing field %v", missing2)
			}
			expect.Hint(t, p, parse.LeaveHint)
		case "Telephone":
			delete(missing, "Telephone")
			expect.Hint(t, p, parse.ValueHint)
			expect.String(t, p, "0127897897")
		}
		h, looperr = p.Next()
	}
	if looperr != nil {
		t.Fatal(looperr)
	}
	if len(missing) > 0 {
		t.Fatalf("missing field %v", missing)
	}
	expect.EOF(t, p)
}
