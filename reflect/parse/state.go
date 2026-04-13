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

import "reflect"

type stateKind rune

const startState = stateKind('S')

const enterStructState = stateKind('(')

const fieldStructState = stateKind('F')

const enterSliceState = stateKind('[')

const fieldSliceState = stateKind('I')

const enterMapState = stateKind('{')

const fieldMapState = stateKind('K')

const valueState = stateKind('V')

const endState = stateKind('$')

type state struct {
	kind      stateKind
	fieldKind fieldKind
	parent    reflect.Value
	value     reflect.Value
	field     int
	maxField  int
	mapIter   *reflect.MapIter
}

type fieldKind rune

const structKind = fieldKind('(')

const sliceKind = fieldKind('[')

const mapKind = fieldKind('{')

const valueKind = fieldKind('v')

func newState(val reflect.Value) state {
	value := deref(val)
	if value.Kind() == reflect.Struct {
		return state{
			kind:      enterStructState,
			parent:    value,
			fieldKind: structKind,
			field:     -1,
			maxField:  value.NumField(),
		}
	} else if isSlice(value) {
		return state{
			kind:      enterSliceState,
			parent:    value,
			fieldKind: sliceKind,
			field:     -1,
			maxField:  value.Len(),
		}
	} else if value.Kind() == reflect.Map {
		return state{
			kind:      enterMapState,
			parent:    value,
			fieldKind: mapKind,
			field:     -1,
			maxField:  value.Len(),
			mapIter:   value.MapRange(),
		}
	}
	return state{
		kind:      valueState,
		fieldKind: valueKind,
		value:     val,
		field:     -1,
		maxField:  1,
	}
}
