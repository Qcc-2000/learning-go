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
	Name   string `cvs:"name"`
	Age    int    `cvs:"age"`
	HasPet bool   `cvs:"has_pet"`
}

// Unmarshal maps all data in a slice of slice of strings into a slice of structs
func Unmarshal(data [][]string, v any) error {
	sliceValPointer := reflect.ValueOf(v)
	if sliceValPointer.Kind() != reflect.Pointer {
		return errors.New("must be a pointer to a slice of structs")
	}
	sliceVal := sliceValPointer.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errors.New("must be a pointer to a slice of structs")
	}
	structT := sliceVal.Type().Elem()
	if structT.Kind() != reflect.Struct {
		return errors.New("must be a pointer to a slice of structs")
	}
	header := data[0]
	namePos := make(map[string]int, len(header))
	for i, name := range header {
		namePos[name] = i
	}
	for _, row := range data[1:] {
		newVal := reflect.New(structT).Elem()
		err := unmarshalOne(row, namePos, newVal)
		if err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, newVal))
	}
	return nil
}
func unmarshalOne(row []string, namePos map[string]int, vv reflect.Value) error {
	vt := vv.Type()
	for i := 0; i < vv.NumField(); i++ {
		typeField := vt.Field(i)
		pos, ok := namePos[typeField.Tag.Get("csv")]
		if !ok {
			continue
		}
		val := row[pos]
		field := vv.Field(i)
		switch field.Kind() {
		case reflect.Int:
			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(i)
		case reflect.String:
			field.SetString(val)
		case reflect.Bool:
			b, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			field.SetBool(b)
		default:
			return fmt.Errorf("cannot handle field type %v", field.Kind())
		}
	}
	return nil
}

// Marshal map the slice of structs into a slice of slice of strings
func Marshal(v any) ([][]string, error) {
	sliceVal := reflect.ValueOf(v)
	if sliceVal.Kind() != reflect.Slice {
		return nil, errors.New("must be a slice")
	}
	structT := sliceVal.Type().Elem()
	if structT.Kind() != reflect.Struct {
		return nil, errors.New("must be a slice")
	}
	var out [][]string
	header := marshalHeader(structT)
	out = append(out, header)
	for i := 0; i < sliceVal.Len(); i++ {
		row, err := marshalOne(sliceVal.Index(i))
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}
func marshalHeader(vt reflect.Type) []string {
	var out []string
	for i := 0; i < vt.NumField(); i++ {
		field := vt.Field(i)
		if curTag, ok := field.Tag.Lookup("csv"); ok {
			out = append(out, curTag)
		}
	}
	return out
}
func marshalOne(vv reflect.Value) ([]string, error) {
	var out []string
	vt := vv.Type()
	for i := 0; i < vt.NumField(); i++ {
		fieldValue := vv.Field(i)
		if _, ok := vt.Field(i).Tag.Lookup("csv"); !ok {
			continue
		}
		switch fieldValue.Kind() {
		case reflect.Int:
			out = append(out, strconv.FormatInt(fieldValue.Int(), 10))
		case reflect.String:
			out = append(out, fieldValue.String())
		case reflect.Bool:
			out = append(out, strconv.FormatBool(fieldValue.Bool()))
		default:
			return nil, fmt.Errorf("cannot handle field of kind %v", fieldValue.Kind())
		}
	}
	return out, nil
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
	out, err := Marshal(entries)
	if err != nil {
		panic(err)
	}
	sb := &strings.Builder{}
	w := csv.NewWriter(sb)
	w.WriteAll(out)
	fmt.Println(sb)

}
