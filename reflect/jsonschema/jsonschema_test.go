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

package jsonschema

import (
	"reflect"
	"testing"

	reflectparser "github.com/katydid/parser-go-reflect/reflect/parse"
	"github.com/katydid/parser-go/expect"
	"github.com/katydid/parser-go/parse"
)

type TestStruct struct {
	A string
	B *int64
	C []string
	M map[string]int64
}

func ptrTo[A any](a A) *A {
	return &a
}

func TestJSONSChema(t *testing.T) {
	input := &TestStruct{A: "a", B: ptrTo(int64(1)), C: []string{"c0", "c1"}, M: map[string]int64{"m123": 123}}
	reflectParser := reflectparser.NewParser()
	p := NewJSONSchemaParser(reflectParser)
	reflectParser.Init(reflect.ValueOf(input))

	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.Tag(t, p, "object")
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "A")
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "a")

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "B")
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 1)

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "C")
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.Tag(t, p, "array")
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 0)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "c0")

	expect.Hint(t, p, parse.FieldHint)
	expect.Int(t, p, 1)
	expect.Hint(t, p, parse.ValueHint)
	expect.String(t, p, "c1")

	expect.Hint(t, p, parse.LeaveHint)
	expect.Hint(t, p, parse.LeaveHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "M")
	expect.Hint(t, p, parse.EnterHint)
	expect.Hint(t, p, parse.FieldHint)
	expect.Tag(t, p, "object")
	expect.Hint(t, p, parse.EnterHint)

	expect.Hint(t, p, parse.FieldHint)
	expect.String(t, p, "m123")
	expect.Hint(t, p, parse.ValueHint)
	expect.Int(t, p, 123)

	expect.Hint(t, p, parse.LeaveHint)
	expect.Hint(t, p, parse.LeaveHint)

	expect.Hint(t, p, parse.LeaveHint)
	expect.Hint(t, p, parse.LeaveHint)
	expect.EOF(t, p)
}
