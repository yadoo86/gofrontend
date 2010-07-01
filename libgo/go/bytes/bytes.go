// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The bytes package implements functions for the manipulation of byte slices.
// Analagous to the facilities of the strings package.
package bytes

import (
	"unicode"
	"utf8"
)

// Compare returns an integer comparing the two byte arrays lexicographically.
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b
func Compare(a, b []byte) int {
	m := len(a)
	if m > len(b) {
		m = len(b)
	}
	for i, ac := range a[0:m] {
		bc := b[i]
		switch {
		case ac > bc:
			return 1
		case ac < bc:
			return -1
		}
	}
	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	}
	return 0
}

// Equal returns a boolean reporting whether a == b.
func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, c := range a {
		if c != b[i] {
			return false
		}
	}
	return true
}

// explode splits s into an array of UTF-8 sequences, one per Unicode character (still arrays of bytes),
// up to a maximum of n byte arrays. Invalid UTF-8 sequences are chopped into individual bytes.
func explode(s []byte, n int) [][]byte {
	if n <= 0 {
		n = len(s)
	}
	a := make([][]byte, n)
	var size int
	na := 0
	for len(s) > 0 {
		if na+1 >= n {
			a[na] = s
			na++
			break
		}
		_, size = utf8.DecodeRune(s)
		a[na] = s[0:size]
		s = s[size:]
		na++
	}
	return a[0:na]
}

// Count counts the number of non-overlapping instances of sep in s.
func Count(s, sep []byte) int {
	if len(sep) == 0 {
		return utf8.RuneCount(s) + 1
	}
	c := sep[0]
	n := 0
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i] == c && (len(sep) == 1 || Equal(s[i:i+len(sep)], sep)) {
			n++
			i += len(sep) - 1
		}
	}
	return n
}

// Index returns the index of the first instance of sep in s, or -1 if sep is not present in s.
func Index(s, sep []byte) int {
	n := len(sep)
	if n == 0 {
		return 0
	}
	c := sep[0]
	for i := 0; i+n <= len(s); i++ {
		if s[i] == c && (n == 1 || Equal(s[i:i+n], sep)) {
			return i
		}
	}
	return -1
}

// IndexByte returns the index of the first instance of c in s, or -1 if c is not present in s.
func IndexByte(s []byte, c byte) int // asm_$GOARCH.s

func indexBytePortable(s []byte, c byte) int {
	for i, b := range s {
		if b == c {
			return i
		}
	}
	return -1
}

// LastIndex returns the index of the last instance of sep in s, or -1 if sep is not present in s.
func LastIndex(s, sep []byte) int {
	n := len(sep)
	if n == 0 {
		return len(s)
	}
	c := sep[0]
	for i := len(s) - n; i >= 0; i-- {
		if s[i] == c && (n == 1 || Equal(s[i:i+n], sep)) {
			return i
		}
	}
	return -1
}

// IndexAny interprets s as a sequence of UTF-8 encoded Unicode code points.
// It returns the byte index of the first occurrence in s of any of the Unicode
// code points in chars.  It returns -1 if chars is empty or if there is no code
// point in common.
func IndexAny(s []byte, chars string) int {
	if len(chars) > 0 {
		var rune, width int
		for i := 0; i < len(s); i += width {
			rune = int(s[i])
			if rune < utf8.RuneSelf {
				width = 1
			} else {
				rune, width = utf8.DecodeRune(s[i:])
			}
			for _, r := range chars {
				if rune == r {
					return i
				}
			}
		}
	}
	return -1
}

