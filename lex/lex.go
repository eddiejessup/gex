package lex

import (
	"fmt"
)

type LexError struct {
	msg string
}

func (p LexError) Error() string {
	return p.msg
}

type ReadState string

const (
	LineBegin      ReadState = "LineBegin"
	LineMiddle     ReadState = "LineMiddle"
	SkippingBlanks ReadState = "SkippingBlanks"
)
type Token interface {
}

type ControlSequenceCall struct {
	Name string
	ReaderName string
	Position int
	LineNr int
	ColNr      int
	Length     int
}

type Lexer struct {
	catter    Catter
	readState ReadState
}

func NewLexer(r Catter) *Lexer {
	return &Lexer{r, LineBegin}
}

var tokeniseCats = map[CatCode]bool {
    BeginGroup: true,
    EndGroup: true,
    MathShift: true,
    AlignTab: true,
    Parameter: true,
    Superscript: true,  
    Subscript: true,
    Letter: true,
    Other: true,
    Active: true,
}

func (p *Lexer) ReadToken() (tok Token, err error) {
	for {
		cc, err := p.catter.ReadCharCatTrio()
		if err != nil {
			return cc, err
		}
		switch {
			case cc.Cat == Comment:
				for {
					ccComment, err := p.catter.ReadCharCat()
					if err != nil || ccComment.Cat == EndOfLine {
						break
					}
				}
				p.readState = LineBegin
			case cc.Cat == Escape:
				tokLength := cc.Length
				ccNameFirst, err := p.catter.ReadCharCatTrio()
				// If escape character is at end of file, no idea what to do.
				if err != nil {
					return tok, err
				}
				tokLength += ccNameFirst.Length
				callName := string(ccNameFirst.Char)
            	// If first character of call is non-letter, make a control
            	// sequence of that single character.
				if ccNameFirst.Cat == Space {
					p.readState = SkippingBlanks
				} else if ccNameFirst.Cat != Letter {
					p.readState = LineMiddle
            	// If first character of call is letter, keep reading control
            	// sequence name until we get a non-letter.
				} else {
					for {
            			// Peek to see if next (possibly triod) character is a
            			// letter.
						_, catNext, err := p.catter.PeekCharCatTrio()
            			// (If we exhaust reader, then finish the control
            			// sequence.)
						if err != nil {
							break
						}
						// If it is a letter, read it and add it to the list of
            			// control sequence characters.
						if catNext == Letter {
							ccNameNext, err := p.catter.ReadCharCatTrio()
							if err != nil {
								break
							}
							callName += string(ccNameNext.Char)
							tokLength += ccNameNext.Length
						} else {
							break
						}
					}
					p.readState = SkippingBlanks
				}
				tok := ControlSequenceCall{
					Name: callName,
					ReaderName: cc.ReaderName,
					Position: cc.Position,
					LineNr: cc.LineNr,
					ColNr: cc.ColNr,
					Length: tokLength,
				}
				return tok, nil
			case tokeniseCats[cc.Cat]:
				p.readState = LineMiddle
				return cc, nil
	        // If TeX sees a character of category [space], the action
	        // depends on the current state.
			case cc.Cat == Space:
				// If TeX is in state [new line] or [skipping blanks],
		        // The character is simply passed by, and TeX remains in the same
		        // state.
		        // Otherwise TeX is in state [line middle].
		        // The character is converted to a token of category [space] whose
		        // character code is [' '], and TeX enters state [skipping blanks].
		        // The character code in a space token is always [' '].
				if p.readState == LineMiddle {
					cc.Char = ' '
					p.readState = SkippingBlanks
					return cc, nil
				}
			case cc.Cat == EndOfLine:
	        	// [...] if TeX is in state [new line],
	    		// the end-of-line character is converted to the control
	    		// sequence token 'par' (end of paragraph).
				if p.readState == LineBegin {
					tok = ControlSequenceCall{
						Name: "par",
						ReaderName: cc.ReaderName,
						Position: cc.Position,
						LineNr: cc.LineNr,
						ColNr: cc.ColNr,
						Length: cc.Length,
					}
					return tok, nil
	        	// if TeX is in state [mid-line],
	    		// the end-of-line character is converted to a token for
	    		// character [' '] of category [space].
				} else if p.readState == LineMiddle {
					cc.Char = ' '
					cc.Cat = Space
		        	// "At the beginning of every line [TeX is] in state [new line]".
	            	p.readState = LineBegin
					return cc, nil
	        	// If TeX is in state [skipping blanks],
	    		// the end-of-line character is simply dropped.				
				} else if p.readState == SkippingBlanks {
				} else {
					panic(fmt.Sprintf("Unknown Read State %v", p.readState))
				}
			default:
				panic(fmt.Sprintf("Unknown category '%v'", cc.Cat))
		}

	}
}
