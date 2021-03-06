package main

import "fmt"

const (
	mlInvalid mltt = iota
	mlVoid
	mlArray
	mlScalar
	mlSlice
)

const (
	rsInvalid rstate = iota
	rsInt
	rsRange

	rsMloc
	rsString
	rsMulti
	rsBool
)

type rstate int
type mltt int

type mloc struct {
	fc     bool
	i      int
	len    int
	mlt    mltt
	rs     rstate
	ranged bool
	b      bool
}

func fromKind(k string) rstate {
	switch k {
	case "int":
		return rsInt
	case "string":
		return rsString
	case "ptr":
		return rsMloc
	case "bool":
		return rsBool
	}
	return rsInvalid
}

func newSent(r rstate) *mloc {
	rt := new(mloc)
	rt.rs = r
	return rt
}

func (ml *mloc) String() string {
	rt := "G"
	if ml.fc {
		rt = "L"
	}
	ap := ""
	if ml.mlt == mlArray {
		ap = fmt.Sprintf("[%v]", ml.len)
	}
	rt = fmt.Sprintf("%v%v%v%v%v", rt, ml.i, ap, ml.mlt, ml.rs)
	return rt

}

func (r rstate) String() string {
	switch r {
	case rsInvalid:
		return "X"
	case rsInt:
		return "I"
	case rsRange:
		return "R"
	case rsMloc:
		return "M"
	case rsString:
		return "S"
	case rsMulti:
		return "U"
	case rsBool:
		return "B"
	default:
		panic("")
	}
}

func (m mltt) String() string {
	switch m {
	case mlInvalid:
		return "X"
	case mlVoid:
		return "V"
	case mlArray:
		return "A"
	case mlScalar:
		return "I"
	case mlSlice:
		return "S"
	}
	return "oth"
}

func (m *mloc) typeOk(a *mloc) bool {
	/*
		if m.fc != a.fc {
			return false
		}
	*/
	if m.mlt == mlSlice {
		return a.mlt == mlArray && m.rs == a.rs
	}
	if m.mlt != a.mlt {
		return false
	}
	if m.rs != a.rs {
		return false
	}
	if m.mlt == mlArray {
		return m.len == a.len
	}
	return true
}

func (m *mloc) check() bool {
	if m.mlt == mlInvalid {
		return false
	}
	if m.rs == rsInvalid {
		return false
	}
	if m.mlt == mlScalar && m.len != -1 {
		return false
	}
	return true
}

func (m *mloc) init(fc bool, mlt mltt) {
	m.fc = fc
	m.mlt = mlt
	m.i, m.len = -1, -1
}
