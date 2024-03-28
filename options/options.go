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

	ALL TypeOptions = 1<<iota - 1
)
