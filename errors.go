package jsonvalidator

import "fmt"

type KeywordValidationError struct {
	keyword string
	reason  string
}

func (e KeywordValidationError) Error() string {
	return fmt.Sprintf("\"" + e.keyword + "\" validation failed, reason: " + e.reason)
}

type SchemaValidationError struct {
	path string
	err  string
}

func (e SchemaValidationError) Error() string {
	var jsonPath string
	if e.path == "" {
		jsonPath = "/"
	} else {
		jsonPath = e.path
	}

	return fmt.Sprintf("validation failed in path " +
		jsonPath +
		": " +
		e.err)
}

type SchemaCompilationError struct {
	path string
	err  string
}

func (e SchemaCompilationError) Error() string {
	return fmt.Sprintf("schema compilation failed in path " + e.path + ": " + e.err)
}

type InvalidDraftError string

func (e InvalidDraftError) Error() string {
	return fmt.Sprintf("draft " + string(e) + " is not supported by JsonValidator")
}

type InvalidReferenceError struct {
	schemaURI string
	fragment  string
	err       string
}

func (e InvalidReferenceError) Error() string {
	fragment := "/"
	if e.fragment != "" {
		fragment = e.fragment
	}

	return fmt.Sprintf(e.err + ": schema id - " + e.schemaURI + ", fragment - " + fragment)
}
