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
	"io"
	"reflect"

	"github.com/katydid/parser-go/cast"
	"github.com/katydid/parser-go/parse"
)

type parser struct {
	options
	state
	original reflect.Value
	alloc    func(size int) []byte
	stack    []state
}

// Parser is a parser for a reflected go structure.
type Parser interface {
	parse.Parser
	//Init initialises the parser with a value of reflected go structure.
	Init(value reflect.Value)
	Reset() error
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
	return
}

func (p *parser) Reset() error {
	p.Init(p.original)
	return nil
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
	panic("unreachable")
}

func (p *parser) Next() (parse.Hint, error) {
	switch p.state.kind {
	case startState:
		p.state.kind = endState
		s := newState(p.original)
		switch s.fieldKind {
		case structKind:
			s.kind = enterStructState
		case sliceKind:
			s.kind = enterSliceState
		case mapKind:
			s.kind = enterMapState
		}
		p.down(s)
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
		s := newState(fieldValue)
		p.down(s)
		switch s.fieldKind {
		case structKind, sliceKind, mapKind:
			return parse.EnterHint, nil
		case valueKind:
			return parse.ValueHint, nil
		}
		panic("unreachable")
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
		fieldValue := p.state.parent.Index(p.field)
		p.state.kind = enterSliceState
		s := newState(fieldValue)
		p.down(s)
		switch s.fieldKind {
		case structKind, sliceKind, mapKind:
			return parse.EnterHint, nil
		case valueKind:
			return parse.ValueHint, nil
		}
		panic("unreachable")
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
		s := newState(fieldValue)
		p.down(s)
		switch s.fieldKind {
		case structKind, sliceKind, mapKind:
			return parse.EnterHint, nil
		case valueKind:
			return parse.ValueHint, nil
		}
		panic("unreachable")
	case valueState:
		if err := p.up(); err != nil {
			return parse.UnknownHint, err
		}
		// cheat and use one of the enter... states to call `nextField` and check if `up` should be called again.
		return p.Next()
	}
	panic("unreachable")
}

func (p *parser) getToken(val reflect.Value) (parse.Kind, []byte, error) {
	val = deref(val)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return parse.Int64Kind, cast.FromInt64(val.Int(), p.alloc), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return parse.Float64Kind, cast.FromFloat64(float64(val.Uint()), p.alloc), nil
	case reflect.Float32, reflect.Float64:
		return parse.Float64Kind, cast.FromFloat64(val.Float(), p.alloc), nil
	case reflect.String:
		return parse.StringKind, []byte(val.String()), nil
	case reflect.Bool:
		if val.Bool() {
			return parse.TrueKind, nil, nil
		}
		return parse.FalseKind, nil, nil
	}
	panic("unreachable")
}

func (p *parser) Token() (parse.Kind, []byte, error) {
	switch p.state.kind {
	case fieldStructState:
		fieldType := p.parent.Type().Field(p.field)
		return parse.StringKind, []byte(fieldType.Name), nil
	case fieldSliceState:
		return parse.Int64Kind, cast.FromInt64(int64(p.field), p.alloc), nil
	case fieldMapState:
		keyValue := p.mapIter.Key()
		return p.getToken(keyValue)
	case valueState:
		return p.getToken(p.value)
	}
	return parse.UnknownKind, nil, nil
}

func (p *parser) Skip() error {
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

func (p *parser) down(state state) {
	// Append the current state to the stack.
	p.stack = append(p.stack, p.state)
	// Create a new state.
	p.state = state
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
