// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fmt_test

import (
	. "fmt"
	"io"
	"math"
	"runtime" // for the malloc count test only
	"strings"
	"testing"
)

type (
	renamedBool       bool
	renamedInt        int
	renamedInt8       int8
	renamedInt16      int16
	renamedInt32      int32
	renamedInt64      int64
	renamedUint       uint
	renamedUint8      uint8
	renamedUint16     uint16
	renamedUint32     uint32
	renamedUint64     uint64
	renamedUintptr    uintptr
	renamedString     string
	renamedBytes      []byte
	renamedFloat      float
	renamedFloat32    float32
	renamedFloat64    float64
	renamedComplex    complex
	renamedComplex64  complex64
	renamedComplex128 complex128
)

func TestFmtInterface(t *testing.T) {
	var i1 interface{}
	i1 = "abc"
	s := Sprintf("%s", i1)
	if s != "abc" {
		t.Errorf(`Sprintf("%%s", empty("abc")) = %q want %q`, s, "abc")
	}
}

type fmtTest struct {
	fmt string
	val interface{}
	out string
}

const b32 uint32 = 1<<32 - 1
const b64 uint64 = 1<<64 - 1

var array = []int{1, 2, 3, 4, 5}
var iarray = []interface{}{1, "hello", 2.5, nil}

type A struct {
	i int
	j uint
	s string
	x []int
}

type I int

func (i I) String() string { return Sprintf("<%d>", int(i)) }

type B struct {
	i I
	j int
}

type C struct {
	i int
	B
}

type F int

func (f F) Format(s State, c int) {
	Fprintf(s, "<%c=F(%d)>", c, int(f))
}

type G int

func (g G) GoString() string {
	return Sprintf("GoString(%d)", int(g))
}

type S struct {
	f F // a struct field that Formats
	g G // a struct field that GoStrings
}

// A type with a String method with pointer receiver for testing %p
type P int

var pValue P

func (p *P) String() string {
	return "String(p)"
}

var b byte

