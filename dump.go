package goini

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
)

func Dump(i interface{}) (string, error) {
	s := new(bytes.Buffer)
	err := write(s, i)
	if err != nil {
		return "", err
	}
	return s.String(), nil
}

func write(w io.Writer, i interface{}) error {
	v := reflect.ValueOf(i)
	return writeValue(w, v, 0)
}

func writeValue(w io.Writer, v reflect.Value, deep int) error {
	if deep > 2 {
		return fmt.Errorf("not support nested map or struct more than 2 level")
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Map:
		if err := encodeMap(w, v, deep+1); err != nil {
			return err
		}
	case reflect.Struct:
		if err := encodeStruct(w, v, deep+1); err != nil {
			return err
		}
	default:
		return fmt.Errorf("dump only support map and struct")
	}
	return nil
}

func isBaseKind(k reflect.Kind, includeBool bool) bool {
	switch k {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return true
	case reflect.Bool:
		return includeBool
	}
	return false
}

func parseTag(tag string) (string, []string) {
	splitter := regexp.MustCompile(`\s*,\s*`)
	commentSplitter := regexp.MustCompile(`\s*;\s*`)
	parts := splitter.Split(tag, 2)
	if len(parts) == 1 {
		return parts[0], nil
	}
	comments := commentSplitter.Split(parts[1], -1)
	return parts[0], comments
}

func writeSection(w io.Writer, name string, first bool) {
	if !first {
		fmt.Fprintln(w, "")
	}
	fmt.Fprintf(w, "[%s]", name)
	fmt.Fprintln(w, "")
}

func writeItem(w io.Writer, name string, v reflect.Value, deep int, first bool) (err error) {
	k := v.Type().Kind()
	if isBaseKind(k, true) {
		_, err = fmt.Fprintln(w, name, "=", v.Interface())
	} else if k == reflect.Slice {
		err = encodeSlice(w, name, v.Type().Elem().Kind(), v)
	} else {
		writeSection(w, name, first)
		err = writeValue(w, v, deep+1)
	}
	return
}

func encodeMap(w io.Writer, v reflect.Value, deep int) (err error) {
	for idx, i := range v.MapKeys() {
		vv := v.MapIndex(i)
		if vv.Type().Kind() == reflect.Interface {
			vv = vv.Elem()
		}
		err = writeItem(w, i.String(), vv, deep, idx == 0)
		if err != nil {
			return
		}
	}
	return
}

func encodeSlice(w io.Writer, name string, k reflect.Kind, v reflect.Value) error {
	if !isBaseKind(k, false) {
		return fmt.Errorf("not support %s", v.Type())
	}
	fmt.Fprintf(w, "%s = ", name)
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			fmt.Fprint(w, ", ", v.Index(i).Interface())
		} else {
			fmt.Fprint(w, v.Index(i).Interface())
		}
	}
	fmt.Fprintln(w, "")
	return nil
}

func encodeStruct(w io.Writer, v reflect.Value, deep int) error {
	t := v.Type()
	var err error
	for i := 0; i < t.NumField(); i++ {
		s := t.Field(i)
		name, comments := parseTag(s.Tag.Get("ini"))
		if name == "" {
			name = s.Name
		}
		if name == "-" || strings.ToLower(s.Name[:1]) == s.Name[:1] {
			continue
		}
		if comments != nil {
			for _, c := range comments {
				fmt.Fprintln(w, ";", c)
			}
		}
		err = writeItem(w, name, v.Field(i), deep, i == 0)
		if err != nil {
			return err
		}
	}
	return nil
}
