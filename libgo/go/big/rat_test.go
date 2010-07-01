// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import "testing"


type setStringTest struct {
	in, out string
	ok      bool
}

var setStringTests = []setStringTest{
	setStringTest{"0", "0", true},
	setStringTest{"-0", "0", true},
	setStringTest{"1", "1", true},
	setStringTest{"-1", "-1", true},
	setStringTest{in: "r", ok: false},
	setStringTest{in: "a/b", ok: false},
	setStringTest{in: "a.b", ok: false},
	setStringTest{"-0.1", "-1/10", true},
	setStringTest{"-.1", "-1/10", true},
	setStringTest{"2/4", "1/2", true},
	setStringTest{".25", "1/4", true},
	setStringTest{"-1/5", "-1/5", true},
	setStringTest{"884243222337379604041632732738665534", "884243222337379604041632732738665534", true},
	setStringTest{"53/70893980658822810696", "53/70893980658822810696", true},
	setStringTest{"106/141787961317645621392", "53/70893980658822810696", true},
	setStringTest{"204211327800791583.81095", "4084226556015831676219/20000", true},
}

func TestRatSetString(t *testing.T) {
	for i, test := range setStringTests {
		x, ok := new(Rat).SetString(test.in)

		if ok != test.ok || ok && x.String() != test.out {
			t.Errorf("#%d got %s want %s", i, x.String(), test.out)
		}
	}
}


type floatStringTest struct {
	in   string
	prec int
	out  string
}

var floatStringTests = []floatStringTest{
	floatStringTest{"0", 0, "0"},
	floatStringTest{"0", 4, "0"},
	floatStringTest{"1", 0, "1"},
	floatStringTest{"1", 2, "1"},
	floatStringTest{"-1", 0, "-1"},
	floatStringTest{".25", 2, "0.25"},
	floatStringTest{".25", 1, "0.3"},
	floatStringTest{"-1/3", 3, "-0.333"},
	floatStringTest{"-2/3", 4, "-0.6667"},
}

func TestFloatString(t *testing.T) {
	for i, test := range floatStringTests {
		x, _ := new(Rat).SetString(test.in)

		if x.FloatString(test.prec) != test.out {
			t.Errorf("#%d got %s want %s", i, x.FloatString(test.prec), test.out)
		}
	}
}


type ratCmpTest struct {
	rat1, rat2 string
	out        int
}

var ratCmpTests = []ratCmpTest{
	ratCmpTest{"0", "0/1", 0},
	ratCmpTest{"1/1", "1", 0},
	ratCmpTest{"-1", "-2/2", 0},
	ratCmpTest{"1", "0", 1},
	ratCmpTest{"0/1", "1/1", -1},
	ratCmpTest{"-5/1434770811533343057144", "-5/1434770811533343057145", -1},
	ratCmpTest{"49832350382626108453/8964749413", "49832350382626108454/8964749413", -1},
	ratCmpTest{"-37414950961700930/7204075375675961", "37414950961700930/7204075375675961", -1},
	ratCmpTest{"37414950961700930/7204075375675961", "74829901923401860/14408150751351922", 0},
}

func TestRatCmp(t *testing.T) {
	for i, test := range ratCmpTests {
		x, _ := new(Rat).SetString(test.rat1)
		y, _ := new(Rat).SetString(test.rat2)

		out := x.Cmp(y)
		if out != test.out {
			t.Errorf("#%d got out = %v; want %v", i, out, test.out)
		}
	}
}


type ratBinFun func(z, x, y *Rat) *Rat
type ratBinArg struct {
	x, y, z string
}

func testRatBin(t *testing.T, i int, name string, f ratBinFun, a ratBinArg) {
	x, _ := NewRat(0, 1).SetString(a.x)
	y, _ := NewRat(0, 1).SetString(a.y)
	z, _ := NewRat(0, 1).SetString(a.z)
	out := f(NewRat(0, 1), x, y)

	if out.Cmp(z) != 0 {
		t.Errorf("%s #%d got %s want %s", name, i, out, z)
	}
}


type ratBinTest struct {
	x, y      string
	sum, prod string
}