var fmttests = []fmtTest{
	fmtTest{"%d", 12345, "12345"},
	fmtTest{"%v", 12345, "12345"},
	fmtTest{"%t", true, "true"},

	// basic string
	fmtTest{"%s", "abc", "abc"},
	fmtTest{"%x", "abc", "616263"},
	fmtTest{"%x", "xyz", "78797a"},
	fmtTest{"%X", "xyz", "78797A"},
	fmtTest{"%q", "abc", `"abc"`},

	// basic bytes
	fmtTest{"%s", []byte("abc"), "abc"},
	fmtTest{"%x", []byte("abc"), "616263"},
	fmtTest{"% x", []byte("abc"), "61 62 63"},
	fmtTest{"%x", []byte("xyz"), "78797a"},
	fmtTest{"%X", []byte("xyz"), "78797A"},
	fmtTest{"%q", []byte("abc"), `"abc"`},

	// escaped strings
	fmtTest{"%#q", `abc`, "`abc`"},
	fmtTest{"%#q", `"`, "`\"`"},
	fmtTest{"1 %#q", `\n`, "1 `\\n`"},
	fmtTest{"2 %#q", "\n", `2 "\n"`},
	fmtTest{"%q", `"`, `"\""`},
	fmtTest{"%q", "\a\b\f\r\n\t\v", `"\a\b\f\r\n\t\v"`},
	fmtTest{"%q", "abc\xffdef", `"abc\xffdef"`},
	fmtTest{"%q", "\u263a", `"\u263a"`},
	fmtTest{"%q", "\U0010ffff", `"\U0010ffff"`},

	// width
	fmtTest{"%5s", "abc", "  abc"},
	fmtTest{"%2s", "\u263a", " \u263a"},
	fmtTest{"%-5s", "abc", "abc  "},
	fmtTest{"%05s", "abc", "00abc"},

	// integers
	fmtTest{"%d", 12345, "12345"},
	fmtTest{"%d", -12345, "-12345"},
	fmtTest{"%10d", 12345, "     12345"},
	fmtTest{"%10d", -12345, "    -12345"},
	fmtTest{"%+10d", 12345, "    +12345"},
	fmtTest{"%010d", 12345, "0000012345"},
	fmtTest{"%010d", -12345, "-000012345"},
	fmtTest{"%-10d", 12345, "12345     "},
	fmtTest{"%010.3d", 1, "       001"},
	fmtTest{"%010.3d", -1, "      -001"},
	fmtTest{"%+d", 12345, "+12345"},
	fmtTest{"%+d", -12345, "-12345"},
	fmtTest{"%+d", 0, "+0"},
	fmtTest{"% d", 0, " 0"},
	fmtTest{"% d", 12345, " 12345"},

	// floats
	fmtTest{"%+.3e", 0.0, "+0.000e+00"},
	fmtTest{"%+.3e", 1.0, "+1.000e+00"},
	fmtTest{"%+.3f", -1.0, "-1.000"},
	fmtTest{"% .3E", -1.0, "-1.000E+00"},
	fmtTest{"% .3e", 1.0, " 1.000e+00"},
	fmtTest{"%+.3g", 0.0, "+0"},
	fmtTest{"%+.3g", 1.0, "+1"},
	fmtTest{"%+.3g", -1.0, "-1"},
	fmtTest{"% .3g", -1.0, "-1"},
	fmtTest{"% .3g", 1.0, " 1"},

	// complex values
	fmtTest{"%+.3e", 0i, "(+0.000e+00+0.000e+00i)"},
	fmtTest{"%+.3f", 0i, "(+0.000+0.000i)"},
	fmtTest{"%+.3g", 0i, "(+0+0i)"},
	fmtTest{"%+.3e", 1 + 2i, "(+1.000e+00+2.000e+00i)"},
	fmtTest{"%+.3f", 1 + 2i, "(+1.000+2.000i)"},
	fmtTest{"%+.3g", 1 + 2i, "(+1+2i)"},
	fmtTest{"%.3e", 0i, "(0.000e+00+0.000e+00i)"},
	fmtTest{"%.3f", 0i, "(0.000+0.000i)"},
	fmtTest{"%.3g", 0i, "(0+0i)"},
	fmtTest{"%.3e", 1 + 2i, "(1.000e+00+2.000e+00i)"},
	fmtTest{"%.3f", 1 + 2i, "(1.000+2.000i)"},
	fmtTest{"%.3g", 1 + 2i, "(1+2i)"},
	fmtTest{"%.3e", -1 - 2i, "(-1.000e+00-2.000e+00i)"},
	fmtTest{"%.3f", -1 - 2i, "(-1.000-2.000i)"},
	fmtTest{"%.3g", -1 - 2i, "(-1-2i)"},
	fmtTest{"% .3E", -1 - 2i, "(-1.000E+00-2.000E+00i)"},
	fmtTest{"%+.3g", complex64(1 + 2i), "(+1+2i)"},
	fmtTest{"%+.3g", complex128(1 + 2i), "(+1+2i)"},

	// erroneous formats
	fmtTest{"", 2, "?(extra int=2)"},
	fmtTest{"%d", "hello", "%d(string=hello)"},

	// old test/fmt_test.go
	fmtTest{"%d", 1234, "1234"},
	fmtTest{"%d", -1234, "-1234"},
	fmtTest{"%d", uint(1234), "1234"},
	fmtTest{"%d", uint32(b32), "4294967295"},
	fmtTest{"%d", uint64(b64), "18446744073709551615"},
	fmtTest{"%o", 01234, "1234"},
	fmtTest{"%#o", 01234, "01234"},
	fmtTest{"%o", uint32(b32), "37777777777"},
	fmtTest{"%o", uint64(b64), "1777777777777777777777"},
	fmtTest{"%x", 0x1234abcd, "1234abcd"},
	fmtTest{"%#x", 0x1234abcd, "0x1234abcd"},
	fmtTest{"%x", b32 - 0x1234567, "fedcba98"},
	fmtTest{"%X", 0x1234abcd, "1234ABCD"},
	fmtTest{"%X", b32 - 0x1234567, "FEDCBA98"},
	fmtTest{"%#X", 0, "0X0"},
	fmtTest{"%x", b64, "ffffffffffffffff"},
	fmtTest{"%b", 7, "111"},
	fmtTest{"%b", b64, "1111111111111111111111111111111111111111111111111111111111111111"},
	fmtTest{"%b", -6, "-110"},
	fmtTest{"%e", float64(1), "1.000000e+00"},
	fmtTest{"%e", float64(1234.5678e3), "1.234568e+06"},
	fmtTest{"%e", float64(1234.5678e-8), "1.234568e-05"},
	fmtTest{"%e", float64(-7), "-7.000000e+00"},
	fmtTest{"%e", float64(-1e-9), "-1.000000e-09"},
	fmtTest{"%f", float64(1234.5678e3), "1234567.800000"},
	fmtTest{"%f", float64(1234.5678e-8), "0.000012"},
	fmtTest{"%f", float64(-7), "-7.000000"},
	fmtTest{"%f", float64(-1e-9), "-0.000000"},
	fmtTest{"%g", float64(1234.5678e3), "1.2345678e+06"},
	fmtTest{"%g", float32(1234.5678e3), "1.2345678e+06"},
	fmtTest{"%g", float64(1234.5678e-8), "1.2345678e-05"},
	fmtTest{"%g", float64(-7), "-7"},
	fmtTest{"%g", float64(-1e-9), "-1e-09"},
	fmtTest{"%g", float32(-1e-9), "-1e-09"},
	fmtTest{"%E", float64(1), "1.000000E+00"},
	fmtTest{"%E", float64(1234.5678e3), "1.234568E+06"},
	fmtTest{"%E", float64(1234.5678e-8), "1.234568E-05"},
	fmtTest{"%E", float64(-7), "-7.000000E+00"},
	fmtTest{"%E", float64(-1e-9), "-1.000000E-09"},
	fmtTest{"%G", float64(1234.5678e3), "1.2345678E+06"},
	fmtTest{"%G", float32(1234.5678e3), "1.2345678E+06"},
	fmtTest{"%G", float64(1234.5678e-8), "1.2345678E-05"},
	fmtTest{"%G", float64(-7), "-7"},
	fmtTest{"%G", float64(-1e-9), "-1E-09"},
	fmtTest{"%G", float32(-1e-9), "-1E-09"},
	fmtTest{"%c", 'x', "x"},
	fmtTest{"%c", 0xe4, "ä"},
	fmtTest{"%c", 0x672c, "本"},
	fmtTest{"%c", '日', "日"},
	fmtTest{"%20.8d", 1234, "            00001234"},
	fmtTest{"%20.8d", -1234, "           -00001234"},
	fmtTest{"%20d", 1234, "                1234"},
	fmtTest{"%-20.8d", 1234, "00001234            "},
	fmtTest{"%-20.8d", -1234, "-00001234           "},
	fmtTest{"%-#20.8x", 0x1234abc, "0x01234abc          "},
	fmtTest{"%-#20.8X", 0x1234abc, "0X01234ABC          "},
	fmtTest{"%-#20.8o", 01234, "00001234            "},
	fmtTest{"%.20b", 7, "00000000000000000111"},
	fmtTest{"%20.5s", "qwertyuiop", "               qwert"},
	fmtTest{"%.5s", "qwertyuiop", "qwert"},
	fmtTest{"%-20.5s", "qwertyuiop", "qwert               "},
	fmtTest{"%20c", 'x', "                   x"},
	fmtTest{"%-20c", 'x', "x                   "},
	fmtTest{"%20.6e", 1.2345e3, "        1.234500e+03"},
	fmtTest{"%20.6e", 1.2345e-3, "        1.234500e-03"},
	fmtTest{"%20e", 1.2345e3, "        1.234500e+03"},
	fmtTest{"%20e", 1.2345e-3, "        1.234500e-03"},
	fmtTest{"%20.8e", 1.2345e3, "      1.23450000e+03"},
	fmtTest{"%20f", float64(1.23456789e3), "         1234.567890"},
	fmtTest{"%20f", float64(1.23456789e-3), "            0.001235"},
	fmtTest{"%20f", float64(12345678901.23456789), "  12345678901.234568"},
	fmtTest{"%-20f", float64(1.23456789e3), "1234.567890         "},
	fmtTest{"%20.8f", float64(1.23456789e3), "       1234.56789000"},
	fmtTest{"%20.8f", float64(1.23456789e-3), "          0.00123457"},
	fmtTest{"%g", float64(1.23456789e3), "1234.56789"},
	fmtTest{"%g", float64(1.23456789e-3), "0.00123456789"},
	fmtTest{"%g", float64(1.23456789e20), "1.23456789e+20"},
	fmtTest{"%20e", math.Inf(1), "                +Inf"},
	fmtTest{"%-20f", math.Inf(-1), "-Inf                "},
	fmtTest{"%20g", math.NaN(), "                 NaN"},

	// arrays
	fmtTest{"%v", array, "[1 2 3 4 5]"},
	fmtTest{"%v", iarray, "[1 hello 2.5 <nil>]"},
	fmtTest{"%v", &array, "&[1 2 3 4 5]"},
	fmtTest{"%v", &iarray, "&[1 hello 2.5 <nil>]"},

	// complexes with %v
	fmtTest{"%v", 1 + 2i, "(1+2i)"},
	fmtTest{"%v", complex64(1 + 2i), "(1+2i)"},
	fmtTest{"%v", complex128(1 + 2i), "(1+2i)"},

	// structs
	fmtTest{"%v", A{1, 2, "a", []int{1, 2}}, `{1 2 a [1 2]}`},
	fmtTest{"%+v", A{1, 2, "a", []int{1, 2}}, `{i:1 j:2 s:a x:[1 2]}`},

	// +v on structs with Stringable items
	fmtTest{"%+v", B{1, 2}, `{i:<1> j:2}`},
	fmtTest{"%+v", C{1, B{2, 3}}, `{i:1 B:{i:<2> j:3}}`},

	// q on Stringable items
	fmtTest{"%s", I(23), `<23>`},
	fmtTest{"%q", I(23), `"<23>"`},
	fmtTest{"%x", I(23), `3c32333e`},
	fmtTest{"%d", I(23), `%d(string=<23>)`},

	// go syntax
	fmtTest{"%#v", A{1, 2, "a", []int{1, 2}}, `fmt_test.A{i:1, j:0x2, s:"a", x:[]int{1, 2}}`},
	fmtTest{"%#v", &b, "(*uint8)(PTR)"},
	fmtTest{"%#v", TestFmtInterface, "(func(*testing.T))(PTR)"},
	fmtTest{"%#v", make(chan int), "(chan int)(PTR)"},
	fmtTest{"%#v", uint64(1<<64 - 1), "0xffffffffffffffff"},
	fmtTest{"%#v", 1000000000, "1000000000"},
	fmtTest{"%#v", map[string]int{"a": 1, "b": 2}, `map[string] int{"a":1, "b":2}`},
	fmtTest{"%#v", map[string]B{"a": B{1, 2}, "b": B{3, 4}}, `map[string] fmt_test.B{"a":fmt_test.B{i:1, j:2}, "b":fmt_test.B{i:3, j:4}}`},
	fmtTest{"%#v", []string{"a", "b"}, `[]string{"a", "b"}`},

	// slices with other formats
	fmtTest{"%#x", []int{1, 2, 15}, `[0x1 0x2 0xf]`},
	fmtTest{"%x", []int{1, 2, 15}, `[1 2 f]`},
	fmtTest{"%q", []string{"a", "b"}, `["a" "b"]`},

	// renamings
	fmtTest{"%v", renamedBool(true), "true"},
	fmtTest{"%d", renamedBool(true), "%d(fmt_test.renamedBool=true)"},
	fmtTest{"%o", renamedInt(8), "10"},
	fmtTest{"%d", renamedInt8(-9), "-9"},
	fmtTest{"%v", renamedInt16(10), "10"},
	fmtTest{"%v", renamedInt32(-11), "-11"},
	fmtTest{"%X", renamedInt64(255), "FF"},
	fmtTest{"%v", renamedUint(13), "13"},
	fmtTest{"%o", renamedUint8(14), "16"},
	fmtTest{"%X", renamedUint16(15), "F"},
	fmtTest{"%d", renamedUint32(16), "16"},
	fmtTest{"%X", renamedUint64(17), "11"},
	fmtTest{"%o", renamedUintptr(18), "22"},
	fmtTest{"%x", renamedString("thing"), "7468696e67"},
	fmtTest{"%q", renamedBytes([]byte("hello")), `"hello"`},
	fmtTest{"%v", renamedFloat(11), "11"},
	fmtTest{"%v", renamedFloat32(22), "22"},
	fmtTest{"%v", renamedFloat64(33), "33"},
	fmtTest{"%v", renamedComplex(7 + .2i), "(7+0.2i)"},
	fmtTest{"%v", renamedComplex64(3 + 4i), "(3+4i)"},
	fmtTest{"%v", renamedComplex128(4 - 3i), "(4-3i)"},

	// Formatter
	fmtTest{"%x", F(1), "<x=F(1)>"},
	fmtTest{"%x", G(2), "2"},
	fmtTest{"%+v", S{F(4), G(5)}, "{f:<v=F(4)> g:5}"},

	// GoStringer
	fmtTest{"%#v", G(6), "GoString(6)"},
	fmtTest{"%#v", S{F(7), G(8)}, "fmt_test.S{f:<v=F(7)>, g:GoString(8)}"},

	// %T
	fmtTest{"%T", (4 - 3i), "complex"},
	fmtTest{"%T", renamedComplex128(4 - 3i), "fmt_test.renamedComplex128"},
	fmtTest{"%T", intVal, "int"},
	fmtTest{"%6T", &intVal, "  *int"},

	// %p
	fmtTest{"p0=%p", new(int), "p0=PTR"},
	fmtTest{"p1=%s", &pValue, "p1=String(p)"}, // String method...
	fmtTest{"p2=%p", &pValue, "p2=PTR"},       // ... not called with %p

	// %p on non-pointers
	fmtTest{"%p", make(chan int), "PTR"},
	fmtTest{"%p", make(map[int]int), "PTR"},
	fmtTest{"%p", make([]int, 1), "PTR"},
	fmtTest{"%p", 27, "%p(int=27)"}, // not a pointer at all

	// erroneous things
	fmtTest{"%d", "hello", "%d(string=hello)"},
	fmtTest{"no args", "hello", "no args?(extra string=hello)"},
	fmtTest{"%s", nil, "%s(<nil>)"},
	fmtTest{"%T", nil, "<nil>"},
	fmtTest{"%-1", 100, "%1(int=100)"},
}

