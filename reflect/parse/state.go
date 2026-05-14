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
)

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

func (p *parser) newState(val reflect.Value) {
	value := deref(val)
	if value.Kind() == reflect.Struct {
		p.state.kind = enterStructState
		p.state.parent = value
		p.state.fieldKind = structKind
		p.state.field = -1
		p.state.maxField = value.NumField()
	} else if isSlice(value) {
		p.state.kind = enterSliceState
		p.state.parent = value
		p.state.fieldKind = sliceKind
		p.state.field = -1
		p.state.maxField = value.Len()
	} else if value.Kind() == reflect.Map {
		p.state.kind = enterMapState
		p.state.parent = value
		p.state.fieldKind = mapKind
		p.state.field = -1
		p.state.maxField = value.Len()
		p.state.mapIter = value.MapRange()
	} else {
		p.state.kind = valueState
		p.state.fieldKind = valueKind
		p.state.value = val
		p.state.field = -1
		p.state.maxField = 1
	}
}
