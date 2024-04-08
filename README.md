# partial-json-parser

Inspired by a reddit post in r/golang and a python library with the same name. This library helps parse and display partial json inputs, trying it's best to fix the missing pieces.

Sometimes when you're streaming json data , there could be lag. And before we receive the last bit of data, the JSON is broken and malformed. But we still might want to display it. This is where the partial-json-parser comes in.

Check out the demo [here](https://blaze2305.github.io/partial-json-parser/)

## Installation
```bash
# Install the lib.
go get -u github.com/blaze2305/partial-json-parser
```
## Usage

Add the import at the top 
```go
import "github.com/blaze2305/partial-json-parser"
```
and use the module as `partialparser`.

Eg
```go
package main

import (
	"fmt"
	"os"

	"github.com/blaze2305/partial-json-parser"
	"github.com/blaze2305/partial-json-parser/options"
)

func main() {
	str := `{"foo":"bar`

	value,err := partialparser.ParseMalformedString(str, options.ALL)
	if err!=nil{
	    fmt.Println("Could not parse json; ERR:",err);
	    os.Exit(1)
	}
	fmt.Println(value)
}
```

Running the above will give you `{"foo":"bar"}`

## Parsing options

When parsing the partial JSON , if you want to check if the string can be converted into one or more "type(s)", you can pass that as an option to the `ParseMalformedString` function. 

The available options can be found in the `options` package and are 

- `STR` : Allows the parser to try fixing string issues
- `NUM` : Allows it to try fixinf number issues
- `ARR` : Allows it to try fixing array issues
- `OBJ` : Allows it to try fixing Object issues
- `NULL` : Allows it to fix NULL
- `BOOL` : Allows it to fix boolean values
- `NAN` : Allows it to try fixing NAN
- `INFINITY` : ALlows it to try fixing Infinity
- `NEG_INFINITY` : Allows it to try fixing Negative Infinity
- `INF` : A combination of INFINITY and NEG_INFINITY
- `SPECIAL` : A combination of NULL BOOL INF and NAN
- `ATOM` : A combination of STR NUM and SPECIAL
- `COLLECTION` : A combination of ARR and OBJ
- `ALL` : All possible types

The values can be accessed by importing `github.com/blaze2305/partial-json-parser/options` and using `options.<whatever>` eg : `options.STR`.

#### NOTE: 
You can create your own combinations by doing a bitwise or `|` amongst the options
eg : `options.NUM | options.ARR | options.OBJ` will allow the parser to consider `NUM` , `ARR` or `OBJ` when parsing.


Eg:
```golang
package main

import (
	"fmt"
	"os"

	"github.com/blaze2305/partial-json-parser"
	"github.com/blaze2305/partial-json-parser/options"
)

func main() {
	str := `["a",{"a":123`

	value, err := partialparser.ParseMalformedString(str, options.NUM|options.ARR|options.OBJ)
	if err != nil {
		fmt.Println("Could not parse json; ERR:", err)
		os.Exit(1)
	}
	fmt.Println(value)
}
```
The above , when run, will output `["a",{"a":123}]`

If you don't allow a type to be fixed by the parser, if/when it encounters that issue it will error out.

