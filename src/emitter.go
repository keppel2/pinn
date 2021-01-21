//OutARM
package main

import "fmt"
import "reflect"
import "math/rand"

import "os"

var _ = os.Stderr

type emitter struct {
	rMap     map[string]*mloc
	cbranch  branch
	ebranch  branch
	ebranch2 branch
	moff     int
	soff     int
	lstack   [][2]branch
	fc       bool
	f        *FuncDecl
	fexitm   map[string]branch
	fexit    branch
	lst      Node
	st       Node
	file     *File
	p        *phys
}

func (e *emitter) checks() {
	for k, v := range e.rMap {
		_ = k
		if !v.check() {
			//e.err(k)
		}
	}
}
func (e *emitter) clearL() {
	for k, v := range e.rMap {
		if v.fc {
			delete(e.rMap, k)
		}
	}
}

func (e *emitter) emitArrayExpr(ae *ArrayExpr) *mloc {
	ml := e.newArml(len(ae.EL))
	for key, expr := range ae.EL {
		rt := e.assignToReg(expr)
		ml.rs = rt.rs
		e.p.mov(TR3, key)
		e.iStore(TR2, TR3, ml)
	}
	return ml
}

func (e *emitter) pushSoff(a int) {
	e.soff += a
	e.p.add(TR9, a)
}

func (e *emitter) newSlc() *mloc {
	ml := new(mloc)
	e.p.mov(TR5, 0)
	ml.init(e.fc, mlSlice)
	if ml.fc {
		e.pushSoff(2)
		ml.i = e.soff
		e.p.push(TR5)
		e.p.push(TR5)
	} else {
		ml.i = e.moff
		e.moff++
		e.moff++
		e.p.mov(TR2, 1)
		e.iStore(TR5, TR2, ml)
		e.p.mov(TR2, 2)
		e.iStore(TR5, TR2, ml)
	}
	return ml
}

func (e *emitter) newIntml() *mloc {
	e.p.emitC("nim")
	ml := new(mloc)
	ml.init(e.fc, mlInt)
	ml.rs = rsInt
	e.p.mov(TR5, 0)
	if ml.fc {
		e.pushSoff(1)
		ml.i = e.soff
		e.p.push(TR5)
	} else {
		ml.i = e.moff
		e.moff++
	}
	e.storeml(ml, TR5)
	return ml
}

func (e *emitter) newArml(len int) *mloc {
	ml := new(mloc)
	ml.init(e.fc, mlArray)
	ml.rs = rsInt
	ml.len = len //atoi(e, t.Len.(*NumberExpr).Il.Value)
	if e.fc {
		e.pushSoff(ml.len)
		ml.i = e.soff
		e.p.mov(TR2, 0)
		for i := 0; i < ml.len; i++ {
			e.p.push(TR2)
		}
	} else {
		ml.i = e.moff
		e.moff += ml.len
		e.p.mov(TR2, 0)
		for i := 0; i < ml.len; i++ {
			e.p.mov(TR3, i)

			e.iStore(TR2, TR3, ml)
		}
	}
	return ml

}

func (e *emitter) newVar(s string, k Kind) {
	if s == "_" {
		e.err(s)
	}

	if _, ok := e.rMap[s]; ok {
		e.err(s)
	}
	switch t := k.(type) {
	case *SKind:
		ml := e.newIntml()
		e.rMap[s] = ml
		if t.Wl.Value == "void" {
			e.rMap[s].mlt = mlVoid
		}

	case *ArKind:
		ml := e.newArml(atoi(e, t.Len.(*NumberExpr).Il.Value))
		ml.rs = fromKind(t.K.(*SKind).Wl.Value)
		e.rMap[s] = ml
	case *SlKind:
		ml := e.newSlc()
		ml.rs = fromKind(t.K.(*SKind).Wl.Value)
		e.rMap[s] = ml
	default:
		e.err(s)
	}
}

func (e *emitter) resetRegs() {
	for i := TR2; i <= TR10; i++ {
		e.p.mov(i, 0)
	}

}

func (e *emitter) pushAll() {
	for _, r := range ptr {
		e.p.push(r)
	}
}
func (e *emitter) popAll() {
	for i := len(ptr) - 1; i >= 0; i-- {
		e.p.pop(ptr[i])
	}
}

