package jsonvalidator

import (
	"encoding/json"

	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/itayankri/go-jsonvalidator/formatchecker"
)

/*
Implemented keywordValidators:
> enum: 					V
> _const: 					V
> _type: 					V ***
> minLength: 				V
> maxLength: 				V
> pattern: 					V
> format: 					X
> multipleOf: 				V
> minimum: 					V
> maximum: 					V
> exclusiveMinimum: 		V
> exclusiveMaximum: 		V
> properties: 				V
> additionalProperties: 	V
> required: 				V
> propertyNames: 			V
> dependencies: 			V
> patternProperties: 		V
> minProperties: 			V
> maxProperties: 			V
> items: 					V ***
> contains: 				V
> additionalItems: 			V
> minItems: 				V
> maxItems: 				V
> uniqueItems: 				V
> anyOf: 					V
> allOf: 					V
> oneOf: 					V
> not: 						V
> _if: 						V
> _then: 					V
> _else: 					V

*** These keywords are being un-marshaled in their validate() function.
	We need to find a way to do that on startup and not on runtime.

*/

// Valid values for "format" fields
const (
	FORMAT_DATE_TIME             = "date-time"
	FORMAT_TIME                  = "time"
	FORMAT_DATE                  = "date"
	FORMAT_EMAIL                 = "email"
	FORMAT_IDN_EMAIL             = "idn-email"
	FORMAT_HOSTNAME              = "hostname"
	FORMAT_IDN_HOSTNAME          = "idn-hostname"
	FORMAT_IPV4                  = "ipv4"
	FORMAT_IPV6                  = "ipv6"
	FORMAT_URI                   = "uri"
	FORMAT_URI_REFERENCE         = "uri-reference"
	FORMAT_IRI                   = "iri"
	FORMAT_IRI_REFERENCE         = "iri-reference"
	FORMAT_URI_TEMPLATE          = "uri-template"
	FORMAT_JSON_POINTER          = "json-pointer"
	FORMAT_RELATIVE_JSON_POINTER = "relative-json-pointer"
	FORMAT_REGEX                 = "regex"
)

type keywordValidator interface {
	validate(string, jsonData, string) error
}

/*****************/
/** Annotations **/
/*****************/

type ref string

func (r ref) validateByRef(jsonPath string, jsonData []byte, rootSchemaID string) error {
	splittedRef := strings.Split(string(r), "#")
	schemaURI := splittedRef[0]
	fragment := splittedRef[1]

	// If the schemaURI is empty string it means that the reference points to a schema
	// in the local schema (for example #/definitions/x), so we want to use the rootSchemaID
	// in order to get the current root-schema from the rootSchemaPool.
	if schemaURI == "" {
		schemaURI = rootSchemaID
	}

	// If the root-schema exists in the rootSchemaPool, validate the data according to the
	// fragment.
	// Else, return an error
	if rootSchema, ok := rootSchemaPool[schemaURI]; ok {
		// If the fragment is an empty fragment, validate the data against the root-schema.
		// Else, validate the data against the sub-schema that the fragment points to.
		if fragment != "" {
			// If the referenced sub-schema exists, validate the data against it.
			// Else, return an error
			if subSchema, ok := rootSchema.subSchemaMap[fragment]; ok {
				return subSchema.validateJsonData(jsonPath, jsonData, rootSchemaID)
			} else {
				return InvalidReferenceError{
					schemaURI: schemaURI,
					fragment:  fragment,
					err:       "could not find fragment in the referenced root schema",
				}
			}
		} else {
			return rootSchema.validateJsonData(jsonPath, jsonData, rootSchemaID)
		}
	} else {
		return InvalidReferenceError{
			schemaURI: schemaURI,
			fragment:  fragment,
			err:       "could not find the referenced root schema",
		}
	}
}

type schema string
type id string
type comment string
type title string
type description string
type examples []interface{}
type definitions map[string]*JsonSchema
type _default json.RawMessage

func (d *_default) UnmarshalJSON(data []byte) error {
	*d = data
	return nil
}

/**********************/
/** Generic Keywords **/
/**********************/

type _type json.RawMessage