// Generic split: splits after each instance of sep,
// including sepSave bytes of sep in the subarrays.
func genSplit(s, sep []byte, sepSave, n int) [][]byte {
	if len(sep) == 0 {
		return explode(s, n)
	}
	if n <= 0 {
		n = Count(s, sep) + 1
	}
	c := sep[0]
	start := 0
	a := make([][]byte, n)
	na := 0
	for i := 0; i+len(sep) <= len(s) && na+1 < n; i++ {
		if s[i] == c && (len(sep) == 1 || Equal(s[i:i+len(sep)], sep)) {
			a[na] = s[start : i+sepSave]
			na++
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	a[na] = s[start:]
	return a[0 : na+1]
}

// Split splits the array s around each instance of sep, returning an array of subarrays of s.
// If sep is empty, Split splits s after each UTF-8 sequence.
// If n > 0, Split splits s into at most n subarrays; the last subarray will contain an unsplit remainder.
func Split(s, sep []byte, n int) [][]byte { return genSplit(s, sep, 0, n) }

// SplitAfter splits the array s after each instance of sep, returning an array of subarrays of s.
// If sep is empty, SplitAfter splits s after each UTF-8 sequence.
// If n > 0, SplitAfter splits s into at most n subarrays; the last subarray will contain an
// unsplit remainder.
func SplitAfter(s, sep []byte, n int) [][]byte {
	return genSplit(s, sep, len(sep), n)
}

// Fields splits the array s around each instance of one or more consecutive white space
// characters, returning a slice of subarrays of s or an empty list if s contains only white space.
func Fields(s []byte) [][]byte {
	n := 0
	inField := false
	for i := 0; i < len(s); {
		rune, size := utf8.DecodeRune(s[i:])
		wasInField := inField
		inField = !unicode.IsSpace(rune)
		if inField && !wasInField {
			n++
		}
		i += size
	}

	a := make([][]byte, n)
	na := 0
	fieldStart := -1
	for i := 0; i <= len(s) && na < n; {
		rune, size := utf8.DecodeRune(s[i:])
		if fieldStart < 0 && size > 0 && !unicode.IsSpace(rune) {
			fieldStart = i
			i += size
			continue
		}
		if fieldStart >= 0 && (size == 0 || unicode.IsSpace(rune)) {
			a[na] = s[fieldStart:i]
			na++
			fieldStart = -1
		}
		if size == 0 {
			break
		}
		i += size
	}
	return a[0:na]
}

// Join concatenates the elements of a to create a single byte array.   The separator
// sep is placed between elements in the resulting array.
func Join(a [][]byte, sep []byte) []byte {
	if len(a) == 0 {
		return []byte{}
	}
	if len(a) == 1 {
		return a[0]
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}

	b := make([]byte, n)
	bp := 0
	for i := 0; i < len(a); i++ {
		s := a[i]
		for j := 0; j < len(s); j++ {
			b[bp] = s[j]
			bp++
		}
		if i+1 < len(a) {
			s = sep
			for j := 0; j < len(s); j++ {
				b[bp] = s[j]
				bp++
			}
		}
	}
	return b
}

// HasPrefix tests whether the byte array s begins with prefix.
func HasPrefix(s, prefix []byte) bool {
	return len(s) >= len(prefix) && Equal(s[0:len(prefix)], prefix)
}

// HasSuffix tests whether the byte array s ends with suffix.
func HasSuffix(s, suffix []byte) bool {
	return len(s) >= len(suffix) && Equal(s[len(s)-len(suffix):], suffix)
}

// Map returns a copy of the byte array s with all its characters modified
// according to the mapping function. If mapping returns a negative value, the character is
// dropped from the string with no replacement.  The characters in s and the
// output are interpreted as UTF-8 encoded Unicode code points.
func Map(mapping func(rune int) int, s []byte) []byte {
	// In the worst case, the array can grow when mapped, making
	// things unpleasant.  But it's so rare we barge in assuming it's
	// fine.  It could also shrink but that falls out naturally.
	maxbytes := len(s) // length of b
	nbytes := 0        // number of bytes encoded in b
	b := make([]byte, maxbytes)
	for i := 0; i < len(s); {
		wid := 1
		rune := int(s[i])
		if rune >= utf8.RuneSelf {
			rune, wid = utf8.DecodeRune(s[i:])
		}
		rune = mapping(rune)
		if rune >= 0 {
			if nbytes+utf8.RuneLen(rune) > maxbytes {
				// Grow the buffer.
				maxbytes = maxbytes*2 + utf8.UTFMax
				nb := make([]byte, maxbytes)
				for i, c := range b[0:nbytes] {
					nb[i] = c
				}
				b = nb
			}
			nbytes += utf8.EncodeRune(rune, b[nbytes:maxbytes])
		}
		i += wid
	}
	return b[0:nbytes]
}

// Repeat returns a new byte slice consisting of count copies of b.
func Repeat(b []byte, count int) []byte {
	nb := make([]byte, len(b)*count)
	bp := 0
	for i := 0; i < count; i++ {
		for j := 0; j < len(b); j++ {
			nb[bp] = b[j]
			bp++
		}
	}
	return nb
}

// ToUpper returns a copy of the byte array s with all Unicode letters mapped to their upper case.
func ToUpper(s []byte) []byte { return Map(unicode.ToUpper, s) }

// ToUpper returns a copy of the byte array s with all Unicode letters mapped to their lower case.
func ToLower(s []byte) []byte { return Map(unicode.ToLower, s) }

// ToTitle returns a copy of the byte array s with all Unicode letters mapped to their title case.
func ToTitle(s []byte) []byte { return Map(unicode.ToTitle, s) }

// TrimLeftFunc returns a subslice of s by slicing off all leading UTF-8 encoded
// Unicode code points c that satisfy f(c).
func TrimLeftFunc(s []byte, f func(r int) bool) []byte {
	var start, wid int
	for start = 0; start < len(s); start += wid {
		wid = 1
		rune := int(s[start])
		if rune >= utf8.RuneSelf {
			rune, wid = utf8.DecodeRune(s[start:])
		}
		if !f(rune) {
			break
		}
	}
	return s[start:]
}

// TrimRightFunc returns a subslice of s by slicing off all trailing UTF-8
// encoded Unicode code points c that satisfy f(c).
func TrimRightFunc(s []byte, f func(r int) bool) []byte {
	var end, wid int
	for end = len(s); end > 0; end -= wid {
		wid = 1
		rune := int(s[end-wid])
		if rune >= utf8.RuneSelf {
			// Back up & look for beginning of rune. Mustn't pass start.
			for wid = 2; end-wid >= 0 && !utf8.RuneStart(s[end-wid]); wid++ {
			}
			if end-wid < 0 { // invalid UTF-8 sequence; stop processing
				break
			}
			rune, wid = utf8.DecodeRune(s[end-wid : end])
		}
		if !f(rune) {
			break
		}
	}
	return s[0:end]
}

// TrimFunc returns a subslice of s by slicing off all leading and trailing
// UTF-8 encoded Unicode code points c that satisfy f(c).
func TrimFunc(s []byte, f func(r int) bool) []byte {
	return TrimRightFunc(TrimLeftFunc(s, f), f)
}

func makeCutsetFunc(cutset string) func(rune int) bool {
	return func(rune int) bool {
		for _, c := range cutset {
			if c == rune {
				return true
			}
		}
		return false
	}
}

// Trim returns a subslice of s by slicing off all leading and
// trailing UTF-8 encoded Unicode code points contained in cutset.
func Trim(s []byte, cutset string) []byte {
	return TrimFunc(s, makeCutsetFunc(cutset))
}

// TrimLeft returns a subslice of s by slicing off all leading
// UTF-8 encoded Unicode code points contained in cutset.
func TrimLeft(s []byte, cutset string) []byte {
	return TrimLeftFunc(s, makeCutsetFunc(cutset))
}

// TrimRight returns a subslice of s by slicing off all trailing
// UTF-8 encoded Unicode code points that are contained in cutset.
func TrimRight(s []byte, cutset string) []byte {
	return TrimRightFunc(s, makeCutsetFunc(cutset))
}

// TrimSpace returns a subslice of s by slicing off all leading and
// trailing white space, as as defined by Unicode.
func TrimSpace(s []byte) []byte {
	return TrimFunc(s, unicode.IsSpace)
}

// How big to make a byte array when growing.
// Heuristic: Scale by 50% to give n log n time.
func resize(n int) int {
	if n < 16 {
		n = 16
	}
	return n + n/2
}

// Add appends the contents of t to the end of s and returns the result.
// If s has enough capacity, it is extended in place; otherwise a
// new array is allocated and returned.
func Add(s, t []byte) []byte {
	lens := len(s)
	lent := len(t)
	if lens+lent <= cap(s) {
		s = s[0 : lens+lent]
	} else {
		news := make([]byte, lens+lent, resize(lens+lent))
		copy(news, s)
		s = news
	}
	copy(s[lens:lens+lent], t)
	return s
}

// AddByte appends byte b to the end of s and returns the result.
// If s has enough capacity, it is extended in place; otherwise a
// new array is allocated and returned.
func AddByte(s []byte, t byte) []byte {
	lens := len(s)
	if lens+1 <= cap(s) {
		s = s[0 : lens+1]
	} else {
		news := make([]byte, lens+1, resize(lens+1))
		copy(news, s)
		s = news
	}
	s[lens] = t
	return s
}

// Runes returns a slice of runes (Unicode code points) equivalent to s.
func Runes(s []byte) []int {
	t := make([]int, utf8.RuneCount(s))
	i := 0
	for len(s) > 0 {
		r, l := utf8.DecodeRune(s)
		t[i] = r
		i++
		s = s[l:]
	}
	return t
}