func TestSprintf(t *testing.T) {
	for _, tt := range fmttests {
		s := Sprintf(tt.fmt, tt.val)
		if i := strings.Index(s, "0x"); i >= 0 && strings.Index(tt.out, "PTR") >= 0 {
			j := i + 2
			for ; j < len(s); j++ {
				c := s[j]
				if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
					break
				}
			}
			s = s[0:i] + "PTR" + s[j:]
		}
		if s != tt.out {
			if _, ok := tt.val.(string); ok {
				// Don't requote the already-quoted strings.
				// It's too confusing to read the errors.
				t.Errorf("Sprintf(%q, %q) = <%s> want <%s>", tt.fmt, tt.val, s, tt.out)
			} else {
				t.Errorf("Sprintf(%q, %v) = %q want %q", tt.fmt, tt.val, s, tt.out)
			}
		}
	}
}

func BenchmarkSprintfEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sprintf("")
	}
}

func BenchmarkSprintfString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sprintf("%s", "hello")
	}
}

func BenchmarkSprintfInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sprintf("%d", 5)
	}
}

func BenchmarkSprintfIntInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sprintf("%d %d", 5, 6)
	}
}

func TestCountMallocs(t *testing.T) {
	mallocs := 0 - runtime.MemStats.Mallocs
	for i := 0; i < 100; i++ {
		Sprintf("")
	}
	mallocs += runtime.MemStats.Mallocs
	Printf("mallocs per Sprintf(\"\"): %d\n", mallocs/100)
	mallocs = 0 - runtime.MemStats.Mallocs
	for i := 0; i < 100; i++ {
		Sprintf("xxx")
	}
	mallocs += runtime.MemStats.Mallocs
	Printf("mallocs per Sprintf(\"xxx\"): %d\n", mallocs/100)
	mallocs = 0 - runtime.MemStats.Mallocs
	for i := 0; i < 100; i++ {
		Sprintf("%x", i)
	}
	mallocs += runtime.MemStats.Mallocs
	Printf("mallocs per Sprintf(\"%%x\"): %d\n", mallocs/100)
	mallocs = 0 - runtime.MemStats.Mallocs
	for i := 0; i < 100; i++ {
		Sprintf("%x %x", i, i)
	}
	mallocs += runtime.MemStats.Mallocs
	Printf("mallocs per Sprintf(\"%%x %%x\"): %d\n", mallocs/100)
}

