package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type MyData struct {
	Name   string `csv:"name"`
	Age    int    `csv:"age"`
	HasPet bool   `csv:"has_pet"`
}

const (
	TAG_PREFIX = "csv"
)

// Marshal maps all the structs in a slice into a slice of strings.
func Marshal(v any) ([][]string, error) {
	sliceVal := reflect.ValueOf(v)
	if sliceVal.Kind() != reflect.Slice {
		return nil, errors.New("input should be a slice of structs")
	}
	structType := sliceVal.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return nil, errors.New("input should be a slice of structs")
	}
	var out [][]string
	header := marshalHeader(structType)
	out = append(out, header)
	for i := 0; i < sliceVal.Len(); i++ {
		rowVal := sliceVal.Index(i)
		row, err := marshalOne(rowVal)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

// marshalHeader is used to put all the tag in the field of the struct into a slice of string.
func marshalHeader(t reflect.Type) []string {
	var out []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag, ok := field.Tag.Lookup(TAG_PREFIX); ok {
			out = append(out, tag)
		}
	}
	return out
}

// marshalOne is used to map one struct value, represented with reflect.Value into []string as one line.
func marshalOne(v reflect.Value) ([]string, error) {
	var out []string
	vt := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		if _, ok := vt.Field(i).Tag.Lookup(TAG_PREFIX); !ok {
			continue
		}
		// turn the reflect.Value -> string
		switch fieldVal.Kind() {
		case reflect.String:
			out = append(out, fieldVal.String())
		case reflect.Bool:
			out = append(out, strconv.FormatBool(fieldVal.Bool()))
		case reflect.Int:
			out = append(out, strconv.FormatInt(fieldVal.Int(), 10))
		default:
			return nil, fmt.Errorf("cannot handle field of kind %v", fieldVal.Kind())
		}
	}
	return out, nil
}

// Unmarshal maps all the rows of data in a slice of slice string into slice of structs
// First row is assumed to be the header with the column names
// Input v expected to the in the format of *[]struct
func Unmarshal(data [][]string, v any) error {
	sliceValPointer := reflect.ValueOf(v)
	if sliceValPointer.Kind() != reflect.Pointer {
		return errors.New("must be a pointer to a slice of structs")
	}
	// dereference the pointer
	sliceVal := sliceValPointer.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errors.New("must be a pointer to a slice of structs")
	}
	structType := sliceVal.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return errors.New("must be a pointer to a slice of structs")
	}
	header := data[0]
	namePos := make(map[string]int)
	for i, name := range header {
		namePos[name] = i
	}
	// assign new value to the type of structType
	// reflect.New creates a pointer to a scalar type,
	for _, row := range data[1:] {
		newVal := reflect.New(structType).Elem()
		err := unmarshalOne(row, namePos, newVal)
		if err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, newVal))
	}
	return nil
}

func unmarshalOne(row []string, namePos map[string]int, newVal reflect.Value) error {
	vt := newVal.Type()
	for i := 0; i < newVal.NumField(); i++ {
		typeField := vt.Field(i)
		field := newVal.Field(i)
		pos, ok := namePos[typeField.Tag.Get(TAG_PREFIX)]
		if !ok {
			continue
		}
		val := row[pos]
		switch field.Kind() {
		case reflect.String:
			field.SetString(val)
		case reflect.Int:
			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(i)
		case reflect.Bool:
			b, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			field.SetBool(b)
		default:
			return fmt.Errorf("cannot handle field of %v", field.Kind())
		}
	}
	return nil
}
func main() {
	data := `name,age,has_pet
Jon,"100",true
"Fred ""The Hammer"" Smith",42,false
Martha,37,"true"
`
	r := csv.NewReader(strings.NewReader(data))
	allData, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	var entries []MyData
	Unmarshal(allData, &entries)
	fmt.Println(entries)

	//now to turn entries into output
	out, err := Marshal(entries)
	if err != nil {
		panic(err)
	}
	sb := &strings.Builder{}
	w := csv.NewWriter(sb)
	w.WriteAll(out)
	fmt.Println(sb)
}
