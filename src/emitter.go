package main

import "math/rand"
import "fmt"
import "reflect"
import "strconv"

var L = false

var RP = "x"

const OS = ", "

type branch int

func (_ branch) aBranch() {}

type branchi interface {
	aBranch()
}

type reg int

type regi interface {
	aReg()
}

type regOrConst interface {
}

func ff(a reg) {}

/*
func (e *emitter) _f() {
	e.mov(5, TR1)
}
*/

func (r reg) aReg() {}

var rs []string = []string{"TR1", "TR2", "TR3", "TR4", "TR5", "TR6", "TR7", "TR8", "TR9", "TRV", "TMAIN", "TBP", "TSP", "TSS"}

var irs []string = []string{
	"ax", "bx", "cx", "dx", "si", "di", "bp", "sp", "8", "9", "10", "11", "12", "13", "14", "15"}

const (
	TR1 reg = iota
	TR2
	TR3
	TR4
	TR5
	TR6
	TR7
	TR8
	TR9
	TRV
	TMAIN
	TBP
	TSP
	TSS

	RMAX
)

const (
	LR reg = 30
	SP reg = LR + 1 + iota
	XZR
)

/*
const (
	R0 reg = iota
	R1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
)
*/

const BP = ".br"
const FP = ".f"

var RB reg = TSS + 1
var IR reg = -1

/*
const (
	MLinvalid = iota
	MLreg
	MLstack
	MLheap
)
*/

type mloc struct {
	fc  bool
	i   int
	len int
	r   reg
}

type emitter struct {
	src     string
	rMap    map[string]*mloc
	rAlloc  [LR + 1]string
	cbranch branch
	ebranch branch
	moff    int
	soff    int
	lstack  [][2]branch
	fc      bool
	fexitm  map[string]branch
	fexit   branch
	lst     Node
	st      Node
}

func (m *mloc) String() string {
	return fmt.Sprintf("%#v", m)
}

func (m *mloc) init(fc bool) {
	m.r = IR
	m.fc = fc
}

func (e *emitter) findReg() reg {
	for k, v := range e.rAlloc[RB:] {
		if v == "" {
			return reg(k) + RB
		}
	}
	return e.freeReg()
}

func moffOff(a int) int {
	return a * 8
}

func (e *emitter) storeAll() {
	for k, _ := range e.rMap {
		e.toStore(k)
	}
}
func (e *emitter) clearL() {
	for k, v := range e.rAlloc {
		_ = v
		e.rAlloc[k] = ""
	}
	for k, v := range e.rMap {
		if v.fc {
			delete(e.rMap, k)
		} else {
			v.r = IR
		}
	}
}

func (e *emitter) newVar(s string, k Kind) {
	switch t := k.(type) {
	case *SKind:
	case *ArKind:
		if _, ok := e.rMap[s]; ok {
			e.err(s)
		}
		ml := new(mloc)
		ml.init(e.fc)
		ml.len = e.atoi(t.Len.(*NumberExpr).Il.Value)
		if e.fc {
			e.soff += ml.len
			ml.i = e.soff
			for i := 0; i < ml.len; i++ {
				e.push(XZR)
			}
		} else {
			ml.i = e.moff
			e.moff += ml.len
		}
		e.rMap[s] = ml
	}
}

func (e *emitter) push(r reg) {
	e.str(ATpre, r, TSP, -8)
}

func (e *emitter) pop(r reg) {
	e.emit("ldr", makeReg(r), "["+makeReg(TSP)+"]", makeConst(8))
}

func (e *emitter) popx() {
	e.add(TSP, 8)
}
func (e *emitter) pushAll() {
	for i := TR2; i <= TR7; i++ {
		if i != TSP {
			e.push(i)
		}
	}

}
func (e *emitter) popAll() {
	for i := TR7; i >= TR2; i-- {
		if i != TSP {
			e.pop(i)
		}
	}
}
func (e *emitter) setIndex(index reg, m *mloc) {
	e.lsl(index, 3)
	if m.fc {
		e.sub(index, moffOff(m.i))
	} else {
		e.add(index, moffOff(m.i))
	}
}

func (e *emitter) iStore(dest reg, index reg, m *mloc) {
	e.setIndex(index, m)
	if m.fc {
		e.str(ATeq, dest, TSS, index)
	} else {
		e.str(ATeq, dest, TBP, index)
	}
}
func (e *emitter) iLoad(dest reg, index reg, m *mloc) {
	e.setIndex(index, m)
	if m.fc {
		e.ldr(ATeq, dest, TSS, index)
	} else {
		e.ldr(ATeq, dest, TBP, index)
	}
}

