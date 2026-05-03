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

package reflect_test

import (
	"errors"
	"reflect"
	"testing"

	reflectparser "github.com/katydid/parser-go-reflect/reflect"
	"github.com/katydid/parser-go/cast"
	"github.com/katydid/parser-go/parse"
)

type MyStruct struct {
	MyField    string
	OtherField string
}

func TestReadme(t *testing.T) {

	mystruct := &MyStruct{MyField: "myvalue", OtherField: "othervalue"}
	parser := reflectparser.NewParser()
	parser.Init(reflect.ValueOf(mystruct))
	myvalue, err := GetMyField(parser)
	if err != nil {
		panic(err)
	}
	println(myvalue)

	if myvalue != "myvalue" {
		t.Fatalf("not myvalue, but got %v", myvalue)
	}
}

func GetMyField(p parse.Parser) (string, error) {
	hint, err := p.Next()
	if hint != parse.EnterHint {
		return "", errors.New("expected object")
	}
	if err != nil {
		return "", err
	}
	for {
		hint, err = p.Next()
		if err != nil {
			return "", err
		}
		if hint != parse.FieldHint {
			return "", errors.New("expected field")
		}
		kind, fieldName, err := p.Token()
		if err != nil {
			return "", err
		}
		if kind != parse.StringKind {
			return "", errors.New("expected string")
		}
		if cast.ToString(fieldName) == "MyField" {
			hint, err = p.Next()
			if err != nil {
				return "", err
			}
			if hint != parse.ValueHint {
				return "", errors.New("expected field")
			}
			kind, val, err := p.Token()
			if err != nil {
				return "", err
			}
			if kind != parse.StringKind {
				return "", errors.New("expected string")
			}
			return cast.ToString(val), nil
		} else {
			p.Skip()
		}
	}
}
