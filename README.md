### goini

#### Feature

* create ini by dumping a struct
* parse ini file into a struct or map

#### example

```go
package main

import (
    "fmt"
    "time"

    "github.com/jedy/goini"
)

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

func main() {
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
    s, _ := goini.Dump(a)
    fmt.Println("----------- dump")
    fmt.Println(s)

    fmt.Println("----------- load")
    d, _ := goini.Load(s)
    fmt.Println(d.Get("r1").MustInt(0))

    var i I2
    d.Get("section2").Mapto(&i)
    fmt.Println(i)
}
```

output:

```
----------- dump
; root item 1
r1 = 1
R2 = test
R5 = 1m0s

[S1]
A = 10
B = 20.1

[section2]
A = hello
B = Tim, Tom

----------- load
1
{hello [Tim Tom]}
```

