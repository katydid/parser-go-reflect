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
	"reflect"
	"testing"

	"katydid.org.za/go/parser-go/debug"
	"katydid.org.za/go/parser-go/log"

	"katydid.org.za/go/parser-go/example"
	"katydid.org.za/go/parser-go/rand"
)

// func TestDebug(t *testing.T) {
// 	p := NewParser()
// 	p.Init(reflect.ValueOf(example.Input))
// 	m, err := hedge.ParseInto(p)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !m.Equal(example.Output) {
// 		t.Fatalf("expected %#v but got %#v", example.Output, m)
// 	}
// }

func TestRandomDebug(t *testing.T) {
	p := NewParser()
	for i := 0; i < 10; i++ {
		p.Init(reflect.ValueOf(example.Input))
		l := log.WrapParser(p)
		err := debug.RandomWalk(l, rand.NewRand(), 10, 3)
		if err != nil {
			t.Fatal(err)
		}
	}
}
