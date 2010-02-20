// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xml

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

const testInput = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN"
  "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<body xmlns:foo="ns1" xmlns="ns2" xmlns:tag="ns3" ` +
	"\r\n\t" + `  >
  <hello lang="en">World &lt;&gt;&apos;&quot; &#x767d;&#40300;翔</hello>
  <goodbye />
  <outer foo:attr="value" xmlns:tag="ns4">
    <inner/>
  </outer>
  <tag:name>
    <![CDATA[Some text here.]]>
  </tag:name>
</body><!-- missing final newline -->`

var rawTokens = []Token{
	CharData(strings.Bytes("\n")),
	ProcInst{"xml", strings.Bytes(`version="1.0" encoding="UTF-8"`)},
	CharData(strings.Bytes("\n")),
	Directive(strings.Bytes(`DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN"
  "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"`)),
	CharData(strings.Bytes("\n")),
	StartElement{Name{"", "body"}, []Attr{Attr{Name{"xmlns", "foo"}, "ns1"}, Attr{Name{"", "xmlns"}, "ns2"}, Attr{Name{"xmlns", "tag"}, "ns3"}}},
	CharData(strings.Bytes("\n  ")),
	StartElement{Name{"", "hello"}, []Attr{Attr{Name{"", "lang"}, "en"}}},
	CharData(strings.Bytes("World <>'\" 白鵬翔")),
	EndElement{Name{"", "hello"}},
	CharData(strings.Bytes("\n  ")),
	StartElement{Name{"", "goodbye"}, nil},
	EndElement{Name{"", "goodbye"}},
	CharData(strings.Bytes("\n  ")),
	StartElement{Name{"", "outer"}, []Attr{Attr{Name{"foo", "attr"}, "value"}, Attr{Name{"xmlns", "tag"}, "ns4"}}},
	CharData(strings.Bytes("\n    ")),
	StartElement{Name{"", "inner"}, nil},
	EndElement{Name{"", "inner"}},
	CharData(strings.Bytes("\n  ")),
	EndElement{Name{"", "outer"}},
	CharData(strings.Bytes("\n  ")),
	StartElement{Name{"tag", "name"}, nil},
	CharData(strings.Bytes("\n    ")),
	CharData(strings.Bytes("Some text here.")),
	CharData(strings.Bytes("\n  ")),
	EndElement{Name{"tag", "name"}},
	CharData(strings.Bytes("\n")),
	EndElement{Name{"", "body"}},
	Comment(strings.Bytes(" missing final newline ")),
}

var cookedTokens = []Token{
	CharData(strings.Bytes("\n")),
	ProcInst{"xml", strings.Bytes(`version="1.0" encoding="UTF-8"`)},
	CharData(strings.Bytes("\n")),
	Directive(strings.Bytes(`DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN"
  "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"`)),
	CharData(strings.Bytes("\n")),
	StartElement{Name{"ns2", "body"}, []Attr{Attr{Name{"xmlns", "foo"}, "ns1"}, Attr{Name{"", "xmlns"}, "ns2"}, Attr{Name{"xmlns", "tag"}, "ns3"}}},
	CharData(strings.Bytes("\n  ")),
	StartElement{Name{"ns2", "hello"}, []Attr{Attr{Name{"", "lang"}, "en"}}},
	CharData(strings.Bytes("World <>'\" 白鵬翔")),
	EndElement{Name{"ns2", "hello"}},
	CharData(strings.Bytes("\n  ")),
	StartElement{Name{"ns2", "goodbye"}, nil},
	EndElement{Name{"ns2", "goodbye"}},
	CharData(strings.Bytes("\n  ")),
	StartElement{Name{"ns2", "outer"}, []Attr{Attr{Name{"ns1", "attr"}, "value"}, Attr{Name{"xmlns", "tag"}, "ns4"}}},
	CharData(strings.Bytes("\n    ")),
	StartElement{Name{"ns2", "inner"}, nil},
	EndElement{Name{"ns2", "inner"}},
	CharData(strings.Bytes("\n  ")),
	EndElement{Name{"ns2", "outer"}},
	CharData(strings.Bytes("\n  ")),
	StartElement{Name{"ns3", "name"}, nil},
	CharData(strings.Bytes("\n    ")),
	CharData(strings.Bytes("Some text here.")),
	CharData(strings.Bytes("\n  ")),
	EndElement{Name{"ns3", "name"}},
	CharData(strings.Bytes("\n")),
	EndElement{Name{"ns2", "body"}},
	Comment(strings.Bytes(" missing final newline ")),
}