type flagPrinter struct{}

func (*flagPrinter) Format(f State, c int) {
	s := "%"
	for i := 0; i < 128; i++ {
		if f.Flag(i) {
			s += string(i)
		}
	}
	if w, ok := f.Width(); ok {
		s += Sprintf("%d", w)
	}
	if p, ok := f.Precision(); ok {
		s += Sprintf(".%d", p)
	}
	s += string(c)
	io.WriteString(f, "["+s+"]")
}

type flagTest struct {
	in  string
	out string
}

var flagtests = []flagTest{
	flagTest{"%a", "[%a]"},
	flagTest{"%-a", "[%-a]"},
	flagTest{"%+a", "[%+a]"},
	flagTest{"%#a", "[%#a]"},
	flagTest{"% a", "[% a]"},
	flagTest{"%0a", "[%0a]"},
	flagTest{"%1.2a", "[%1.2a]"},
	flagTest{"%-1.2a", "[%-1.2a]"},
	flagTest{"%+1.2a", "[%+1.2a]"},
	flagTest{"%-+1.2a", "[%+-1.2a]"},
	flagTest{"%-+1.2abc", "[%+-1.2a]bc"},
	flagTest{"%-1.2abc", "[%-1.2a]bc"},
}

func TestFlagParser(t *testing.T) {
	var flagprinter flagPrinter
	for _, tt := range flagtests {
		s := Sprintf(tt.in, &flagprinter)
		if s != tt.out {
			t.Errorf("Sprintf(%q, &flagprinter) => %q, want %q", tt.in, s, tt.out)
		}
	}
}

