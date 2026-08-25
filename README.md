## parser-go-reflect

Reflection based parser for Go.

We can dynamically walk over reflecting Go `structs` using the [Parser interface](https://git.katydid.org.za/parser-go):

```go
import (
    "reflect"
    reflectparser "katydid.org.za/go/parser-go-reflect/reflect"
)

func main() {
	mystruct := &MyStruct{MyField: "myvalue", OtherField: "othervalue"}
    parser := reflectparser.NewParser()
	parser.Init(reflect.ValueOf(mystruct))
	myvalue, err := GetMyField(parser)
	if err != nil {
		panic(err)
	}
	println(myvalue)
}
```

We can then use the parser to decode only `MyField` and skip over other fields and return `"myvalue"`:

```go
import (
	"errors"
	"katydid.org.za/go/parser-go/cast"
	"katydid.org.za/go/parser-go/parse"
)

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
```