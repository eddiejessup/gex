package lex

import (
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

type Lexer struct {
	catter    Catter
	readState ReadState
}

func NewLexer(r Catter) *Lexer {
	return &Lexer{r, LineBegin}
}