func TestStructPrinter(t *testing.T) {
	var s struct {
		a string
		b string
		c int
	}
	s.a = "abc"
	s.b = "def"
	s.c = 123
	type Test struct {
		fmt string
		out string
	}
	var tests = []Test{
		Test{"%v", "{abc def 123}"},
		Test{"%+v", "{a:abc b:def c:123}"},
	}
	for _, tt := range tests {
		out := Sprintf(tt.fmt, s)
		if out != tt.out {
			t.Errorf("Sprintf(%q, &s) = %q, want %q", tt.fmt, out, tt.out)
		}
	}
}

// Check map printing using substrings so we don't depend on the print order.
func presentInMap(s string, a []string, t *testing.T) {
	for i := 0; i < len(a); i++ {
		loc := strings.Index(s, a[i])
		if loc < 0 {
			t.Errorf("map print: expected to find %q in %q", a[i], s)
		}
		// make sure the match ends here
		loc += len(a[i])
		if loc >= len(s) || (s[loc] != ' ' && s[loc] != ']') {
			t.Errorf("map print: %q not properly terminated in %q", a[i], s)
		}
	}
}

func TestMapPrinter(t *testing.T) {
	m0 := make(map[int]string)
	s := Sprint(m0)
	if s != "map[]" {
		t.Errorf("empty map printed as %q not %q", s, "map[]")
	}
	m1 := map[int]string{1: "one", 2: "two", 3: "three"}
	a := []string{"1:one", "2:two", "3:three"}
	presentInMap(Sprintf("%v", m1), a, t)
	presentInMap(Sprint(m1), a, t)
}

