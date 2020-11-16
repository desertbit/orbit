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

package lexer

// TokenType identifies the type of lex tokens.
type TokenType int

const (
	// Special Tokens

	ILLEGAL TokenType = iota
	EOF

	// Literals

	literalBegin
	IDENT     // sendData
	INT       // 123
	RAWSTRING // `rawstring`
	BYTESIZE  // 1MB
	DURATION  // 80ns
	literalEnd

	// Keywords

	keywordBegin
	VERSION
	ERRORS
	ENUM
	TYPE
	SERVICE
	CALL
	STREAM
	ASYNC
	ARG
	RET
	MAXARGSIZE
	MAXRETSIZE
	TIMEOUT
	MAP
	keywordEnd

	// Delimiters

	delimBegin
	LPAREN // (
	RPAREN // )
	LBRACE // {
	RBRACE // }
	LBRACK // [
	RBRACK // ]
	COLON  // :
	EQUAL  // =
	delimEnd
)

func (tt TokenType) String() string {
	if tt == ILLEGAL {
		return "illegal"
	} else if tt == EOF {
		return "eof"
	} else if literalBegin < tt && tt < literalEnd {
		switch tt {
		case IDENT:
			return "ident"
		case INT:
			return "int"
		case RAWSTRING:
			return "rawstring"
		case BYTESIZE:
			return "bytesize"
		case DURATION:
			return "duration"
		}
	} else if keywordBegin < tt && tt < keywordEnd {
		for k, v := range keywordTokenTypes {
			if v == tt {
				return k
			}
		}
	} else {
		for k, v := range delimTokenTypes {
			if v == tt {
				return string(k)
			}
		}
	}

	return "__UNKNOWN__"
}

// Token is an item returned from the scanner.
type Token struct {
	Type  TokenType
	Value string
	Pos   Pos
}

// IsLiteral returns true, if the token has
// type literal.
func (t Token) IsLiteral() bool {
	return literalBegin < t.Type && t.Type < literalEnd
}

// IsKeyword returns true, if the token has
// type keyword.
func (t Token) IsKeyword() bool {
	return keywordBegin < t.Type && t.Type < keywordEnd
}

// IsDelimiter returns true, if the token has
// type delimiter.
func (t Token) IsDelimiter() bool {
	return delimBegin < t.Type && t.Type < delimEnd
}

// Pos describes a position in a text file.
type Pos struct {
	Line   int
	Column int
}

const (
	// Only needed during lexing, no tokens.
	eof      rune = -1
	hyphen        = '-'
	backtick      = '`'
	newline       = '\n'
	slash         = '/'
	asterisk      = '*'
)

var keywordTokenTypes = map[string]TokenType{
	"version":    VERSION,
	"errors":     ERRORS,
	"enum":       ENUM,
	"type":       TYPE,
	"service":    SERVICE,
	"call":       CALL,
	"stream":     STREAM,
	"async":      ASYNC,
	"arg":        ARG,
	"ret":        RET,
	"maxArgSize": MAXARGSIZE,
	"maxRetSize": MAXRETSIZE,
	"timeout":    TIMEOUT,
	"map":        MAP,
}

func toKeywordTokenType(s string) TokenType {
	t, ok := keywordTokenTypes[s]
	if !ok {
		return ILLEGAL
	}
	return t
}

var delimTokenTypes = map[rune]TokenType{
	'(': LPAREN,
	')': RPAREN,
	'{': LBRACE,
	'}': RBRACE,
	'[': LBRACK,
	']': RBRACK,
	':': COLON,
	'=': EQUAL,
}

func isDelim(r rune) (ok bool) {
	_, ok = delimTokenTypes[r]
	return
}

func toDelimTokenType(r rune) TokenType {
	t, ok := delimTokenTypes[r]
	if !ok {
		return ILLEGAL
	}
	return t
}
