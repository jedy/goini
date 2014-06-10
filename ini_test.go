package goini

import (
	"strings"
	"testing"
	"time"
)

const content = `
root1 = 1
root2 = me ss age
# comment
; comment
root 3 = true # comment
root4 = 10s ; comment

[section1]
Sec1 = 10.9
Sec2 = 1,2,3,4

[section2]
sec = message
sec2 = false
`

func TestRead(t *testing.T) {
	r := strings.NewReader(content)
	root, err := read(r)
	if err != nil {
		t.Error(err)
	}

	if len(root.children) != 6 {
		t.Errorf("%d", len(root.children))
	}

	if len(root.children["section1"].children) != 2 {
		t.Errorf("%d", len(root.children["section1"].children))
	}

	if root.children["root1"].data != "1" {
		t.Errorf("%s", root.children["root1"].data)
	}

	if root.children["root2"].data != "me ss age" {
		t.Errorf("%s", root.children["root2"].data)
	}

}

func TestParse(t *testing.T) {
	r := strings.NewReader(content)
	root, err := read(r)
	if err != nil {
		t.Error(err)
		return
	}

	n := root.Get("root1").MustInt(0)
	if n != 1 {
		t.Errorf("%d", n)
	}

	s, err := root.Get("root2").Value()
	if err != nil || s != "me ss age" {
		t.Error(err)
	}

	b, err := root.Get("root 3").Bool()
	if err != nil || b != true {
		t.Error(err)
	}

	d, err := root.Get("root4").Duration()
	if err != nil || d != time.Second*10 {
		t.Error(err)
	}

	ns, err := root.Get("section1").Get("Sec2").Ints()
	if err != nil || len(ns) != 4 || ns[0] != 1 {
		t.Error(err)
	}

	b, err = root.Get("section2", "sec2").Bool()
	if err != nil || b != false {
		t.Error(err)
	}

	_, err = root.Get("noexists").Int()
	if err == nil {
		t.Error("should be error")
	}
}

func TestMap(t *testing.T) {
	r := strings.NewReader(content)
	root, err := read(r)
	if err != nil {
		t.Error(err)
		return
	}

	var n uint16
	err = root.Get("root1").Mapto(&n)
	if err != nil || n != 1 {
		t.Error(err)
	}

	m := make(map[int]int)
	err = root.Mapto(&m)
	if err.Error() != "can't map to map[int]int" {
		t.Error(err)
	}

	m2 := make(map[string]string)
	err = root.Get("section2").Mapto(&m2)
	if err != nil || m2["sec2"] != "false" {
		t.Error(err)
	}

	var m3 map[string]interface{}
	err = root.Get("section2").Mapto(&m3)
	if err != nil {
		t.Error(err)
	}
	if node, ok := m3["sec2"].(Node); !ok || node.MustBool(true) != false {
		t.Errorf("%s", m3)
	}

	var l []int
	err = root.Get("section1", "Sec2").Mapto(&l)
	if err != nil {
		t.Error(err)
	}
	if len(l) != 4 || l[0] != 1 || l[3] != 4 {
		t.Error(l)
	}

	var s struct {
		Root1    int           `ini:"root1"`
		Root2    string        `ini:"root2"`
		Root3    bool          `ini:"root 3"`
		Root4    time.Duration `ini:"root4"`
		Section1 struct {
			Sec1 float64
			Sec2 []int
		} `ini:"section1"`
	}
	err = root.Mapto(&s)
	if err != nil {
		t.Error(err)
	}
	if s.Root1 != 1 ||
		s.Section1.Sec1 != 10.9 ||
		s.Section1.Sec2[0] != 1 ||
		s.Root4 != time.Second*10 {
		t.Error(s)
	}
}

func TestComment(t *testing.T) {
	a := struct {
		A int `ini:"a, comment"`
		B int `ini:"b, comment1; comment2"`
	}{}
	s, err := Dump(a)
	if err != nil {
		t.Error(err)
	}
	if s != `; comment
a = 0
; comment1
; comment2
b = 0
` {
		t.Error(s)
	}
}

func TestDumpMap(t *testing.T) {
	a := make(map[string]interface{})
	a["item1"] = 1
	a["item2"] = "test"
	a["item3"] = false
	s, err := Dump(a)
	if err != nil {
		t.Error(err)
	}
	d, err := Load(s)
	if err != nil {
		t.Error(err)
	}
	if d.Get("item1").MustInt(0) != 1 ||
		d.Get("item2").MustValue("") != "test" ||
		d.Get("item3").MustBool(true) != false {
		t.Error(s)
	}
}

func TestDumpStruct(t *testing.T) {
	type I1 struct {
		A int
		B float64
	}
	type I2 struct {
		A string
		B []string
	}
	type I struct {
		R1 int `ini:"r1, root item 1"`
		R2 string
		r3 float64
		R4 int `ini:"-"`
		R5 time.Duration
		S1 I1
		S2 I2 `ini:"section2"`
	}
	a := I{
		R1: 1,
		R2: "test",
		r3: 1.23,
		R4: 10,
		R5: time.Minute,
		S1: I1{
			A: 10,
			B: 20.1,
		},
		S2: I2{
			A: "hello",
			B: []string{"Tim", "Tom"},
		},
	}
	s, err := Dump(&a)
	if err != nil {
		t.Error(err)
	}
	n, err := Load(s)
	if err != nil {
		t.Error(err)
	}
	var b I
	err = n.Mapto(&b)
	if err != nil {
		t.Error(err)
	}
	if b.R1 != 1 ||
		b.R4 != 0 ||
		b.R5 != time.Minute ||
		b.S1.A != 10 ||
		len(b.S2.B) != 2 ||
		b.S2.B[0] != "Tim" ||
		b.S2.B[1] != "Tom" {
		t.Error(b)
	}
}