func (t *_type) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	var data interface{}

	// First we need to unmarshal the json data.
	err := json.Unmarshal(*t, &data)
	if err != nil {
		return err
	}

	// The "type" field in json schema can be represented by two different values:
	// - string - the inspected value can be only one json type.
	// - array - the inspected value can be a variety of json types.
	// - default - the schema is incorrect.
	switch typeFromSchema := data.(type) {
	case []interface{}:
		{
			// If we arrived this loop, it means "type" is an array of types.
			// We need to go over the existing types and perform
			// "json type assertion" of jsonData and the current json type.
			for _, typeFromList := range typeFromSchema {
				// A json type must be represented by a string.
				if v, ok := typeFromList.(string); ok {
					// Perform the "json type assertion"
					err := assertJsonType(v, jsonData.value)

					// If the assertion succeeded, return true
					if err == nil {
						return nil
					}
				} else {
					return KeywordValidationError{
						"type",
						"\"type\" field in schema must be string or array of strings",
					}
				}
			}

			// JsonTypeMismatchError
			return KeywordValidationError{
				"type",
				"inspected value does not match any of the valid types in the schema",
			}
		}
	case string:
		{
			// In this case, there is only one valid type, so we
			// perform "json type assertion" of the json type and jsonData.
			return assertJsonType(typeFromSchema, jsonData.value)
		}
	default:
		{
			return KeywordValidationError{
				"type",
				"\"type\" field in schema must be string or array of strings",
			}
		}
	}
}

// assertJsonType is a function that gets a jsonType and some jsonData and
// returns true if the value belongs to the type.
// If it is not, the function will return an appropriate error.
func assertJsonType(jsonType string, jsonData interface{}) error {
	switch jsonType {
	case TYPE_OBJECT:
		{
			if _, ok := jsonData.(map[string]interface{}); ok {
				return nil
			} else {
				return KeywordValidationError{
					"type",
					"inspected value expected to be a json object",
				}
			}
		}
	case TYPE_ARRAY:
		{
			if _, ok := jsonData.([]interface{}); ok {
				return nil
			} else {
				return KeywordValidationError{
					"type",
					"inspected value expected to be a json array",
				}
			}
		}
	case TYPE_STRING:
		{
			if _, ok := jsonData.(string); ok {
				return nil
			} else {
				return KeywordValidationError{
					"type",
					"inspected value expected to be a json string",
				}
			}
		}
	case TYPE_INTEGER:
		{
			if value, ok := jsonData.(float64); ok && value == float64(int(value)) {
				return nil
			} else {
				return KeywordValidationError{
					"type",
					"inspected value expected to be a json integer",
				}
			}
		}
	case TYPE_NUMBER:
		{
			if _, ok := jsonData.(float64); ok {
				return nil
			} else {
				return KeywordValidationError{
					"type",
					"inspected value expected to be a json number",
				}
			}
		}
	case TYPE_BOOLEAN:
		{
			if _, ok := jsonData.(bool); ok {
				return nil
			} else {
				return KeywordValidationError{
					"type",
					"inspected value expected to be a json boolean",
				}
			}
		}
	case TYPE_NULL:
		{
			if jsonData == nil {
				return nil
			} else {
				return KeywordValidationError{
					"type",
					"inspected value expected to be a json null",
				}
			}
		}
	default:
		{
			return KeywordValidationError{
				"type",
				"invalid json type " + jsonType,
			}
		}
	}
}

func (t *_type) UnmarshalJSON(data []byte) error {
	*t = data
	return nil
}

func (t *_type) MarshalJSON() ([]byte, error) {
	return []byte(*t), nil
}

type enum []interface{}

func (e enum) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// Iterate over the items in "enum" array.
	for _, item := range e {
		// Marshal the item from "enum" array back comparable value that does
		// not require type assertion.
		rawEnumItem, err := json.Marshal(item)
		if err != nil {
			return nil
		}

		// Convert both of the byte arrays to string for more convenient
		// comparison. If they are equal, the data is valid against "enum".
		if string(rawEnumItem) == string(jsonData.raw) {
			return nil
		}
	}

	// If we arrived here it means that the inspected value is not equal
	// to any of the values in "enum".
	return KeywordValidationError{
		"enum",
		"inspected value does not match any of the items in \"enum\" array",
	}
}

type _const json.RawMessage

