// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv_test

import (
	"math"
	. "strconv"
	"testing"
)

type ftoaTest struct {
	f    float64
	fmt  byte
	prec int
	s    string
}

func fdiv(a, b float64) float64 { return a / b } // keep compiler in the dark

const (
	below1e23 = 99999999999999974834176
	above1e23 = 100000000000000008388608
)

var ftoatests = []ftoaTest{
	ftoaTest{1, 'e', 5, "1.00000e+00"},
	ftoaTest{1, 'f', 5, "1.00000"},
	ftoaTest{1, 'g', 5, "1"},
	ftoaTest{1, 'g', -1, "1"},
	ftoaTest{20, 'g', -1, "20"},
	ftoaTest{1234567.8, 'g', -1, "1.2345678e+06"},
	ftoaTest{200000, 'g', -1, "200000"},
	ftoaTest{2000000, 'g', -1, "2e+06"},

	// g conversion and zero suppression
	ftoaTest{400, 'g', 2, "4e+02"},
	ftoaTest{40, 'g', 2, "40"},
	ftoaTest{4, 'g', 2, "4"},
	ftoaTest{.4, 'g', 2, "0.4"},
	ftoaTest{.04, 'g', 2, "0.04"},
	ftoaTest{.004, 'g', 2, "0.004"},
	ftoaTest{.0004, 'g', 2, "0.0004"},
	ftoaTest{.00004, 'g', 2, "4e-05"},
	ftoaTest{.000004, 'g', 2, "4e-06"},

	ftoaTest{0, 'e', 5, "0.00000e+00"},
	ftoaTest{0, 'f', 5, "0.00000"},
	ftoaTest{0, 'g', 5, "0"},
	ftoaTest{0, 'g', -1, "0"},

	ftoaTest{-1, 'e', 5, "-1.00000e+00"},
	ftoaTest{-1, 'f', 5, "-1.00000"},
	ftoaTest{-1, 'g', 5, "-1"},
	ftoaTest{-1, 'g', -1, "-1"},

	ftoaTest{12, 'e', 5, "1.20000e+01"},
	ftoaTest{12, 'f', 5, "12.00000"},
	ftoaTest{12, 'g', 5, "12"},
	ftoaTest{12, 'g', -1, "12"},

	ftoaTest{123456700, 'e', 5, "1.23457e+08"},
	ftoaTest{123456700, 'f', 5, "123456700.00000"},
	ftoaTest{123456700, 'g', 5, "1.2346e+08"},
	ftoaTest{123456700, 'g', -1, "1.234567e+08"},

	ftoaTest{1.2345e6, 'e', 5, "1.23450e+06"},
	ftoaTest{1.2345e6, 'f', 5, "1234500.00000"},
	ftoaTest{1.2345e6, 'g', 5, "1.2345e+06"},

	ftoaTest{1e23, 'e', 17, "9.99999999999999916e+22"},
	ftoaTest{1e23, 'f', 17, "99999999999999991611392.00000000000000000"},
	ftoaTest{1e23, 'g', 17, "9.9999999999999992e+22"},

	ftoaTest{1e23, 'e', -1, "1e+23"},
	ftoaTest{1e23, 'f', -1, "100000000000000000000000"},
	ftoaTest{1e23, 'g', -1, "1e+23"},

	ftoaTest{below1e23, 'e', 17, "9.99999999999999748e+22"},
	ftoaTest{below1e23, 'f', 17, "99999999999999974834176.00000000000000000"},
	ftoaTest{below1e23, 'g', 17, "9.9999999999999975e+22"},

	ftoaTest{below1e23, 'e', -1, "9.999999999999997e+22"},
	ftoaTest{below1e23, 'f', -1, "99999999999999970000000"},
	ftoaTest{below1e23, 'g', -1, "9.999999999999997e+22"},

	ftoaTest{above1e23, 'e', 17, "1.00000000000000008e+23"},
	ftoaTest{above1e23, 'f', 17, "100000000000000008388608.00000000000000000"},
	ftoaTest{above1e23, 'g', 17, "1.0000000000000001e+23"},

	ftoaTest{above1e23, 'e', -1, "1.0000000000000001e+23"},
	ftoaTest{above1e23, 'f', -1, "100000000000000010000000"},
	ftoaTest{above1e23, 'g', -1, "1.0000000000000001e+23"},

	ftoaTest{fdiv(5e-304, 1e20), 'g', -1, "5e-324"},
	ftoaTest{fdiv(-5e-304, 1e20), 'g', -1, "-5e-324"},

	ftoaTest{32, 'g', -1, "32"},
	ftoaTest{32, 'g', 0, "3e+01"},

	ftoaTest{100, 'x', -1, "%x"},

	ftoaTest{math.NaN(), 'g', -1, "NaN"},
	ftoaTest{-math.NaN(), 'g', -1, "NaN"},
	ftoaTest{math.Inf(0), 'g', -1, "+Inf"},
	ftoaTest{math.Inf(-1), 'g', -1, "-Inf"},
	ftoaTest{-math.Inf(0), 'g', -1, "-Inf"},

	ftoaTest{-1, 'b', -1, "-4503599627370496p-52"},

	// fixed bugs
	ftoaTest{0.9, 'f', 1, "0.9"},
	ftoaTest{0.09, 'f', 1, "0.1"},
	ftoaTest{0.0999, 'f', 1, "0.1"},
	ftoaTest{0.05, 'f', 1, "0.1"},
	ftoaTest{0.05, 'f', 0, "0"},
	ftoaTest{0.5, 'f', 1, "0.5"},
	ftoaTest{0.5, 'f', 0, "0"},
	ftoaTest{1.5, 'f', 0, "2"},
}

func TestFtoa(t *testing.T) {
	if FloatSize != 32 {
		println("floatsize: ", FloatSize)
		panic("floatsize")
	}
	for i := 0; i < len(ftoatests); i++ {
		test := &ftoatests[i]
		s := Ftoa64(test.f, test.fmt, test.prec)
		if s != test.s {
			t.Error("test", test.f, string(test.fmt), test.prec, "want", test.s, "got", s)
		}
		s = FtoaN(test.f, test.fmt, test.prec, 64)
		if s != test.s {
			t.Error("testN=64", test.f, string(test.fmt), test.prec, "want", test.s, "got", s)
		}
		if float64(float32(test.f)) == test.f && test.fmt != 'b' {
			s := Ftoa32(float32(test.f), test.fmt, test.prec)
			if s != test.s {
				t.Error("test32", test.f, string(test.fmt), test.prec, "want", test.s, "got", s)
			}
			s = FtoaN(test.f, test.fmt, test.prec, 32)
			if s != test.s {
				t.Error("testN=32", test.f, string(test.fmt), test.prec, "want", test.s, "got", s)
			}
		}
	}
}
