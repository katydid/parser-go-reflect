//  Copyright 2013 Walter Schulze
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Package reflect contains an implementation of a parser for a reflected go structure.
package reflect

import (
	"reflect"

	"github.com/katydid/parser-go-json/json/tag"
	"github.com/katydid/parser-go-reflect/reflect/parse"
	goparse "github.com/katydid/parser-go/parse"
)

// Parser is a parser for a reflected go structure.
type Parser interface {
	goparse.Parser
	//Init initialises the parser with a value of reflected go structure.
	Init(value reflect.Value)
	Reset()
}

// NewParser returns a new reflect parser.
func NewParser(options ...Option) Parser {
	return parse.NewParser()
}

// NewJSONSchemaParser returns a new reflect parser that tags objects and arrays, so that the types can be checked by JSONSchema.
// The following json: `{"a": ["b", "c"]}`
// is parsed as: `{"object": {"a": {"array": {0: "b", 1: "c"}}}}`.
// The kind returned from the Token method for "object" and "array" will be parse.TagKind.
func NewJSONSchemaParser() Parser {
	parser := parse.NewParser()
	taggedParser := tag.NewTagger(parser, tag.WithTags())
	return &reflectParser{Parser: taggedParser, underlying: parser}
}

type reflectParser struct {
	tag.Parser
	underlying parse.Parser
}

func (r *reflectParser) Init(value reflect.Value) {
	r.Parser.Reset()
	r.underlying.Init(value)
}