func (c *_const) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// Convert both of the byte arrays to string for more convenient
	// comparison. If they are equal, the data is valid against "const".
	if string(*c) == string(jsonData.raw) {
		return nil
	} else {
		return KeywordValidationError{
			"const",
			"inspected value not equal to \"" + string(*c) + "\"",
		}
	}
}

func (c *_const) UnmarshalJSON(data []byte) error {
	// In this function we Unmarshal and then Marshal again
	// the argument data in order to remove special characters
	// like \n \t \r etc.

	var unmarshaledData interface{}

	err := json.Unmarshal(data, &unmarshaledData)
	if err != nil {
		return err
	}

	rawConst, err := json.Marshal(unmarshaledData)
	if err != nil {
		return err
	}

	*c = rawConst
	return nil
}

/*********************/
/** String Keywords **/
/*********************/

type minLength int

func (ml *minLength) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// If jsonData is a string, validate its length,
	// else, return a KeywordValidationError
	if v, ok := jsonData.value.(string); ok {
		if len(v) >= int(*ml) {
			return nil
		} else {
			return KeywordValidationError{
				"minLength",
				"inspected string is less than " + strconv.Itoa(int(*ml)),
			}
		}
	}

	return nil
}

type maxLength int

func (ml *maxLength) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// If jsonData is a string, validate its length,
	// else, return a KeywordValidationError
	if v, ok := jsonData.value.(string); ok {
		if len(v) <= int(*ml) {
			return nil
		} else {
			return KeywordValidationError{
				"maxLength",
				"inspected string is greater than " + strconv.Itoa(int(*ml)),
			}
		}
	}

	return nil
}

type pattern string

func (p *pattern) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// If jsonData is a string, validate its length,
	// else, return a KeywordValidationError
	if v, ok := jsonData.value.(string); ok {
		match, err := regexp.MatchString(string(*p), v)

		// The pattern or the value is not in the right format (string)
		if err != nil {
			return KeywordValidationError{
				"pattern",
				err.Error(),
			}
		}

		if match {
			return nil
		} else {
			return KeywordValidationError{
				"pattern",
				"value " + v + " does not match to pattern" + string(*p),
			}
		}
	}

	return nil
}

type format string

