//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	partialparser "github.com/blaze2305/partial-json-parser"
	"github.com/blaze2305/partial-json-parser/options"
)

func ParsePartialJson(this js.Value, inputs []js.Value) interface{} {

	jsonString := inputs[0].String()

	output, err := partialparser.ParseMalformedString(jsonString, options.ALL, false)
	if err != nil {
		return fmt.Sprintf("Error parsing : %s", err)
	}

	return output
}

func main() {
	c := make(chan bool)

	js.Global().Set("parsePartialJson", js.FuncOf(ParsePartialJson))

	<-c
}
