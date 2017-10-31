package main

import (
    "fmt"
    // "io/ioutil"
    "github.com/eddiejessup/gnex/read"
    "github.com/eddiejessup/gnex/lex"
)


func readTest() {
    r := read.NestedByteReaderFromPath("outer.txt")
    r.Insert(read.NestedByteReaderFromPath("inner.txt"))

    v, err := r.PeekByte(1)
    if err != nil {
        panic(err.Error())
    } else {
        fmt.Printf("Peeked %#U\n", v)
    }

    fmt.Println()

    for {
        v, err := r.ReadByte()
        if err != nil {
            break
        }
        fmt.Printf("%v:%v\t%#U, %v\n", r.LineNr, r.ColNr, v, err)
    }
}

func defaultCatCodes() map[byte]lex.CatCode {
    catCodes := make(map[byte]lex.CatCode)
    for i := byte(0); i < 128; i++ {
        var cat lex.CatCode
        switch {
            case i == '\\':
                cat = lex.Escape
            case i == ' ':
                cat = lex.Space
            case i == '%':
                cat = lex.Comment
            // Null.
            case i == 0:
                cat = lex.Ignored
            // Line feed.
            // NON-STANDARD
            case i == 10:
                cat = lex.EndOfLine
            // Carriage return.
            case i == 13:
                cat = lex.EndOfLine
            // Delete.
            case i == 127:
                cat = lex.Invalid
            // ASCII letter.
            case ((i >= 65) && (i <= 90)) || ((i >= 97) && (i <= 122)):
                cat = lex.Letter
            default:
                cat = lex.Other
        }
        catCodes[i] = cat
    }
    catCodes['^'] = lex.Superscript
    return catCodes
}

func catterTest() {
    r := read.NestedByteReaderFromPath("outer.txt")
    catCodes := defaultCatCodes()
    catter := lex.NewCatter(r, catCodes)

    for {
        charCat, err := catter.ReadCharCatTrio()
        // _, err := catter.ReadCharCatTrio()
        if err != nil {
            fmt.Println(err)
            break
        }
        fmt.Printf("%#U, %v, %v\n", charCat.Char, charCat.Cat, err)
    }
}

func lexerTest() {
    r := read.NestedByteReaderFromPath("outer.txt")
    catCodes := defaultCatCodes()
    catter := lex.NewCatter(r, catCodes)
    lexer := lex.NewLexer(*catter)

    for {
        tok, err := lexer.ReadToken()
        if err != nil {
            fmt.Println(err)
            break
        }
        fmt.Printf("%v %v\n", tok, err)
    }
}

type Result interface {

}

type YaLexer struct {
    lexer lex.Lexer
    result Result
}


func (ya *YaLexer) Lex(lval *yySymType) int {
    if ya.result != nil {
        return 0
    }
    tok, err := ya.lexer.ReadToken()
    if err != nil {
        return 0
    }
    if call, ok := tok.(lex.ControlSequenceCall); ok {
        lval.valCall = call
        switch call.Name {
            case "end":
                return END
            case "par":
                return PAR
            default:
                return CONTROL_SEQUENCE
        }
    } else if cc, ok := tok.(lex.CharCat); ok {
        lval.valCharCat = cc
        return CHAR_CAT
    } else {
        panic("Unknown token type")
    }
}

func (l *YaLexer) Error(s string) {
    fmt.Printf("syntax error: %s\n", s)
}



func yaccTest() {
    r := read.NestedByteReaderFromPath("parsetest.txt")
    catCodes := defaultCatCodes()
    catter := lex.NewCatter(r, catCodes)
    lexer := lex.NewLexer(*catter)

    for {
        parser := yyNewParser()
        yaLexer := &YaLexer{lexer: *lexer}
        parser.Parse(yaLexer)
        fmt.Printf("%v\n", yaLexer.result)
        if yaLexer.result == nil {
            break
        }
    }

}


func main() {
    // catterTest()
    // lexerTest()
    yaccTest()
}
