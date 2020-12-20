package main

import "fmt"
import "strings"

type phys struct {
	sb strings.Builder
	ug *emitter
}

func (p *phys) init(u *emitter) {
	p.ug = u
	p.sb.WriteString("//phys\n")
}

func (p *phys) padd(i string) {
	p.sb.WriteString(i)
}

func (p *phys) emit(i string, ops ...string) {
	const ind = "  "
	const AM = " "
	p.padd(ind + i + AM)
	if ops != nil {
		p.padd(ops[0])
		for _, s := range ops[1:] {
			p.padd(OS + s)
		}
	}
	p.padd("//" + p.ug.dString() + "\n")
}

func (p *phys) label(s string) {
	p.padd(s + ":\n")
}

func (p *phys) flabel(s string) {
	p.label(fmake(s))
}

func (p *phys) makeLabel(i branchi) {
	p.label(fmt.Sprintf("%v%v", BP, i))
}

func (p *phys) nativeOp(op string, a regi, b regOrConst) {
	if L {
		p.emitR(op, b, a)
	} else {
		p.emitR(op, a, a, b)
	}
}
func (p *phys) emitR(i string, ops ...regOrConst) {
	sa := []string{}
	for _, s := range ops {

		sa = append(sa, makeRC(s, true))
	}
	p.emit(i, sa...)
}

func (p *phys) cmp(a regi, b regOrConst) {
	if L {
		p.emitR("cmpq", b, a)
	} else {
		p.emitR("cmp", a, b)
	}
}
func (p *phys) push(r regi) {
	p.str(ATpre, r, TSP, -8)
}
func (p *phys) str(t atype, d regi, base regi, offset ...regOrConst) {
	if len(offset) == 1 {
		switch t {
		case ATeq:
			if L {
				p.emit("mov", makeReg(d), fmt.Sprintf("%v(%v)", makeRC(offset[0], false), makeReg(base)))
			} else {
				p.emit("str", makeReg(d), offSet(makeReg(base), makeRC(offset[0], true)))
			}

		case ATpre:
			if L {
				p.add(base, offset[0])
				p.emit("mov", makeReg(d), fmt.Sprintf("(%v)", makeReg(base)))
			} else {
				p.emit("str", makeReg(d), fmt.Sprintf("[%v%v%v]!", makeReg(base), OS, makeRC(offset[0], true)))
			}
		case ATpost:
			if L {
				p.emit("mov", makeReg(d), fmt.Sprintf("(%v)", makeReg(base)))
				p.add(base, offset[0])
			} else {
				p.emit("str", makeReg(d), fmt.Sprintf("[%v]%v%v", makeReg(base), OS, makeRC(offset[0], true)))
			}
		}
	} else {
		if L {
			p.emit("mov", makeReg(d), fmt.Sprintf("(%v)", makeReg(base)))
		} else {
			p.emit("str", makeReg(d), fmt.Sprintf("[%v]", makeReg(base)))
		}
	}
}
func (p *phys) ldr(t atype, d regi, base regi, offset ...regOrConst) {
	if len(offset) == 1 {
		switch t {
		case ATeq:
			if L {
				p.emit("mov", fmt.Sprintf("%v(%v)", makeConst(offset[0].(int), false), makeReg(base)), makeReg(d))
				//p.emit("mov", fmt.Sprintf("%v(%v,%v,8)", 0, makeReg(base), makeRC(offset[0], false)), makeReg(d))
			} else {
				p.emit("ldr", makeReg(d), offSet(makeReg(base), makeRC(offset[0], true)))
			}
		case ATpre:
			if L {
				p.add(base, offset[0])
				p.emit("mov", fmt.Sprintf("(%v)", makeReg(base)), makeReg(d))
			} else {
				p.emit("ldr", makeReg(d), fmt.Sprintf("[%v%v%v]!", makeReg(base), OS, makeRC(offset[0], true)))
			}
		case ATpost:
			if L {
				p.emit("mov", fmt.Sprintf("(%v)", makeReg(base)), makeReg(d))
				p.add(base, offset[0])
			} else {
				p.emit("ldr", makeReg(d), fmt.Sprintf("[%v]%v%v", makeReg(base), OS, makeRC(offset[0], true)))
			}
		}
	} else {
		if L {
			p.emit("mov", fmt.Sprintf("(%v)", makeReg(base)), makeReg(d))
		} else {
			p.emit("ldr", makeReg(d), fmt.Sprintf("[%v]", makeReg(base)))
		}
	}
}

