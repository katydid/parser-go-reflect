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

	goparse "github.com/katydid/parser-go/parse"
)

type TestStruct struct {
	A string
	B *int64
	C []string
	M map[string]int64
}

func BenchmarkWithRandomMapTestStruct(b *testing.B) {
	num := 1000
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ps := make([]Parser, num)
	for i := 0; i < num; i++ {
		s := randMap(r, reflect.TypeOf(make(map[string]*TestStruct)))
		ps[i] = NewParser()
		ps[i].Init(reflect.ValueOf(s))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := ps[i%num]
		p.Reset()
		walk(p)
	}
}

func walk(p goparse.Parser) error {
	for {
		_, err := p.Next()
		if err != nil && err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
