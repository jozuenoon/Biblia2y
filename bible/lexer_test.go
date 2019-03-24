package bible

import (
	"reflect"
	"strings"
	"testing"
)

func TestScanner_peekNRunes(t *testing.T) {

	testscan1 := NewScanner(strings.NewReader("ap 3,20"))

	b, err := testscan1.peekNRunes(1)
	if err != nil {
		t.Fatal(err)
	}

	if b != "a" {
		t.Fatal("wrong literal want a have", b)
	}

	b, err = testscan1.peekNRunes(2)
	if err != nil {
		t.Fatal(err)
	}

	if b != "ap" {
		t.Fatal("wrong literal want ap have", b)
	}

	b, err = testscan1.peekNRunes(3)
	if err != nil {
		t.Fatal(err)
	}

	if b != "ap " {
		t.Fatal("wrong literal want ap_ have", b)
	}
}

func TestScanner_Scan1(t *testing.T) {

	testscan1 := NewScanner(strings.NewReader("ap 3,20"))

	var tokens []Token
	var literals []string
	for {
		tok, lit := testscan1.Scan()
		if tok == EOF {
			break
		}
		tokens = append(tokens, tok)
		literals = append(literals, lit)
	}

	if !reflect.DeepEqual(tokens, []Token{BOOK, WS, NEXT_NUM, COMMA, NEXT_NUM}) {
		t.Errorf("wrong tokens")
	}
}

func TestScanner_Scan2(t *testing.T) {

	testscan1 := NewScanner(strings.NewReader("ef 4,1-5,20"))

	var tokens []Token
	var literals []string
	for {
		tok, lit := testscan1.Scan()
		if tok == EOF {
			break
		}
		tokens = append(tokens, tok)
		literals = append(literals, lit)
	}

	if !reflect.DeepEqual(tokens, []Token{BOOK, WS, NEXT_NUM, COMMA, NEXT_NUM, DASH, NEXT_NUM, COMMA, NEXT_NUM}) {
		t.Errorf("wrong tokens")
	}
}

func TestScanner_Scan4(t *testing.T) {

	testscan1 := NewScanner(strings.NewReader("ef 4,1-5,2"))

	var tokens []Token
	var literals []string
	for {
		tok, lit := testscan1.Scan()
		if tok == EOF {
			break
		}
		tokens = append(tokens, tok)
		literals = append(literals, lit)
	}

	if !reflect.DeepEqual(tokens, []Token{BOOK, WS, NEXT_NUM, COMMA, NEXT_NUM, DASH, NEXT_NUM, COMMA, NEXT_NUM}) {
		t.Errorf("wrong tokens")
	}
}

func TestScanner_Scan3(t *testing.T) {

	testscan1 := NewScanner(strings.NewReader("dz 9-10"))

	var tokens []Token
	var literals []string
	for {
		tok, lit := testscan1.Scan()
		if tok == EOF {
			break
		}
		tokens = append(tokens, tok)
		literals = append(literals, lit)
	}

	if !reflect.DeepEqual(tokens, []Token{BOOK, WS, NEXT_NUM, DASH, NEXT_NUM}) {
		t.Errorf("wrong tokens")
	}
}

func TestScanner_ScanChapterRange1(t *testing.T) {

	testscan1 := NewScanner(strings.NewReader("1kor 1"))

	var tokens []Token
	var literals []string
	for {
		tok, lit := testscan1.Scan()
		if tok == EOF {
			break
		}
		tokens = append(tokens, tok)
		literals = append(literals, lit)
	}

	if !reflect.DeepEqual(tokens, []Token{BOOK, WS, NEXT_NUM}) {
		t.Errorf("wrong tokens")
	}
}