func offSet(a, b string) string {
	return fmt.Sprintf("[%v%v%v]", a, OS, b)
}

func (e *emitter) toStore(id string) {
	ml := e.rMap[id]
	if ml.r == IR {
		return //e.err("")
	}
	if ml.i == -1 {
		if ml.fc {
			e.pushReg(ml)
		} else {
			ml.i = e.moff
			e.str(ATeq, ml.r, TBP, moffOff(ml.i))
			e.moff++
		}
	} else {

		if ml.fc {
			e.str(ATeq, ml.r, TSS, -moffOff(ml.i))
		} else {
			e.str(ATeq, ml.r, TBP, moffOff(ml.i))
		}
	}
	e.rAlloc[ml.r] = ""
	ml.r = IR
}

func (e *emitter) freeReg() reg {
	k := rand.Intn(int(RMAX + 1 - RB))
	k += int(RB)
	s := e.rAlloc[k]
	e.toStore(s)
	return reg(k)
}

func (e emitter) dString() string {
	return fmt.Sprint(e.rMap, e.rAlloc, e.st, reflect.TypeOf(e.st))
}

func (e *emitter) emit(i string, ops ...string) {
	const ind = "  "
	const AM = " "
	e.src += ind + i + AM
	if ops != nil {
		e.src += ops[0]
		for _, s := range ops[1:] {
			e.src += OS + s
		}
	}
	e.src += "//" + e.dString() + "\n"
}

func (e *emitter) emitR(i string, ops ...regOrConst) {
	sa := []string{}
	for _, s := range ops {

		sa = append(sa, makeRC(s))
	}
	e.emit(i, sa...)
}

func (e *emitter) init() {
	if L {
		RP = "%r"
	}
	rand.Seed(42)
	e.rMap = make(map[string]*mloc)
	e.fexitm = make(map[string]branch)
	e.cbranch = 1
}

/*
func (e *emitter) varAt(i int) string {
	for k, v := range e.rMap {
		if v == i {
			return k
		}
	}
	return ""
}
*/

func (e *emitter) clab() branch {
	rt := e.cbranch
	e.cbranch++
	return rt
}

func makeReg(i regi) string {
	if i.(reg) >= TR1 && i.(reg) <= TSS {
		return rs[i.(reg)]
	}

	if i == LR {
		return "lr"
	}
	if i == SP {
		return "sp"
	}
	if i == XZR {
		return "xzr"
	}

	return fmt.Sprintf("%v%v", RP, i)
}

func (e *emitter) atoi(s string) int {
	x, err := strconv.Atoi(s)
	if err != nil {
		e.err(err.Error())
	}
	return x

}

func makeConst(i int) string {
	if L {
		return fmt.Sprintf("$%v", i)
	} else {
		return fmt.Sprintf("#%v", i)
	}
}

func makeBranch(i branchi) string {
	return fmt.Sprintf("%v%v", BP, i)
}

func (e *emitter) label(s string) {
	e.src += s + ":\n"
}

func (e *emitter) makeLabel(i branchi) {
	e.label(fmt.Sprintf("%v%v", BP, i))
}

func (e *emitter) pushloop(a, b branch) {
	e.lstack = append(e.lstack, [2]branch{a, b})
}

func (e *emitter) poploop() [2]branch {
	rt := e.lstack[len(e.lstack)-1]
	e.lstack = e.lstack[0 : len(e.lstack)-1]
	return rt
}

func (e *emitter) peekloop() [2]branch {
	return e.lstack[len(e.lstack)-1]
}

func (e *emitter) err(msg string) {

	panic(fmt.Sprintln(msg, e.dString(), e.src))
}

