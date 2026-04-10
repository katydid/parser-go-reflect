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
	"iter"
	"maps"
	"reflect"
	"testing"
	_ "unsafe"
)

func rangeMapChan(m map[string]any) func() (string, error) {
	c := make(chan string)
	go func() {
		for key := range m {
			c <- key
		}
		close(c)
	}()
	return func() (string, error) {
		k, ok := <-c
		if ok {
			return k, nil
		}
		return "", io.EOF
	}
}

func BenchmarkRangeMapChan(b *testing.B) {
	m := map[string]any{
		"a": 1,
		"b": 1,
		"c": 1,
		"d": 1,
		"e": 1,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := rangeMapChan(m)
		_, err := f()
		for err == nil {
			_, err = f()
		}
	}
}

func rangeMapIter(m map[string]any) func() (string, error) {
	next, stop := iter.Pull(maps.Keys(m))
	return func() (string, error) {
		v, ok := next()
		if !ok {
			stop()
			return "", io.EOF
		}
		return v, nil
	}
}

func BenchmarkRangeMapIter(b *testing.B) {
	m := map[string]any{
		"a": 1,
		"b": 1,
		"c": 1,
		"d": 1,
		"e": 1,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := rangeMapIter(m)
		_, err := f()
		for err == nil {
			_, err = f()
		}
	}
}

func rangeMapReflect(m map[string]any) func() (string, error) {
	r := reflect.ValueOf(m)
	mapIter := r.MapRange()
	return func() (string, error) {
		if !mapIter.Next() {
			return "", io.EOF
		}
		return mapIter.Key().Interface().(string), nil
	}
}

func BenchmarkRangeReflect(b *testing.B) {
	m := map[string]any{
		"a": 1,
		"b": 1,
		"c": 1,
		"d": 1,
		"e": 1,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := rangeMapReflect(m)
		_, err := f()
		for err == nil {
			_, err = f()
		}
	}
}