func (e *emitter) getaddr(m *mloc) {
	if m.fc {
		e.p.add(TR2, TSS)
	} else {
		e.p.add(TR2, TBP)
	}
}
func (e *emitter) setIndex(index regi, m *mloc) {
	e.p.lsl(index, 3)
	if m.fc {
		e.p.sub(index, moffOff(m.i))
	} else {
		e.p.add(index, moffOff(m.i))
	}
}

func (e *emitter) iStore(dest regi, index regi, m *mloc) {
	if m.mlt == mlVoid {
		e.loadml(m, TR1)
		e.p.emit("mov", makeReg(dest), fmt.Sprintf("%v(%v,%v,8)", 0, makeReg(TR1), makeReg(index)))
		return
	}
	if m.fc {
		e.p.emit("mov", makeReg(dest), fmt.Sprintf("%v(%v,%v,8)", -moffOff(m.i), makeReg(TSS), makeReg(index)))
	} else {
		e.p.emit("mov", makeReg(dest), fmt.Sprintf("%v(%v,%v,8)", moffOff(m.i), makeReg(TBP), makeReg(index)))
	}
}

func (e *emitter) iLoad(dest regi, index regi, m *mloc) {
	if m.mlt == mlInt {
		e.err("")
	}
	if m.mlt == mlVoid {
		e.loadml(m, TR1)
		e.p.emit("mov", fmt.Sprintf("%v(%v,%v,8)", 0, makeReg(TR1), makeReg(index)), makeReg(dest))
		return
	}
	if m.fc {
		e.p.emit("mov", fmt.Sprintf("%v(%v,%v,8)", -moffOff(m.i), makeReg(TSS), makeReg(index)), makeReg(dest))
	} else {
		e.p.emit("mov", fmt.Sprintf("%v(%v,%v,8)", moffOff(m.i), makeReg(TBP), makeReg(index)), makeReg(dest))
	}
}

func (e *emitter) dString() string {
	return fmt.Sprint(e.st, reflect.TypeOf(e.st), e.rMap)
}

func (e *emitter) rangeCheck(ml *mloc) {
	if ml.mlt == mlVoid {
		return
	}
	if ml.mlt == mlSlice {
		e.p.mov(TR5, 0)
		e.iLoad(TR3, TR5, ml)
		e.p.cmp(TR2, TR3)
	} else {
		e.p.cmp(TR2, ml.len)
	}

	lab := e.clab()
	e.p.br(lab, "lt")
	e.p.emit2Print()
	e.p.mov(TR2, ml.len)
	e.p.emit2Print()
	e.p.emit2Prints("range, line")
	ln := e.st.Gpos().Line
	e.p.mov(TR2, ln)
	e.p.emit2Printd()
	e.p.emit2Prints(".EXIT.")
	e.p.mov(TR1, 7)
	e.p.emitExit()

	e.p.makeLabel(lab)
}

func (e *emitter) init(f *File) {
	RP = "%r"
	rand.Seed(42)
	e.p = new(phys)
	e.p.init(e)
	e.moff = 1
	e.rMap = make(map[string]*mloc)
	e.fexitm = make(map[string]branch)
	e.cbranch = 1
	e.fexit = e.clab()
	e.file = f
}

func (e *emitter) clab() branch {
	rt := e.cbranch
	e.cbranch++
	return rt
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
	panic(msg)
}

func (e *emitter) loadml(ml *mloc, r regi) {
	if ml.mlt == mlArray {
		e.err(fmt.Sprint(ml.mlt))
	}
	if ml.fc {
		e.p.ldr(ATeq, r, TSS, -moffOff(ml.i))
	} else {
		e.p.ldr(ATeq, r, TBP, moffOff(ml.i))
	}
}

func (e *emitter) storeml(ml *mloc, r regi) {
	if ml.mlt == mlArray {
		e.err("")
	}
	if ml.fc {
		e.p.str(ATeq, r, TSS, -moffOff(ml.i))
	} else {
		e.p.str(ATeq, r, TBP, moffOff(ml.i))
	}
}

func (e *emitter) loadId(v string, r regi) {
	ml, ok := e.rMap[v]
	if !ok {
		e.err(v)
	}
	e.loadml(ml, r)
}

func (e *emitter) storeInt(v string, r regi) {
	ml, ok := e.rMap[v]
	if !ok {
		e.err(v)
	}
	e.storeml(ml, r)
}

