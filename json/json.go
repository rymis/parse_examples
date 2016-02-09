package main

import (
	"github.com/rymis/parse"
	"fmt"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"time"
	"runtime/pprof"
	"log"
	"os"
)

/* Grammar based on images from json.org: */
/*
	object <- '{' members? '}'
	members <- pair (',' pair)*
	pair <- string ':' value
	array <- '[' elements? ']'
	elements <- value (',' value)*
	value <- string / object / array / number / 'true' / 'false' / 'null'
	string <- '"' char* '"'
	char <- ('\\' ["\\/bfnrt]) / '\\' [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F] / [^"\\]
	number <- int frac? exp?
	int <- '-'? [0-9]+       // Actually ([1-9][0-9]* / '0')
	frac <- '.' [0-9]+
	exp  <- [eE] '+-'? [0-9]+
*/

// object <- '{' members? '}'
type Object struct {
	_          string     `literal:"{"`
	Members  []Pair `repeat:"*" delimiter:","`
	_          string     `literal:"}"`
}

// pair <- string ':' value
type Pair struct {
	Name     String
	_        string `literal:":"`
	Value    Value
}

// array <- '[' elements? ']'
type Array struct {
	_           string      `literal:"["`
	Elements  []Value       `repeat:"*" delimiter:","`
	_           string      `literal:"]"`
}

// value <- string / object / array / number / 'true' / 'false' / 'null'
type Value struct {
	parse.FirstOf
	Object Object
	Array  Array
	String String
	Number Number
	True   string `literal:"true"`
	False  string `literal:"false"`
	Null   string `literal:"null"`
}

// string <- '"' char* '"'
// char <- ('\\' ["\\/bfnrt]) / '\\' [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F] / [^"\\]
type String struct {
	String string // TODO: Go string parses more then JSON string expect
}

// number <- int frac? exp?
// int <- '-'? [0-9]+       // Actually ([1-9][0-9]* / '0')
// frac <- '.' [0-9]+
// exp  <- [eE] '+-'? [0-9]+
type Number struct {
	Num   string   `regexp:"-?([1-9][0-9]*|0)(\\.[0-9]+)?([eE][-+]?[0-9]+)?"`
}

// Converters:
func (self *Object) Map() map[string]interface{} {
	res := make(map[string]interface{})

	for _, pair := range(self.Members) {
		res[pair.Name.String] = pair.Value.Value()
	}

	return res
}

func (self *Value) Value() interface{} {
	switch self.Field {
	case "String":
		return self.String.String
	case "Object":
		return self.Object.Map()
	case "Array":
		return self.Array.Array()
	case "Number":
		return self.Number.Number()
	case "True":
		return true
	case "False":
		return false
	case "Null":
		return nil
	}

	panic("Invalid Field in FirstOf")
}

func (self *Number) Number() float64 {
	f, _ := strconv.ParseFloat(self.Num, 64)
	return f
}

func (self *Array) Array() []interface{} {
	res := make([]interface{}, 0)
	for _, v := range(self.Elements) {
		res = append(res, v.Value())
	}

	return res
}

func ParseJSON(json []byte) (res map[string]interface{}, err error) {
	var obj Object
	_, err = parse.Parse(&obj, json, &parse.Options{ PackratEnabled: false, SkipWhite: parse.SkipSpaces })
	if err != nil {
		return
	}

	return obj.Map(), nil
}

func main() {
	res, err := ParseJSON([]byte(`  {
		"test": 123,
		"obj": {
			"bool": false,
			"nil": null
		},
		"array": [
			1234.5435e-2,
			{
				"xxx": "yyy"
			}
		]
}`))
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	v, err := json.Marshal(res)
	println(string(v))

	if true { // code.json from Go distribution
		data, err := ioutil.ReadFile("code.json")
		if err != nil {
			fmt.Printf("Error: %v", err)
			return
		}

		f, err := os.Create("json.prof")
		if err != nil {
			log.Fatal(err)
		}

		tm := time.Now()
		pprof.StartCPUProfile(f)
		res, err = ParseJSON(data)
		pprof.StopCPUProfile()
		if err != nil {
			fmt.Printf("Error: %v", err)
			return
		}
		d := time.Since(tm)

		println(res)
		println("Parsed in ", d.Nanoseconds(), " nanosecs")

		j2 := make(map[string]interface{})
		tm = time.Now()
		json.Unmarshal(data, &j2)
		d2 := time.Since(tm)
		println("Parsed in ", d2.Nanoseconds(), " nanosecs")

		println(float64(d.Nanoseconds()) / float64(d2.Nanoseconds()), " times faster")
	}
}

