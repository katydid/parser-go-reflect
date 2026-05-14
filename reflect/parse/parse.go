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
	"fmt"
	"io"
	"reflect"

	"github.com/katydid/parser-go-json/json/jsonschema"
	"github.com/katydid/parser-go/cast"
	"github.com/katydid/parser-go/parse"
)

type parser struct {
	options
	state
	original reflect.Value
	alloc    func(size int) []byte
	stack    []state

	// cache tokens
	tokenKind parse.Kind
	tokenVal  []byte
}

// Parser is a parser for a reflected go structure.
type Parser interface {
	parse.Parser
	//Init initialises the parser with a value of reflected go structure.
	jsonschema.JSONSchemaAble
	Init(value reflect.Value)
	Reset()
}

// NewParser returns a new reflect parser.
func NewParser(options ...Option) Parser {
	return &parser{options: newOptions(options...), stack: make([]state, 0, 10)}
}

func (p *parser) Init(value reflect.Value) {
	p.state = state{}
	p.state.kind = startState
	p.original = value
	p.alloc = func(size int) []byte { return make([]byte, size) }
	p.stack = p.stack[:0]
	p.tokenKind = parse.UnknownKind
	p.tokenVal = nil
	return
}

func (p *parser) Reset() {
	p.Init(p.original)
	return
}

func (p *parser) nextField(fieldKind fieldKind) bool {
	switch fieldKind {
	case mapKind:
		p.field++
		return p.mapIter.Next()
	case structKind:
		p.field++
		if p.field >= p.maxField {
			return false
		}
		// skip over nil fields that are nil, contain empty slices or contain empty maps
		value := p.parent.Field(p.field)
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				return p.nextField(structKind)
			}
		} else if value.Kind() == reflect.Slice {
			if value.IsNil() || value.Len() == 0 {
				return p.nextField(structKind)
			}
		} else if value.Kind() == reflect.Map {
			if value.IsNil() || value.Len() == 0 {
				return p.nextField(structKind)
			}
		}
		return p.field < p.maxField
	case sliceKind:
		p.field++
		return p.field < p.maxField
	case valueKind:
		p.field++
		return false
	}
	panic(fmt.Sprintf("unreachable fieldKind %v", fieldKind))
}

func (p *parser) Next() (parse.Hint, error) {
	p.tokenKind = parse.UnknownKind
	p.tokenVal = nil
	switch p.state.kind {
	case startState:
		p.state.kind = endState
		p.down(p.original)
		switch p.state.fieldKind {
		case structKind:
			p.state.kind = enterStructState
		case sliceKind:
			p.state.kind = enterSliceState
		case mapKind:
			p.state.kind = enterMapState
		case valueKind:
			p.state.kind = valueState
			return parse.ValueHint, nil
		}
		return parse.EnterHint, nil
	case endState:
		return parse.UnknownHint, io.EOF
	case enterStructState:
		ok := p.nextField(structKind)
		if !ok {
			if err := p.up(); err != nil {
				return parse.UnknownHint, err
			}
			return parse.LeaveHint, nil
		}
		p.state.kind = fieldStructState
		return parse.FieldHint, nil
	case fieldStructState:
		fieldValue := p.parent.Field(p.field)
		p.state.kind = enterStructState
		p.down(fieldValue)
		switch p.state.fieldKind {
		case structKind, sliceKind, mapKind:
			return parse.EnterHint, nil
		case valueKind:
			return parse.ValueHint, nil
		}
		panic(fmt.Sprintf("unreachable fieldKind %v", p.state.fieldKind))
	case enterSliceState:
		ok := p.nextField(sliceKind)
		if !ok {
			if err := p.up(); err != nil {
				return parse.UnknownHint, err
			}
			return parse.LeaveHint, nil
		}
		p.state.kind = fieldSliceState
		return parse.FieldHint, nil
	case fieldSliceState:
		fieldValue := p.parent.Index(p.field)
		p.state.kind = enterSliceState
		p.down(fieldValue)
		switch p.state.fieldKind {
		case structKind, sliceKind, mapKind:
			return parse.EnterHint, nil
		case valueKind:
			return parse.ValueHint, nil
		}
		panic(fmt.Sprintf("unreachable fieldKind %v", p.state.fieldKind))
	case enterMapState:
		ok := p.nextField(mapKind)
		if !ok {
			if err := p.up(); err != nil {
				return parse.UnknownHint, err
			}
			return parse.LeaveHint, nil
		}
		p.state.kind = fieldMapState
		return parse.FieldHint, nil
	case fieldMapState:
		fieldValue := p.mapIter.Value()
		p.state.kind = enterMapState
		p.down(fieldValue)
		switch p.state.fieldKind {
		case structKind, sliceKind, mapKind:
			return parse.EnterHint, nil
		case valueKind:
			return parse.ValueHint, nil
		}
		panic(fmt.Sprintf("unreachable fieldKind %v", p.state.fieldKind))
	case valueState:
		if err := p.up(); err != nil {
			return parse.UnknownHint, err
		}
		// cheat and use one of the enter... states to call `nextField` and check if `up` should be called again.
		return p.Next()
	}
	panic(fmt.Sprintf("unreachable stateKind %v", p.state.kind))
}

