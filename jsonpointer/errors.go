package jsonwalker

import "fmt"

type JsonArrayIndexError int

func (e JsonArrayIndexError) Error() string {
	return fmt.Sprintf("Index %d out of range", e)
}

type JsonPointerSyntaxError struct {
	err  string
	path string
}

func (e JsonPointerSyntaxError) Error() string {
	return fmt.Sprintf("JsonPointer syntax error for \"%s\" - %s", e.path, e.err)
}

type InvalidJsonPointerError struct {
	path   string
	reason string
}

func (e InvalidJsonPointerError) Error() string {
	return fmt.Sprintf("invalid json pointer \"%s\": %s", e.path, e.reason)
}

type MissingJsonTokenError string

func (e MissingJsonTokenError) Error() string {
	return fmt.Sprintf("token \"" + string(e) + "\" is missing")
}
