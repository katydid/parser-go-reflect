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

type options struct {
	jsonNumber bool
}

type Option = func(o *options)

func newOptions(opts ...Option) options {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return *o
}

// WithJsonNumber assumes that this is a reflection of JSON that was unmarshaled into a dynamic map.
// When unmarshaling from JSON into a dynamic map integers become doubles, but we still want to validate them as integers.
// This option makes Ints and Uints also return Doubles.
func WithJsonNumber(o *options) {
	o.jsonNumber = true
}
