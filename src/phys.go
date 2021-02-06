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

func (p *phys) emitC(i string) {
	p.emit("//ec," + i)
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
	p.emitR(op, b, a)
}
func (p *phys) emitR(i string, ops ...regOrConst) {
	sa := []string{}
	for _, s := range ops {

		sa = append(sa, makeRC(s, true))
	}
	p.emit(i, sa...)
}

func (p *phys) cmp(a regi, b regOrConst) {
	p.emitR("cmpq", b, a)
}
func (p *phys) stackup(i int) {
	p.add(TSP, 8*i)
}
func (p *phys) push(r regi) {
	p.stackup(-1)
	p.str(ATeq, r, TSP)
}
func (p *phys) str(t atype, d regi, base regi, offset ...regOrConst) {
	if len(offset) == 1 {
		switch t {
		case ATeq:
			p.emit("mov", makeReg(d), fmt.Sprintf("%v(%v)", makeRC(offset[0], false), makeReg(base)))

		case ATpre:
			p.add(base, offset[0])
			p.emit("mov", makeReg(d), fmt.Sprintf("(%v)", makeReg(base)))
		case ATpost:
			p.emit("mov", makeReg(d), fmt.Sprintf("(%v)", makeReg(base)))
			p.add(base, offset[0])
		}
	} else {
		p.emit("mov", makeReg(d), fmt.Sprintf("(%v)", makeReg(base)))
	}
}
func (p *phys) ldr(t atype, d regi, base regi, offset ...regOrConst) {
	if len(offset) == 1 {
		switch t {
		case ATeq:
			p.emit("mov", fmt.Sprintf("%v(%v)", makeConst(offset[0].(int), false), makeReg(base)), makeReg(d))
		case ATpre:
			p.add(base, offset[0])
			p.emit("mov", fmt.Sprintf("(%v)", makeReg(base)), makeReg(d))
		case ATpost:
			p.emit("mov", fmt.Sprintf("(%v)", makeReg(base)), makeReg(d))
			p.add(base, offset[0])
		}
	} else {
		p.emit("mov", fmt.Sprintf("(%v)", makeReg(base)), makeReg(d))
	}
}

func (p *phys) tspchk() {
	/*
		p.cmp(TSP, TMAIN)
		lab := p.ug.clab()
		p.br(lab, "le")
		p.mov(TR1, 3)
		p.emitExit()

		p.makeLabel(lab)
	*/

}
func (p *phys) pop(r regi) {
	p.ldr(ATeq, r, TSP)
	p.stackup(1)
	p.tspchk()

}

func (p *phys) peek(r regi) {
	p.ldr(ATeq, r, TSP)
}

func (p *phys) br(b branchi, s ...string) {
	br := "jmp"
	if len(s) == 1 {
		br = "j" + localCond(s[0])
	}
	p.emit(br, makeBranch(b.(branch)))
	return
}

func (p *phys) pushTen() {
	for i := TR1; i <= TR9; i++ {
		p.push(i)
	}
}

func (p *phys) popTen() {
	for i := TR9; i >= TR1; i-- {
		p.pop(i)
	}
}
func (p *phys) emitPrint(ugly *emitter) {
	p.flabel("printch")
	p.pushTen()
	p.push(TR1)
	p.emitSprint(1, TSP)
	p.pop(TR1)
	p.popTen()
	p.emitRet()

	p.flabel("printdec")
	p.pop(TR5)
	p.pushTen()
	p.mov(TR10, TSP)
	labpd := p.ug.clab()
	p.mov(TR1, 0)
	p.makeLabel(labpd)
	p.mov(TR3, 10)
	p.mov(TR2, TR5)
	p.ug.doOp(TR2, TR3, "%")
	p.add(TR2, int('0'))
	p.push(TR2)
	p.add(TR1, 1)
	p.div(TR5, TR3)
	p.cmp(TR5, 0)
	p.br(labpd, "ne")
	p.push(TR1)
	p.fcall("printchar")
	p.mov(TSP, TR10)
	p.popTen()
	p.emitRet()

	p.flabel("printchar")
	p.pop(TR2)
	p.mov(TR10, TSP)
	p.pushTen()
	eplab := p.ug.clab()
	eplab2 := p.ug.clab()
	p.makeLabel(eplab)
	p.cmp(TR2, 0)
	p.br(eplab2, "eq")
	p.sub(TR2, 1)
	p.emitSprint(1, TR10)
	p.add(TR10, 8)
	p.br(eplab)
	p.makeLabel(eplab2)
	p.popTen()
	p.mov(TSP, TR10)
	p.emitRet()

	p.flabel("println")
	p.mov(TR1, int('\n'))
	p.fcall("printch")
	p.emitRet()

	p.flabel("print")
	p.mov(TR1, int('('))
	p.fcall("printch")
	p.pop(TR1)
	p.pushTen()
	p.mov(TR5, TSP)
	p.sub(TR5, 16)
	p.mov(TR2, 0)
	p.mov(TR3, 0)
	lab := ugly.clab()
	lab2 := ugly.clab()
	lab3 := ugly.clab()
	p.makeLabel(lab)
	p.mov(TR4, TR1)
	p.and(TR4, 0xf)
	p.cmp(TR4, 10)
	p.br(lab2, "lt")
	p.add(TR4, int('a'-':'))
	p.makeLabel(lab2)
	p.lsr(TR1, 4)
	p.add(TR4, int('0'))
	p.lsl(TR2, 8)
	p.add(TR2, TR4)
	p.cmp(TR3, 7)
	p.br(lab3, "ne")
	p.str(ATeq, TR2, TR5, 8)
	p.mov(TR2, 0)
	p.makeLabel(lab3)
	p.add(TR3, 1)
	p.cmp(TR3, 16)
	p.br(lab, "ne")
	p.str(ATeq, TR2, TR5)
	p.emitSprint(16, TR5)
	p.mov(TR1, int('.'))
	p.fcall("printch")
	p.popTen()
	p.emitRet()
}

