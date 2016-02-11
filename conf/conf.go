package main

// Simple configuration files parser.

import (
	"github.com/rymis/parse"
	"bytes"
	"fmt"
)

// Value could be string, boolean, integer or identifier:
type Value struct {
	parse.FirstOf
	Str      string
	Bool     bool
	Int      int64
	Id       string `regexp:"[-a-zA-Z0-9_]+"`
}

// KeyValue contains name and value pair.
type KeyValue struct {
	Key     string `regexp:"[a-zA-Z_][-a-zA-Z0-9_]*"`
	Value   Value
	_       string `literal:";"`
}

// KeyValueOrSection:
type KVOrSection struct {
	parse.FirstOf
	KV  KeyValue
	Section Section
}

// Section is named group of KeyValues or SubSections:
type Section struct {
	Name     string `regexp:"[a-zA-Z_][-a-zA-Z0-9_]*"`
	_        string `literal:"{"`
	Values []KVOrSection
	_        string `literal:"}"`
}

// Configuration is one or more sections:
type Configuration struct {
	Sections []Section `parse:"+"`
	// EOF could be parsed like here:
	_          string `regexp:".|\\n" parse:"!"`
}

func ParseConfiguration(data []byte) (*Configuration, error) {
	opts := &parse.Options{SkipWhite: func (str []byte, loc int) int {
		return parse.SkipAll(str, loc, parse.SkipShellComment, parse.SkipSpaces)
	}}
	cfg := new(Configuration)
	_, err := parse.Parse(cfg, data, opts)
	return cfg, err
}

func (self Value) String() string {
	switch self.Field {
	case "Str":
		return fmt.Sprintf("`%s`", self.Str)
	case "Int":
		return fmt.Sprintf("%d", self.Int)
	case "Bool":
		return fmt.Sprintf("%v", self.Bool)
	case "Id":
		return self.Id
	default:
		panic("Invalid value")
	}
}

func (self KeyValue) String() string {
	return self.Key + " " + self.Value.String()
}

func (self KVOrSection) String() string {
	if self.Field == "KV" {
		return self.KV.String()
	} else if self.Field == "Section" {
		return self.Section.String()
	} else {
		panic("Invalid selection")
	}
}

func (self Section) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%s {\n", self.Name)
	for i := range(self.Values) {
		fmt.Fprintf(buf, "\t%s\n", self.Values[i].String())
	}
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

func (self Configuration) String() string {
	buf := new(bytes.Buffer)
	for i := range(self.Sections) {
		fmt.Fprintf(buf, "%s", self.Sections[i].String())
	}
	return buf.String()
}

func main() {
	cfg, err := ParseConfiguration([]byte(`# This is test of configuration
section0 {
	string "string";
	flag    true;
	num     100;
	id      section1;
	innersection {
		name "You can use internal sections :)";
	}
}
section1 {
	xxx -1;
}
`))
	if err != nil {
		println(err.Error())
		return
	}

	println(cfg.String())
}

