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

func (m *mloc) String() string {
	return fmt.Sprintf("%#v", m)
}

func (m *mloc) init(fc bool, mlt int) {
	m.fc = fc
	m.mlt = mlt
	m.i, m.len = -1, -1
}
