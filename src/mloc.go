package main

import "fmt"

const (
	mlInvalid mltt = iota
	mlVoid
	mlArray
	mlInt
	mlSlice
)

const (
	rsInvalid rstate = iota
	rsInt
	rsRange
	rsMloc
	rsString
)

type rstate int
type mltt int

type mloc struct {
	fc  bool
	i   int
	len int
	mlt mltt
	rs  rstate
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
	rt = fmt.Sprintf("%v%v%v%v", rt, ml.i, ap, ml.mlt)
	return rt

}

func (m mltt) String() string {
	switch m {
	case mlInvalid:
		return "X"
	case mlVoid:
		return "V"
	case mlArray:
		return "A"
	case mlInt:
		return "I"
	case mlSlice:
		return "S"
	}
	return "oth"
}

func (m *mloc) typeOk(a *mloc) bool {
	if m.fc != a.fc {
		return false
	}
	if m.mlt != a.mlt {
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
	if m.mlt == mlInt && m.len != -1 {
		return false
	}
	return true
}

func (m *mloc) init(fc bool, mlt mltt) {
	m.fc = fc
	m.mlt = mlt
	m.i, m.len = -1, -1
}