func TestEmptyMap(t *testing.T) {
	const emptyMapStr = "map[]"
	var m map[string]int
	s := Sprint(m)
	if s != emptyMapStr {
		t.Errorf("nil map printed as %q not %q", s, emptyMapStr)
	}
	m = make(map[string]int)
	s = Sprint(m)
	if s != emptyMapStr {
		t.Errorf("empty map printed as %q not %q", s, emptyMapStr)
	}
}

// Check that Sprint (and hence Print, Fprint) puts spaces in the right places,
// that is, between arg pairs in which neither is a string.
func TestBlank(t *testing.T) {
	got := Sprint("<", 1, ">:", 1, 2, 3, "!")
	expect := "<1>:1 2 3!"
	if got != expect {
		t.Errorf("got %q expected %q", got, expect)
	}
}

// Check that Sprintln (and hence Println, Fprintln) puts spaces in the right places,
// that is, between all arg pairs.
func TestBlankln(t *testing.T) {
	got := Sprintln("<", 1, ">:", 1, 2, 3, "!")
	expect := "< 1 >: 1 2 3 !\n"
	if got != expect {
		t.Errorf("got %q expected %q", got, expect)
	}
}


// Check Formatter with Sprint, Sprintln, Sprintf
func TestFormatterPrintln(t *testing.T) {
	f := F(1)
	expect := "<v=F(1)>\n"
	s := Sprint(f, "\n")
	if s != expect {
		t.Errorf("Sprint wrong with Formatter: expected %q got %q\n", expect, s)
	}
	s = Sprintln(f)
	if s != expect {
		t.Errorf("Sprintln wrong with Formatter: expected %q got %q\n", expect, s)
	}
	s = Sprintf("%v\n", f)
	if s != expect {
		t.Errorf("Sprintf wrong with Formatter: expected %q got %q\n", expect, s)
	}
}
