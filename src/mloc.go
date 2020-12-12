package main

import "fmt"

const (
	mlInvalid = iota
	mlVoid
	mlArray
	mlInt
)

type mloc struct {
	fc  bool
	i   int
	len int
	mlt int
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
func (m *mloc) String() string {
	return fmt.Sprintf("%#v", m)
}

func (m *mloc) init(fc bool, mlt int) {
	m.fc = fc
	m.mlt = mlt
	m.i, m.len = -1, -1
}
