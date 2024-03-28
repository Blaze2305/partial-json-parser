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

func parseJson(jsonString string, allowed options.TypeOptions) (string, error) {
	completion, err := completeAny(jsonString, allowed)

	if err != nil {
		return "", fmt.Errorf("not enough data to fix json string")
	}

	return jsonString[:completion.index] + completion.string, nil
}

func completeAny(jsonString string, allowed options.TypeOptions) (*jsonCompletion, error) {
	value := strings.TrimLeft(jsonString, " ")

	switch char := value[0]; {
	case char == '"':
		return completeString(value, allowed)
	case strings.ContainsRune("0123456789", rune(char)):
		return completeNumber(value, allowed)
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
			index:  index + 1,
			string: "",
		}, nil
	}

	if options.STR&allowed != allowed {
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
	return nil, nil
}

func completeNumber(jsonString string, allowed options.TypeOptions) (*jsonCompletion, error) {
	return nil, nil
}
