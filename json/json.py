#!/usr/bin/env python

from pyparsing import *
import time

# The same parser with Python

p_True = Keyword("true")
p_False = Keyword("false")
p_Null = Keyword("null")
p_Number = Combine(Regex(r'-?([1-9][0-9]*|0)(\.[0-9]+)?([eE][-+]?[0-9]+)?'))
p_String = dblQuotedString.copy()
p_Value = Forward()
p_Array = Literal("[").suppress() + Optional(p_Value + ZeroOrMore(Literal(',').suppress() + p_Value)) + Literal("]").suppress()
p_Pair = Group(p_String + Literal(":").suppress() + p_Value)
p_Object = Literal("{").suppress() + Optional(p_Pair + ZeroOrMore(Literal(',').suppress() + p_Pair)) + Literal("}").suppress()
p_Value << (p_Object | p_Array | p_String | p_Number | p_True | p_False | p_Null)

p_True.setParseAction(lambda s, l, t: True)
p_False.setParseAction(lambda s, l, t: False)
p_Null.setParseAction(lambda s, l, t: [ None ])
p_Number.setParseAction(lambda s, l, t: float(t[0]))
# p_String.setParseAction(lambda s, l, t: eval(t[0])) # TODO: it is fucking bad
p_String.setParseAction(lambda s, l, t: t[0])
p_Array.setParseAction(lambda s, l, t: t[:] if len(t) > 0 else [[]])

def obj(s, l, t):
    if len(t) == 0:
        return [{}]
    r = {}
    for p in t:
        r[p[0]] = p[1]
    return r
p_Object.setParseAction(obj)


test_1 = """
{
    "test": 123,
    "obj": {
        "bool": false,
        "nil": null
    }, 
    "array": [
        1234.5435e-2,
        {
            "xxx": "yyy"
        },
        [],
        {}
    ]
}
"""

if __name__ == '__main__':
    print "TEST:"
    try:
        print p_Object.parseString(test_1, True)
    except ParseException, e:
        print e
        print e.markInputline()

    f = open("code.json", "rb")
    json = f.read()
    f.close()

    try:
        t1 = time.time()
        res = p_Object.parseString(json, True)
        t2 = time.time()
        print "%f seconds elapsed" % (t2 - t1)
    except ParseException, e:
        print e
        print e.markInputline()



