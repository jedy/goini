package goini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Node struct {
	data     string
	children map[string]Node
}

var durationType = reflect.TypeOf(time.Second)

func Load(data string) (*Node, error) {
	r := strings.NewReader(data)
	return read(r)
}

func LoadFile(path string) (*Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return read(file)
}

func LoadTo(data string, i interface{}) error {
	node, err := Load(data)
	if err != nil {
		return err
	}
	return node.Mapto(i)
}

func LoadFileTo(path string, i interface{}) error {
	node, err := LoadFile(path)
	if err != nil {
		return err
	}
	return node.Mapto(i)
}

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
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Map:
		for _, i := range v.MapKeys() {
			fmt.Fprintln(w, i, "=", v.MapIndex(i).Interface())
		}
	case reflect.Struct:
		if err := encodeStruct(w, v); err != nil {
			return nil
		}
	default:
		return fmt.Errorf("dump only support map and struct")
	}
	return nil
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

func encodeStruct(w io.Writer, v reflect.Value) error {
	t := v.Type()
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
		f := v.Field(i)
		switch s.Type.Kind() {
		case reflect.Bool,
			reflect.Int,
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
			fmt.Fprintln(w, name, "=", f.Interface())
		case reflect.Slice:
			switch s.Type.Elem().Kind() {
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
				fmt.Fprintf(w, "%s = ", name)
				for i := 0; i < f.Len(); i++ {
					if i > 0 {
						fmt.Fprint(w, ", ", f.Index(i).Interface())
					} else {
						fmt.Fprint(w, f.Index(i).Interface())
					}
				}
				fmt.Fprintln(w, "")
			default:
				return fmt.Errorf("not support %s", s.Type)
			}
		case reflect.Struct:
			fmt.Fprintln(w, "")
			fmt.Fprintf(w, "[%s]", name)
			fmt.Fprintln(w, "")
			err := encodeStruct(w, f)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("not support %s", s.Type)
		}
	}
	return nil
}

func read(r io.Reader) (*Node, error) {
	s := bufio.NewScanner(r)
	root := Node{"", make(map[string]Node)}
	cur := root
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		switch {
		case line == "":
			continue
		case line[0] == '#' || line[0] == ';':
			continue
		case line[0] == '[':
			length := len(line)
			if line[length-1] == ']' {
				section := strings.TrimSpace(line[1 : length-1])
				cur = Node{"", make(map[string]Node)}
				root.children[section] = cur
			} else {
				return nil, fmt.Errorf("[ should match with ], %s", line)
			}
		default:
			parts := strings.SplitN(line, "=", 2)
			if len(parts) < 2 {
				return nil, fmt.Errorf("only support key = value, %s", line)
			}
			key := strings.TrimSpace(parts[0])
			value := parts[1]
			if i := strings.IndexAny(value, ";#"); i != -1 {
				value = value[:i]
			}
			value = strings.TrimSpace(value)
			cur.children[key] = Node{value, nil}
		}
	}
	return &root, nil
}

func (n Node) Mapto(i interface{}) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("can't map to non-pointer value")
	}
	v = v.Elem()
	k := v.Kind()

	if v.Type() == durationType {
		t, err := time.ParseDuration(n.data)
		if err != nil {
			return nil
		}
		v.Set(reflect.ValueOf(t))
		return nil
	}

	var err error
	switch k {
	case reflect.Bool:
		var b bool
		b, err = n.Bool()
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var d int64
		d, err = strconv.ParseInt(n.data, 10, 64)
		if err == nil {
			if v.OverflowInt(d) {
				err = fmt.Errorf("%d overflow", d)
			}
		}
		v.SetInt(d)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var u uint64
		u, err = strconv.ParseUint(n.data, 10, 64)
		if err == nil {
			if v.OverflowUint(u) {
				err = fmt.Errorf("%d overflow", u)
			}
		}
		v.SetUint(u)
	case reflect.Float32, reflect.Float64:
		var f float64
		f, err = strconv.ParseFloat(n.data, 64)
		if err == nil {
			if v.OverflowFloat(f) {
				err = fmt.Errorf("%f overflow", f)
			}
		}
		v.SetFloat(f)
	case reflect.String:
		var s string
		s, err = n.Value()
		v.SetString(s)
	case reflect.Map:
		err = n.decodeMap(v)
	case reflect.Slice:
		err = n.decodeSlice(v)
	case reflect.Struct:
		err = n.decodeStruct(v)
	case reflect.Interface:
		v.Set(reflect.ValueOf(n))
	default:
		return fmt.Errorf("can't map to %s", v.Kind())
	}
	if err != nil {
		return err
	}
	return nil
}