func (p *phys) pnull() {
	p.add(TSP, 8)
}

func (p *phys) pop(r regi) {
	p.ldr(ATpost, r, TSP, 8)
}
func (p *phys) peek(r regi) {
	p.ldr(ATeq, r, TSP)
}

func (p *phys) br(b branchi, s ...string) {
	if L {
		br := "jmp"
		if len(s) == 1 {
			br = "j" + localCond(s[0])
		}
		p.emit(br, makeBranch(b.(branch)))
		return
	}
	br := "b"
	if len(s) == 1 {
		br += "." + s[0]
	}
	p.emit(br, makeBranch(b.(branch)))
}

func (p *phys) emitPrint(ugly *emitter) {
	p.flabel("printdec")
	p.peek(TR3)
	p.mov(TSS, TSP)
	labpd := p.ug.clab()
	p.mov(TR10, 0)
	p.makeLabel(labpd)
	p.mov(TR6, 10)
	p.mov(TR2, TR3)
	p.ug.doOp(TR2, TR6, "%")
	p.add(TR2, int('0'))
	p.push(TR2)
	p.add(TR10, 1)
	p.div(TR3, TR6)
	p.cmp(TR3, 0)
	p.br(labpd, "ne")
	p.fcall("printchar")
	p.mov(TSP, TSS)
	p.emit("ret")

	//  p.emitR("syscall")

	p.flabel("printchar")
	eplab := ugly.clab()
	eplab2 := ugly.clab()
	p.mov(TSS, TSP)
	p.makeLabel(eplab)
	p.cmp(TR10, 0)
	p.br(eplab2, "eq")
	p.sub(TR10, 1)

	if L {
		p.mov(TR1, 0x2000004) //SYSCALL 1 on linux
		p.mov(TR6, 1)         //STDOUT
		p.mov(TR4, 1)         //1 byte
		p.mov(TR5, TSS)
	} else {
		p.mov(TR1, 1) //STDOUT
		p.mov(TR2, TSS)
		p.mov(TR3, 1) //1 byte
		p.mov(TR9, 64)
	}
	p.syscall()
	p.add(TSS, 8)
	p.br(eplab)
	p.makeLabel(eplab2)
	p.emit("ret")
	p.flabel("println")
	if L {
		p.mov(TR8, int('\n'))
		p.push(TR8)
		p.mov(TR1, 0x2000004) //SYSCALL 1 on linux
		p.mov(TR6, 1)         //STDOUT
		p.mov(TR4, 1)         //1 byte
		p.mov(TR5, TSP)
		p.syscall()
		p.add(TSP, 8)
		p.emit("ret")
	} else {
		p.mov(TR1, int('\n'))
		p.push(TR1)
		p.mov(TR1, 1)
		p.mov(TR2, TSP)
		p.mov(TR3, 1)
		p.mov(TR9, 64)
		p.syscall()
		p.add(TSP, 8)
		p.emit("ret")
	}

	p.flabel("print")
	//p.mov(TR1, TR2)
	//p.emitExit()
	p.mov(TSS, TSP)
	p.ldr(ATeq, TR5, TSS)

	p.sub(TSS, 17)
	p.mov(TR3, int(','))
	p.str(ATeq, TR3, TSS)
	p.mov(TR2, 0)
	p.mov(TR3, 0)

	lab := ugly.clab()
	lab2 := ugly.clab()
	lab3 := ugly.clab()
	p.makeLabel(lab)
	p.mov(TR4, TR5)
	p.and(TR4, 0xf)
	p.cmp(TR4, 10)
	p.br(lab2, "lt")
	p.add(TR4, int('a'-':'))
	p.makeLabel(lab2)
	p.lsr(TR5, 4)
	p.add(TR4, int('0'))
	p.lsl(TR2, 8)
	p.add(TR2, TR4)
	p.cmp(TR3, 7)
	p.br(lab3, "ne")
	p.str(ATeq, TR2, TSS, 9)
	p.mov(TR2, 0)
	p.makeLabel(lab3)
	p.add(TR3, 1)
	p.cmp(TR3, 16)
	p.br(lab, "ne")
	p.str(ATeq, TR2, TSS, 1)
	if L {
		p.mov(TR1, 0x2000004)
		p.mov(TR6, 1)
		p.mov(TR4, 17)
		p.mov(TR5, TSS)

	} else {

		p.mov(TR1, 1)
		p.mov(TR2, TSS)
		p.mov(TR3, 17)
		p.mov(TR9, 64)
	}
	p.syscall()
	p.mov(TSS, TSP)
	p.emit("ret")
}

