package main

import (
    "io"
    "encoding/binary"
    "math"
)

var FixWordScale float64 = math.Pow(2, -20)

func read2bsi(r io.Reader) (v int16, err error) {
    err = binary.Read(r, binary.BigEndian, &v)
    return
}

func read2bsiFrom(r io.ReadSeeker, position uint16) (v int16, err error) {
    r.Seek(int64(position), io.SeekStart)
    return read2bsi(r)
}

func read4bsi(r io.Reader) (v int32, err error) {
    err = binary.Read(r, binary.BigEndian, &v)
    return
}

func read1bui(r io.Reader) (v uint8, err error) {
    err = binary.Read(r, binary.BigEndian, &v)
    return
}

func read1buiFrom(r io.ReadSeeker, position uint16) (v uint8, err error) {
    r.Seek(int64(position), io.SeekStart)
    return read1bui(r)
}

func read2bui(r io.Reader) (v uint16, err error) {
    err = binary.Read(r, binary.BigEndian, &v)
    return
}

func read4bui(r io.Reader) (v uint32, err error) {
    err = binary.Read(r, binary.BigEndian, &v)
    return
}

func readFixWord(r io.Reader) (v float64, err error) {
    x, err := read4bsi(r)
    v = FixWordScale * float64(x)
    return
}

func readBCPL(r io.Reader) (s string, err error) {
    length, err := read1bui(r)
    if (err != nil) {
        return
    }
    sb := make([]byte, length, length)
    err = binary.Read(r, binary.BigEndian, &sb)
    s = string(sb)
    return
}

func readBCPLFrom(r io.ReadSeeker, position uint16) (s string, err error) {
    r.Seek(int64(position), io.SeekStart)
    return readBCPL(r)
}

func currentPosition(r io.Seeker) (p uint16, err error) {
    pb, err := r.Seek(0, io.SeekCurrent)
    p = uint16(pb)
    return
}
