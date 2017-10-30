package main

import (
    "fmt"
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


    // @staticmethod
    // def default_initial_cat_codes():
    //     char_to_cat = get_unset_ascii_char_dict()
    //     char_to_cat.update({c: CatCode.other for c in ascii_characters})
    //     char_to_cat.update({let: CatCode.letter for let in ascii_letters})

    //     char_to_cat['\\'] = CatCode.escape
    //     char_to_cat[' '] = CatCode.space
    //     char_to_cat['%'] = CatCode.comment
    //     char_to_cat[WeirdChar.null.value] = CatCode.ignored
    //     # NON-STANDARD
    //     char_to_cat[WeirdChar.line_feed.value] = CatCode.end_of_line
    //     char_to_cat[WeirdChar.carriage_return.value] = CatCode.end_of_line
    //     char_to_cat[WeirdChar.delete.value] = CatCode.invalid
    //     return char_to_cat
// class WeirdChar(Enum):
//     null = chr(0)
//     line_feed = chr(10)
//     carriage_return = chr(13)
//     delete = chr(127)

func catterTest() {
    r := read.NestedByteReaderFromPath("outer.txt")
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

func main() {
    catterTest()
}