func (e *emitter) storeId(v string, r regi) {
	if v == "_" {
		return
	}
	_, ok := e.rMap[v]
	if !ok {
		e.err(v)
	}

	e.storeInt(v, r)

}
func (e *emitter) doOp(dest, b regi, op string) {
	switch op {
	case "+":
		e.p.add(dest, b)
		return
	case "-":
		e.p.sub(dest, b)
		return
	case "*":
		e.p.mul(dest, b)
		return
	case "/":
		e.p.div(dest, b)
		return
	case "%":
		e.p.rem(dest, b)
		return
	case "@":
		e.p.add(b, 1)
		return
	case ":":
		return
	default:
		e.err(op)
	}
}
func (e *emitter) condExpr(dest branch, be *BinaryExpr) {
	if be.op == "||" {
		e.condExpr(dest, be.LHS.(*BinaryExpr))
		e.condExpr(dest, be.RHS.(*BinaryExpr))
	} else if be.op == "&&" {
		lab := e.clab()
		lab2 := e.clab()
		e.condExpr(lab, be.LHS.(*BinaryExpr))
		e.p.br(lab2)
		e.p.makeLabel(lab)
		e.condExpr(dest, be.RHS.(*BinaryExpr))
		e.p.makeLabel(lab2)
	} else if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == "<=" || be.op == ">" || be.op == ">=" {
		lh := e.assignToReg(be.LHS)
		if lh.rs != rsInt {
			e.err(be.op)
		}
		e.p.push(TR2)
		rh := e.assignToReg(be.RHS)
		if rh.rs != rsInt {
			e.err(be.op)
		}
		e.p.pop(TR4)
		e.p.cmp(TR4, TR2)
		bi := ""
		switch be.op {
		case "==":
			bi = "eq"
		case "!=":
			bi = "ne"
		case "<":
			bi = "lt"
		case "<=":
			bi = "le"
		case ">":
			bi = "gt"
		case ">=":
			bi = "ge"
		default:
			e.err(be.op)
		}
		e.p.br(branch(dest), bi)
		return
	} else {
		e.err(be.op)
	}

}

func (e *emitter) binaryExpr(be *BinaryExpr) *mloc {
	var rt *mloc

	lh := e.assignToReg(be.LHS)
	if lh.rs == rsInvalid {
		e.err("")
	}
	e.p.push(TR2)
	rh := e.assignToReg(be.RHS)
	if rh.rs == rsInvalid {
		e.err("")
	}
	e.p.mov(TR3, TR2)
	e.p.pop(TR2)
	e.doOp(TR2, TR3, be.op)
	if be.op == ":" || be.op == "@" {
		rt = newSent(rsRange)
	} else {
		rt = newSent(rsInt)
	}
	return rt
}

func (e *emitter) emitFunc(f *FuncDecl) {
	e.f = f
	e.p.flabel(f.Wl.Value)
	e.soff = 0
	e.p.mov(TSS, TSP)
	e.p.mov(TR9, 0)
	e.p.add(TSS, moffOff(f.PSize))
	for _, nt := range f.NTlist {
		if nt.N.Value == "_" {
			continue
		}
		if _, ok := e.rMap[nt.N.Value]; ok {
			e.err(nt.N.Value)
		}
		if ark, ok := nt.K.(*ArKind); ok {
			ml := new(mloc)
			ml.init(e.fc, mlArray)
			ml.rs = fromKind(ark.K.(*SKind).Wl.Value)
			plen := atoi(e, ark.Len.(*NumberExpr).Il.Value)
			e.pushSoff(plen)
			ml.len = plen
			ml.i = e.soff
			e.rMap[nt.N.Value] = ml
		} else {
			ml := new(mloc)
			ml.init(e.fc, mlInt)
			ml.rs = rsInt
			e.pushSoff(1)
			ml.i = e.soff
			e.rMap[nt.N.Value] = ml
		}
	}
	lab := e.clab()
	lab2 := e.clab()
	e.ebranch = lab
	e.ebranch2 = lab2
	e.emitStmt(f.B)
	if f.K != nil {
		e.p.mov(TR1, 8)
		e.p.emitExit()
	}

	e.p.makeLabel(lab)
	e.p.emitScheck()
	e.p.mov(TSP, TSS)
	e.p.tspchk()
	e.p.makeLabel(lab2)
	e.p.emitRet()
	e.checks()
	e.clearL()
}

