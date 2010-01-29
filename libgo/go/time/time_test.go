// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"os"
	"testing"
	"testing/quick"
	. "time"
)

func init() {
	// Force US Pacific time for daylight-savings
	// tests below (localtests).  Needs to be set
	// before the first call into the time library.
	os.Setenv("TZ", "America/Los_Angeles")
}

type TimeTest struct {
	seconds int64
	golden  Time
}

var utctests = []TimeTest{
	TimeTest{0, Time{1970, 1, 1, 0, 0, 0, Thursday, 0, "UTC"}},
	TimeTest{1221681866, Time{2008, 9, 17, 20, 4, 26, Wednesday, 0, "UTC"}},
	TimeTest{-1221681866, Time{1931, 4, 16, 3, 55, 34, Thursday, 0, "UTC"}},
	TimeTest{-11644473600, Time{1601, 1, 1, 0, 0, 0, Monday, 0, "UTC"}},
	TimeTest{599529660, Time{1988, 12, 31, 0, 1, 0, Saturday, 0, "UTC"}},
	TimeTest{978220860, Time{2000, 12, 31, 0, 1, 0, Sunday, 0, "UTC"}},
	TimeTest{1e18, Time{31688740476, 10, 23, 1, 46, 40, Friday, 0, "UTC"}},
	TimeTest{-1e18, Time{-31688736537, 3, 10, 22, 13, 20, Tuesday, 0, "UTC"}},
	TimeTest{0x7fffffffffffffff, Time{292277026596, 12, 4, 15, 30, 7, Sunday, 0, "UTC"}},
	TimeTest{-0x8000000000000000, Time{-292277022657, 1, 27, 8, 29, 52, Sunday, 0, "UTC"}},
}

var localtests = []TimeTest{
	TimeTest{0, Time{1969, 12, 31, 16, 0, 0, Wednesday, -8 * 60 * 60, "PST"}},
	TimeTest{1221681866, Time{2008, 9, 17, 13, 4, 26, Wednesday, -7 * 60 * 60, "PDT"}},
}

func same(t, u *Time) bool {
	return t.Year == u.Year &&
		t.Month == u.Month &&
		t.Day == u.Day &&
		t.Hour == u.Hour &&
		t.Minute == u.Minute &&
		t.Second == u.Second &&
		t.Weekday == u.Weekday &&
		t.ZoneOffset == u.ZoneOffset &&
		t.Zone == u.Zone
}

func TestSecondsToUTC(t *testing.T) {
	for i := 0; i < len(utctests); i++ {
		sec := utctests[i].seconds
		golden := &utctests[i].golden
		tm := SecondsToUTC(sec)
		newsec := tm.Seconds()
		if newsec != sec {
			t.Errorf("SecondsToUTC(%d).Seconds() = %d", sec, newsec)
		}
		if !same(tm, golden) {
			t.Errorf("SecondsToUTC(%d):", sec)
			t.Errorf("  want=%+v", *golden)
			t.Errorf("  have=%+v", *tm)
		}
	}
}

func TestSecondsToLocalTime(t *testing.T) {
	for i := 0; i < len(localtests); i++ {
		sec := localtests[i].seconds
		golden := &localtests[i].golden
		tm := SecondsToLocalTime(sec)
		newsec := tm.Seconds()
		if newsec != sec {
			t.Errorf("SecondsToLocalTime(%d).Seconds() = %d", sec, newsec)
		}
		if !same(tm, golden) {
			t.Errorf("SecondsToLocalTime(%d):", sec)
			t.Errorf("  want=%+v", *golden)
			t.Errorf("  have=%+v", *tm)
		}
	}
}

func TestSecondsToUTCAndBack(t *testing.T) {
	f := func(sec int64) bool { return SecondsToUTC(sec).Seconds() == sec }
	f32 := func(sec int32) bool { return f(int64(sec)) }
	cfg := &quick.Config{MaxCount: 10000}

	// Try a reasonable date first, then the huge ones.
	if err := quick.Check(f32, cfg); err != nil {
		t.Fatal(err)
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Fatal(err)
	}
}

type TimeFormatTest struct {
	time           Time
	formattedValue string
}

var iso8601Formats = []TimeFormatTest{
	TimeFormatTest{Time{2008, 9, 17, 20, 4, 26, Wednesday, 0, "UTC"}, "2008-09-17T20:04:26Z"},
	TimeFormatTest{Time{1994, 9, 17, 20, 4, 26, Wednesday, -18000, "EST"}, "1994-09-17T20:04:26-0500"},
	TimeFormatTest{Time{2000, 12, 26, 1, 15, 6, Wednesday, 15600, "OTO"}, "2000-12-26T01:15:06+0420"},
}

func TestISO8601Conversion(t *testing.T) {
	for _, f := range iso8601Formats {
		if f.time.ISO8601() != f.formattedValue {
			t.Error("ISO8601():")
			t.Errorf("  want=%+v", f.formattedValue)
			t.Errorf("  have=%+v", f.time.ISO8601())
		}
	}
}

func BenchmarkSeconds(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Seconds()
	}
}

func BenchmarkNanoseconds(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Nanoseconds()
	}
}
