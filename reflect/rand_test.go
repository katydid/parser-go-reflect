//  Copyright 2017 Walter Schulze
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

package reflect

import (
	"reflect"
	"strings"
)

type randy interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

// random returns a randomly populated pointer to a struct.
// This is a reflect rewrite of the deprecated code generator https://github.com/gogo/protobuf/blob/master/plugin/populate/populate.go
func random(r randy, i interface{}) interface{} {
	typ := reflect.TypeOf(i)
	return randStruct(r, typ)
}

func randStruct(r randy, typ reflect.Type) interface{} {
	if typ.Kind() != reflect.Ptr {
		panic("only pointers to structs are supported, but this is not a pointer")
	}
	typ = typ.Elem()
	if typ.Kind() != reflect.Struct {
		panic("only pointers to structs are supported, but this is not a struct")
	}
	strct := reflect.New(typ).Elem()
	for i := 0; i < strct.NumField(); i++ {
		fieldName := strct.Type().Field(i).Name
		if strings.HasPrefix(fieldName, "XXX_") {
			continue
		}
		if !strct.Type().Field(i).IsExported() {
			continue
		}
		fieldType := typ.Field(i).Type
		isPointer := fieldType.Kind() == reflect.Ptr
		// is a slice, but not a byte slice
		isSlice := fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() != reflect.Uint8
		isMap := fieldType.Kind() == reflect.Map
		if !isSlice && !isMap {
			if isPointer && r.Intn(5) == 0 {
				continue
			}
			if isPointer {
				fieldType = fieldType.Elem()
			}
			fieldKind := fieldType.Kind()
			switch fieldKind {
			case reflect.Struct:
				setValue(strct.Field(i), randStruct(r, typ.Field(i).Type))
			case reflect.Float64:
				setValue(strct.Field(i), randFloat64(r))
			case reflect.Float32:
				setValue(strct.Field(i), randFloat32(r))
			case reflect.Int64:
				setValue(strct.Field(i), randInt64(r))
			case reflect.Int32:
				isEnum := fieldType.String() != "int32"
				if isEnum {
					setValue(strct.Field(i), randEnum(r))
				} else {
					setValue(strct.Field(i), randInt32(r))
				}
			case reflect.Uint64:
				setValue(strct.Field(i), randUint64(r))
			case reflect.Uint32:
				setValue(strct.Field(i), randUint32(r))
			case reflect.Bool:
				setValue(strct.Field(i), randBool(r))
			case reflect.String:
				setValue(strct.Field(i), randString(r))
			case reflect.Slice:
				setValue(strct.Field(i), randBytes(r))
			default:
				panic("unsupported type: " + fieldType.String())
			}
		} else if isSlice {
			var size int
			if fieldType.Elem().Kind() == reflect.Ptr {
				// generate smaller lists for structs
				size = r.Intn(5)
			} else {
				size = r.Intn(10)
			}
			slice := reflect.MakeSlice(fieldType, size, size)
			elemType := fieldType.Elem()
			elemKind := elemType.Kind()
			for i := 0; i < size; i++ {
				item := slice.Index(i)
				switch elemKind {
				case reflect.Ptr:
					setValue(item, randStruct(r, elemType))
				case reflect.Float64:
					setValue(item, randFloat64(r))
				case reflect.Float32:
					setValue(item, randFloat32(r))
				case reflect.Int64:
					setValue(item, randInt64(r))
				case reflect.Int32:
					isEnum := elemType.String() != "int32"
					if isEnum {
						setValue(item, randEnum(r))
					}
					setValue(item, randInt32(r))
				case reflect.Uint64:
					setValue(item, randUint64(r))
				case reflect.Uint32:
					setValue(item, randUint32(r))
				case reflect.Bool:
					setValue(item, randBool(r))
				case reflect.String:
					setValue(item, randString(r))
				case reflect.Slice:
					setValue(item, randBytes(r))
				default:
					panic("unsupported type: " + elemType.String())
				}
			}
			setValue(strct.Field(i), slice.Interface())
		} else if isMap {
			setValue(strct.Field(i), randMap(r, fieldType))
		}
	}
	return strct.Addr().Interface()
}

func randMap(r randy, typ reflect.Type) any {
	size := r.Intn(10)
	mappy := reflect.MakeMapWithSize(typ, size)
	keyType := typ.Key()
	valueType := typ.Elem()
	for i := 0; i < size; i++ {
		key := randValue(r, keyType)
		value := randValue(r, valueType)
		mappy.SetMapIndex(key, value)
	}
	return mappy.Interface()
}

func randValue(r randy, typ reflect.Type) reflect.Value {
	switch typ.Kind() {
	case reflect.Ptr:
		return reflect.ValueOf(randStruct(r, typ))
	case reflect.Float64:
		return reflect.ValueOf(randFloat64(r))
	case reflect.Float32:
		return reflect.ValueOf(randFloat32(r))
	case reflect.Int64:
		return reflect.ValueOf(randInt64(r))
	case reflect.Int32:
		isEnum := typ.String() != "int32"
		if isEnum {
			return reflect.ValueOf(randEnum(r))
		}
		return reflect.ValueOf(randInt32(r))
	case reflect.Uint64:
		return reflect.ValueOf(randUint64(r))
	case reflect.Uint32:
		return reflect.ValueOf(randUint32(r))
	case reflect.Bool:
		return reflect.ValueOf(randBool(r))
	case reflect.String:
		return reflect.ValueOf(randString(r))
	case reflect.Slice:
		return reflect.ValueOf(randBytes(r))
	default:
		panic("unsupported type: " + typ.String())
	}
}

func setValue(assignToValue reflect.Value, value interface{}) {
	newValue := reflect.ValueOf(value)
	if newValue.Type().Kind() != reflect.Ptr && assignToValue.Type().Kind() == reflect.Ptr {
		assignToPtr := reflect.New(assignToValue.Type().Elem())
		assignToValue.Set(assignToPtr)
		assignToValue = assignToPtr.Elem()
	}
	if assignToValue.Type().AssignableTo(newValue.Type()) {
		assignToValue.Set(newValue)
		return
	}
	// assign enums
	assignToValue.SetInt(newValue.Int())
}

func randFloat64(r randy) float64 {
	v := r.Float64()
	if r.Intn(2) == 0 {
		return v * -1
	}
	return v
}

func randFloat32(r randy) float32 {
	v := r.Float32()
	if r.Intn(2) == 0 {
		return v * -1
	}
	return v
}

func randInt64(r randy) int64 {
	v := r.Int63()
	if r.Intn(2) == 0 {
		return v * -1
	}
	return v
}

func randUint64(r randy) uint64 {
	return uint64(r.Uint32())
}

func randInt32(r randy) int32 {
	v := r.Int31()
	if r.Intn(2) == 0 {
		return v * -1
	}
	return v
}

// This is an approximation, since we can't know the range of enum values, we guess that there are about 5.
func randEnum(r randy) int32 {
	return int32(r.Intn(5))
}

func randUint32(r randy) uint32 {
	return r.Uint32()
}

func randBool(r randy) bool {
	return bool(r.Intn(2) == 0)
}

func randBytes(r randy) []byte {
	count := r.Intn(100)
	bytes := make([]byte, count)
	for i := range bytes {
		bytes[i] = byte(r.Intn(256))
	}
	return bytes
}

func randString(r randy) string {
	v6 := r.Intn(10)
	tmps := make([]rune, v6)
	for i := 0; i < v6; i++ {
		tmps[i] = randUTF8Rune(r)
	}
	return string(tmps)
}

func randUTF8Rune(r randy) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
