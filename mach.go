package main

import (
	"bytes"
	"debug/macho"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
)

var pFlag = flag.Bool("p", false, "Print")
var fFlag = flag.String("f", "si.outx", "File")
var oFlag = flag.String("o", "si.out2", "Out")

func main() {
	flag.Parse()
	// var bs []byte
	f, err := macho.Open(*fFlag)
	if err != nil {
		panic(err)
	}
	ss, _ := f.ImportedLibraries()
	fmt.Println(ss)
	ss, _ = f.ImportedSymbols()
	fmt.Println(ss)

	fmt.Printf("%#v\n", f)
	for _, v := range f.Loads {
		fmt.Printf("%#v\n", v)
	}
	if *pFlag {
		return
	}
	ret5 := []byte{0x48, 0xc7, 0xc0, 0x05, 0, 0, 0, 0xc3}
	_ = ret5
	// f.FileHeader.Ncmd = 1
	// f.FileHeader.Cmdsz = uint32(len(f.Loads[0].Raw()))
	mb := new(bytes.Buffer)
	ncmd := 0
	cmdsz := 0
	for _, v := range f.Loads {
		ncmd++
		cmdsz += len(v.Raw())
	}
	f.FileHeader.Cmdsz = uint32(cmdsz)
	f.FileHeader.Ncmd = uint32(ncmd)
	binary.Write(mb, binary.LittleEndian, f.FileHeader)
	mb.Write([]byte{0, 0, 0, 0})
	for _, v := range f.Loads {
		mb.Write(v.Raw())
	}
	offset := mb.Len()

	for offset != 0x3fb0 {
		mb.WriteByte(0)
		offset = mb.Len()
	}

	// mb.Write(f.Loads[0].Raw())
	os.WriteFile(*oFlag, mb.Bytes(), 0777)
}
