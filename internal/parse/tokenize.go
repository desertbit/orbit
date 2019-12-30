/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2019 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2019 Sebastian Borchers <sebastian[at]desertbit.com>
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

package parse

import (
	"io/ioutil"
	"unicode"
)

type token struct {
	value string
	line  int
}

func tokenize(filePath string) (tks []*token, err error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	var (
		value    string
		skipNext bool

		lines = 1
		s     = string(data)
		sr    = []rune(s)
	)
	tks = make([]*token, 0)

	for i, r := range s {
		if skipNext {
			skipNext = false
			continue
		}

		if r == '\n' {
			lines++
		} else if unicode.IsSpace(r) {
			if value == "" {
				continue
			}

			tks = append(tks, &token{value: value, line: lines})
			value = ""
		} else if r == '{' || r == '}' || r == '(' || r == ')' || r == '[' || r == ']' {
			if value != "" {
				tks = append(tks, &token{value: value, line: lines})
				value = ""
			}

			// Group array brackets together.
			if r == '[' && i+1 < len(sr) && sr[i+1] == ']' {
				skipNext = true
				tks = append(tks, &token{value: "[]", line: lines})
			} else {
				tks = append(tks, &token{value: string(r), line: lines})
			}
		} else {
			value += string(r)
		}
	}
	return
}
