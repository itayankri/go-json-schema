package jsonwalker

import (
	"encoding/json"
	"strconv"
	"strings"
)

// JsonPointer is a type that represents a json pointer by keeping
// a list of json tokens. A json token may be a key of json object
// or an index of json array.
type JsonPointer []string

// NewJsonPointer is a function that create a JsonPointer according
// to a specific json pointer of type string.
// It returns a JsonPointerSyntaxError if the string does not have
// a '/' prefix.
func NewJsonPointer(path string) (JsonPointer, error) {
	// If path equals to "", return an empty-reference JsonPointer.
	if len(path) == 0 || path == "/" {
		return JsonPointer{}, nil
	}

	// If the first character of path is not '/' return JsonPointerSyntaxError
	if path[0] != '/' {
		return nil, JsonPointerSyntaxError{
			"first character of non-empty reference must be '/'",
			path,
		}
	}

	// Split path by '/' in order to get a []string of json tokens
	tokens := strings.Split(path, "/")

	// Convert the []string to JonPointer and omit the first string
	// in the slice because when the delimiter is the first character
	// in a string, Split return "" in the slice's first cell.
	return JsonPointer(tokens[1:]), nil
}

// Evaluate is a receiver function that searches for the JsonPointer's data
// in a given json value.
func (jp JsonPointer) Evaluate(jsonData json.RawMessage) (interface{}, error) {
	var data interface{}

	// Unmarshal jsonData (which under the hood is a slice of bytes).
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	// If the JsonPointer is an empty reference, return the whole data.
	if len(jp) == 0 {
		return data, nil
	}

	// Evaluate each token and put the returned value is data in order
	// to evaluate the next token.
	for _, token := range jp {
		data, err = evaluateToken(token, data)
		if err != nil {
			return nil, InvalidJsonPointerError{
				"/" + strings.Join(jp, "/"),
				err.Error(),
			}
		}
	}

	return data, nil
}

// evaluateToken is a function that get a json token and some json data and
// returns the correct value from the json data.
func evaluateToken(token string, jsonData interface{}) (interface{}, error) {
	// Json values in go can be represented in three ways.
	// - Json object is represented as map[string]interface{}
	// - Json array is represented as []interface{}
	// - Atomic value (which is the default case here)
	switch v := jsonData.(type) {
	case map[string]interface{}:
		{
			if prop, ok := v[token]; ok {
				return prop, nil
			}

			return nil, MissingJsonTokenError(token)
		}
	case []interface{}:
		{
			index, err := strconv.Atoi(token)
			if err != nil {
				return nil, err
			}

			return v[index], nil
		}
	default:
		{
			return nil, MissingJsonTokenError(token)
		}
	}
}
