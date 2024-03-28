package partialparser

import (
	"fmt"
	"strings"

	"github.com/blaze2305/partial-json-parser/options"
)

type jsonCompletion struct {
	index  int
	string string
}

func ParseMalformedString(malformed string, options options.TypeOptions) (string, error) {

	str := strings.TrimSpace(malformed)
	if len(str) == 0 {
		return "", fmt.Errorf("string is empty; cannot parse")
	}

	return parseJson(str, options)

}

func skipBlank(text string, index int) int {
	i := index
	for i < len(text) && text[i] == ' ' {
		i += 1
	}
	return i
}

func parseJson(jsonString string, allowed options.TypeOptions) (string, error) {
	completion, err := completeAny(jsonString, allowed, true)

	if err != nil {
		return "", fmt.Errorf("not enough data to fix json string")
	}

	return jsonString[:completion.index] + completion.string, nil
}

func completeAny(jsonString string, allowed options.TypeOptions, topLevel bool) (*jsonCompletion, error) {
	value := strings.TrimLeft(jsonString, " ")
	switch char := value[0]; {
	case char == '"':
		return completeString(value, allowed)
	case strings.ContainsRune("0123456789", rune(char)):
		return completeNumber(value, allowed, topLevel)
	case char == '[':
		return completeArray(value, allowed)
	case char == '{':
		return completeObject(value, allowed)
	}

	return nil, fmt.Errorf("MalformedJSON(unexpected char %c)", value[0])

}

func completeObject(jsonString string, allowed options.TypeOptions) (*jsonCompletion, error) {
	return nil, nil
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

	for j < len(jsonString) {
		j = skipBlank(jsonString, j)
		if j >= len(jsonString) {
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
		}

		// if the string in the result has some char in it, that means we can add the final ] and complete the array, because it means that all the item(s) in the array is fine
		if result.string != "" {
			if options.ARR|allowed == allowed {
				return &jsonCompletion{
					index:  j + result.index,
					string: result.string + "]",
				}, nil
			}
			return nil, fmt.Errorf("cannot parse string with given options")
		}

		// this means that the first item in the array is fine, but we need to check the other items
		j += result.index
		i = j

		j = skipBlank(jsonString, j)
		if j >= len(jsonString) {
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
	return nil, fmt.Errorf("cannot parse string with given options")
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