func (p *parser) getToken(val reflect.Value) (parse.Kind, []byte, error) {
	val = deref(val)
	if val.Kind() == reflect.Invalid {
		return parse.NullKind, nil, nil
	}
	if val.CanInterface() {
		ival := val.Interface()
		switch x := ival.(type) {
		case json.Number:
			vint, err := x.Int64()
			if err == nil {
				return parse.Int64Kind, cast.FromInt64(vint, p.alloc), nil
			}
			vfloat, err := x.Float64()
			if err == nil {
				return parse.Float64Kind, cast.FromFloat64(vfloat, p.alloc), nil
			}
		}
	}
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return parse.Int64Kind, cast.FromInt64(val.Int(), p.alloc), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return parse.Float64Kind, cast.FromFloat64(float64(val.Uint()), p.alloc), nil
	case reflect.Float32, reflect.Float64:
		return parse.Float64Kind, cast.FromFloat64(val.Float(), p.alloc), nil
	case reflect.String:
		return parse.StringKind, cast.FromString(val.String(), p.alloc), nil
	case reflect.Bool:
		if val.Bool() {
			return parse.TrueKind, nil, nil
		}
		return parse.FalseKind, nil, nil
	}
	panic(fmt.Sprintf("unreachable val.Kind %v", val.Kind()))
}

func (p *parser) Token() (parse.Kind, []byte, error) {
	if p.tokenKind != parse.UnknownKind {
		return p.tokenKind, p.tokenVal, nil
	}
	switch p.state.kind {
	case fieldStructState:
		p.tokenKind = parse.StringKind
		fieldType := p.parent.Type().Field(p.field)
		p.tokenVal = cast.FromString(fieldType.Name, p.alloc)
	case fieldSliceState:
		p.tokenKind = parse.Int64Kind
		p.tokenVal = cast.FromInt64(int64(p.field), p.alloc)
	case fieldMapState:
		keyValue := p.mapIter.Key()
		tokenKind, tokenVal, err := p.getToken(keyValue)
		if err != nil {
			return tokenKind, tokenVal, err
		}
		p.tokenKind = tokenKind
		p.tokenVal = tokenVal
	case valueState:
		tokenKind, tokenVal, err := p.getToken(p.value)
		if err != nil {
			return tokenKind, tokenVal, err
		}
		p.tokenKind = tokenKind
		p.tokenVal = tokenVal
	default:
		return parse.UnknownKind, nil, nil
	}
	return p.tokenKind, p.tokenVal, nil
}

func (p *parser) JSONSchemaType() jsonschema.JSONSchemaType {
	switch p.state.kind {
	case enterStructState, enterMapState:
		return jsonschema.JSONSchemaTypeObject
	case enterSliceState:
		return jsonschema.JSONSchemaTypeArray
	}
	return jsonschema.JSONSchemaTypeUnknown
}

func (p *parser) Skip() error {
	p.tokenKind = parse.UnknownKind
	p.tokenVal = nil
	switch p.state.kind {
	case startState:
		_, err := p.Next()
		if err != nil {
			return err
		}
	case enterStructState:
		return p.up()
	case fieldStructState:
		p.state.kind = enterStructState
		return nil
	case enterSliceState:
		return p.up()
	case fieldSliceState:
		p.state.kind = enterSliceState
		return nil
	case enterMapState:
		return p.up()
	case fieldMapState:
		p.state.kind = enterMapState
		return nil
	case valueState:
		if err := p.up(); err != nil {
			return err
		}
		if len(p.stack) > 0 {
			return p.up()
		}
		return io.EOF
	case endState:
		_, err := p.Next()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) down(val reflect.Value) {
	// Append the current state to the stack.
	p.stack = append(p.stack, p.state)
	// Create a new state.
	p.newState(val)
}

func (p *parser) up() error {
	if len(p.stack) == 0 {
		return errUnexpectedClose
	}
	top := len(p.stack) - 1
	// Set the current state to the state on top of the stack.
	p.state = p.stack[top]
	// Remove the state on the top the stack from the stack,
	// but do it in a way that keeps the capacity,
	// so we can reuse it the next time down is called.
	p.stack = p.stack[:top]
	return nil
}

func deref(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	} else if v.Kind() == reflect.Interface {
		// get underlying type of interface
		return v.Elem()
	}
	return v
}

func isSlice(v reflect.Value) bool {
	return v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8
}