func (e *emitter) getRs(ex Expr) rstate {
	switch t := ex.(type) {
	case *ArrayExpr:
		rs := e.getRs(t.EL[0])
		return rs
	case *CallExpr:
		ID := makeVar(t.ID)
		if ID == "malloc" {
			return rsMloc
		}
		if len(e.file.getFunc(ID).K) > 1 {
			return rsMulti
		}
		ks := e.file.getFunc(ID).K[0].(*SKind).Wl.Value
		return fromKind(ks)
	case *IndexExpr:
		if v, ok := t.E.(*BinaryExpr); ok {
			if v.op == ":" || v.op == "@" {
				return rsInt
			}
		}
		return rsInt

	default:
		return rsInt
	}

}

func (e *emitter) getType(ex Expr) *mloc {
	switch t := ex.(type) {
	case *ArrayExpr:
		mloc := e.newArml(len(t.EL))
		mloc.rs = e.getType(t.EL[0]).rs
		e.p.emitC(fmt.Sprint(mloc))
		return mloc
	case *CallExpr:
		ID := makeVar(t.ID)
		if ID == "malloc" {
			rt := e.newIntml()
			rt.mlt = mlVoid
			return rt
		}
		if len(e.file.getFunc(ID).K) > 1 {
			return newSent(rsMulti)
		}

	case *IndexExpr:
		if v, ok := t.E.(*BinaryExpr); ok {
			if v.op == ":" || v.op == "@" {
				rt := e.newSlc()
				return rt
			}
		}

	}
	return e.newIntml()
}

func (e *emitter) assignToReg(ex Expr) *mloc {
	var rt *mloc
	e.lst = e.st
	e.st = ex
	defer func() { e.st = e.lst }()
	switch t2 := ex.(type) {
	case *ArrayExpr:
		rt = e.emitArrayExpr(t2)
		return rt
	case *NumberExpr:
		e.p.mov(TR2, atoi(e, t2.Il.Value))
		return newSent(rsInt)
	case *StringExpr:
		return newSent(rsString)
	case *VarExpr:
		if t2.Wl.Value == "_" {
			e.err(t2.Wl.Value)
		}
		rt = e.rMap[t2.Wl.Value]
		if rt.mlt != mlArray && rt.mlt != mlSlice {
			e.loadId(t2.Wl.Value, TR2)
		}
	case *BinaryExpr:
		rt = e.binaryExpr(t2)
	case *UnaryExpr:
		rt = newSent(rsInt)
		if t2.op == "-" {
			e.assignToReg(t2.E)
			e.p.mov(TR5, -1)
			e.p.mul(TR2, TR5)
		} else if t2.op == "&" {
			rt = newSent(rsMloc)
			switch t3 := t2.E.(type) {
			case *VarExpr:
				v := t3.Wl.Value
				if v == "_" {
					e.err(v)
				}
				ml := e.rMap[v]
				e.p.mov(TR2, 0)
				e.setIndex(TR2, ml)
				e.getaddr(ml)
			case *IndexExpr:
				v := t3.X.(*VarExpr).Wl.Value
				if v == "_" {
					e.err(v)
				}
				ml := e.rMap[v]
				e.assignToReg(t3.E)
				e.rangeCheck(ml)
				e.setIndex(TR2, ml)
				e.getaddr(ml)
			}
		} else if t2.op == "*" {
			e.assignToReg(t2.E)
			e.p.ldr(ATeq, TR2, TR2)
		} else if t2.op == "+" {
			e.assignToReg(t2.E)
		} else {
			e.err(t2.op)
		}
	case *TrinaryExpr:
		lab := e.clab()
		lab2 := e.clab()
		lab3 := e.clab()
		e.condExpr(lab, t2.LHS.(*BinaryExpr))
		e.p.br(lab3)
		e.p.makeLabel(lab)
		rt2 := e.assignToReg(t2.MS)
		e.p.br(lab2)
		e.p.makeLabel(lab3)
		rt3 := e.assignToReg(t2.RHS)
		if !rt2.typeOk(rt3) {
			e.err("")
		}
		e.p.makeLabel(lab2)
		rt = rt2

	case *CallExpr:
		rt = e.emitCall(t2)
	case *IndexExpr:
		v := t2.X.(*VarExpr).Wl.Value
		if v == "_" {
			e.err(v)
		}
		ml := e.rMap[v]
		ert := e.assignToReg(t2.E)
		if ert == nil {
			e.err(v)
		}
		if ert.rs == rsRange {
			if ml.mlt != mlArray {
				e.err(v)
			}
			rt = ml
			rt.rs = rsRange
			break
		}
		if ml.mlt == mlSlice {
			e.rangeCheck(ml)
			e.p.mov(TR3, TR2)
			e.p.lsl(TR3, 3)
			e.p.mov(TR5, 1)
			e.iLoad(TR2, TR5, ml)
			e.p.add(TR2, TR3)
			e.p.ldr(ATeq, TR2, TR2)
			break

		}
		e.rangeCheck(ml)
		e.iLoad(TR2, TR2, ml)
		rt = newSent(rsInt)
	default:
		e.err("")
	}
	return rt
}