var ratBinTests = []ratBinTest{
	ratBinTest{"0", "0", "0", "0"},
	ratBinTest{"0", "1", "1", "0"},
	ratBinTest{"-1", "0", "-1", "0"},
	ratBinTest{"-1", "1", "0", "-1"},
	ratBinTest{"1", "1", "2", "1"},
	ratBinTest{"1/2", "1/2", "1", "1/4"},
	ratBinTest{"1/4", "1/3", "7/12", "1/12"},
	ratBinTest{"2/5", "-14/3", "-64/15", "-28/15"},
	ratBinTest{"4707/49292519774798173060", "-3367/70976135186689855734", "84058377121001851123459/1749296273614329067191168098769082663020", "-1760941/388732505247628681598037355282018369560"},
	ratBinTest{"-61204110018146728334/3", "-31052192278051565633/2", "-215564796870448153567/6", "950260896245257153059642991192710872711/3"},
	ratBinTest{"-854857841473707320655/4237645934602118692642972629634714039", "-18/31750379913563777419", "-27/133467566250814981", "15387441146526731771790/134546868362786310073779084329032722548987800600710485341"},
	ratBinTest{"618575745270541348005638912139/19198433543745179392300736", "-19948846211000086/637313996471", "27674141753240653/30123979153216", "-6169936206128396568797607742807090270137721977/6117715203873571641674006593837351328"},
	ratBinTest{"-3/26206484091896184128", "5/2848423294177090248", "15310893822118706237/9330894968229805033368778458685147968", "-5/24882386581946146755650075889827061248"},
	ratBinTest{"26946729/330400702820", "41563965/225583428284", "1238218672302860271/4658307703098666660055", "224002580204097/14906584649915733312176"},
	ratBinTest{"-8259900599013409474/7", "-84829337473700364773/56707961321161574960", "-468402123685491748914621885145127724451/396955729248131024720", "350340947706464153265156004876107029701/198477864624065512360"},
	ratBinTest{"575775209696864/1320203974639986246357", "29/712593081308", "410331716733912717985762465/940768218243776489278275419794956", "808/45524274987585732633"},
	ratBinTest{"1786597389946320496771/2066653520653241", "6269770/1992362624741777", "3559549865190272133656109052308126637/4117523232840525481453983149257", "8967230/3296219033"},
	ratBinTest{"-36459180403360509753/32150500941194292113930", "9381566963714/9633539", "301622077145533298008420642898530153/309723104686531919656937098270", "-3784609207827/3426986245"},
}

func TestRatBin(t *testing.T) {
	for i, test := range ratBinTests {
		arg := ratBinArg{test.x, test.y, test.sum}
		testRatBin(t, i, "Add", (*Rat).Add, arg)

		arg = ratBinArg{test.y, test.x, test.sum}
		testRatBin(t, i, "Add symmetric", (*Rat).Add, arg)

		arg = ratBinArg{test.sum, test.x, test.y}
		testRatBin(t, i, "Sub", (*Rat).Sub, arg)

		arg = ratBinArg{test.sum, test.y, test.x}
		testRatBin(t, i, "Sub symmetric", (*Rat).Sub, arg)

		arg = ratBinArg{test.x, test.y, test.prod}
		testRatBin(t, i, "Mul", (*Rat).Mul, arg)

		arg = ratBinArg{test.y, test.x, test.prod}
		testRatBin(t, i, "Mul symmetric", (*Rat).Mul, arg)

		if test.x != "0" {
			arg = ratBinArg{test.prod, test.x, test.y}
			testRatBin(t, i, "Quo", (*Rat).Quo, arg)
		}

		if test.y != "0" {
			arg = ratBinArg{test.prod, test.y, test.x}
			testRatBin(t, i, "Quo symmetric", (*Rat).Quo, arg)
		}
	}
}


func TestIssue820(t *testing.T) {
	x := NewRat(3, 1)
	y := NewRat(2, 1)
	z := y.Quo(x, y)
	q := NewRat(3, 2)
	if z.Cmp(q) != 0 {
		t.Errorf("got %s want %s", z, q)
	}

	y = NewRat(3, 1)
	x = NewRat(2, 1)
	z = y.Quo(x, y)
	q = NewRat(2, 3)
	if z.Cmp(q) != 0 {
		t.Errorf("got %s want %s", z, q)
	}

	x = NewRat(3, 1)
	z = x.Quo(x, x)
	q = NewRat(3, 3)
	if z.Cmp(q) != 0 {
		t.Errorf("got %s want %s", z, q)
	}
}
