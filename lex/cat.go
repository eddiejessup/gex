package lex

import (
	"fmt"
	"github.com/eddiejessup/gnex/read"
)

type CatterError struct {
	msg string
}

func (p CatterError) Error() string {
	return p.msg
}

type CatCode string

const (
	Escape      CatCode = "Escape"
	BeginGroup  CatCode = "BeginGroup"
	EndGroup    CatCode = "EndGroup"
	MathShift   CatCode = "MathShift"
	AlignTab    CatCode = "AlignTab"
	EndOfLine   CatCode = "EndOfLine"
	Parameter   CatCode = "Parameter"
	Superscript CatCode = "Superscript"
	Subscript   CatCode = "Subscript"
	Ignored     CatCode = "Ignored"
	Space       CatCode = "Space"
	Letter      CatCode = "Letter"
	Other       CatCode = "Other"
	Active      CatCode = "Active"
	Comment     CatCode = "Comment"
	Invalid     CatCode = "Invalid"
)

type Catter struct {
	reader     read.FancyByteReader
	CatCodeMap map[byte]CatCode
}

type CharCat struct {
	Char       byte
	Cat        CatCode
	ReaderName string
	Position   int
	LineNr     int
	ColNr      int
	Length     int
}

func NewCatter(r read.FancyByteReader, catCodes map[byte]CatCode) *Catter {
	return &Catter{reader: r, CatCodeMap: catCodes}
}

func (p *Catter) CharToCat(char byte) (cat CatCode, err error) {
	cat, ok := p.CatCodeMap[char]
	if !ok {
		err = CatterError{msg: fmt.Sprintf("Character has no category assigned: '%#U'", char)}
	}
	return
}

func (p *Catter) ReadCharCat() (cc CharCat, err error) {
	fancyByte, err := p.reader.ReadFancyByte()
	if err != nil {
		return
	}
	cat, err := p.CharToCat(fancyByte.B)
	if err != nil {
		return
	}
	cc = CharCat{
		Char:       fancyByte.B,
		Cat:        cat,
		ReaderName: fancyByte.ReaderName,
		Position:   fancyByte.Position,
		LineNr:     fancyByte.LineNr,
		ColNr:      fancyByte.ColNr,
		Length: 1,
	}
	return
}

func (p *Catter) PeekCharCat(n int) (char byte, cat CatCode, err error) {
	char, err = p.reader.PeekByte(n)
	if err != nil {
		return
	}
	cat, err = p.CharToCat(char)
	if err != nil {
		return
	}
	return
}

func (p *Catter) peekCharCatTrio() (char byte, cat CatCode, triod bool, err error) {
	char1, cat1, err := p.PeekCharCat(1)
	if err != nil {
		return
	}
	char2, cat2, err2 := p.PeekCharCat(2)
	char3, cat3, err3 := p.PeekCharCat(3)

	// For trioing to be happening, requires all of:
	// - Can peek to next three CharCats without an error such as end-of-file.
	// - Next two characters have category 'superscript'
	// - Next two characters are the same character
	// - Third character does not have category 'end-of-line'
	triod = (err2 == nil && err3 == nil &&
		cat1 == Superscript && cat2 == cat1 &&
		char2 == char1 && cat3 != EndOfLine)
	if triod {
		char = char3
		if char >= 64 {
			char -= 64
		} else {
			char += 64
		}
		cat, err = p.CharToCat(char)
		if err != nil {
			return
		}
	} else {
		char, cat = char1, cat1
	}
	return
}

func (p *Catter) ReadCharCatTrio() (cc CharCat, err error) {
	char, cat, triod, err := p.peekCharCatTrio()
	// Above function only peeks, so now actually advance by the correct number
	// of bytes.
	cc, err = p.ReadCharCat()
	if triod {
		cc.Char = char
		cc.Cat = cat
		cc.Length = 3

		p.ReadCharCat()
		p.ReadCharCat()
	} else {
		if cc.Char != char || cc.Cat != cat {
			panic("Peeking and reading did not return same results")
		}
	}
	return
}

func (p *Catter) PeekCharCatTrio() (char byte, cat CatCode, err error) {
	char, cat, _, err = p.peekCharCatTrio()
	return
}