func (e *emitter) emitPrint() {
	e.label("println")
	e.mov(TR1, int('\n'))
	e.push(TR1)
	e.mov(TR1, 1)
	e.mov(TR2, TSP)
	e.mov(TR3, 1)
	e.mov(TR9, 64)
	e.emitR("svc", 0)
	e.pop(TR1)
	e.emit("ret")

	e.label("print")

	e.sub(TSP, 17)
	e.mov(TR3, int(','))
	e.str(ATeq, TR3, TSP)
	e.mov(TR2, 0)
	e.mov(TR3, 0)

	lab := e.clab()
	lab2 := e.clab()
	lab3 := e.clab()
	e.makeLabel(lab)
	e.mov(TR4, TR1)
	e.and(TR4, 0xf)
	e.emitR("cmp", TR4, 10)
	e.br(lab2, "lt")
	e.add(TR4, int('a'-':'))
	e.makeLabel(lab2)
	e.lsr(TR1, 4)
	e.add(TR4, int('0'))
	e.lsl(TR2, 8)
	e.add(TR2, TR4)
	e.emitR("cmp", TR3, 7)
	e.br(lab3, "ne")
	e.str(ATeq, TR2, TSP, 9)
	e.mov(TR2, 0)
	e.makeLabel(lab3)
	e.add(TR3, 1)
	e.emitR("cmp", TR3, 16)
	e.br(lab, "ne")
	e.str(ATeq, TR2, TSP, 1)
	e.mov(TR1, 1)
	e.mov(TR2, TSP)
	e.mov(TR3, 17)
	e.mov(TR9, 64)
	e.emitR("svc", 0)
	e.add(TSP, 17)
	e.emit("ret")
}

func (e *emitter) pushReg(m *mloc) {
	e.soff++
	m.i = e.soff
	e.push(m.r)
}

func (e *emitter) loadId(v string, r reg) {
	ml, ok := e.rMap[v]
	if !ok {
		e.err(v)
	}
	if ml.fc {
		e.ldr(ATeq, r, TSS, -moffOff(ml.i))
	} else {
		e.ldr(ATeq, r, TBP, moffOff(ml.i))
	}
}

func (e *emitter) storeId(v string, r reg) {
	ml, ok := e.rMap[v]
	if ok {
		if ml.fc {
			e.str(ATeq, r, TSS, -moffOff(ml.i))
		} else {
			e.str(ATeq, r, TBP, moffOff(ml.i))
		}
	} else {
		ml := new(mloc)
		ml.init(e.fc)
		if ml.fc {
			e.push(r)
			e.soff++
			ml.i = e.soff
		} else {
			ml.i = e.moff
			e.str(ATeq, r, TBP, moffOff(ml.i))
			e.moff++
		}
		e.rMap[v] = ml
	}

}

func (e *emitter) forceReg(v string, r reg) {
	if _, ok := e.rMap[v]; ok {
		e.err(v)
	}
	if e.rAlloc[r] != "" {
		e.err(v)
	}
	e.rAlloc[r] = v
	ml := new(mloc)
	ml.init(e.fc)
	if !e.fc {
		e.err("")
	}
	ml.r = r
	e.pushReg(ml)
	e.rMap[v] = ml

}

func (e *emitter) fillReg(s string, create bool) reg {
	ml, ok := e.rMap[s]
	if ok && ml.r != IR {
		//    e.rAlloc[ml.r] = ""
		//    ml.r = IR
		return ml.r
	}
	if !ok && !create {
		e.err(s)
	}
	k := e.findReg()
	e.rAlloc[k] = s
	if !ok {
		ml = new(mloc)
		ml.init(e.fc)
		ml.r = k
		e.rMap[s] = ml
	} else {
		if ml.fc {
			e.ldr(ATeq, k, TSS, -moffOff(ml.i))
		} else {
			e.ldr(ATeq, k, TBP, moffOff(ml.i))
		}
		ml.r = k
	}
	return k
}

func makeRC(a regOrConst) string {
	if a2, ok := a.(reg); ok {
		return makeReg(a2)
	}
	return makeConst(a.(int))
}

func (e *emitter) br(b branchi, s ...string) {
	if L {
		br := "jmp"
		if len(s) == 1 {
			br = "j" + s[0]
		}
		e.emit(br, makeBranch(b.(branch)))
		return
	}
	br := "b"
	if len(s) == 1 {
		br += "." + s[0]
	}
	e.emit(br, makeBranch(b.(branch)))
}

type atype int

const (
	ATinvalid atype = iota
	ATeq
	ATpre
	ATpost
)

func (e *emitter) str(t atype, d regi, base regi, offset ...regOrConst) {
	if len(offset) == 1 {
		switch t {
		case ATeq:
			e.emit("str", makeReg(d), offSet(makeReg(base), makeRC(offset[0])))
		case ATpre:
			e.emit("str", makeReg(d), fmt.Sprintf("[%v%v%v]!", makeReg(base), OS, makeRC(offset[0])))
		case ATpost:
			e.emit("str", makeReg(d), fmt.Sprintf("[%v]%v%v", makeReg(base), OS, makeRC(offset[0])))
		}
	} else {
		e.emit("str", makeReg(d), fmt.Sprintf("[%v]", makeReg(base)))
	}
}