func (e *emitter) emitAssign(as *AssignStmt) {
	mts := make([]*mloc, len(as.RHSa))
	e.p.emitC("ea")

	if as.Op == "+=" || as.Op == "-=" || as.Op == "/=" || as.Op == "*=" || as.Op == "%=" {
		if len(as.RHSa) != 1 || len(as.LHSa) != 1 {
			e.err(as.Op)
		}
	}
	if as.Op == "--" || as.Op == "++" {
		if len(as.RHSa) != 0 || len(as.LHSa) != 1 {
			e.err(as.Op)
		}
	}

	if as.Op == ":=" {
		if e.getRs(as.RHSa[0]) == rsMulti {
			for _, v := range as.LHSa {
				id := makeVar(v)
				if id == "_" {
					e.err(id)
				}
				if e.rMap[id] != nil {
					e.err(id)
				}
				ml := e.newIntml()
				e.rMap[id] = ml
			}
		} else {
			for k, v := range as.LHSa {

				id := makeVar(v)
				if id == "_" {
					e.err(id)
				}
				if e.rMap[id] != nil {
					e.err(id)
				}
				e.p.emitC("eaea")
				ml := e.getType(as.RHSa[k])
				if ml.rs == rsMulti {
					if len(as.RHSa) != 1 {
						e.err("")
					}
				}

				if len(as.RHSa) == 1 && e.getType(as.RHSa[0]).rs == rsMulti {
					ml := e.newIntml()
					e.rMap[id] = ml
				} else {
					e.p.emitC("gd")
					ml := e.getType(as.RHSa[k])
					e.p.emitC("gd")
					e.rMap[id] = ml
					e.p.emitC("dg")
				}
			}
		}
	}

	for k, v := range as.RHSa {
		mts[k] = e.assignToReg(v)
		if mts[k].rs == rsInvalid {
			e.err(as.Op)
		}
		if mts[k].rs == rsMulti {
			if len(as.RHSa) != 1 {
				e.err("")
			}
			//	e.p.stackup(-len(ptr))
			//		e.p.stackup(-mts[0].len)
		} else {
			e.p.push(TR2)
		}
	}
	for k, v := range as.LHSa {
		if len(mts) > 0 {
			e.p.stackup(len(as.LHSa) - k - 1)
			e.p.tspchk()
			e.p.peek(TR2)
			e.p.stackup(-(len(as.LHSa) - k - 1))
			e.p.tspchk()
		}
		switch lh2 := v.(type) {
		case *UnaryExpr:
			if lh2.op != "*" {
				e.err(lh2.op)
			}
			e.p.push(TR2)
			e.assignToReg(lh2.E)
			e.p.mov(TR1, TR2)
			e.p.pop(TR2)
			e.p.str(ATeq, TR2, TR1)
		case *VarExpr:
			id := lh2.Wl.Value
			if as.Op == "=" && e.rMap[id] == nil {
				if id != "_" {
					e.err(id)
				}
			}
			if as.Op == "+=" || as.Op == "-=" || as.Op == "/=" || as.Op == "*=" || as.Op == "%=" || as.Op == "++" || as.Op == "--" {
				if id == "_" {
					e.err(id)
				}
				e.loadId(id, TR3)
				if as.Op[1:2] == "=" {
				} else {
					e.p.mov(TR2, 1)
				}
				e.doOp(TR3, TR2, as.Op[0:1])
				e.storeInt(id, TR3)
				break
			}
			var ml *mloc
			if len(mts) == 1 && mts[0].rs == rsMulti {
				ml = newSent(rsInt)
			} else {
				ml = mts[k]
			}
			if ml.mlt == mlInvalid && ml.rs == rsRange {
				e.err(id)
			} else if ml.rs == rsMloc {
				if e.rMap[id].mlt != mlVoid {
					e.err(id)
				}
				e.storeId(id, TR2)
				break
			} else if ml.mlt == mlArray && ml.rs == rsRange {
				mls := e.rMap[id]
				if mls == nil {
					mls = e.newSlc()
					e.rMap[id] = mls
				}
				if mls.mlt != mlSlice {
					e.err(id)
				}
				e.p.mov(TR4, TR3)
				e.p.sub(TR4, TR2)
				e.p.mov(TR5, 0)
				e.iStore(TR4, TR5, mls)
				e.p.mov(TR5, 1)
				e.setIndex(TR2, ml)
				e.getaddr(ml)
				e.iStore(TR2, TR5, mls)
				break
			}
			if ml.mlt == mlArray {
				if e.rMap[id] != nil && !e.rMap[id].typeOk(ml) {
					e.err(id)
				}
				nml := e.newArml(ml.len)
				e.rMap[id] = nml

				lab := e.clab()
				lab2 := e.clab()

				e.p.mov(TR2, 0)
				e.p.makeLabel(lab)
				e.p.cmp(TR2, ml.len)
				e.p.br(lab2, "ge")
				e.iLoad(TR3, TR2, ml)
				e.iStore(TR3, TR2, nml)
				e.p.add(TR2, 1)
				e.p.br(lab)
				e.p.makeLabel(lab2)
				break
			}

			e.storeId(id, TR2)

		case *IndexExpr:
			if as.Op == "+=" || as.Op == "-=" || as.Op == "/=" || as.Op == "*=" || as.Op == "%=" || as.Op == "++" || as.Op == "--" {
				if as.Op[1:2] == "=" {
					e.p.push(TR2)
					e.assignToReg(lh2)
					e.p.pop(TR3)
				} else {
					e.assignToReg(lh2)
					e.p.mov(TR3, 1)
				}
				e.doOp(TR2, TR3, as.Op[0:1])
			}

			v := makeVar(lh2.X)
			ml := e.rMap[v]

			e.p.push(TR2)
			e.assignToReg(lh2.E)
			e.rangeCheck(ml)
			e.p.pop(TR3)
			e.iStore(TR3, TR2, ml)
		default:
			e.err("")
		}

	}
	if len(mts) == 1 && mts[0].rs == rsMulti {
		e.p.add(TSP, TR7)
		e.p.stackup(len(ptr))
	} else {
		e.p.stackup(len(as.RHSa))
	}
	e.p.tspchk()

}
func (e *emitter) emitCall(ce *CallExpr) *mloc {
	var rt *mloc
	e.lst = e.st
	e.st = ce
	defer func() { e.st = e.lst }()
	//e.p.emitC(e.st)
	ID := makeVar(ce.ID)
	if ff, ok := fmap[ID]; ok {
		rt := ff(e, ce)
		return rt
	}

	fun := e.file.getFunc(ID)
	if fun == nil {
		e.err("Function not found: " + ID)
	}
	if fun.PCount == -1 {
		e.err("Internal function: " + ID)
	}
	if len(fun.K) == 0 {
		rt = newSent(rsInvalid)
	} else if len(fun.K) == 1 {
		skind := fun.K[0].(*SKind).Wl.Value
		rt = newSent(fromKind(skind))
	} else {
		rt = newSent(rsMulti)
		rt.len = len(fun.K)
		rt.i = fun.PSize
	}
	if len(ce.Params) != fun.PCount {
		e.err(ID)
	}

	e.pushAll()
	ssize := fun.PSize
	_ = ssize

	for k, v := range ce.Params {
		if v, ok := v.(*StringExpr); ok {
			sl := len(v.W.Value)
			for _, r := range revString(v.W.Value) {
				e.p.mov(TR2, int(r))
				e.p.push(TR2)
			}
			e.p.mov(TR2, sl)
			e.p.push(TR2)
			ssize = sl
			break
		}
		var kind Kind
		if len(fun.NTlist) != 0 {
			kind = fun.NTlist[k].K
		}
		if ie, ok := v.(*VarExpr); ok && e.rMap[ie.Wl.Value].len > 0 {
			if ie.Wl.Value == "_" {
				e.err("+_")
			}
			if atoi(e, kind.(*ArKind).Len.(*NumberExpr).Il.Value) != e.rMap[ie.Wl.Value].len {
				e.err(ID)
			}
			ml := e.rMap[ie.Wl.Value]
			for i := ml.len - 1; i >= 0; i-- {
				e.p.mov(TR2, i)
				e.iLoad(TR2, TR2, ml)
				e.p.push(TR2)
			}
		} else {
			if kind != nil {
				if _, ok := kind.(*SKind); !ok {
					e.err(ID)
				}
			}
			e.assignToReg(v)
			e.p.push(TR2)
		}
	}
	e.p.add(TMAIN, 1)
	e.p.cmp(TMAIN, 0x100)
	labm := e.clab()
	e.p.br(labm, "le")
	e.p.mov(TR1, 3)
	e.p.emitExit()
	e.p.makeLabel(labm)

	e.p.fcall(ID)

	e.p.sub(TMAIN, 1)

	if len(fun.K) == 1 {
		e.p.pop(TR2)
	} else if len(fun.K) > 1 {
		e.p.add(TSP, TR7)
		e.popAll()
		e.p.stackup(-len(ptr))
		e.p.sub(TSP, TR7)
		return rt
	}
	e.popAll()

	return rt

}

