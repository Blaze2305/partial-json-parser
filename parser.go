package partialparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/blaze2305/partial-json-parser/options"
)

type jsonCompletion struct {
	index  int
	string string
}

func formatJson(jsonString string) string {
	dest := &bytes.Buffer{}
	json.Indent(dest, []byte(jsonString), "", " ")
	return dest.String()
}

// ParseMalformedString - Parses a malformed / incomplete string and tries to see if it can be compelted. Takes the string, options on what types to consider and formating enable bool
//
// If the string cannot be parsed/fixed, returns an error , else returns the "complete" string
//
// If the format bool is set to true , golangs internal encoding/json parser is using to parse the fixed string and pretty print it using json.Indent.
// NOTE: if your json contains +/- INF or NaN , and format is `true` The field will be missing as go does not support that in. See notes in the json.Marshall function [here](https://pkg.go.dev/encoding/json#Marshal)
//
// Options is a list of all the types the parser can try using to fix the json. Check the Readme for more info on how to use it.
func ParseMalformedString(malformed string, options options.TypeOptions, format bool) (string, error) {

	str := strings.TrimSpace(malformed)
	if len(str) == 0 {
		return "", fmt.Errorf("string is empty; cannot parse")
	}

	jsonString, err := parseJson(malformed, options)
	if err != nil {
		return "", err
	}

	if format {
		return formatJson(jsonString), nil
	} else {
		return jsonString, err
	}
}

func skipBlank(text string, index int) int {
	i := index
	for i < len(text) && (unicode.IsSpace(rune(text[i])) || text[i] == '\n') {
		i += 1
	}
	return i
}

func parseJson(jsonString string, allowed options.TypeOptions) (string, error) {

	i := skipBlank(jsonString, 0)
	value := jsonString[i:]

	completion, err := completeAny(value, allowed, true)

	if err != nil {
		return "", fmt.Errorf("not enough data to fix json string %s", err)
	}

	return value[:completion.index] + completion.string, nil
}

func completeAny(jsonString string, allowed options.TypeOptions, topLevel bool) (*jsonCompletion, error) {
	switch char := jsonString[0]; {
	case char == '"':
		return completeString(jsonString, allowed)
	case strings.ContainsRune("0123456789", rune(char)):
		return completeNumber(jsonString, allowed, topLevel)
	case char == '[':
		return completeArray(jsonString, allowed)
	case char == '{':
		return completeObject(jsonString, allowed)
	case char == '-': // handles negative numbers
		if len(jsonString) == 1 {
			return nil, fmt.Errorf("cannot parse singular '-'")
		} else if jsonString[1] != 'I' { // this just checks if we're dealing with negative inf
			return completeNumber(jsonString, allowed, topLevel)
		}

	}

	// handle NULL
	if strings.HasPrefix(jsonString, "null") {
		return &jsonCompletion{
			index: 4,
		}, nil
	}

	if strings.HasPrefix("null", jsonString) {
		if options.NULL|allowed == allowed {
			return &jsonCompletion{
				index: 0,
			}, nil
		}
		return nil, fmt.Errorf("cannot parse null with given options")
	}

	// handle boolean true
	if strings.HasPrefix(jsonString, "true") {
		return &jsonCompletion{
			index: 4,
		}, nil
	}

	if strings.HasPrefix("true", jsonString) {
		if options.BOOL|allowed == allowed {
			return &jsonCompletion{
				index: 0,
			}, nil
		}
		return nil, fmt.Errorf("cannot parse bool with given options")
	}

	// handle boolean false
	if strings.HasPrefix(jsonString, "false") {
		return &jsonCompletion{
			index: 5,
		}, nil
	}

	if strings.HasPrefix("false", jsonString) {
		if options.BOOL|allowed == allowed {
			return &jsonCompletion{
				index: 0,
			}, nil
		}
		return nil, fmt.Errorf("cannot parse bool with given options")
	}

	// handle infinity
	if strings.HasPrefix(jsonString, "Infinity") {
		return &jsonCompletion{
			index: 8,
		}, nil
	}

	if strings.HasPrefix("Infinity", jsonString) {
		if options.INFINITY|allowed == allowed {
			return &jsonCompletion{
				index: 0,
			}, nil
		}
		return nil, fmt.Errorf("cannot parse Infinity with given options")
	}

	// handle negative infinity
	if strings.HasPrefix(jsonString, "-Infinity") {
		return &jsonCompletion{
			index: 9,
		}, nil
	}

	if strings.HasPrefix("-Infinity", jsonString) {
		if options.NEG_INFINITY|allowed == allowed {
			return &jsonCompletion{
				index: 0,
			}, nil
		}
		return nil, fmt.Errorf("cannot parse -Infinity with given options")
	}

	// handle NaN
	if strings.HasPrefix(jsonString, "NaN") {
		return &jsonCompletion{
			index: 3,
		}, nil
	}

	if strings.HasPrefix("NaN", jsonString) {
		if options.NAN|allowed == allowed {
			return &jsonCompletion{
				index: 0,
			}, nil
		}
		return nil, fmt.Errorf("cannot parse NaN with given options")
	}

	return nil, fmt.Errorf("MalformedJSON(unexpected char %c)", jsonString[0])

}

