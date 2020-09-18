package main

var n = struct{}{}

func tmOk(s string) bool {
	_, ok := tokenMap[s]
	return ok
}

func omOk(s string) bool {
	_, ok := opMap[s]
	return ok
}

var tokenMap = map[string]bool{
	"EOF":         false,
	"name":        false,
	"literal":     false,
	"op":          false,
	"op=":         false,
	"opop":        false,
	"=":           false,
	":=":          false,
	"(":           false,
	"[":           false,
	"{":           false,
	")":           false,
	"]":           false,
	"}":           false,
	",":           false,
	";":           false,
	":":           false,
	"?":           false,
	"@":           false,
	".":           false,
	"...":         false,
	"break":       false,
	"case":        false,
	"const":       true,
	"continue":    true,
	"default":     false,
	"else":        false,
	"fallthrough": true,
	"for":         true,
	"func":        false,
	"if":          true,
	"map":         false,
	"range":       false,
	"return":      true,
	"type":        true,
	"var":         true,
}

type LitKind uint8

const (
	IntLit LitKind = iota
	FloatLit
	StringLit
)

type Operator uint

var opMap = map[string]struct{}{
	"!":  n,
	"||": n,
	"&&": n,
	"==": n,
	"!=": n,
	"<":  n,
	"<=": n,
	">":  n,
	">=": n,
	"+":  n,
	"-":  n,
	"|":  n,
	"^":  n,
	"*":  n,
	"/":  n,
	"%":  n,
	"&":  n,
	"<<": n,
	">>": n,
}
