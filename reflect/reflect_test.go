//  Copyright 2015 Walter Schulze
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

package reflect

import (
	"io"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/katydid/parser-go/parser/debug"
)

func TestDebug(t *testing.T) {
	p := NewReflectParser()
	p.Init(reflect.ValueOf(debug.Input))
	m, err := debug.Parse(p)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Equal(debug.Output) {
		t.Fatalf("expected %s but got %s", debug.Output, m)
	}
}

func TestRandomDebug(t *testing.T) {
	p := NewReflectParser()
	for i := 0; i < 10; i++ {
		p.Init(reflect.ValueOf(debug.Input))
		//l := debug.NewLogger(p, debug.NewLineLogger())
		err := debug.RandomWalk(p, debug.NewRand(), 10, 3)
		if err != nil {
			t.Fatal(err)
		}
		//t.Logf("original %v vs random %v", debug.Output, m)
	}
}

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

func checkTestStruct(t *testing.T, s *TestStruct, p ReflectParser) {
	if err := p.Next(); err != nil {
		t.Fatal(err)
	}
	if v, err := p.String(); err != nil || v != "A" {
		t.Fatalf("expected field A, but got %s, %s", v, err)
	}
	p.Down()
	if err := p.Next(); err != nil {
		t.Fatal(err)
	}
	if v, err := p.String(); err != nil || v != s.A {
		t.Fatalf("expected field %s, but got %s, %s", s.A, v, err)
	}
	p.Up()

	if s.B != nil {
		if err := p.Next(); err != nil {
			t.Fatal(err)
		}
		if v, err := p.String(); err != nil || v != "B" {
			t.Fatalf("expected field B, but got %s, %s", v, err)
		}
		p.Down()
		if err := p.Next(); err != nil {
			t.Fatal(err)
		}
		if v, err := p.Int(); err != nil || v != *s.B {
			t.Fatalf("expected field %d, but got %d, %s", *s.B, v, err)
		}
		p.Up()
	}

	if len(s.C) > 0 {
		if err := p.Next(); err != nil {
			t.Fatal(err)
		}
		if v, err := p.String(); err != nil || v != "C" {
			t.Fatalf("expected field C, but got %s, %s", v, err)
		}
		p.Down()
		for i, c := range s.C {
			if err := p.Next(); err != nil {
				t.Fatal(err)
			}
			if v, err := p.Int(); err != nil || v != int64(i) {
				t.Fatalf("expected index %d, but got %d, %s", i, v, err)
			}
			p.Down()
			if err := p.Next(); err != nil {
				t.Fatal(err)
			}
			if v, err := p.String(); err != nil || v != c {
				t.Fatalf("expected field %s, but got %s, %s", c, v, err)
			}
			p.Up()
		}
		p.Up()
	}

	if len(s.M) > 0 {
		if err := p.Next(); err != nil {
			t.Fatal(err)
		}
		if v, err := p.String(); err != nil || v != "M" {
			t.Fatalf("expected field M, but got %s, %s", v, err)
		}
		p.Down()
		checkMap(t, reflect.ValueOf(s.M), p)
		p.Up()
	}
}

func testWithRandomStruct(t *testing.T, r *rand.Rand) {
	s := RandomTestStruct(r)
	t.Logf("%#v", s)
	p := NewReflectParser()
	p.Init(reflect.ValueOf(s))
	checkTestStruct(t, s, p)
	if err := p.Next(); err != io.EOF {
		t.Fatal(err)
	}
}

func TestWithRandomStruct(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100; i++ {
		testWithRandomStruct(t, r)
	}
}

func contains(xs []string, x string) bool {
	for i := 0; i < len(xs); i++ {
		if xs[i] == x {
			return true
		}
	}
	return false
}

func checkMap(t *testing.T, m reflect.Value, p ReflectParser) {
	keys := m.MapKeys()
	ks := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		ks[i] = keys[i].String()
	}
	for i := 0; i < m.Len(); i++ {
		if err := p.Next(); err != nil {
			t.Fatal(err)
		}
		k, err := p.String()
		if err != nil {
			t.Fatal(err)
		}
		if ok := contains(ks, k); !ok {
			t.Fatalf("expected key in map keys %v, but got %s, %s", ks, k, err)
		}
		p.Down()
		if m.Type().Elem().Kind() == reflect.Int64 {
			if err := p.Next(); err != nil {
				t.Fatal(err)
			}
			mapvalue := m.MapIndex(reflect.ValueOf(k)).Interface().(int64)
			if v, err := p.Int(); err != nil || v != mapvalue {
				t.Fatalf("expected value %d, but got %d, %s", mapvalue, v, err)
			}
		} else {
			checkTestStruct(t, m.MapIndex(reflect.ValueOf(k)).Interface().(*TestStruct), p)
		}
		p.Up()
	}
}

func testWithRandomMapInt64(t *testing.T, r *rand.Rand) {
	s := randMap(r, reflect.TypeOf(make(map[string]int64)))
	p := NewReflectParser()
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
	p := NewReflectParser()
	p.Init(reflect.ValueOf(s))
	checkMap(t, reflect.ValueOf(s), p)
}

func TestWithRandomMapTestStruct(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100; i++ {
		testWithRandomMapTestStruct(t, r)
	}
}

func TestMapOfInterface(t *testing.T) {
	m := map[string]interface{}{"MenuPaperclip": []interface{}{"a", "b", "c"}}
	p := NewReflectParser()
	p.Init(reflect.ValueOf(m))
	if err := p.Next(); err != nil {
		t.Fatal(err)
	}
	if v, err := p.String(); err != nil || v != "MenuPaperclip" {
		t.Fatalf("expected field MenuPaperclip, but got %s, %s", v, err)
	}
	p.Down()
	for i := 0; i < len(m["MenuPaperclip"].([]interface{})); i++ {
		if err := p.Next(); err != nil {
			t.Fatal(err)
		}
		index, err := p.Int()
		if err != nil {
			t.Fatal(err)
		}
		if index != int64(i) {
			t.Fatalf("expected index %d, but got %d", i, index)
		}
		p.Down()
		if err := p.Next(); err != nil {
			t.Fatal(err)
		}
		value := m["MenuPaperclip"].([]interface{})[index].(string)
		if v, err := p.String(); err != nil || v != value {
			t.Fatalf("expected value %s, but got %s, %s", value, v, err)
		}
		p.Up()
	}
	p.Up()
	if err := p.Next(); err != io.EOF {
		t.Fatal(err)
	}
}
