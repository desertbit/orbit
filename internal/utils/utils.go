/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2018 Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/*
Package utils is the common sin of every Go programmer, including functions that
seem to be usable everywhere, but do not share the same functionality.
*/
package utils

import (
	"crypto/rand"
	"strings"
	"unicode"
)

const (
	// randStrChars defines the possible characters for the RandomString() function.
	// RandomString() expects only ASCII characters, therefore, no char spanning
	// more than 1 byte in UTF-8 encoding must be used.
	randStrChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!?$#:;=&.,"
	// randStrCharsLen stores the length of randStrChars for performance reasons.
	randStrCharsLen = byte(len(randStrChars))
)

// RandomString generates a random string with len n using the crypto/rand RNG.
// The returned string contains only chars defined in the randStrChars constant.
func RandomString(n uint) (string, error) {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = randStrChars[b%randStrCharsLen]
	}
	return string(bytes), nil
}

// NoTitle performs the opposite operation of the strings.Title() func.
// It ensures the first char of the given string is lowercase.
func NoTitle(s string) string {
	done := false
	return strings.Map(func(r rune) rune {
		if !done {
			done = true
			return unicode.ToLower(r)
		}
		return r
	}, s)
}
