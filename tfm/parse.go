package main

import (
	"fmt"
	"io"
	"os"
)

type Table int
const (
	Header Table = iota
	CharacterInfo
	Width
	Height
	Depth
	ItalicCorrection
	LigKern
	Kern
	ExtensibleCharacter
	FontParameter
)
const (
	NrTables = FontParameter + 1
	HeaderDataLengthWordsMin = 18
    // The header starts at 24 bytes.
	HeaderPointer = 24
	CharacterCodingSchemeLength = 40
    FamilyLength = 20
)

type MathSymbolParams struct {
	Num1 float64
	Num2 float64
	Num3 float64
	Denom1 float64
	Denom2 float64
	Sup1 float64
	Sup2 float64
	Sup3 float64
	Sub1 float64
	Sub2 float64
	Supdrop float64
	Subdrop float64
	Delim1 float64
	Delim2 float64
	AxisHeight float64
}

type MathExtensionParams struct {
	DefaultRuleThickness float64
	BigOpSpacing [5]float64
}

type TFM struct {
	fileLengthWords uint16
	headerDataLengthWords uint16
	smallestCharCode uint16
	largestCharCode uint16
	tableLengthsWords []uint16
	tablePointers []uint16
	checksum uint32
	designFontSize float64
	characterCodingScheme string
	family string

	slant float64
	spacing float64
	spaceStretch float64
	spaceShrink float64
	xHeight float64
	quad float64
	extraSpace float64

	MathSymbolParams MathSymbolParams
	MathExtensionParams MathExtensionParams
}

func (tfm *TFM) PositionInTable(table Table, indexWords uint16) uint16 {
	return tfm.tablePointers[table] + 4 * indexWords
}

func NewTFM(path string) *TFM {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	fileLengthWords, _ := read2bui(file)
	headerDataLengthWords, _ := read2bui(file)
	smallestCharCode, _ := read2bui(file)
	largestCharCode, _ := read2bui(file)

	// Set table lengths.
	tableLengthsWords := make([]uint16, NrTables, NrTables)
	if headerDataLengthWords < HeaderDataLengthWordsMin {
		headerDataLengthWords = HeaderDataLengthWordsMin
	}
	tableLengthsWords[Header] = headerDataLengthWords

	nrChars := largestCharCode - smallestCharCode + 1
	tableLengthsWords[CharacterInfo] = nrChars

	for table := Width; table < NrTables; table++ {
		tableLength, _ := read2bui(file)
		tableLengthsWords[table] = tableLength
	}

	tablePointers := make([]uint16, NrTables, NrTables)

	tfm := TFM{
		fileLengthWords: fileLengthWords,
		headerDataLengthWords: headerDataLengthWords,
		smallestCharCode: smallestCharCode,
		largestCharCode: largestCharCode,
		tableLengthsWords: tableLengthsWords,
		tablePointers: tablePointers,
		MathSymbolParams: nil,
	}

	// Infer table pointers from table lengths.
	tfm.tablePointers[Header] = HeaderPointer
	for table := Header; table < FontParameter; table++ {
		tfm.tablePointers[table + 1] = tfm.PositionInTable(table, tableLengthsWords[table])
	}

	validationFileLength := tfm.PositionInTable(FontParameter, tableLengthsWords[FontParameter])
    if validationFileLength != fileLengthWords * 4 {
        panic("Bad TFM file")
    }

    // Read header.
	file.Seek(int64(tfm.tablePointers[Header]), io.SeekStart)

    // Read header[0 ... 1].
    tfm.checksum, _ = read4bui(file)
    tfm.designFontSize, _ = readFixWord(file)

    // Read header[2 ... 11] if present.
    position, err := currentPosition(file)
    if (err != nil) {
    	panic(err)
    }
    characterInfoTablePosition := tfm.tablePointers[CharacterInfo]
    if position < characterInfoTablePosition {
        tfm.characterCodingScheme, err = readBCPL(file)
        if (err != nil) {
        	panic(err)
        }
    }

    // Read header[12 ... 16] if present.
    position += CharacterCodingSchemeLength
    if position < characterInfoTablePosition {
        tfm.family, err = readBCPLFrom(file, position)
        if (err != nil) {
        	panic(err)
        }
    }

    // Read header[12 ... 16] if present.
    position += FamilyLength
    if position < characterInfoTablePosition {
        // Seven bit safe flag; unused.
        read1buiFrom(file, position)
        // Unknown.
        read2bui(file)
        // Face; unused.
        read1bui(file)
    }

    // Read font parameters.
    file.Seek(int64(tfm.tablePointers[FontParameter]), io.SeekStart)

    if tfm.characterCodingScheme == "TeX math italic" {
        panic("Unsupported character coding scheme")
    }

    tfm.slant, _ = readFixWord(file)
    tfm.spacing, _ = readFixWord(file)
    tfm.spaceStretch, _ = readFixWord(file)
    tfm.spaceShrink, _ = readFixWord(file)
    tfm.xHeight, _ = readFixWord(file)
    tfm.quad, _ = readFixWord(file)
    tfm.extraSpace, _ = readFixWord(file)

    switch tfm.characterCodingScheme {
    case "TeX math symbols":
        // Read the additional 15 fix-word parameters.
        num1, _ := readFixWord(file)
		num2, _ := readFixWord(file)
		num3, _ := readFixWord(file)
		denom1, _ := readFixWord(file)
		denom2, _ := readFixWord(file)
		sup1, _ := readFixWord(file)
		sup2, _ := readFixWord(file)
		sup3, _ := readFixWord(file)
		sub1, _ := readFixWord(file)
		sub2, _ := readFixWord(file)
		supdrop, _ := readFixWord(file)
		subdrop, _ := readFixWord(file)
		delim1, _ := readFixWord(file)
		delim2, _ := readFixWord(file)
		axisHeight, _ := readFixWord(file)
        tfm.MathSymbolParams = MathSymbolParams{
	        Num1: num1,
	        Num2: num2,
	        Num3: num3,
	        Denom1: denom1,
	        Denom2: denom2,
	        Sup1: sup1,
	        Sup2: sup2,
	        Sup3: sup3,
	        Sub1: sub1,
	        Sub2: sub2,
	        Supdrop: supdrop,
	        Subdrop: subdrop,
	        Delim1: delim1,
	        Delim2: delim2,
	        AxisHeight: axisHeight,
        }
    case "TeX math extension", "euler substitutions only":
        // Read the additional 6 fix-word parameters.
    	defaultRuleThickness, _ := readFixWord(file)
    	var bigOpSpacing [5]float64
    	for i := range(bigOpSpacing) {
    		bigOpSpacing[i], _ = readFixWord(file)
    	}
    	tfm.MathExtensionParams = MathExtensionParams{
    		DefaultRuleThickness: defaultRuleThickness,
    		BigOpSpacing: bigOpSpacing,
    	}
    }
	return &tfm
}

func main() {
	tfm := NewTFM("cmr10.tfm")
    fmt.Printf("%v", tfm)
}
