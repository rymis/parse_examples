# Using the Go parse library.

## Abstract
This document contains several examples of usage of the github.com/rymis/parse library.

This library is lightweight but powerful PEG (Parsing Expression Grammars) parser implementation with
clean Go that uses reflection to define grammar. This method of grammar definition allows you to
construct AST (Abstract Syntax Tree) in the same time when you are contstructing grammar. It is fast
way to create DSL (Domain Specific Language), small parsers and your own languages.

Library code could be found at https://github.com/rymis/parse.

## Parsing Expression Grammars
This section describes PEG parsers. You can find more info in Wikipedia (https://en.wikipedia.org/wiki/PEG)
and in articles from external links.

Original description of PEG was published in the work of Bryan Ford (http://portal.acm.org/citation.cfm?id=964001.964011).
Here I will describe it in several sentences. Each grammar contains:
 * Finite set of terminal symbols T
 * Finite set of non-terminal symbols N
 * Set of rules P in form described bellow
 * A symbol S from T named start rule.

Each rule has form `S <- {expression}` where expression is:
 1. Empty expression --- parses empty string
 2. A terminal symbol t --- parses this symbol
 3. A non-terminal symbol e --- parses string using rule for s
 4. Sequence e<sub>1</sub> e<sub>2</sub> --- parses e<sub>1</sub> and then e<sub>2</sub>
 5. Ordered choice e<sub>1</sub> / e<sub>2</sub> --- tryes to parse e<sub>1</sub> and if failed tryes to parse e<sub>2</sub> at the same place.
 6. Zero or more e* --- parses e zero or more times.
 7. One or more e+ --- parses e one or more times.
 8. Optional e? --- parses e zero or one times.
 9. And predicate &e --- parses e but does not increase position.
 1. Not predicate !e --- parses e and returns error if parsed at position.

This rules could be simply compiled to parser. Let declare functions with following type:
``` Go
// Parse function.
// This function returns new location on success or -1 on error:
type Parser func (str []byte, location int) int
```
Now we can define functions with simple algorithm.

Empty expression will be always parsed:
``` Go
func parse_empty(str []byte, location int) int {
	return location
}
```

Terminal parsing function for terminal `t` is:
``` Go
func parse_t(str []byte, location int) int {
	if location < len(str) && str[location] == t {
		return location + 1
	} else {
		return -1
	}
}
```

Non-terminal parsing is function call.

Sequence parse function:
``` Go
func parse_sequence(str []byte, location int) int {
	l := parse_e1(str, location)
	if l < 0 {
		return -1
	}
	return parse_e2(str, location)
}
```

Ordered choice parsing function is simple:
``` Go
func parse_choice(str []byte, location int) int {
	l := parse_e1(str, location)
	if l >= 0 {
		return l
	}
	return parse_e2(str, location)
}
```

Zero or more, one or more and optional functions could be written in simple way:
``` Go
func parse_zero_or_more(str []byte, location int) int {
	var l int
	for {
		l = parse_e(str, location)
		if l < 0 {
			return location
		}
		location = l
	}
}

func parse_one_or_more(str []byte, location int) int {
	l := parse_e(str, location)
	if l < 0 {
		return -1
	}
	for {
		l = parse_e(str, location)
		if l < 0 {
			return location
		}
		location = l
	}
}

func parse_optional(str []byte, location int) int {
	l = parse_e(str, location)
	if l < 0 {
		return location
	}
	return l
}
```

And finally predicates:
``` Go
func parse_and(str []byte, location int) int {
	if parse_e(str, location) < 0 {
		return -1
	}
	return location
}

func parse_not(str []byte, location int) int {
	if parse_e(str, location) < 0 {
		return location
	}
	return -1
}
```

You can see that these functions look like recursive descent parser and it is the truth. But
it is not always good to write recursive descent parsers manually. Parsing expression grammars
allows you to use them simple way.

## Constructing parsers with parse library
This section contains information about parsers construction using parse library.

You do not need to define grammar separatly instead you must define Go types with special tags.
This allows you to construct AST (Abstract Syntax Tree) in the same time you are defining grammar.

### Parse function
Parse is main function of parse library. It is defined as:
``` Go
func Parse(result interface{}, str []byte, params *Options) (new_location int, err error)
```

First argument of this function is pointer to structure you want to parse. Second argument is string to parse
and third argument is optional options. Function returns length of parsed value. If it could not be parsed
function returns error. If error has type parse.Error it contains information about position of error and
message. Convertion of this error to string will give you informative error description.

Typical usage of this function looks like:
``` Go
var v MyType
l, err := parse.Parse(&v, ".....", nil)
```

Function could return length less then `len(str)` so if you need to parse full string you may use several methods.
First and most intuitive way is check if `l == len(str)`. Second method is to add end of file parser at the end of
grammar. You can see both ways in examples so I will not describe them here.

Following sections describe type construction for use with Parse function.

### Parsers for Go builtin types
This library contains parsers for most Go builtin types. Here I will describe them and show several examples of usage.

Using of builtin types parsers allows you to write your own configuration parsers faster.

#### Boolean values
Parser will parse to boolean values: `"true"` and `"false"`. Both this values must be last in string or followed by
not letter or number or `_`. So parser will return error if input string looks like: `trueOrFalseValue`.

``` Go
var b bool
l, err := parse.Parse(&b, []byte("true"), nil)
```

#### Integer values
Integer values parser uses Go syntax for integers plus optional sign. So integers could be decimal (`123`), octal (`0123`),
hexadecimal (`0x123`), positive(`10`) and negative (`-10`). Parser parses values and checks for overflow so you can not
parse `1000` with uint8 parser. For int32 values Go contains synonym rune and there are no way to determine is this int32
or rune so parser will try also to parse value in unicode letter form: `'c'`. Syntax of character is the same Go lexer uses.
``` Go
var i int
var u uint
l, err := parse.Parse(&i, []byte("-1000"), nil)
l, err := parse.Parse(&u, []byte("1000"), nil)
```

#### Floating point values
Floating point values parser uses Go float syntax. Actually strconv.ParseFloat function used to parse values after determining
is it floating number or not. Integer values is floating points too so you can parse for example `100`. Both float32 and float64
types are supported.

``` Go
var f float64
l, err := parse.Parse(&f, []byte("12.33333e-2"), nil)
```

#### Strings
If string field does not contain tag it will be parsed as Go string. Both `"string"` and `string` variants are supported.

``` Go
var s string
l, err := parse.Parse(&s, []byte(`"string"`), nil)
```

### Complex types
Parsing basic types is not enough to parse real data. To combine several types you need to use structures and slices.

#### Sequences
If you need to parse several tokens following one by one you can use struct. Each member of struct will be parsed
in string if name is not private and there are no `skip` tag set to `true`. If Options contains SkipWhite function
parser will skip white space before parsing next member. You can define additional parameters to parser for some
types and they will be parsed tag-specific. Before continue I will show simple example:

``` Go
type TwoNumbers struct {
	I int64
	F float64
}

var t TwoNumbers
_, err := parse.Parse(&t, []byte("10 -2.444e-1"), nil)
```

In this example parser will parse two numbers: interger `I` and floating point `F` and set this fields in structure. Example
works because default skip spaces function is used and this function will skip space chars.

All private fields will be skipped but anonymous fields. Anonymous fields allows you to add some specific elements
like braces, commas, ... Public members could be skipped if you'll add tag `parse:"skip"`.

``` Go
type Vector2 struct {
	// This is left brace literal:
	_ string `literal:"("`
	// X coordinate
	X float64
	// Comma
	_ string `literal:","`
	// Y coordinate
	Y float64
	_ string `literal:")"`
	// Next public field will be skipped (not parsed)
	Norm float64 `parse:"skip"`
	// Also skipped because private
	someField SomeType
}
```

You can find additional information about literals in following sections.

#### Choice
In PEG you can find ordered choice operation `/`. If you need to construct this type of parser you need to create structure
with first field of type `parse.FirstOf`. For example:

``` Go
type StringOrNumber struct {
	parse.FirstOf
	String string
	Number float64
}
...
var t StringOrNumber
_, err := parse.Parse(&t, []byte("10"), nil)
if err == nil {
	if t.Field == "String" { // String was parsed
	} else if t.Field == "Number" { // Number was parsed
	}
}
```

As showed in this example you can determine what field was parsed by checking `Field` field of structure. Parser will try to
parse fields one by one until find the good one. If no type could be parsed parser will try to determine what error was most
apropriate to this location. Actually it is error with maximum location.

#### Repetitions
To parse "zero or more" or "one or more" elements of one type you can use slices. For example:
``` Go
type Lists struct {
	L1 []Type1
	L2 []Type2 `parse:"+"`
}
```

Field `L1` is zero or more elements of `Type1`. If you set tag `parse` to field you can control is it zero or more (`*`) or
one or more (`+`) so field `L2` is one or more elements of `Type2`.

Also you could add `delimiter:"literal"` to tags and parser will parse list separated by delimiter.

#### Optional fields
In many cases field will be not of type `T` but of type `*T` --- pointer to `T`. In this case you can specify tag `parse:"?"`
and parse will set pointer to `nil` when can't parse it. For example:
``` Go
type T struct {
	Number    float64
	// Next field is optional:
	Comment  *string `parse:"?"`
	// But dot is not optional!
	Dot      *string `literal:"."`
}
...
var t T
_, err := parse.Parse(&t, str, nil)
if t.Comment != nil {
	...
}
```
When parser processes pointer to type it will pass all tags to the type parser itself so `Dot` will be parsed correctly.

### Strings parsing

#### Literal parsing
As was showed in section Sequences you can set tag `literal:"something"` to string fields of structure. In this case parser will
parse exactly this literal at position. It is useful for parsing specials, like braces or keywords. For example:
``` Go
type IfStatement struct {
	_ string `literal:"if"`
	_ string `literal:"("`
	Predicate Expression
	_ string `literal:")"`
	Code      Block
	Else     *struct {
		_ string `literal:"else"`
		Code Block
	} `parse:"?"`
}
```

#### Regular expression parsing
If string contains `regexp:"rx"` tag parser will try to parse string matched by this regular expression at position. This tag
could be used to parse identifiers, numbers as strings, ... For example:
``` Go
type KeyValue struct {
	// Key here is identifier in meaning of must languages
	Key string `regexp:"[a-zA-Z][a-zA-Z0-9_]*"`
	_   string `literal:":"`
	Val Value
}
```

You can define literal and regular expression in the same time. In this case regular expression will be used for parsing but
literal for output. It is useful when you are parsing anonymous field and parser can't get the value.

### NotAny and FollowedBy parsers
If field has got tag `parse:"!"` parser will try to parse this field and will return error if it is parsed. If value is not
parsed parser will continue to parse structure.

If field has got tag `parse:"&"` parser will try to parse this field and will return error if can't. After this parser
will continue to parse structure from the same location.

This attributes makes it possible to parse context-specific grammars in addition to context free grammars.

### Additional tags
Full tags information could be found in package documentation.

### Parse options
You can specify some options to parser. First you can define skip function: function for skipping white spaces. There are several
functions present in library with names like SkipSpaces. This functions could be combined using SkipAll function.

Parameter EnablePackrat allows you to enable or disable packrat parsing. By default this parameter is disabled but packrat
table will be used for left recursion detection. In real life it is resonably good solution.

Parameter Debug enables debug messages. It could be used to write call information for each parser call.

### Output functions
In addition to Parse function library contains Write and Append functions that allows you to serialize message. There are some
additional restrictions to serialization: all the anonymous fields of structure must have type string and contain tag `literal`.
This allows us to determine values of the anonymous fields and produce correct output. There are no guarantie that value written
with Write method will be parsed with Parse, but it is possible to make it correct if you want to.

## Examples of usage
This section contains several simple examples of usage in real applications.

### Calculator with parse library
This example could be used as Quick Start Guide. Sometimes it is very useful to have your own mathmatics expressions parsing capability.
I will try to show how to implement one fast and easy using parse library. Look at code:
``` Go
type Expression struct {
	parse.FirstOf
	Expr struct {
		First *Expression
		Op    string `regexp:"[-+]"`
		Second Production
	}
	Prod Production
}

type Production struct {
	parse.FirstOf
	Prod struct {
		First *Production
		Op    string `regexp:"[/%*]"`
		Second Atom
	}
	Atom Atom
}

type Atom struct {
	parse.FirstOf
	Expr struct {
		_ string `literal:"("`
		Expr *Expression
		_ string `literal:")"`
	}
	Number float64
}
```

In this code I have defined three types: Expression, Production and Atom. Expression is additive expression it could be
defined as PEG in form:
```
Expression <- Expression [+-] Production / Production
```
Definition as type looks very simmilar. As you can see this is left-recursive rule and it is very rare PEG compiler that
could parse one. It is one of the great disadvantages of PEG parsers because grammar
```
Expression <- Production ([+-] Production)*
```
doesn't give you pairs that you can calculate immediatly: you must use loops.  But this library is able to parse LR
grammars so you can write this simple rule and don't worry about calculations order.

And finally Atom is braced expression or floating point number. And now you want to add variables. Ok, I will change
only Atom:
``` Go
type Atom struct {
	parse.FirstOf
	Expr struct {
		_ string `literal:"("`
		Expr *Expression
		_ string `literal:")"`
	}
	Number float64
	// This is the typical identifier parser:
	Var    string `regexp:"[a-zA-Z][a-zA-Z0-9_]*"`
}
```

After this modification you could parse expressions like `x + 2.546 * y`. But of cause you'd like to have functions like
sin or cos. Ok, we will add functions too:
``` Go
type Atom struct {
	parse.FirstOf
	Expr struct {
		_ string `literal:"("`
		Expr *Expression
		_ string `literal:")"`
	}
	Number float64
	Func   FunctionCall
	Var    string `regexp:"[a-zA-Z][a-zA-Z0-9_]*"`
}
type FunctionCall struct {
	Name     string `regexp:"[a-zA-Z][a-zA-Z0-9_]*"`
	_        string `literal:"("`
	Args   []Expression `parse:"*" delimiter:","`
	_        string `literal:")"`
}
```
As you can see I have added FunctionCall type that allows you to parse function calls with zero or more
argumets delimited with ",". Classical PEG grammar for this call is
```
FunctionCall <- [a-zA-Z][a-zA-Z0-9_]* "(" (Expression ("," Expression)*)? ")"
```

And one moment: FunctionCall is placed before Var in Atom structure. It is not random place: PEG has only
ordinal choice operator so I try to place longer variants with the same prefix before shorter ones. Moreover
I recomend you to use something like this for keywords:
``` Go
type If struct {
	_      string `literal:"if"`
	_      string `regexp:"[^a-zA-Z0-9_]|$" parse:"&"`
}
```
This type will parse "if" only if it is followed by end of string or special symbol. This allows you to
simplify grammar in some cases.

Full code of calculator you can see in `calc` subdirectory.

### JSON parser
Of cause this example is not intendent for real life usage. If you need to parse JSON use `encoding/json`
library. But it is good example of parse library usage and I will write JSON parser here.

You can look to [json.org](http://json.org) and see grammar as diagrams. I will convert one into PEG:
```
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
```
Ok now we are ready to write code:
``` Go
type Object struct {
	_          string     `literal:"{"`
	Members  []Pair `repeat:"*" delimiter:","`
	_          string     `literal:"}"`
}
type Pair struct {
	Name     String
	_        string `literal:":"`
	Value    Value
}
type Array struct {
	_           string      `literal:"["`
	Elements  []Value       `repeat:"*" delimiter:","`
	_           string      `literal:"]"`
}
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
type String struct {
	// Go string could parse JSON string
	String string
}
type Number struct {
	Num   string   `regexp:"-?([1-9][0-9]*|0)(\\.[0-9]+)?([eE][-+]?[0-9]+)?"`
}
```
This code is simple PEG grammar translation into Go code. You can define a couple of functions:
``` Go
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
```
and you'll have easy to use (but very very slow) JSON parser.

### Configuration files parser
TODO

## Conclusion
This is cool library :)
TODO...

## External links