func (e *emitter) ldr(t atype, d regi, base regi, offset ...regOrConst) {
	if len(offset) == 1 {
		switch t {
		case ATeq:
			e.emit("ldr", makeReg(d), offSet(makeReg(base), makeRC(offset[0])))
		case ATpre:
			e.emit("ldr", makeReg(d), fmt.Sprintf("[%v%v%v]!", makeReg(base), OS, makeRC(offset[0])))
		case ATpost:
			e.emit("ldr", makeReg(d), fmt.Sprintf("[%v]%v%v", makeReg(base), OS, makeRC(offset[0])))
		}
	} else {
		e.emit("ldr", makeReg(d), fmt.Sprintf("[%v]", makeReg(base)))
	}
}

func (e *emitter) nativeOp(op string, a regi, b regOrConst) {
	if L {
		e.emitR(op, b, a)
	} else {
		e.emitR(op, a, a, b)
	}
}

func (e *emitter) sub(a regi, b regOrConst) {
	e.nativeOp("sub", a, b)
}
func (e *emitter) add(a regi, b regOrConst) {
	e.nativeOp("add", a, b)
}
func (e *emitter) mul(a regi, b regOrConst) {
	e.nativeOp("mul", a, b)
}
func (e *emitter) and(a regi, b regOrConst) {
	e.nativeOp("and", a, b)
}
func (e *emitter) lsl(a regi, b regOrConst) {
	e.nativeOp("lsl", a, b)
}
func (e *emitter) lsr(a regi, b regOrConst) {
	e.nativeOp("lsr", a, b)
}
func (e *emitter) mov(a regi, b regOrConst) {
	if L {
		e.emitR("mov", b, a)
	} else {
		e.emitR("mov", a, b)
	}
	/*
		a2, ok := a.(reg)
		if !ok {
			e.err("")
		}
	*/

	//	sb := makeRC(b)
	//	e.emit("mov", makeReg(a.(reg)), sb)//.(reg)), sb)
}

func (e *emitter) regLoad(ex Expr) reg {
	var rt reg
	switch t := ex.(type) {
	case *NumberExpr:
		e.mov(TR1, e.atoi(t.Il.Value))
		rt = TR1
	case *VarExpr:
		i := e.fillReg(t.Wl.Value, false)
		rt = i
	default:
		e.err("")
	}
	return rt

}

func (e *emitter) doOp(dest, b reg, op string) {
	mn := ""
	switch op {
	case "+":
		e.add(dest, b)
		return
	case "-":
		e.sub(dest, b)
		return
	case "*":
		e.mul(dest, b)
		return
	case "/":
		mn = "udiv"
	}
	if mn != "" {
		e.emitR(mn, dest, dest, b)
		return
	}
	switch op {
	case "%":
		e.mov(TR5, dest)
		e.emitR("udiv", dest, TR5, b)
		e.emitR("msub", dest, dest, b, TR5)
		return
	default:
		e.err(op)
	}

}

func (e *emitter) binaryExpr(dest reg, be *BinaryExpr) {
	e.lst = e.st
	e.st = be
	defer func() { e.st = e.lst }()
	if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == "<=" || be.op == ">" || be.op == ">=" {
		e.assignToReg(TR4, be.LHS)
		e.assignToReg(TR5, be.RHS)
		e.emitR("cmp", TR4, TR5)
		bi := ""
		switch be.op {
		case "==":
			bi = "ne"
		case "!=":
			bi = "eq"
		case "<":
			bi = "ge"
		case "<=":
			bi = "gt"
		case ">":
			bi = "le"
		case ">=":
			bi = "lt"
		}
		e.br(branch(dest), bi)
		return
	}
	switch t := be.LHS.(type) {
	case *NumberExpr, *VarExpr:
		e.assignToReg(dest, t)
	case *BinaryExpr:
		e.binaryExpr(dest, t)
	case *CallExpr:
		e.emitCall(t)
		e.mov(dest, TR1)
	}
	//	op := ""
	if t, ok := be.RHS.(*CallExpr); ok {
		e.emitCall(t)
		e.mov(TR3, TR1)
	} else {
		e.assignToReg(TR3, be.RHS)
	}
	e.doOp(dest, TR3, be.op)
}

