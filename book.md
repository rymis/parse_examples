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

TODO: describe it

## Constructing parsers with parse library
This section contains information about parsers construction using parse library.

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
This library contains parsers for most Go buildin types. Here I will describe them and show several examples of usage.

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
like braces, commas, ... Public members could be skipped if you'll add tag `skip="true"`.

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
	Norm float64 `skip:"true"`
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
	L2 []Type2 `repeat:"+"`
}
```

Field `L1` is zero or more elements of `Type1`. If you set tag `repeat` to field you can control is it zero or more (`*`) or
one or more (`+`) so field `L2` is one or more elements of `Type2`.

#### Optional fields
In many cases field will be not of type `T` but of type `*T` --- pointer to `T`. In this case you can specify tag `optional:"true"`
and parse will set pointer to `nil` when can't parse it. For example:
``` Go
type T struct {
	Number    float64
	// Next field is optional:
	Comment  *string `optional:"true"`
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
	} `optional:"true"`
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
If field has got tag `not_any="true"` parser will try to parse this field and will return error if it is parsed. If value is not
parsed parser will continue to parse structure.

If field has got tag `followed_by="true"` parser will try to parse this field and will return error if can't. After this parser
will continue to parse structure from the same location.

This attributes makes it possible to parse context-specific grammars in addition to context free grammars.

### Additional tags
TODO

### Parse options
You can specify some options to parser. First you can define skip function: function for skipping white spaces. There are several
functions present in library with names like SkipSpaces. This functions could be combined using SkipAll function.

Parameter EnablePackrat allows you to enable or disable packrat parsing. By default this parameter is disabled but packrat
table will be used for left recursion detection.

Parameter Debug enables debug messages. It could be used to write call information for each parser call.

### Output functions
In addition to Parse function library contains Write and Append functions that allows you to serialize message. There are some
additional restrictions to serialization: all the anonymous fields of structure must have type string and contain tag `literal`.
This allows us to determine values of the anonymous fields and produce correct output. There are no guarantie that value written
with Write method will be parsed with Parse, but it is possible to make it correct if you want to.

## Examples of usage
This section contains several simple examples of usage in real applications.

### Calculator with parse library
TODO

### JSON parser
TODO

### Configuration files parser
TODO

## Conclusion
This is cool library :)
TODO...

## External links

