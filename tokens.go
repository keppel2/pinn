package main

var n = struct{}{}

func tmOk(s string) bool {
	_, ok := tokenMap[s]
	return ok
}

// True are keywords that can be fast-forwarded to in case of error.
var tokenMap = map[string]bool{
	"EOF":         false,
	"name":        false,
	"literal":     false,
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

	"!":  false,
	"||": false,
	"&&": false,
	"==": false,
	"!=": false,
	"<":  false,
	"<=": false,
	">":  false,
	">=": false,
	"+":  false,
	"-":  false,
	"|":  false,
	"^":  false,
	"*":  false,
	"/":  false,
	"%":  false,
	"&":  false,
	"<<": false,
	">>": false,
}

type LitKind uint8

const (
	IntLit LitKind = iota + 1
	FloatLit
	StringLit
)
