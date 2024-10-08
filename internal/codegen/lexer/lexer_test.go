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

package lexer_test

import (
	"testing"

	"github.com/desertbit/orbit/internal/codegen/lexer"
	r "github.com/stretchr/testify/require"
)

func TestLexer_Next(t *testing.T) {
	t.Run("ok", testLextNextOk)
	t.Run("fail", testLexerNextErr)
}

func testLextNextOk(t *testing.T) {
	t.Parallel()

	const input = ` this 
is some e-x123ample{ text


[ to test ]{: =: } the*
    tokenizer   汉字 }haha
` + "`[hello==:{}-[]  `" + ` hee
` + "``" + ` yaa
service{ asynct async ` + "maxArgSize``:" + `
-58 1ms 0 85MiB 999 0B 0ns // async 99 =: 汉
// service{}:
*url
/*service // bla blub 
999 // /**/
*/errors: iAmAnError, anotherError,`

	cases := []struct {
		val  string
		typ  lexer.TokenType
		line int
		col  int
	}{
		{val: "this", typ: lexer.IDENT, line: 1, col: 2},
		{val: "is", typ: lexer.IDENT, line: 2, col: 1},
		{val: "some", typ: lexer.IDENT, line: 2, col: 4},
		{val: "e-x123ample", typ: lexer.IDENT, line: 2, col: 9},
		{val: "{", typ: lexer.LBRACE, line: 2, col: 20},
		{val: "text", typ: lexer.IDENT, line: 2, col: 22},
		{val: "[", typ: lexer.LBRACK, line: 5, col: 1},
		{val: "to", typ: lexer.IDENT, line: 5, col: 3},
		{val: "test", typ: lexer.IDENT, line: 5, col: 6},
		{val: "]", typ: lexer.RBRACK, line: 5, col: 11},
		{val: "{", typ: lexer.LBRACE, line: 5, col: 12},
		{val: ":", typ: lexer.COLON, line: 5, col: 13},
		{val: "=", typ: lexer.EQUAL, line: 5, col: 15},
		{val: ":", typ: lexer.COLON, line: 5, col: 16},
		{val: "}", typ: lexer.RBRACE, line: 5, col: 18},
		{val: "the", typ: lexer.IDENT, line: 5, col: 20},
		{val: "*", typ: lexer.ASTERISK, line: 5, col: 23},
		{val: "tokenizer", typ: lexer.IDENT, line: 6, col: 5},
		{val: "汉字", typ: lexer.IDENT, line: 6, col: 17},
		{val: "}", typ: lexer.RBRACE, line: 6, col: 20},
		{val: "haha", typ: lexer.IDENT, line: 6, col: 21},
		{val: "[hello==:{}-[]  ", typ: lexer.RAWSTRING, line: 7, col: 2},
		{val: "hee", typ: lexer.IDENT, line: 7, col: 20},
		{val: "", typ: lexer.RAWSTRING, line: 8, col: 2},
		{val: "yaa", typ: lexer.IDENT, line: 8, col: 4},
		{val: "service", typ: lexer.SERVICE, line: 9, col: 1},
		{val: "{", typ: lexer.LBRACE, line: 9, col: 8},
		{val: "asynct", typ: lexer.IDENT, line: 9, col: 10},
		{val: "async", typ: lexer.ASYNC, line: 9, col: 17},
		{val: "maxArgSize", typ: lexer.MAXARGSIZE, line: 9, col: 23},
		{val: "", typ: lexer.RAWSTRING, line: 9, col: 34},
		{val: ":", typ: lexer.COLON, line: 9, col: 35},
		{val: "-58", typ: lexer.INT, line: 10, col: 1},
		{val: "1ms", typ: lexer.IDENT, line: 10, col: 5},
		{val: "0", typ: lexer.INT, line: 10, col: 9},
		{val: "85MiB", typ: lexer.IDENT, line: 10, col: 11},
		{val: "999", typ: lexer.INT, line: 10, col: 17},
		{val: "0B", typ: lexer.IDENT, line: 10, col: 21},
		{val: "0ns", typ: lexer.IDENT, line: 10, col: 24},
		{val: "*", typ: lexer.ASTERISK, line: 12, col: 1},
		{val: "url", typ: lexer.IDENT, line: 12, col: 2},
		{val: "errors", typ: lexer.ERRORS, line: 15, col: 3},
		{val: ":", typ: lexer.COLON, line: 15, col: 9},
		{val: "iAmAnError", typ: lexer.IDENT, line: 15, col: 11},
		{val: ",", typ: lexer.COMMA, line: 15, col: 21},
		{val: "anotherError", typ: lexer.IDENT, line: 15, col: 23},
		{val: ",", typ: lexer.COMMA, line: 15, col: 35},
	}

	l := lexer.Lex(input)
	for i, c := range cases {
		tk := l.Next()
		if tk.Type == lexer.EOF {
			// Test for premature EOF.
			r.True(t, i == len(cases)-1, "case %d", i)
			return
		}

		r.Exactly(t, c.val, tk.Value, "case %d", i)
		r.Exactly(t, c.line, tk.Pos.Line, "case %d", i)
		r.Exactly(t, c.col, tk.Pos.Column, "case %d", i)
	}
}

func testLexerNextErr(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		line  int
		col   int
	}{
		{input: "`\ntest`", line: 1, col: 2}, // 0
		{input: "`test", line: 1, col: 2},
		{input: "88MiB6", line: 1, col: 1},
		{input: "-", line: 1, col: 1},
		{input: "88.6", line: 1, col: 1},
	}

	for i, c := range cases {
		l := lexer.Lex(c.input)
		tk := l.Next()
		r.Exactly(t, lexer.ILLEGAL, tk.Type, "case %d", i)
		r.Exactly(t, c.line, tk.Pos.Line, "case %d", i)
		r.Exactly(t, c.col, tk.Pos.Column, "case %d", i)
	}
}