func (e *emitter) emitFunc(f *FuncDecl) {
	e.label(FP + f.Wl.Value)
	e.soff = 0
	e.mov(TSS, TSP)
	for _, vd := range f.PList {
		for _, vd2 := range vd.List {

			if _, ok := e.rMap[vd2.Value]; ok {
				e.err(vd2.Value)
			}
			ml := new(mloc)
			ml.init(e.fc)
			e.soff++
			ml.i = -(f.PCount - e.soff)
			e.rMap[vd2.Value] = ml
		}
	}
	e.soff = 0
	lab := e.clab()
	e.ebranch = lab
	e.emitStmt(f.B)
	e.makeLabel(lab)

	e.mov(TSP, TSS)
	e.emit("ret")
	e.clearL()
}

func (e *emitter) assignToReg(r reg, ex Expr) {
	switch t2 := ex.(type) {
	case *NumberExpr:
		e.mov(r, e.atoi(t2.Il.Value))
	case *VarExpr:
		e.loadId(t2.Wl.Value, r)
	case *BinaryExpr:
		e.binaryExpr(r, t2)
	case *CallExpr:
		e.emitCall(t2)
		e.mov(r, TR1)
	case *IndexExpr:
		v := t2.X.(*VarExpr).Wl.Value
		ml := e.rMap[v]
		e.assignToReg(TR1, t2.E)
		e.iLoad(r, TR1, ml)
	default:
		e.err("")
	}

}

func (e *emitter) emitCall(ce *CallExpr) {
	e.st = ce
	ID := ce.ID.(*VarExpr).Wl.Value

	fn := FP + ID
	if ID == "assert" {
		e.assignToReg(TR2, ce.Params[0])
		e.assignToReg(TR3, ce.Params[1])
		e.emitR("cmp", TR2, TR3)
		lab := e.clab()
		e.br(lab, "eq")
		ln := e.st.Gpos().Line
		e.mov(TR1, ln)
		e.mov(LR, TMAIN)
		e.emit("ret")
		e.makeLabel(lab)
		return
	} else if ID == "bad" {
		e.mov(TR1, 7)
		e.mov(LR, TMAIN)
		e.emit("ret")
		return
	} else if ID == "exit" {
		e.assignToReg(TR1, ce.Params[0])
		e.mov(LR, TMAIN)
		e.emit("ret")
		return
	} else if ID == "print" {
		e.assignToReg(TR1, ce.Params[0])
		fn = ID
		didPrint = true
	} else if ID == "println" {
		didPrint = true
		fn = ID
	}

	// e.pushP()
	e.pushAll()
	e.push(TSS)
	e.push(LR)

	for _, v := range ce.Params {
		//		e.push(1 + reg(k))

		e.assignToReg(TR1, v)
		e.push(TR1)
	}

	e.emit("bl", fn)
	e.add(TSP, moffOff(len(ce.Params)))
	e.pop(LR)
	e.pop(TSS)

	e.popAll()

}

