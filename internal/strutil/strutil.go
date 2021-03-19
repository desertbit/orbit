/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
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
Package strutil is the common sin of every Go programmer.
It contains utility functions around strings.
*/
package strutil

import (
	"crypto/rand"
	"unicode"
)

const (
	// randStrChars defines the possible characters for the RandomString() function.
	// RandomString() expects only ASCII characters, therefore, no char spanning
	// more than 1 byte in UTF-8 encoding must be used.
	randStrChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// randStrCharsLen stores the length of randStrChars.
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

// FirstUpper returns a copy of s, where the first rune is
// guaranteed to be an uppercase letter, like unicode.ToUpper
// suggests.
func FirstUpper(s string) string {
	for i, v := range s {
		return string(unicode.ToUpper(v)) + s[i+1:]
	}
	return ""
}

// FirstLower returns a copy of s, where the first rune is
// guaranteed to be a lowercase letter, like unicode.ToLower
// suggests.
func FirstLower(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return ""
}
