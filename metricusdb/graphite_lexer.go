package metricusdb

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Pos int

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemBool                  // boolean constant
	itemEOF
	itemIdentifier // alphanumeric identifier not starting with '.'
	itemLeftParen  // '(' inside action
	itemNumber     // simple number, including imaginary
	itemComplex
	itemRightParen // ')' inside action
	itemSeparator  // Comma with any number of spaces on each side
	itemString     // quoted string (includes quotes)
	itemFunction   //Graphite functions
	itemTarget     //Graphite series target
)

var itemLookup = map[itemType]string{
	itemFunction:   `itemFunction`,
	itemTarget:     `itemTarget`,
	itemEOF:        `itemEOF`,
	itemLeftParen:  `itemLeftParen`,
	itemRightParen: `itemRightParen`,
	itemError:      `itemError`,
	itemString:     `itemString`,
	itemNumber:     `itemNumber`,
	itemSeparator:  `itemSeparator`,
}

var key = map[string]itemType{}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	leftDelim  string    // start of action
	rightDelim string    // end of action
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent item returned by nextItem
	items      chan item // channel of scanned items
	parenDepth int       // nesting depth of ( ) exprs
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// lex creates a new scanner for the input string.
func lex(name, input, left, right string) *lexer {
	if left == "" {
		left = leftDelim
	}
	if right == "" {
		right = rightDelim
	}
	l := &lexer{
		name:       name,
		input:      input,
		leftDelim:  left,
		rightDelim: right,
		items:      make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexTarget; l.state != nil; {
		l.state = l.state(l)
	}
}

// state functions

const (
	leftDelim    = "{{"
	rightDelim   = "}}"
	leftComment  = "/*"
	rightComment = "*/"
)

func lexTarget(l *lexer) stateFn {
	for {
		r := l.next()
		if r == eof || isEndOfLine(r) {
			l.emit(itemTarget)
			l.emit(itemEOF)
			break
		}
		switch {
		case isSeparator(r):
			l.backup()
			l.emit(itemTarget)
			return lexInsideFunction
		case r == '(':
			fmt.Println("Emitting function")
			l.backup()
			l.emit(itemFunction)
			return lexInsideFunction
		case r == ')':
			l.backup()
			l.emit(itemTarget)
			return lexInsideFunction
		}
	}
	return nil
}

func lexInsideFunction(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || isEndOfLine(r):
			if l.parenDepth == 0 {
				l.emit(itemEOF)
				return nil
			} else {
				return l.errorf("unclosed action")
			}
		case r == '+' || r == '-' || ('0' <= r && r <= '9'):
			l.backup()
			return lexNumber
		case r == '\'' || r == '"':
			l.ignore()
			return lexQuote
		case isSeparator(r):
			return lexSeparator
		case r == '(':
			l.emit(itemLeftParen)
			l.parenDepth++
			return lexInsideFunction
		case r == ')':
			l.emit(itemRightParen)
			l.parenDepth--
			if l.parenDepth < 0 {
				return l.errorf("unexpected right paren %#U", r)
			}
			return lexInsideFunction
		case isAlphaNumeric(r):
			l.backup()
			return lexTarget
		}
	}
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSeparator(l *lexer) stateFn {
	for isSeparator(l.peek()) {
		l.next()
	}
	l.emit(itemSeparator)
	return lexInsideFunction
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	if sign := l.peek(); sign == '+' || sign == '-' {
		// Complex: 1+2i. No spaces, must end in 'i'.
		if !l.scanNumber() || l.input[l.pos-1] != 'i' {
			return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
		}
		l.emit(itemComplex)
	} else {
		l.emit(itemNumber)
	}
	return lexInsideFunction
}

func (l *lexer) scanNumber() bool {
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	// Is it imaginary?
	l.accept("i")
	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return false
	}
	return true
}

// lexQuote scans a quoted string.
func lexQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '"', '\'':
			break Loop
		}
	}
	l.backup()
	l.emit(itemString)
	l.next()
	l.ignore()
	return lexInsideFunction
}

// isSpace reports whether r is a space character.
func isSeparator(r rune) bool {
	return r == ' ' || r == '\t' || r == ','
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