func (e *emitter) emitStmt(s Stmt) {
	e.st = s
	e.storeAll()
	e.emit("//")
	switch t := s.(type) {
	case *ExprStmt:
		e.assignToReg(TR1, t.Expr)
	case *BlockStmt:
		for _, s := range t.SList {
			e.emitStmt(s)
		}
	case *ContinueStmt:
		e.br(e.peekloop()[0])
	case *BreakStmt:
		e.br(e.peekloop()[1])
	case *LoopStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab, lab2)
		e.emitStmt(t.B)
		e.br(lab)
		e.makeLabel(lab2)
		e.poploop()
	case *WhileStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab, lab2)
		e.binaryExpr(reg(lab2), t.Cond.(*BinaryExpr))
		e.emitStmt(t.B)
		e.br(lab)
		e.makeLabel(lab2)

	case *IfStmt:
		lab := e.clab()
		if t.Else == nil {
			e.binaryExpr(reg(lab), t.Cond.(*BinaryExpr))
			e.emitStmt(t.Then)
		} else {
			lab2 := e.clab()
			e.binaryExpr(reg(lab2), t.Cond.(*BinaryExpr))
			e.emitStmt(t.Then)
			e.br(lab)
			e.makeLabel(lab2)
			e.emitStmt(t.Else)
		}
		e.makeLabel(lab)

	case *ReturnStmt:
		if t.E != nil {
			e.assignToReg(TRV, t.E)
		} else {
			e.mov(TRV, 5)
		}
		if L {
			e.emitR("mov", TRV, TR1)
		} else {
			e.mov(TR1, TRV)
		}
		e.br(e.ebranch)
	case *AssignStmt:
		lh := t.LHSa[0]
		switch lh2 := lh.(type) {
		case *VarExpr:
			id := lh2.Wl.Value
			if t.Op == "+=" || t.Op == "-=" || t.Op == "/=" || t.Op == "*=" || t.Op == "%=" {
				//lhi := e.fillReg(id, false)
				//e.forceReg(id, TR5)
				e.loadId(id, TR2)
				e.assignToReg(TR3, t.RHSa[0])
				//				e.mov(TR2, lhi)
				e.doOp(TR2, TR3, t.Op[0:1])
				e.storeId(id, TR2)
				//				e.toStore(id)

				return
			}
			if t.Op == "++" {
				e.loadId(id, TR3)
				//			lhi := e.fillReg(id, false)
				e.mov(TR1, 1)
				e.doOp(TR3, TR1, "+")
				e.storeId(id, TR3)
				return
			} else if t.Op == "--" {
				e.loadId(id, TR3)
				e.mov(TR1, 1)
				e.doOp(TR3, TR1, "-")
				e.storeId(id, TR3)
				return
			}
			//lhi := e.fillReg(id, true)
			//e.assignToReg(lhi, t.RHSa[0])
			e.assignToReg(TR4, t.RHSa[0])
			e.storeId(id, TR4)
			//e.storeAll()
			//e.toStore(id)

		case *IndexExpr:
			if t.Op == "+=" || t.Op == "-=" || t.Op == "/=" || t.Op == "*=" || t.Op == "%=" {
				e.assignToReg(TR4, lh2)
				e.assignToReg(TR3, t.RHSa[0])
				e.doOp(TR4, TR3, t.Op[0:1])
			} else if t.Op == "++" || t.Op == "--" {
				e.assignToReg(TR4, lh2)
				if t.Op == "++" {
					e.mov(TR1, 1)
					e.doOp(TR4, TR1, "+")
				} else {
					e.mov(TR1, 1)
					e.doOp(TR4, TR1, "-")
				}
			} else {
				e.assignToReg(TR4, t.RHSa[0])
			}

			v := lh2.X.(*VarExpr).Wl.Value
			ml := e.rMap[v]
			e.assignToReg(TR1, lh2.E)
			e.iStore(TR4, TR1, ml)
		}

	case *VarStmt:
		for _, v := range t.List {
			e.newVar(v.Value, t.Kind)
		}
	case *ForStmt:
		if t.Inits != nil {
			e.emitStmt(t.Inits)
		}

		lab := e.clab()
		lab2 := e.clab()
		lab3 := e.clab()
		e.pushloop(lab, lab2)
		e.br(lab3)
		e.makeLabel(lab)
		if t.Loop != nil {
			e.emitStmt(t.Loop)
		}
		e.makeLabel(lab3)

		if t.E != nil {
			e.binaryExpr(reg(lab2), t.E.(*BinaryExpr))
		}
		e.emitStmt(t.B)
		e.br(lab)

		e.makeLabel(lab2)

	default:
		e.err("")

	}

}

func (e *emitter) emitDefines() {
	if L {
		for r := TR1; r >= TSS; r-- {
			e.src += "#define " + rs[TR1-r] + " " + fmt.Sprintf("%v%v", RP, rs[r]) + "\n"
		}
	} else {
		for r := TR1; r <= TSS; r++ {
			e.src += "#define " + rs[r] + " " + fmt.Sprintf("%v%v", RP, r) + "\n"
		}
	}
}

var didPrint = false

func (e *emitter) emitF(f *File) {
	e.emitDefines()
	e.src += ".global main\n"
	e.label("main")
	e.mov(TMAIN, LR)
	e.sub(SP, 0x100)
	e.mov(TSP, SP)
	e.mov(TSS, SP)
	e.sub(SP, 0x10000)
	e.mov(TBP, SP)
	lab := e.clab()
	e.ebranch = lab
	for _, s := range f.SList {
		e.emitStmt(s)
	}
	e.mov(TR1, XZR)
	e.makeLabel(lab)
	e.emit("ret")
	e.clearL()
	e.fc = true
	for _, s := range f.FList {

		e.emitFunc(s)
	}
	if didPrint {
		e.emitPrint()
	}
}
