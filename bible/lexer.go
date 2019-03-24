package bible

import (
	"bufio"
	"bytes"
	"io"
)

//go:generate stringer -type=Token
type Token int

const (
	ILLEGAL Token = iota
	BOOK
	NEXT_NUM
	COMMA
	DASH
	EOF
	WS // WhiteSpace

	eof rune = rune(0)
)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isComma(ch rune) bool {
	return ch == ','
}

func isDash(ch rune) bool {
	return (ch == '-' || ch == '—')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) peekNRunes(n int) (string, error) {
	// usually rune is 2 bytes
	b, err := s.r.Peek(n)
	if err != nil {
		return string(b), err
	}
	return string(b), nil
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() { _ = s.r.UnreadRune() }

func (s *Scanner) Scan() (tok Token, lit string) {
	// Read next rune.
	var ch rune
	st, err := s.peekNRunes(1)
	if err != nil {
		ch = eof
	} else {
		ch = rune(st[0])
	}

	switch {
	case isWhitespace(ch):
		return s.scanWhitespace()
	case isLetter(ch):
		return s.scanBook()
	case isDigit(ch):
		// Maybe book starts with number...
		st, err := s.peekNRunes(2)
		if err == nil {
			nextch := rune(st[1])
			if isLetter(nextch) {
				return s.scanBook()
			}
		}
		return s.scanNextNum()
	}

	s.read()
	switch {
	case ch == eof:
		return EOF, ""
	case isComma(ch):
		return COMMA, string(ch)
	case isDash(ch):
		return DASH, string(ch)
	}
	return ILLEGAL, string(ch)
}

// scanWhitespace - consume all whitespaces
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

Loop:
	for {
		ch := s.read()
		switch {
		case ch == eof:
			break Loop
		case !isWhitespace(ch):
			s.unread()
			break Loop
		default:
			_, _ = buf.WriteRune(ch)
		}
	}
	return WS, buf.String()
}

func (s *Scanner) scanBook() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

Loop:
	for {
		ch := s.read()
		switch {
		case ch == eof:
			break Loop
		case !isLetter(ch) && !isDigit(ch):
			s.unread()
			break Loop
		default:
			_, _ = buf.WriteRune(ch)
		}
	}
	return BOOK, buf.String()
}

func (s *Scanner) scanNextNum() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

Loop:
	for {
		ch := s.read()
		switch {
		case ch == eof:
			break Loop
		case !isDigit(ch):
			s.unread()
			break Loop
		default:
			_, _ = buf.WriteRune(ch)
		}
	}
	return NEXT_NUM, buf.String()
}