func (n Node) getValueOfNode(t reflect.Type) (reflect.Value, error) {
	val := reflect.New(t)
	err := n.Mapto(val.Interface())
	if err != nil {
		return val, err
	}
	return val.Elem(), nil
}

func (n Node) decodeStruct(v reflect.Value) error {
	t := v.Type()
	r := regexp.MustCompile(`\s*,\s*`)
	for i := 0; i < t.NumField(); i++ {
		s := t.Field(i)
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		name := r.Split(s.Tag.Get("ini"), 2)[0]
		if name == "" {
			name = s.Name
		}
		if name == "-" {
			continue
		}
		node := n.Get(name)
		if node.IsEmpty() {
			continue
		}
		val, err := node.getValueOfNode(s.Type)
		if err != nil {
			return err
		}
		f.Set(val)
	}
	return nil
}

func (n Node) decodeMap(v reflect.Value) error {
	t := v.Type()
	if t.Key().Kind() != reflect.String {
		err := fmt.Errorf("can't map to map[%s]%s", t.Key(), t.Elem())
		return err
	}
	if n.children == nil {
		return nil
	}
	if v.IsNil() {
		v.Set(reflect.MakeMap(t))
	}
	for i := range n.children {
		val, err := n.children[i].getValueOfNode(t.Elem())
		if err != nil {
			return err
		}
		v.SetMapIndex(reflect.ValueOf(i), val)
	}
	return nil
}

func (n Node) decodeSlice(v reflect.Value) error {
	t := v.Type()
	var values []string
	values, err := n.Values()
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return nil
	}
	if v.IsNil() || v.Len() < len(values) {
		v.Set(reflect.MakeSlice(t, len(values), len(values)))
	}
	for i := range values {
		val, err := (Node{values[i], nil}).getValueOfNode(t.Elem())
		if err != nil {
			return err
		}
		v.Index(i).Set(val)
	}
	return nil
}

func (n Node) Get(k ...string) Node {
	for _, v := range k {
		if n.children == nil {
			return Node{}
		}
		n = n.children[v]
	}
	return n
}

func (n Node) IsEmpty() bool {
	return n.data == "" && n.children == nil
}

func (n Node) Int() (int, error) {
	return strconv.Atoi(n.data)
}

func (n Node) Float() (float64, error) {
	return strconv.ParseFloat(n.data, 64)
}

func (n Node) Value() (string, error) {
	if n.children != nil {
		return "", fmt.Errorf("can't convert map to scalar")
	}
	return n.data, nil
}

func (n Node) Bool() (bool, error) {
	return strconv.ParseBool(n.data)
}

func (n Node) Duration() (time.Duration, error) {
	return time.ParseDuration(n.data)
}

func (n Node) MustDuration(d time.Duration) time.Duration {
	v, e := n.Duration()
	if e != nil {
		return d
	}
	return v
}

func (n Node) MustInt(d int) int {
	v, e := n.Int()
	if e != nil {
		return d
	}
	return v
}

func (n Node) MustFloat(d float64) float64 {
	v, e := n.Float()
	if e != nil {
		return d
	}
	return v
}

func (n Node) MustValue(d string) string {
	v, e := n.Value()
	if e != nil {
		return d
	}
	return v
}

func (n Node) MustBool(d bool) bool {
	v, e := n.Bool()
	if e != nil {
		return d
	}
	return v
}

func (n Node) Values() ([]string, error) {
	if n.children != nil {
		return nil, fmt.Errorf("can't convert map to list")
	}
	if n.data == "" {
		return []string{}, nil
	}
	r := regexp.MustCompile(`\s*,\s*`)
	list := r.Split(n.data, -1)
	return list, nil
}

func (n Node) Ints() ([]int, error) {
	list, err := n.Values()
	if err != nil {
		return nil, err
	}
	l := make([]int, len(list))
	for i := 0; i < len(list); i++ {
		l[i], err = strconv.Atoi(list[i])
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}

func (n Node) Floats() ([]float64, error) {
	list, err := n.Values()
	if err != nil {
		return nil, err
	}
	l := make([]float64, len(list))
	for i := 0; i < len(list); i++ {
		l[i], err = strconv.ParseFloat(list[i], 64)
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}