var xmlInput = []string{
	// unexpected EOF cases
	"<",
	"<t",
	"<t ",
	"<t/",
	"<t/>c",
	"<!",
	"<!-",
	"<!--",
	"<!--c-",
	"<!--c--",
	"<!d",
	"<t></",
	"<t></t",
	"<?",
	"<?p",
	"<t a",
	"<t a=",
	"<t a='",
	"<t a=''",
	"<t/><![",
	"<t/><![C",
	"<t/><![CDATA[d",
	"<t/><![CDATA[d]",
	"<t/><![CDATA[d]]",

	// other Syntax errors
	" ",
	">",
	"<>",
	"<t/a",
	"<0 />",
	"<?0 >",
	//	"<!0 >",	// let the Token() caller handle
	"</0>",
	"<t 0=''>",
	"<t a='&'>",
	"<t a='<'>",
	"<t>&nbspc;</t>",
	"<t a>",
	"<t a=>",
	"<t a=v>",
	//	"<![CDATA[d]]>",	// let the Token() caller handle
	"cdata",
	"<t></e>",
	"<t></>",
	"<t></t!",
	"<t>cdata]]></t>",
}

type stringReader struct {
	s   string
	off int
}

func (r *stringReader) Read(b []byte) (n int, err os.Error) {
	if r.off >= len(r.s) {
		return 0, os.EOF
	}
	for r.off < len(r.s) && n < len(b) {
		b[n] = r.s[r.off]
		n++
		r.off++
	}
	return
}

func (r *stringReader) ReadByte() (b byte, err os.Error) {
	if r.off >= len(r.s) {
		return 0, os.EOF
	}
	b = r.s[r.off]
	r.off++
	return
}

func StringReader(s string) io.Reader { return &stringReader{s, 0} }

func TestRawToken(t *testing.T) {
	p := NewParser(StringReader(testInput))

	for i, want := range rawTokens {
		have, err := p.RawToken()
		if err != nil {
			t.Fatalf("token %d: unexpected error: %s", i, err)
		}
		if !reflect.DeepEqual(have, want) {
			t.Errorf("token %d = %#v want %#v", i, have, want)
		}
	}
}

func TestToken(t *testing.T) {
	p := NewParser(StringReader(testInput))

	for i, want := range cookedTokens {
		have, err := p.Token()
		if err != nil {
			t.Fatalf("token %d: unexpected error: %s", i, err)
		}
		if !reflect.DeepEqual(have, want) {
			t.Errorf("token %d = %#v want %#v", i, have, want)
		}
	}
}

func TestSyntax(t *testing.T) {
	for i := range xmlInput {
		p := NewParser(StringReader(xmlInput[i]))
		var err os.Error
		for _, err = p.Token(); err == nil; _, err = p.Token() {
		}
		if _, ok := err.(SyntaxError); !ok {
			t.Fatalf(`xmlInput "%s": expected SyntaxError not received`, xmlInput[i])
		}
	}
}

type allScalars struct {
	Bool    bool
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint    int
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Uintptr uintptr
	Float   float
	Float32 float32
	Float64 float64
	String  string
}

var all = allScalars{
	Bool: true,
	Int: 1,
	Int8: -2,
	Int16: 3,
	Int32: -4,
	Int64: 5,
	Uint: 6,
	Uint8: 7,
	Uint16: 8,
	Uint32: 9,
	Uint64: 10,
	Uintptr: 11,
	Float: 12.0,
	Float32: 13.0,
	Float64: 14.0,
	String: "15",
}

const testScalarsInput = `<allscalars>
	<bool/>
	<int>1</int>
	<int8>-2</int8>
	<int16>3</int16>
	<int32>-4</int32>
	<int64>5</int64>
	<uint>6</uint>
	<uint8>7</uint8>
	<uint16>8</uint16>
	<uint32>9</uint32>
	<uint64>10</uint64>
	<uintptr>11</uintptr>
	<float>12.0</float>
	<float32>13.0</float32>
	<float64>14.0</float64>
	<string>15</string>
</allscalars>`

func TestAllScalars(t *testing.T) {
	var a allScalars
	buf := bytes.NewBufferString(testScalarsInput)
	err := Unmarshal(buf, &a)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, all) {
		t.Errorf("expected %+v got %+v", a, all)
	}
}

type item struct {
	Field_a string
}

func TestIssue569(t *testing.T) {
	data := `<item><field_a>abcd</field_a></item>`
	var i item
	buf := bytes.NewBufferString(data)
	err := Unmarshal(buf, &i)

	if err != nil || i.Field_a != "abcd" {
		t.Fatalf("Expecting abcd")
	}
}

func TestUnquotedAttrs(t *testing.T) {
	data := "<tag attr=azAZ09:-_\t>"
	p := NewParser(StringReader(data))
	p.Strict = false
	token, err := p.Token()
	if _, ok := err.(SyntaxError); ok {
		t.Errorf("Unexpected error: %v", err)
	}
	if token.(StartElement).Name.Local != "tag" {
		t.Errorf("Unexpected tag name: %v", token.(StartElement).Name.Local)
	}
	attr := token.(StartElement).Attr[0]
	if attr.Value != "azAZ09:-_" {
		t.Errorf("Unexpected attribute value: %v", attr.Value)
	}
	if attr.Name.Local != "attr" {
		t.Errorf("Unexpected attribute name: %v", attr.Name.Local)
	}
}
