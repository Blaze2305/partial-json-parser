package options

// specify what type options are available to parse the partial json
type TypeOptions int

const (
	STR TypeOptions = 1 << iota
	NUM
	ARR
	OBJ
	NULL
	BOOL
	NAN
	INFINITY
	NEG_INFINITY

	INF        = INFINITY | NEG_INFINITY
	SPECIAL    = NULL | BOOL | INF | NAN
	ATOM       = STR | NUM | SPECIAL
	COLLECTION = ARR | OBJ
	ALL        = ATOM | COLLECTION
)
