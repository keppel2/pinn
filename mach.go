package main

import (
	"bytes"
	"amn/macho"
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
	for _, v := range f.Sections {
		fmt.Printf("%#v\n", v)
	}
	if *pFlag {
		return
	}
	baUW, _ := os.ReadFile("si.unwind")
	baLE, _ := os.ReadFile("si.linkedit")
	ret5 := []byte{0x48, 0xc7, 0xc0, 0x05, 0, 0, 0, 0xc3}
	_, _ = ret5, baLE
	// f.FileHeader.Ncmd = 1
	// f.FileHeader.Cmdsz = uint32(len(f.Loads[0].Raw()))
	mb := new(bytes.Buffer)
	ncmd := 0
	cmdsz := 0




	for k, v := range f.Loads {
		ncmd++
		cmdsz += len(v.Raw())
		if k == 1 {
			// ms := v.(*macho.Segment)
			// ms.SegmentHeader.Offset = 0x8000
			// ms.SegmentHeader.Memsz = 0x80000
			// ms.SegmentHeader.Filesz = 0x8000

		}

	}
	f.FileHeader.Cmdsz = uint32(cmdsz)
	f.FileHeader.Ncmd = uint32(ncmd)
	binary.Write(mb, binary.LittleEndian, f.FileHeader)
	mb.Write([]byte{0, 0, 0, 0})
	for k, v := range f.Loads {
    if k == 1 {
      ms := v.(*macho.Segment)
	//   ms.Flat.Addr = 0x10000_8000
	//   ms.Flat.Offset = 0x8000
	  ms.Flat.Memsz = 0x8000
	  ms.Flat.Filesz = 0x8000
      binary.Write(mb, binary.LittleEndian, ms.Flat)
	  for k, v2 := range ms.FlatSections {
		  if k == 0 { // Code
			  v2.Addr = 0x10000_2000
			  v2.Size = 8//0x3000
			  v2.Offset = 0x2000
		  } else if k == 1 {
			v2.Addr = 0x10000_1000
			v2.Size = uint64(len(baUW))
			v2.Offset = 0x1000
		  }
		  binary.Write(mb, binary.LittleEndian, v2)
	  }
	  continue
    } else if k == 2 {
		ms := v.(*macho.Segment)
		ms.Flat.Addr = 0x10000_8000
		ms.Flat.Offset = 0x8000
		binary.Write(mb, binary.LittleEndian, ms.Flat)
		continue
	} else if k == 3 {
		x := v.Raw()
		x[41] = 0x80
		// x[0x11] = 0x80
		mb.Write(x)
		continue

	} else if k == 4 {
			x := v.Raw()
			x[9] = 0x80
			x[0x11] = 0x80
			mb.Write(x)
			continue

	} else if k == 10 {
		x := v.Raw()
		x[8] = 0
		x[9] = 0x20
		mb.Write(x)
		continue
	}
	mb.Write(v.Raw())
	}

	offset := mb.Len()
	for offset != 0x1000 {
		mb.WriteByte(0)
		offset = mb.Len()
	}
	mb.Write(baUW)


	offset = mb.Len()
	for offset != 0x2000 {
		mb.WriteByte(0)
		offset = mb.Len()
	}

	mb.Write(ret5)





	offset = mb.Len()
_ = offset
	for offset != 0x8000 {
		mb.WriteByte(0)
		offset = mb.Len()
	}
	mb.Write(baLE)


	// offset = mb.Len()
	// for offset != 0xC000 {
	// 	mb.WriteByte(0)
	// 	offset = mb.Len()
	// }

	os.WriteFile(*oFlag, mb.Bytes(), 0777)
}
