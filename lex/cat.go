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
	Char byte
	Cat  CatCode
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
	char, err := p.reader.ReadByte()
	if err != nil {
		return
	}
	cat, err := p.CharToCat(char)
	cc = CharCat{char, cat}
	return
}

func (p *Catter) PeekCharCat(n int) (cc CharCat, err error) {
	char, err := p.reader.PeekByte(n)
	if err != nil {
		return
	}
	cat, err := p.CharToCat(char)
	cc = CharCat{char, cat}
	return
}

func (p *Catter) peekCharCatTrio() (cc CharCat, triod bool, err error) {
	cc1, err1 := p.PeekCharCat(1)
	if err1 != nil {
		return cc, false, err1
	}
	cc2, err2 := p.PeekCharCat(2)
	cc3, err3 := p.PeekCharCat(3)

	// For trioing to be happening, requires all of:
	// - Can peek to next three CharCats without an error such as end-of-file.
	// - Next two characters have category 'superscript'
	// - Next two characters are the same character
	// - Third character does not have category 'end-of-line'
	triod = (err2 == nil && err3 == nil &&
		cc1.Cat == Superscript && cc2.Cat == cc1.Cat &&
		cc2.Char == cc1.Char && cc3.Cat != EndOfLine)
	if triod {
		char := cc3.Char
		if char >= 64 {
			char -= 64
		} else {
			char += 64
		}
		cat, err := p.CharToCat(char)
		if err != nil {
			return cc, triod, err
		}
		cc = CharCat{char, cat}
	} else {
		cc = cc1
	}
	return
}

func (p *Catter) ReadCharCatTrio() (cc CharCat, err error) {
	cc, triod, err := p.peekCharCatTrio()
	// Above function only peeks, so now actually advance by the correct number
	// of bytes.
	if triod {
		p.ReadCharCat()
		p.ReadCharCat()
		p.ReadCharCat()
	} else {
		p.ReadCharCat()
	}
	return
}

func (p *Catter) PeekCharCatTrio() (cc CharCat, err error) {
	cc, _, err = p.peekCharCatTrio()
	return
}