func (f *format) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	if v, ok := jsonData.value.(string); ok {
		switch string(*f) {
		case FORMAT_DATE_TIME:
			if err := formatchecker.IsValidDateTime(v); err != nil {
				return KeywordValidationError{
					"format",
					"date-time incorrectly formatted " + err.Error(),
				}
			}
		case FORMAT_DATE:
			if err := formatchecker.IsValidDate(v); err != nil {
				return KeywordValidationError{
					"format",
					"date incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_TIME:
			if err := formatchecker.IsValidTime(v); err != nil {
				return KeywordValidationError{
					"format",
					"time incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_EMAIL:
			if err := formatchecker.IsValidEmail(v); err != nil {
				return KeywordValidationError{
					"format",
					"email incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_IDN_EMAIL:
			if err := formatchecker.IsValidIdnEmail(v); err != nil {
				return KeywordValidationError{
					"format",
					"idn-email incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_HOSTNAME:
			if err := formatchecker.IsValidHostname(v); err != nil {
				return KeywordValidationError{
					"format",
					"hostname incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_IDN_HOSTNAME:
			if err := formatchecker.IsValidIdnHostname(v); err != nil {
				return KeywordValidationError{
					"format",
					"idn-hostname incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_IPV4:
			if err := formatchecker.IsValidIPv4(v); err != nil {
				return KeywordValidationError{
					"format",
					"ipv4 incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_IPV6:
			if err := formatchecker.IsValidIPv6(v); err != nil {
				return KeywordValidationError{
					"format",
					"ipv6 incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_URI:
			if err := formatchecker.IsValidURI(v); err != nil {
				return KeywordValidationError{
					"format",
					"uri incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_URI_REFERENCE:
			if err := formatchecker.IsValidUriRef(v); err != nil {
				return KeywordValidationError{
					"format",
					"uri-reference incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_IRI:
			if err := formatchecker.IsValidIri(v); err != nil {
				return KeywordValidationError{
					"format",
					"iri incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_IRI_REFERENCE:
			if err := formatchecker.IsValidIriRef(v); err != nil {
				return KeywordValidationError{
					"format",
					"iri-reference incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_URI_TEMPLATE:
			if err := formatchecker.IsValidURITemplate(v); err != nil {
				return KeywordValidationError{
					"format",
					"uri-template incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_JSON_POINTER:
			if err := formatchecker.IsValidJSONPointer(v); err != nil {
				return KeywordValidationError{
					"format",
					"json-pointer incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_RELATIVE_JSON_POINTER:
			if err := formatchecker.IsValidRelJSONPointer(v); err != nil {
				return KeywordValidationError{
					"format",
					"relative-json-pointer incorrectly formatted: " + err.Error(),
				}
			}
		case FORMAT_REGEX:
			if err := formatchecker.IsValidRegex(v); err != nil {
				return KeywordValidationError{
					"format",
					"regex incorrectly formatted: " + err.Error(),
				}
			}
		default:
			return nil
		}
	}

	return nil
}

/*********************/
/** Number Keywords **/
/*********************/

type multipleOf float64

func (mo *multipleOf) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// If jsonData is float64, validate it. Else, return KeywordValidationError
	if v, ok := jsonData.value.(float64); ok {
		if math.Mod(v, float64(*mo)) == 0 {
			return nil
		} else {
			return KeywordValidationError{
				"multipleOf",
				"inspected value is not a multiple of " + strconv.FormatFloat(float64(*mo),
					'f',
					6,
					64),
			}
		}
	}

	return nil
}

type minimum float64

func (m *minimum) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// If jsonData is float64, validate it. Else, return KeywordValidationError
	if v, ok := jsonData.value.(float64); ok {
		if v >= float64(*m) {
			return nil
		} else {
			return KeywordValidationError{
				"minimum",
				"inspected value is less than " + strconv.FormatFloat(float64(*m),
					'f',
					6,
					64),
			}
		}
	}

	return nil
}

type maximum float64

func (m *maximum) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// If jsonData is float64, validate it. Else, return KeywordValidationError
	if v, ok := jsonData.value.(float64); ok {
		if v <= float64(*m) {
			return nil
		} else {
			return KeywordValidationError{
				"maximum",
				"inspected value is greater than " + strconv.FormatFloat(float64(*m),
					'f',
					6,
					64),
			}
		}
	}

	return nil
}

type exclusiveMinimum float64

func (em *exclusiveMinimum) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// If jsonData is float64, validate it. Else, return KeywordValidationError
	if v, ok := jsonData.value.(float64); ok {
		if v > float64(*em) {
			return nil
		} else {
			return KeywordValidationError{
				"exclusiveMinimum",
				"inspected value is not greater than " + strconv.FormatFloat(float64(*em),
					'f',
					6,
					64),
			}
		}
	}

	return nil
}

type exclusiveMaximum float64

func (em *exclusiveMaximum) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// If jsonData is float64, validate it. Else, return KeywordValidationError
	if v, ok := jsonData.value.(float64); ok {
		if v < float64(*em) {
			return nil
		} else {
			return KeywordValidationError{
				"exclusiveMaximum",
				"inspected value is not less than " + strconv.FormatFloat(float64(*em),
					'f',
					6,
					64),
			}
		}
	}

	return nil
}

/*********************/
/** Object Keywords **/
/*********************/

type properties map[string]*JsonSchema

func (p properties) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we need to verify that jsonData is a json object
	if object, ok := jsonData.value.(map[string]interface{}); ok {
		// For each "property" validate it according to its JsonSchema.
		for key, value := range p {
			// Before we try to validate the data against the schema,
			// we make sure that the data actually contains the property.
			if _, ok := object[key]; ok {
				err := value.validateJsonData(jsonPath+"/"+key, jsonData.raw, rootSchemaId)
				if err != nil {
					return err
				}
			}
		}
	}

	// If we arrived here, the validation of all the properties
	// succeeded.
	return nil
}

type additionalProperties struct {
	JsonSchema
	siblingProperties        *properties
	siblingPatternProperties *patternProperties
}

func (ap *additionalProperties) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First we need to verify that jsonData is a json object.
	if object, isObject := jsonData.value.(map[string]interface{}); isObject {
		// Iterate over the properties of the inspected object.
		for property := range object {
			validatedByProperties := false
			validatedByPatternProperties := false

			// Check if the property validated against a schema in 'properties' field
			if (*ap).siblingProperties != nil {
				if _, ok := (*ap.siblingProperties)[property]; ok {
					validatedByProperties = true
				}
			}

			// Check if the property validated against a schema in 'patternProperties' field
			if (*ap).siblingPatternProperties != nil {
				// Iterate over the patterns in "patternProperties" field.
				for pattern := range *ap.siblingPatternProperties {
					// Check if the inspected property matches to the pattern.
					match, err := regexp.MatchString(pattern, property)

					// The pattern or the value is not in the right format (string)
					if err != nil {
						return KeywordValidationError{
							"additionalProperties",
							err.Error(),
						}
					}

					// If there is no match, validate the value of the property against
					// the given schema in "additionalProperties" field.
					if match {
						validatedByPatternProperties = true
					}
				}
			}

			if !validatedByProperties && !validatedByPatternProperties {
				err := (*ap).validateJsonData(jsonPath+"/"+property, jsonData.raw, rootSchemaId)

				// If the validation fails, return an error.
				if err != nil {
					return KeywordValidationError{
						"additionalProperties",
						"property \"" +
							property +
							"\" failed in validation: \n" + err.Error(),
					}
				}
			}
		}
	}

	// If we arrived here, none of the properties failed in validation,
	// and we return true.
	return nil
}

type required []string

func (r required) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we must verify that jsonData is a json object.
	if object, ok := jsonData.value.(map[string]interface{}); ok {
		// For each property in the required list, check if it exists.
		for _, property := range r {
			if object[property] == nil {
				return KeywordValidationError{
					"required",
					"Missing required property - " + property,
				}
			}
		}
	}

	// Is we arrived here, all the properties exist.
	return nil
}

type propertyNames struct {
	JsonSchema
}

func (pn *propertyNames) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we need to verify that jsonData is a json object
	if object, ok := jsonData.value.(map[string]interface{}); ok {
		// Iterate over the object's properties.
		for property := range object {
			// Validate the property name against the schema stored in "propertyNames" field
			err := pn.validateJsonData("", []byte("\""+property+"\""), rootSchemaId)

			// If the property name could be validated against the scheme return an error
			if err != nil {
				return KeywordValidationError{
					"propertyNames",
					"property name \"" + property + "\" failed in validation: " + err.Error(),
				}
			}
		}
	}

	// If we arrived here it means that all the property names validated successfully against
	// the schema stored in "propertyNames".
	return nil
}

type dependencies map[string]interface{}

func (d dependencies) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First we need to verify that jsonData is a json object.
	if object, ok := jsonData.value.(map[string]interface{}); ok {

		// Iterate over the dependencies object from the schema.
		for propertyName, dependency := range d {
			// A dependency may be a json array (consist of strings) of a json
			// object which is a json schema that the inspected value need to
			// validated against.
			switch v := dependency.(type) {

			// In this case the dependency is a sub-schema.
			case *JsonSchema:
				{
					// Check if the propertyName (which is the key in the "dependencies" object)
					// is present in the data. If it is, validate the whole instance against the
					// sub-schema.
					if _, ok := object[propertyName]; ok {
						// Validate the whole data against the given sub-schema.
						err := v.validateJsonData("", jsonData.raw, rootSchemaId)
						if err != nil {
							return KeywordValidationError{
								"dependencies",
								"inspected value failed in validation against sub-schema given in \"" +
									propertyName +
									"\" dependency: " +
									err.Error(),
							}
						}
					}
				}
			// In this case the dependency is a list of required property names.
			case []interface{}:
				{
					// Iterate over the items in the dependency array.
					for index, value := range v {
						// Verify that the value is actually a string.
						// If not, return an error
						if requiredProperty, ok := value.(string); ok {
							// Check if the required property name is missing. If it is,
							// return an error.
							if _, ok := object[requiredProperty]; !ok {
								return KeywordValidationError{
									"dependencies",
									"missing property \"" +
										requiredProperty +
										"\" although it is required according to \"" +
										propertyName +
										"\" dependency",
								}
							}
						} else {
							return KeywordValidationError{
								"dependencies",
								"all items in dependency array must be strings, item at position " +
									strconv.Itoa(index) +
									" is not a string",
							}
						}
					}
				}
			default:
				{
					return KeywordValidationError{
						"dependencies",
						"dependency value must be a json object or a json array",
					}
				}
			}
		}
	}

	// If we arrived here it means that all the validations succeeded.
	return nil
}

type patternProperties map[string]*JsonSchema

func (pp patternProperties) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First we need to verify that jsonData is a json object.
	if object, ok := jsonData.value.(map[string]interface{}); ok {
		// Iterate over the given patterns.
		for pattern, subSchema := range pp {
			// Iterate over the properties in the inspected value.
			for property := range object {
				// Check if the property matches to the pattern.
				match, err := regexp.MatchString(pattern, property)

				// The pattern or the value is not in the right format (string)
				if err != nil {
					return KeywordValidationError{
						"patternProperties",
						err.Error(),
					}
				}

				// If there is a match, validate the value of the property against
				// the given schema.
				if match {
					err := subSchema.validateJsonData(jsonPath+"/"+property, jsonData.raw, rootSchemaId)

					// If the validation fails, return an error.
					if err != nil {
						return KeywordValidationError{
							"patternProperties",
							"property \"" +
								property +
								"\" that matches the pattern \"" +
								pattern +
								"\" failed in validation: \n" + err.Error(),
						}
					}
				}
			}
		}
	}

	// If we arrived here it means that none of the properties failed in
	// validation against any of the given schemas.
	return nil
}

type minProperties int

func (mp *minProperties) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we must verify that jsonData is a json object.
	// If it is not a json object, we return an error.
	if v, ok := jsonData.value.(map[string]interface{}); ok {
		// If the amount of keys in jsonData equals to or greater
		// than minProperties.
		// Else, return an error.
		if len(v) >= int(*mp) {
			return nil
		} else {
			return KeywordValidationError{
				"minProperties",
				"inspected value must contains at least " + strconv.Itoa(int(*mp)) + " properties",
			}
		}
	}

	return nil
}

type maxProperties int

func (mp *maxProperties) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we must verify that jsonData is a json object.
	// If it is not a json object, we return an error.
	if v, ok := jsonData.value.(map[string]interface{}); ok {
		// If the amount of keys in jsonData equals to or less
		// than maxProperties.
		// Else, return an error.
		if len(v) <= int(*mp) {
			return nil
		} else {
			return KeywordValidationError{
				"minProperties",
				"inspected value may contains at most " +
					strconv.Itoa(int(*mp)) +
					" properties",
			}
		}
	}

	return nil
}

/********************/
/** Array Keywords **/
/********************/

type items json.RawMessage

func (i items) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we need to verify that json Data is an array
	if array, ok := jsonData.value.([]interface{}); ok {
		var data interface{}

		// Unmarshal the value in items in order to figure out if it is a
		// json object or json array
		err := json.Unmarshal(i, &data)
		if err != nil {
			return err
		}

		// Handle the value in items according to its json type.
		switch itemsField := data.(type) {
		// If jsonData is a json object, which means that is holds a single schema,
		// we validate the all the items in the inspected array against the given
		// schema.
		case map[string]interface{}:
			{
				// This is the JsonSchema instance that should hold the schema in
				// "items" field.
				var schema JsonSchema

				// Unmarshal the rawSchema into the JsonSchema struct.
				err = json.Unmarshal(i, &schema)
				if err != nil {
					return err
				}

				// Iterate over the items in the inspected array and validate each
				// item against the schema in "items" field.
				for index := 0; index < len(array); index++ {
					err := schema.validateJsonData(jsonPath+"/"+strconv.Itoa(index), jsonData.raw, rootSchemaId)
					if err != nil {
						return err
					}
				}
			}
		// If jsonData is a json array, which means that is holds multiple json schema objects,
		// we validate each item in the inspected array against the schema at the same position.
		case []interface{}:
			{
				if len(itemsField) > len(array) {
					return KeywordValidationError{
						"items",
						"when \"items\" field contains a list of Json Schema objects, the " +
							"inspected array must contain at least the same amount of items",
					}
				}

				// Iterate over the schemas in "items" field.
				for index, schemaFromItems := range itemsField {
					// Marshal the current schema in "items" field in order to Unmarshal it
					// into JsonSchema instance.
					rawSchema, err := json.Marshal(schemaFromItems)
					if err != nil {
						return err
					}

					// This is the JsonSchema instance that should hold the current
					// working schema.
					var schema JsonSchema

					// Unmarshal the rawSchema into the JsonSchema struct.
					err = json.Unmarshal(rawSchema, &schema)
					if err != nil {
						return err
					}

					// Validate the item against the schema at the same position.
					err = schema.validateJsonData(jsonPath+"/"+strconv.Itoa(index), jsonData.raw, rootSchemaId)
					if err != nil {
						return err
					}
				}
			}
		// The default case indicates that the value in items field is not a json schema or
		// a list of json schema.
		default:
			{
				return KeywordValidationError{
					"items",
					"\"items\" field value in schema must be a valid Json Schema or an array of Json Schema",
				}
			}
		}
	}

	// If we arrived here it means that all the items in the inspected array
	// validated successfully against the given schema.
	return nil
}

func (i *items) UnmarshalJSON(data []byte) error {
	*i = data
	return nil
}

type additionalItems struct {
	JsonSchema
	siblingItems *items
}

func (ai *additionalItems) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// Unmarshal the sibling field "items" in order to check it's json type.
	var siblingItems interface{}
	err := json.Unmarshal(*ai.siblingItems, &siblingItems)
	if err != nil {
		return err
	}

	// If "items" is a json array, "additionalItems" needs to verify the items
	// that the schema in "items" field did not validate.
	if itemsArray, ok := siblingItems.([]interface{}); ok {
		// Check if jsonData is a json array.
		if array, ok := jsonData.value.([]interface{}); ok {
			// Iterate over the inspected array from the position that items stopped
			// validating.
			for index := range array[len(itemsArray):] {
				// Validate the inspected item against the schema given in "additionalItems".
				err := ai.validateJsonData(jsonPath+"/"+strconv.Itoa(index), jsonData.raw, rootSchemaId)
				if err != nil {
					return KeywordValidationError{
						"additionalItems",
						"item at position " +
							strconv.Itoa(index) +
							" failed in validation: " +
							err.Error(),
					}
				}
			}

			// If we arrived here it means that no item failed in validation.
			return nil
		}
	}

	// If "items" field is not an array of json schema, additionalItems
	// is meaningless so we return true.
	return nil
}

type contains struct {
	JsonSchema
}

func (c *contains) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we need to verify that jsonData is a json array.
	if array, ok := jsonData.value.([]interface{}); ok {
		// Go over all the items in the array in order to inspect them.
		for index := range array {
			// If the item is valid against the given schema, which means that
			// the array contains the required value.
			err := (*c).validateJsonData(jsonPath+"/"+strconv.Itoa(index), jsonData.raw, rootSchemaId)
			if err == nil {
				return nil
			}
		}
	}

	// If we arrived here it means that we could not validate any of the array's
	// items against the given schema.
	return KeywordValidationError{
		"contains",
		"could validate any of the inspected array's items against the given schema",
	}
}

type minItems int

func (mi *minItems) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we need to verify that jsonData is an array.
	if v, ok := jsonData.value.([]interface{}); ok {
		// Check that the number of items in the array is equal to
		// or greater than minItems.
		if len(v) >= int(*mi) {
			return nil
		} else {
			return KeywordValidationError{
				"minItems",
				"inspected array must contain at least " + strconv.Itoa(int(*mi)) + " items",
			}
		}
	}

	return nil
}

type maxItems int

func (mi *maxItems) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we need to verify that jsonData is an array.
	if v, ok := jsonData.value.([]interface{}); ok {
		// Check that the number of items in the array is equal to
		// or less than maxItems.
		if len(v) <= int(*mi) {
			return nil
		} else {
			return KeywordValidationError{
				"maxItems",
				"inspected array must contain at most " + strconv.Itoa(int(*mi)) + " items",
			}
		}
	}

	return nil
}

type uniqueItems bool

func (ui *uniqueItems) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// First, we need to verify that jsonData is an array.
	if array, ok := jsonData.value.([]interface{}); ok {
		// Create a map that will help us to check if we already met the
		// item by using the map's hashing mechanism.
		uniqueSet := make(map[string]int)

		// Iterate over the items in the inspected array.
		for index, item := range array {
			// Marshal the item back to hash-able value, because maps (json object)
			// and slices (json arrays) are not a hash-able values.
			rawItem, err := json.Marshal(item)
			if err != nil {
				return err
			}

			// If ok is true it means that the value exists in the map, which means
			// we already met it in one of the previous iterations.
			// Else, insert the item into the map as key, and the index as value.
			if v, ok := uniqueSet[string(rawItem)]; ok {
				return KeywordValidationError{
					"uniqueItems",
					"the inspected array contains two equal items at indices: " +
						strconv.Itoa(v) +
						", " +
						strconv.Itoa(index),
				}
			} else {
				uniqueSet[string(rawItem)] = index
			}
		}

		// If we arrived here it means that we did not meat any item which is
		// similar to another item in the array.
		return nil
	}

	return nil
}

/********************/
/** Other Keywords **/
/********************/

type contentMediaType string
type contentEncoding string

/**************************/
/** Conditional Keywords **/
/**************************/

type anyOf []*JsonSchema

func (af anyOf) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// Validate jsonData.raw against each of the schemas until on of them succeeds.
	for _, schema := range af {
		err := schema.validateJsonData("", jsonData.raw, rootSchemaId)
		if err == nil {
			return nil
		}
	}

	// If we arrived here, the validation of jsonData failed against all schemas.
	return KeywordValidationError{
		"anyOf",
		"inspected value could not be validated against any of the given schemas",
	}
}

type allOf []*JsonSchema

func (af allOf) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// Validate jsonData.raw against each of the schemas.
	// If one of them fails, return error.
	for _, schema := range af {
		err := schema.validateJsonData("", jsonData.raw, rootSchemaId)
		if err != nil {
			return KeywordValidationError{
				"allOf",
				"inspected value could not be validated against all of the given schemas",
			}
		}
	}

	// If we arrived here, the validation of jsonData succeeded against all
	// given schemas.
	return nil
}

type oneOf []*JsonSchema

func (of oneOf) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	var oneValidationAlreadySucceeded bool

	// Validate jsonData.raw against each of the schemas until on of them succeeds.
	for _, schema := range of {
		err := schema.validateJsonData("", jsonData.raw, rootSchemaId)
		if err == nil {
			if oneValidationAlreadySucceeded {
				return KeywordValidationError{
					"oneOf",
					"inspected data is valid against more than one given schema",
				}
			} else {
				oneValidationAlreadySucceeded = true
			}
		}
	}

	if oneValidationAlreadySucceeded {
		return nil
	} else {
		// If we arrived here, the validation of jsonData failed against all schemas.
		return KeywordValidationError{
			"oneOf",
			"inspected value could not be validated against any of the given schemas",
		}
	}
}

type not struct {
	JsonSchema
}

func (n *not) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	err := (*n).validateJsonData(jsonPath, jsonData.raw, rootSchemaId)
	if err != nil {
		return nil
	} else {
		return KeywordValidationError{
			"not",
			"inspected value did not fail on validation against the schema defined by this keyword",
		}
	}
}

type _if struct {
	JsonSchema
	siblingThen *_then
	siblingElse *_else
}

func (i *_if) validate(jsonPath string, jsonData jsonData, rootSchemaId string) error {
	// Validate the data against the given schema in "if".
	err := (*i).validateJsonData("", jsonData.raw, rootSchemaId)

	// If the validation succeeded, validate the data against the given schema
	// in "then".
	// Else, validate the data against the given schema in "else".
	if err == nil {
		if (*i).siblingThen != nil {
			return (*i).siblingThen.validateJsonData(jsonPath, jsonData.raw, rootSchemaId)
		}
	} else {
		if (*i).siblingElse != nil {
			return (*i).siblingElse.validateJsonData(jsonPath, jsonData.raw, rootSchemaId)
		}
	}

	return nil
}

type _then struct {
	JsonSchema
}

type _else struct {
	JsonSchema
}

/****************************/
/** Authorization Keywords **/
/****************************/

type readOnly bool
type writeOnly bool
