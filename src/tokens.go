package main

var n = struct{}{}

func tmOk(s string) bool {
	_, ok := tokenMap[s]
	return ok
}

// True are keywords that can be fast-forwarded to in case of error.
var tokenMap = map[string]int{
	"EOF":         0,
	"name":        0,
	"literal":     0,
	"=":           0,
	":=":          0,
	"(":           0,
	"[":           0,
	"{":           0,
	")":           0,
	"]":           0,
	"}":           0,
	",":           0,
	";":           0,
	":":           0,
	"?":           1,
	"@":           2,
	"#":           2,
	".":           0,
	"...":         0,
	"break":       0,
	"case":        0,
	"const":       0,
	"continue":    0,
	"default":     0,
	"define":      0,
	"else":        0,
	"fallthrough": 0,
	"for":         0,
	"forr":        0,
	"func":        0,
	"if":          0,
	"loop":        0,
	"map":         0,
	"range":       0,
	"return":      0,
	"type":        0,
	"var":         0,
	"while":       0,

	"!":  0,
	"||": 3,
	"&&": 3,
	"==": 4,
	"!=": 4,
	"<":  4,
	"<=": 4,
	">":  4,
	">=": 4,
	"+":  5,
	"+=": 0,
	"++": 0,
	"--": 0,
	"-":  5,
	"-=": 0,
	"|":  5,
	"^":  5,
	"*":  5,
	"*=": 0,
	"/":  5,
	"/=": 0,
	"%":  5,
	"%=": 0,
	"&":  5,
	"<<": 5,
	">>": 5,
}

type LitKind uint8

const (
	IntLit LitKind = iota + 1
	FloatLit
	StringLit
)