func (p *phys) dbgExit() {
	p.push(TR2)
	p.fcall("print")
	didPrint = true
	p.emitExit()
}

func (p *phys) emitScheck() {
	p.emitC("emitSch")
	p.lsl(TR9, 3)
	p.add(TR9, TSP)
	p.cmp(TR9, TSS)
	labx := p.ug.clab()
	p.br(labx, "eq")
	p.emitLC()
	p.emitExit8()

	p.makeLabel(labx)

}
func (p *phys) emitLC() {
	if p.ug.st == nil {
		return
	}
	ln := p.ug.st.Gpos().Line
	p.mov(TR2, ln)
	p.push(TR2)
	p.fcall("printdec")

}
func (p *phys) emit2Printd() {
	p.push(TR2)
	didPrint = true
	p.fcall("printdec")
}

func (p *phys) emit2Print() {
	p.push(TR2)
	didPrint = true
	p.fcall("print")
}

func (p *phys) emit2Prints(s string) {
	for _, r := range s {
		p.mov(TR1, int(r))
		p.fcall("printch")
	}
}

func (p *phys) emitExit8() {
	p.mov(TR1, 8)
	p.emitExit()
}

func (p *phys) emitSprint(count int, source regOrConst) {
	p.mov(TR1, 0x2000004)
	p.mov(TR6, 1) //STDOUT
	p.mov(TR4, count)
	p.mov(TR5, source)
	p.syscall()
}

func (p *phys) emitExit() {
	p.ldr(ATeq, TR2, TBP)
	p.emitR("push", TR2)
	p.emitRet()
}

func (p *phys) sub(a regi, b regOrConst) {
	p.nativeOp("sub", a, b)
}
func (p *phys) add(a regi, b regOrConst) {
	p.nativeOp("add", a, b)
}
func (p *phys) mul(a regi, b regOrConst) {
	p.nativeOp("imul", a, b)
}

func (p *phys) fcall(id string) {
	p.emit("call", fmake(id))
}

func (p *phys) emitRet() {
	p.emit("ret")
}

func (p *phys) syscall() {
	p.emit("syscall")
}

func (p *phys) rem(a regi, b regOrConst) {
	p.push(TR1)
	p.push(TR4)
	p.mov(TR1, a)
	p.mov(TR4, 0)
	p.emitR("div", b)
	p.mov(a, TR4)
	p.pop(TR4)
	p.pop(TR1)

}
func (p *phys) div(a regi, b regOrConst) {
	p.push(TR1)
	p.push(TR4)
	p.mov(TR1, a)
	p.mov(TR4, 0)
	p.emitR("div", b)
	p.mov(a, TR1)
	p.pop(TR4)
	p.pop(TR1)
}
func (p *phys) and(a regi, b regOrConst) {
	p.nativeOp("and", a, b)
}
func (p *phys) lsl(a regi, b regOrConst) {
	p.nativeOp("sal", a, b)
}
func (p *phys) lsr(a regi, b regOrConst) {
	p.nativeOp("shr", a, b)
}
func (p *phys) mov(a regi, b regOrConst) {
	p.emitR("mov", b, a)
}

func (p *phys) emitDefines() {
	for r := TR1; r <= TSS; r++ {
		p.padd("#define " + rs[r] + " " + fmt.Sprintf("%v%v", RP, irs[r]) + "\n")
	}
}

func (p *phys) storeString(s string) {
}