func (e *emitter) emitStmt(s Stmt) {
	e.lst = e.st
	e.st = s
	defer func() { e.st = e.lst }()
	e.p.emit("/*")
	if s != nil {
		v := new(visitor)
		v.visitStmt(s)
		e.p.emit(v.s)
	}
	e.p.emit("*/")
	e.p.emit("//")
	//		  e.p.emit2Prints(".")
	//	  e.p.emit2Print()
	//				  e.p.emit2Prints("<")
	//e.p.emitC("<")
	//				  e.p.emitLC()
	//	e.p.emitC(">")
	//			  e.p.emit2Prints(">")
	switch t := s.(type) {
	case *ExprStmt:
		e.assignToReg(t.Expr)
	case *BlockStmt:
		for _, s := range t.SList {
			e.emitStmt(s)
		}
		return
	case *ContinueStmt:
		e.p.br(e.peekloop()[0])
	case *BreakStmt:
		e.p.br(e.peekloop()[1])
	case *LoopStmt:
		lab := e.clab()
		lab2 := e.clab()
		e.p.makeLabel(lab)
		e.pushloop(lab, lab2)
		e.emitStmt(t.B)
		e.p.br(lab)
		e.p.makeLabel(lab2)
		e.poploop()
	case *WhileStmt:
		lab := e.clab()
		lab2 := e.clab()
		lab3 := e.clab()
		e.p.makeLabel(lab)
		e.pushloop(lab, lab2)
		e.condExpr(lab3, t.Cond.(*BinaryExpr))
		e.p.br(lab2)
		e.p.makeLabel(lab3)
		e.emitStmt(t.B)
		e.p.br(lab)
		e.p.makeLabel(lab2)
		e.poploop()

	case *IfStmt:
		lab := e.clab()
		if t.Else == nil {
			lab2 := e.clab()
			e.condExpr(lab2, t.Cond.(*BinaryExpr))
			e.p.br(lab)
			e.p.makeLabel(lab2)
			e.emitStmt(t.Then)
		} else {
			lab2 := e.clab()
			lab3 := e.clab()
			e.condExpr(lab2, t.Cond.(*BinaryExpr))
			e.p.br(lab3)
			e.p.makeLabel(lab2)
			e.emitStmt(t.Then)
			e.p.br(lab)
			e.p.makeLabel(lab3)
			e.emitStmt(t.Else)
		}
		e.p.makeLabel(lab)
		return

	case *ReturnStmt:
		if len(t.EL) == 0 {
			e.p.br(e.ebranch)
			return
		}
		if !e.fc {
			if len(t.EL) != 1 {
				e.err("")
			}
			e.p.mov(TR1, TR2)
			e.p.br(e.ebranch)
			return
		}
		if len(e.f.K) != len(t.EL) {
			e.err("")
		}

		if len(e.f.K) == 1 {
			e.assignToReg(t.EL[0])
			e.p.mov(TSP, TSS)
			e.p.push(TR2)
			e.p.br(e.ebranch2)
			return
		}
		for _, ex := range t.EL {
			e.p.emitC("tEL")
			e.assignToReg(ex)
			e.p.push(TR2)
			e.p.add(TR9, 1)
		}
		e.p.mov(TR7, TR9)
		e.p.lsl(TR7, 3)
		e.p.br(e.ebranch2)

		return
	case *AssignStmt:
		e.emitAssign(t)

	case *VarStmt:
		for _, v := range t.List {
			e.newVar(v.Value, t.Kind)
		}
		if e.fc {
			return
		}
	case *ForStmt:
		if t.Inits != nil {
			if rs, ok := t.Inits.(*AssignStmt); ok {
				if rs.irange {
					var iter, key *mloc
					if len(rs.LHSa) == 2 {
						first, second := makeVar(rs.LHSa[0]), makeVar(rs.LHSa[1])
						if first != "_" {
							key = e.rMap[rs.LHSa[0].(*VarExpr).Wl.Value]
						}
						if second != "_" {
							iter = e.rMap[rs.LHSa[1].(*VarExpr).Wl.Value]
						}
					} else {
						first := rs.LHSa[0].(*VarExpr).Wl.Value
						if first != "_" {
							iter = e.rMap[rs.LHSa[0].(*VarExpr).Wl.Value]
						}
					}
					var ml *mloc
					ml = e.assignToReg(rs.RHSa[0])
					lab := e.clab()
					lab2 := e.clab()
					e.pushloop(lab, lab2)
					e.p.mov(TR1, 0)
					e.p.makeLabel(lab)
					if key != nil {
						e.storeml(key, TR1)
					}

					if ml.rs != rsRange {
						e.iLoad(TR2, TR1, ml)
					}
					if iter != nil {
						e.storeml(iter, TR2)
					}
					e.p.push(TR2)
					e.p.push(TR3)
					e.p.push(TR1)
					e.p.add(TR9, 3)

					e.emitStmt(t.B)
					e.p.pop(TR1)
					e.p.pop(TR3)
					e.p.pop(TR2)
					e.p.sub(TR9, 3)
					e.p.add(TR1, 1)
					if ml.rs == rsRange {
						e.p.add(TR2, 1)
						e.p.cmp(TR2, TR3)
					} else {
						e.p.cmp(TR1, ml.len)
					}
					e.p.br(lab, "lt")
					labExit := e.clab()
					e.p.br(labExit)
					e.p.makeLabel(lab2)
					e.poploop()
					e.p.stackup(3)
					e.p.sub(TR9, 3)
					e.p.makeLabel(labExit)
					return
				}
			}

			e.emitStmt(t.Inits)
		}

		lab := e.clab()
		lab2 := e.clab()
		lab3 := e.clab()
		lab4 := e.clab()
		lab5 := e.clab()
		e.p.makeLabel(lab)
		e.pushloop(lab, lab2)
		e.p.br(lab3)
		e.p.makeLabel(lab5)
		if t.Loop != nil {
			e.emitStmt(t.Loop)
		}
		e.p.makeLabel(lab3)

		if t.E != nil {
			e.condExpr(lab4, t.E.(*BinaryExpr))
		} else {
			e.p.br(lab4)
		}
		e.p.br(lab2)
		e.p.makeLabel(lab4)
		e.emitStmt(t.B)
		e.p.br(lab5)

		e.p.makeLabel(lab2)
		e.poploop()
		return
	case nil:
	default:
		e.err("")

	}

}