func completeObject(jsonString string, allowed options.TypeOptions) (*jsonCompletion, error) {

	i := 1
	j := 1
	length := len(jsonString)

	for j < length {
		j = skipBlank(jsonString, j)
		if j >= length {
			break
		}

		if jsonString[j] == '}' {
			return &jsonCompletion{
				index: j + 1,
			}, nil
		}

		key, err := completeString(jsonString[j:], allowed)
		if err != nil || key.string != "" {
			// if we can't parsek the key, then we just do the next best thing and complete the obj as an empty obj
			// same if the key is incomplete
			if options.OBJ|allowed == allowed {
				return &jsonCompletion{
					index:  i,
					string: "}",
				}, nil
			}
			return nil, fmt.Errorf("cannot parse object with given options")
		}

		// if we can parse the key, we move the index by that amount and continue parsing
		j += key.index

		j = skipBlank(jsonString, j)
		if j >= length {
			break
		}

		if jsonString[j] != ':' {
			return nil, fmt.Errorf("MalformedJSON( expected \":\" got %c)", jsonString[j])
		}
		j += 1

		j = skipBlank(jsonString, j)
		if j >= length {
			break
		}

		result, err := completeAny(jsonString[j:], allowed, false)
		if err != nil { // cant complete the objects value, so just end it and make it an empty object again
			if options.OBJ|allowed == allowed {
				return &jsonCompletion{
					index:  i,
					string: "}",
				}, nil
			}
			return nil, fmt.Errorf("cannot parse object with given options")
		}

		// if the string in the result has some char in it, that means we can add the final } and complete the object, because it means that all the item(s) in the object are fine
		if result.string != "" {
			if options.OBJ|allowed == allowed {
				return &jsonCompletion{
					index:  j + result.index,
					string: result.string + "}",
				}, nil
			}
			return nil, fmt.Errorf("cannot parse object with given options")
		}

		// this means that the first key value par in the object are fine, but we need to check the other items
		j += result.index
		i = j

		j = skipBlank(jsonString, j)
		if j >= length {
			break
		}

		if jsonString[j] == ',' {
			j += 1
		} else if jsonString[j] == '}' {
			return &jsonCompletion{
				index: j + 1,
			}, nil
		} else {
			return nil, fmt.Errorf("MalformedJSON(expected \",\" or \"}\" got %c)", jsonString[j])
		}
	}

	// if we've reached the end of the string, we throw the ] at the last known good point and return
	if options.OBJ|allowed == allowed {
		return &jsonCompletion{
			index:  i,
			string: "}",
		}, nil
	}
	return nil, fmt.Errorf("cannot parse object with given options")

}

