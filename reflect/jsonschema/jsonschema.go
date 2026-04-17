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

package jsonschema

import (
	"github.com/katydid/parser-go-json/json/tag"
	reflectparser "github.com/katydid/parser-go-reflect/reflect/parse"
	"github.com/katydid/parser-go/parse"
)

func NewJSONSchemaParser(reflectParser reflectparser.Parser) parse.Parser {
	taggedParser := tag.NewTagger(reflectParser, tag.WithTags())
	return taggedParser
}