func (e *emitter) emitDefines() {
	for r := TR1; r <= TSS; r++ {
		e.p.padd("#define " + rs[r] + " " + fmt.Sprintf("%v%v", RP, irs[r]) + "\n")
	}
}

func (e *emitter) emitF() {
	e.p.emitDefines()
	e.p.padd(".global _main\n")
	e.p.label("_main")
	e.p.emitR("pop", TR1)
	e.p.emitR("push", TR1)
	e.p.mov(TSP, SP)
	e.p.sub(TSP, 0x1000)
	e.p.mov(TSS, TSP)
	//e.p.mov(TMAIN, TSP)
	e.p.mov(TMAIN, 0)
	e.p.mov(TBP, TSS)
	e.p.sub(TBP, 0xA0000)
	e.p.str(ATeq, TR1, TBP)
	e.p.mov(THP, TBP)
	e.p.sub(THP, 0x1000)
	lab := e.clab()
	e.ebranch = lab
	e.p.mov(TR9, TSP)
	for _, s := range e.file.SList {
		e.emitStmt(s)
	}
	e.p.emitC("globs")
	e.p.makeLabel(lab)
	e.p.cmp(TR9, TSP)
	tc := e.clab()
	e.p.br(tc, "eq")
	e.p.mov(TR2, TR9)
	e.p.sub(TR2, TSP)
	e.p.emit2Print()
	e.p.emitExit8()
	e.p.makeLabel(tc)
	e.p.mov(TR1, 0)
	e.p.emitRet()
	e.checks()
	e.fc = true
	for _, s := range e.file.FList {
		if s.B != nil {
			e.emitFunc(s)
		}
	}
	if didPrint {
		e.p.emitPrint(e)
	}
	e.p.makeLabel(e.fexit)
	if len(e.lstack) != 0 {
		e.err("Loop stack")
	}
	e.p.emitC("end")
}
