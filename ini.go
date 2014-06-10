package goini

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

type Node struct {
	data     string
	children map[string]Node
}

var durationType = reflect.TypeOf(time.Second)

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
