package main

import "fmt"

type mloc struct {
	fc  bool
	i   int
	len int
}

func (m *mloc) String() string {
	return fmt.Sprintf("%#v", m)
}

func (m *mloc) init(fc bool) {
	m.fc = fc
}