func completeString(jsonString string, allowed options.TypeOptions) (*jsonCompletion, error) {
	index := 1
	charEscaped := false
	stringLength := len(jsonString)

	for index < stringLength && (jsonString[index] != '"' || charEscaped) {
		if jsonString[index] == '\\' {
			charEscaped = !charEscaped
		} else {
			charEscaped = false
		}
		index += 1
	}

	if index < stringLength {
		return &jsonCompletion{
			index: index + 1,
		}, nil
	}

	if options.STR|allowed != allowed {
		return nil, fmt.Errorf("cannot complete malformed json")
	}

	// handle unicode and hex strings
	// handle \uXXXX
	u_index := strings.LastIndex(jsonString, "\\u")
	if u_index != -1 {
		if u_index+6 == stringLength {
			return &jsonCompletion{
				index:  u_index + 6,
				string: "\"",
			}, nil
		}
		return &jsonCompletion{
			index:  u_index,
			string: "\"",
		}, nil
	}

	// handle \UXXXXXXXX
	U_index := strings.LastIndex(jsonString, "\\U")
	if U_index != -1 {
		if U_index+10 == stringLength {
			return &jsonCompletion{
				index:  U_index + 10,
				string: "\"",
			}, nil
		}
		return &jsonCompletion{
			index:  U_index,
			string: "\"",
		}, nil
	}

	// handle \xXX
	x_index := strings.LastIndex(jsonString, "\\x")
	if x_index != -1 {
		if x_index+4 == stringLength {
			return &jsonCompletion{
				index:  x_index + 4,
				string: "\"",
			}, nil
		}
		return &jsonCompletion{
			index:  x_index,
			string: "\"",
		}, nil
	}

	if charEscaped {
		return &jsonCompletion{
			index:  index - 1,
			string: "\"",
		}, nil
	}

	return &jsonCompletion{
		index:  index,
		string: "\"",
	}, nil
}

func completeArray(jsonString string, allowed options.TypeOptions) (*jsonCompletion, error) {
	i := 1
	j := 1
	length := len(jsonString)

	for j < length {
		j = skipBlank(jsonString, j)
		if j >= length {
			break
		}

		if jsonString[j] == ']' {
			return &jsonCompletion{
				index: j + 1,
			}, nil
		}

		result, err := completeAny(jsonString[j:], allowed, false)
		if err != nil { // cant complete the array, so just end it and make it an empty array
			if options.ARR|allowed == allowed {
				return &jsonCompletion{
					index:  i,
					string: "]",
				}, nil
			}
			return nil, fmt.Errorf("cannot parse array with given options")
		}

		// if the string in the result has some char in it, that means we can add the final ] and complete the array, because it means that all the item(s) in the array is fine
		if result.string != "" {
			if options.ARR|allowed == allowed {
				return &jsonCompletion{
					index:  j + result.index,
					string: result.string + "]",
				}, nil
			}
			return nil, fmt.Errorf("cannot parse array with given options")
		}

		// this means that the first item in the array is fine, but we need to check the other items
		j += result.index
		i = j

		j = skipBlank(jsonString, j)
		if j >= length {
			break
		}

		if jsonString[j] == ',' {
			j += 1
		} else if jsonString[j] == ']' {
			return &jsonCompletion{
				index: j + 1,
			}, nil
		} else {
			return nil, fmt.Errorf("MalformedJSON(expected \",\" or \"]\" got %c)", jsonString[j])
		}
	}

	// if we've reached the end of the string, we throw the ] at the last known good point and return
	if options.ARR|allowed == allowed {
		return &jsonCompletion{
			index:  i,
			string: "]",
		}, nil
	}
	return nil, fmt.Errorf("cannot parse array with given options")
}

func completeNumber(jsonString string, allowed options.TypeOptions, topLevel bool) (*jsonCompletion, error) {

	i := 1
	length := len(jsonString)

	// move forwards while we still have nummbers ;
	// NOTE : this includes exponents in the form x-/+ey and decimals
	for i < length && strings.ContainsRune("0123456789.+-eE", rune(jsonString[i])) {
		i += 1
	}

	specialNum := false
	for strings.ContainsRune(".-+eE", rune(jsonString[i-1])) {
		i -= 1
		specialNum = true
	}

	if specialNum || i == length && !topLevel {
		if options.NUM|allowed == allowed {
			return &jsonCompletion{
				index: i,
			}, nil
		}
		return nil, fmt.Errorf("cannot parse number with given options")
	}

	return &jsonCompletion{
		index: i,
	}, nil
}
