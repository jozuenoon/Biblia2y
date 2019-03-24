package bible

import (
	"fmt"
	"io"
)

//go:generate stringer -type=State
type State int

const (
	S_START State = iota
	S_BOOK
	S_CHAPTER
	S_VERSE
	S_NEXT_NUM
	S_NEXT_VERSE
	S_END
)

type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max = 1)
	}
	state     State
	lastState State

	versePointer int

	// The list of references that were discovered.
	verses []string

	bsvc Service
}

func NewParser(r io.Reader, bsvc Service) *Parser {
	return &Parser{s: NewScanner(r), bsvc: bsvc}
}

func (p *Parser) scan() (tok Token, lit string) {
	// If we have token on the buffer then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read next token from scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) updateVerse(idxNum string) {
	v := p.verses[p.versePointer]
	v += p.expandWithZeros(idxNum)
	p.verses[p.versePointer] = v
}

func (p *Parser) formatedBookNumber(bookName string) (string, error) {
	bookNum, err := p.bsvc.GetBookNumber(bookName)
	if err != nil {
		return "", err
	}
	bookNumString := fmt.Sprintf("%03d", bookNum)

	if len(bookNumString) != 3 {
		return "", fmt.Errorf("expected 3 digit book number: %s", bookNumString)
	}

	return bookNumString, nil
}

func (p *Parser) nextVerse() {
	p.verses = append(p.verses, "")
	p.versePointer++
}

func (p *Parser) expandWithZeros(lit string) string {
	switch len(lit) {
	case 0:
		return "000"
	case 1:
		return "00" + lit
	case 2:
		return "0" + lit
	default:
		return lit
	}
}

func (p *Parser) Parse() (*Verse, error) {
	p.versePointer = 0

	p.state = S_START
	p.lastState = S_START

	var tok Token
	var lit string
	// Preserve book.
	var currentBookIdx string
	var currentChapterIdx string

	for {
		//fmt.Println(p.state, tok)
		switch {
		case p.state == S_START:
			// Add empty string so it exist.
			p.verses = append(p.verses, "")
		case p.state == S_BOOK:
			// Append book number to index and preserve book
			// for future verses.
			bookNum, err := p.formatedBookNumber(lit)
			if err != nil {
				return nil, err
			}
			p.updateVerse(bookNum)
			currentBookIdx = bookNum
		case p.state == S_NEXT_NUM && p.lastState == S_VERSE && tok == DASH:
			p.nextVerse()
			_, lit = p.scanIgnoreWhitespace()
			p.updateVerse(currentBookIdx)
			toknext, _ := p.scanIgnoreWhitespace()
			if toknext == COMMA {
				p.unscan()
				p.updateVerse(lit) // add chapter index
			} else if toknext == EOF {
				p.updateVerse(currentChapterIdx)
				p.updateVerse(lit) // add verse index
			}
		case p.state == S_CHAPTER && tok == DASH:
			// Look ahead to get next chapter (next verse)
			p.nextVerse()
			_, lit = p.scanIgnoreWhitespace()
			p.updateVerse(currentBookIdx)
			p.updateVerse(lit) // add chapter index
		case p.state == S_CHAPTER:
			p.updateVerse(lit) // add chapter index
			currentChapterIdx = lit
		case p.state == S_VERSE:
			if tok != NEXT_NUM {
				_, lit = p.scanIgnoreWhitespace()
			}
			p.updateVerse(lit) // add verse index
		case p.state == S_END:
			// We expect 1 or 2 verses.
			if len(p.verses) == 1 {
				return p.bsvc.NewVerseFromSingleLabel(Label(p.verses[0]))
			}
			if len(p.verses) >= 2 {
				return p.bsvc.NewVerseFromDualLabel(Label(p.verses[0]), Label(p.verses[1]))
			}
			return nil, fmt.Errorf("parser error, nothing to create")
		}

		tok, lit = p.scanIgnoreWhitespace()

		// Find next state and buffer last one.
		p.lastState = p.state
		p.state = p.NextState(tok, lit)

	}
}

func (p *Parser) NextState(tok Token, lit string) State {

	switch {
	// Book at start.
	case p.state == S_START && tok == BOOK:
		return S_BOOK
	// After BOOK we expect chapter directly.
	case p.state == S_BOOK:
		return S_CHAPTER
	// After chapter if COMMA we will have verse.
	case p.state == S_CHAPTER && tok == COMMA:
		return S_VERSE
	// If there was chapter last time and DASH pop then we have next chapter.
	case p.state == S_CHAPTER && tok == DASH:
		return S_CHAPTER
	// If there was verse and DASH pop then we expect CHAPTER OR VERSE
	case p.state == S_VERSE && tok == DASH:
		return S_NEXT_NUM
	case p.state == S_NEXT_NUM && tok == COMMA:
		return S_VERSE
	case tok == EOF:
		return S_END
	}
	return S_END
}
