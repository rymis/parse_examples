package main

// This version of JSON parser uses Set tag to create map[string]interface{}

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
	"flag"
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
	_        []Pair `repeat:"*" delimiter:"," set:"SetValues"`
	_          string     `literal:"}"`
	// It is private field so I don't need to specify any tags
	values     map[string]interface{}
}

func (self *Object) SetValues(pairs []Pair) error {
	self.values = make(map[string]interface{})
	for _, p := range(pairs) {
		// TODO: check if value is unique?
		self.values[p.Name.String] = p.Value.Value()
	}
	return nil
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
	_         []Value       `repeat:"*" delimiter:"," set:"SetValues"`
	_           string      `literal:"]"`
	values    []interface{}
}

func (self *Array) SetValues(vals []Value) error {
	self.values = make([]interface{}, 0, len(vals))
	for _, v := range(vals) {
		self.values = append(self.values, v.Value())
	}

	return nil
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
	_     string   `regexp:"-?([1-9][0-9]*|0)(\\.[0-9]+)?([eE][-+]?[0-9]+)?" set:"SetValue"`
	value float64
}

func (self *Number) SetValue(s string) error {
	self.value, _ = strconv.ParseFloat(s, 64)
	return nil
}

// Converters:
func (self *Value) Value() interface{} {
	switch self.Field {
	case "String":
		return self.String.String
	case "Object":
		return self.Object.values
	case "Array":
		return self.Array.values
	case "Number":
		return self.Number.value
	case "True":
		return true
	case "False":
		return false
	case "Null":
		return nil
	}

	panic("Invalid Field in FirstOf")
}

func ParseJSON(json []byte) (res map[string]interface{}, err error) {
	var obj Object
	_, err = parse.Parse(&obj, json, &parse.Options{ PackratEnabled: false, SkipWhite: parse.SkipSpaces })
	if err != nil {
		return
	}

	return obj.values, nil
}

var profile = flag.Bool("profile", false, "Enable profiling")
func main() {
	flag.Parse()

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

		if *profile {
			println("Enabling profiler...")
			f, err := os.Create("json.prof")
			if err != nil {
				log.Fatal(err)
			}

			pprof.StartCPUProfile(f)
		}

		tm := time.Now()
		res, err = ParseJSON(data)
		d := time.Since(tm)

		if *profile {
			pprof.StopCPUProfile()
		}
		if err != nil {
			fmt.Printf("Error: %v", err)
			return
		}

		println(res)
		println("Parsed in ", d.Nanoseconds(), " nanosecs")

		j2 := make(map[string]interface{})
		tm = time.Now()
		json.Unmarshal(data, &j2)
		d2 := time.Since(tm)
		println("Parsed in ", d2.Nanoseconds(), " nanosecs")

		fmt.Printf("%2.6f times faster\n", float64(d.Nanoseconds()) / float64(d2.Nanoseconds()))
	}
}

