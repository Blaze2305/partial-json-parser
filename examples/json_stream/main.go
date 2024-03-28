package main

import (
	"fmt"
	"strings"
	"time"

	partialparser "github.com/blaze2305/partial-json-parser"
	"github.com/blaze2305/partial-json-parser/options"
)

func sendPartialJson(c chan<- string) {
	jsonString := `
	{
		"foo":"bar",
		"length":10,
		"person":{
			"age":100,
			"hp":35.6,
			"def":[
				"item1",
				"item2",
				3,
				4.5e+6
			]
		},
		"pokemon":[
			{
				"name":"cinderace",
				"height":1.4,
				"gamesAvailable":{
					"sw/sh":true,
					"bd/sp":false
				},
				"coolness":Infinity
			}
		]

	}
`

	items := strings.Split(jsonString, "\n")
	for _, item := range items {
		c <- item
		time.Sleep(time.Second * 1)
	}

	c <- "DONESENDING"

}

func main() {
	c := make(chan string)

	go sendPartialJson(c)

	str := ""

loop:
	for {
		select {
		case x := <-c:
			switch x {
			case "DONESENDING":
				break loop
			default:
				fmt.Println("received :", x)
				str += x
				// setting format arg to false here, because the input contains Infinity and go doesnt like that ;-;
				jsonValue, err := partialparser.ParseMalformedString(str, options.ALL, false)
				if err != nil {
					fmt.Println("err", err)
					continue
				}
				fmt.Println("parsed json", jsonValue)
			}
		}
		fmt.Println("--------------\n")
	}
	fmt.Println("DONE")
}
