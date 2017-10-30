package read

import (
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path"
)

type ValueError struct {
    msg string
}

func (p ValueError) Error() string {
    return p.msg
}

type ExhaustedError struct {
    bytesRead int
}

func (p ExhaustedError) Error() string {
    return fmt.Sprintf("Exhausted, read %v bytes", p.bytesRead)
}

type NestedChild interface {}

type fancyByte struct {
    B byte
    Path string
    Position int
    LineNr int
    ColNr int
}

type FancyByteReader interface {
    io.ByteReader
    PeekByte(n int) (byte, error)
    ReadFancyByte() (fancyByte, error)
}

type NestedByteReader struct {
    name string
    contents []NestedChild
    position int
    LineNr int
    ColNr int
}

func NestedByteReaderFromBytes(name string, bs []byte) *NestedByteReader {
    contents := make([]NestedChild, len(bs), len(bs))
    for i, b := range bs {
        contents[i] = b
    }
    return &NestedByteReader{name: name, contents: contents}
}

func NestedByteReaderFromReader(name string, r io.Reader) *NestedByteReader {
    bs, _ := ioutil.ReadAll(r)
    return NestedByteReaderFromBytes(name, bs)
}

func NestedByteReaderFromPath(p string) *NestedByteReader {
    r, _ := os.Open(p)
    name := path.Base(p)
    return NestedByteReaderFromReader(name, r)
}

func (p *NestedByteReader) innerReader() (r NestedChild, err error) {
    if p.position > len(p.contents) - 1 {
        err = ExhaustedError{}
    } else {
        r = p.contents[p.position]
    }
    return
}

func (p *NestedByteReader) ReadFancyByte() (v fancyByte, err error) {
    for {
        innerR, errR := p.innerReader()

        if err, ok := errR.(ExhaustedError); ok {
            return fancyByte{}, err
        }

        if b, ok := innerR.(byte); ok {
            v = fancyByte{B: b, Position: p.position,
                          LineNr: p.LineNr, ColNr: p.ColNr}
            p.position++
            if b == '\n' {
                p.LineNr++
                p.ColNr = 0
            } else {
                p.ColNr++
            }
            return v, nil
        } else if nBR, ok := innerR.(NestedByteReader); ok {
            v, errB := nBR.ReadFancyByte()
            // If the inner reader returns a value, return that.
            if errB == nil {
                return v, errB
            } else if _, ok := errB.(ExhaustedError); ok {
                // We must have exhausted the current inner reader.
                // Move to the next one and try again.
                p.position++            
            } else {
                // Unknown error, return it.
                return v, errB
            }
        } else {
            panic(fmt.Sprintf("Unknown inner reader type '%v'", innerR))
        }
    }
}

func (p *NestedByteReader) ReadByte() (v byte, err error) {
    vF, err := p.ReadFancyByte()
    v = vF.B
    return
}

func (p *NestedByteReader) PeekByte(n int) (v byte, err error) {
    if n < 1 {
        err = ValueError{msg: fmt.Sprintf("Cannot peek %#v bytes, backwards peeking not implemented", n)}
        return
    }

    nToRead := n
    positionTemp := p.position

    for {
        // If we have reached the end of this reader's contents, return an
        // error containing the number of bytes we managed to read.
        if positionTemp > len(p.contents) - 1 {
            nRead := n - nToRead
            err = ExhaustedError{bytesRead: nRead}
            return
        }

        innerRTemp := p.contents[positionTemp]

        // If the current item is a single byte, read that, decrease the number
        // of bytes we must read, and increment our position.
        if vTemp, ok := innerRTemp.(byte); ok {
            nToRead--
            // If we have peeked all the bytes we have to do, return the peeked value.
            if nToRead == 0 {
                return vTemp, nil
            }
            positionTemp++
        // If the current item is a nested byte reader.
        } else if r, ok := innerRTemp.(*NestedByteReader); ok {
            // Try to peek the number of bytes we have yet to read from that
            // reader.
            vTemp, errB := r.PeekByte(nToRead)
            // If we peek the full number of bytes from it, we are done.
            if errB == nil {
                return vTemp, nil
            // Otherwise, decrease the number of bytes by the number we *did*
            // manage to peek, and increment our position.
            } else if e, ok := errB.(ExhaustedError); ok {
                // Given that we returned an exhausted error, we should have
                // read fewer bytes than we requested.
                if e.bytesRead >= nToRead {
                    panic(fmt.Sprintf("Peeking %v bytes returned an error, but also read %v bytes", nToRead, e.bytesRead))
                }
                nToRead -= e.bytesRead
                positionTemp++
            // If we got some other error, just return that.
            } else {
                return vTemp, errB
            }
        } else {
            panic(fmt.Sprintf("Unknown inner reader '%v'", innerRTemp))
        }
    }
}


func (p *NestedByteReader) Insert(r NestedChild) {
    innerR, _ := p.innerReader()
    if v, ok := innerR.(*NestedByteReader); ok {
        v.Insert(r)                
    } else {
        i := p.position
        p.contents = append(p.contents[:i], append([]NestedChild{r}, p.contents[i:]...)...)        
    }
}
