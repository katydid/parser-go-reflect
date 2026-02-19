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

package reflect

import (
	"io"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/katydid/parser-go/parser"
)

func BenchmarkWithRandomMapTestStruct(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ps := make([]ReflectParser, 1000)
	for i := 0; i < 1000; i++ {
		s := randMap(r, reflect.TypeOf(make(map[string]*TestStruct)))
		ps[i] = NewReflectParser()
		ps[i].Init(reflect.ValueOf(s))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := ps[i%1000]
		p.Reset()
		walk(p)
	}
}

func walk(p parser.Interface) error {
	for {
		if err := p.Next(); err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		if p.IsLeaf() {

		} else {
			p.Down()
			err := walk(p)
			if err != nil {
				return err
			}
			p.Up()
		}
	}
	return nil
}