func (p *phys) dbgExit() {
	p.push(TR2)
	p.emit("call", FP+"print")
	didPrint = true
	p.emitExit()
}

func (p *phys) emit2Print() {
	p.push(TR2)
	didPrint = true
	p.emit("call", fmake("print"))
	p.pnull()
}

func (p *phys) emit2Prints(s string) {
	for _, r := range revString(s) {
		p.mov(TR2, int(r))
		p.push(TR2)
		p.mov(TR10, len(s))
	}

	p.emit("call", fmake("printchar"))
	p.add(TSP, moffOff(len(s)))

}

func (p *phys) emitExit() {
	if L {
		p.emitR("push", TMAIN)
	} else {
		p.mov(LR, TMAIN)
	}
	p.emit("ret")
}

func (p *phys) sub(a regi, b regOrConst) {
	p.nativeOp("sub", a, b)
}
func (p *phys) add(a regi, b regOrConst) {
	p.nativeOp("add", a, b)
}
func (p *phys) mul(a regi, b regOrConst) {
	if L {
		p.nativeOp("imul", a, b)
	} else {
		p.nativeOp("mul", a, b)
	}
}

func (p *phys) fcall(id string) {
	if L {
		p.emit("call", fmake(id))
	} else {
		p.emit("bl", fmake(id))
	}
}

func (p *phys) syscall() {
	if L {
		p.emit("syscall")
	} else {
		p.emitR("svc", 0)
	}
}

func (p *phys) rem(a regi, b regOrConst) {
	if L {
		p.mov(TR1, a)
		p.mov(TR4, 0)
		p.emitR("div", b)
		p.mov(a, TR4)
	} else {
		p.mov(TR5, a)
		p.emitR("udiv", a, TR5, b)
		p.emitR("msub", a, a, b, TR5)
	}
}
func (p *phys) div(a regi, b regOrConst) {
	if L {
		p.mov(TR1, a)
		p.mov(TR4, 0)
		p.emitR("div", b)
		p.mov(a, TR1)
	} else {
		p.nativeOp("udiv", a, b)
	}
}
func (p *phys) and(a regi, b regOrConst) {
	p.nativeOp("and", a, b)
}
func (p *phys) lsl(a regi, b regOrConst) {
	if L {
		p.nativeOp("sal", a, b)
	} else {
		p.nativeOp("lsl", a, b)
	}
}
func (p *phys) lsr(a regi, b regOrConst) {
	if L {
		p.nativeOp("shr", a, b)
	} else {
		p.nativeOp("lsr", a, b)
	}
}
func (p *phys) mov(a regi, b regOrConst) {
	if L {
		p.emitR("mov", b, a)
	} else {
		p.emitR("mov", a, b)
	}
}

func (p *phys) emitDefines() {
	if L {
		for r := TR1; r <= TSS; r++ {
			p.padd("#define " + rs[r] + " " + fmt.Sprintf("%v%v", RP, irs[r]) + "\n")
		}
	} else {
		for r := TR1; r <= TSS; r++ {
			p.padd("#define " + rs[r] + " " + fmt.Sprintf("%v%v", RP, ars[r]) + "\n")
		}
	}
}
