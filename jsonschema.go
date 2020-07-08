package jsonvalidator

import (
	"encoding/json"
	"fmt"

	"strconv"
	"strings"

	"github.com/itayankri/gojsonvalidator/jsonpointer"
	"github.com/pkg/errors"
)

// Valid Json Schema types
const (
	TYPE_OBJECT  = "object"
	TYPE_ARRAY   = "array"
	TYPE_STRING  = "string"
	TYPE_NUMBER  = "number"
	TYPE_INTEGER = "integer"
	TYPE_BOOLEAN = "boolean"
	TYPE_NULL    = "null"
)

// Valid values for "contentEncoding" field
const (
	ENCODING_7BIT             = "7bit"
	ENCODING_8bit             = "8bit"
	ENCODING_BINARY           = "binary"
	ENCODING_QUOTED_PRINTABLE = "quoted-printable"
	ENCODING_BASE64           = "base64"
)

type jsonData struct {
	raw   json.RawMessage
	value interface{}
}

type JsonSchema struct {
	// RejectAll is ***not*** a json schema keyword!
	// It is an internal flag for internal use that represents a json schema
	// that suppose to reject all json values.
	// If it is true, all other field will be ignored and validateJsonData()
	// will always return false.
	RejectAll bool `json:"rejectAll,omitempty"`

	// The $schema keyword is used to declare that a JSON fragment is
	// actually a piece of JSON Schema.
	Schema *schema `json:"$schema,omitempty"`

	// The value of $ref is a URI, and the part after # sign is in a format
	// called JSON Pointer.
	Ref *ref `json:"$ref,omitempty"`

	// The $id property is a URI that serves two purposes:
	// It declares a unique identifier for the schema
	// It declares a base URI against which $ref URIs are resolved.
	Id *id `json:"$id,omitempty"`

	// The $comment keyword is strictly intended for adding comments
	// to the JSON schema source. Its value must always be a string.
	Comment *comment `json:"$comment,omitempty"`

	// Title and Description used to describe the schema and not used for
	// validation.
	Title       *title       `json:"title,omitempty"`
	Description *description `json:"description,omitempty"`

	// The value of this keyword MUST be either a string or an array. If it is
	// an array, elements of the array MUST be strings and MUST be unique.
	// String values MUST be one of the six primitive types
	// ("null", "boolean", "object", "array", "number", or "string"),
	// or "integer" which matches any number with a zero fractional part.
	Type *_type `json:"type,omitempty"`

	// The default keyword specifies a default value for an item.
	Default _default `json:"default,omitempty"`

	// The examples keyword is a place to provide an array of examples
	// that validate against the schema.
	Examples examples `json:"examples,omitempty"`

	// The value of this keyword MUST be an array.
	// An instance validates successfully against this keyword if its value is
	// equal to one of the elements in this keyword's array value.
	Enum enum `json:"enum,omitempty"`

	// The value of this keyword MAY be of any type, including null.
	// An instance validates successfully against this keyword if its value is
	// equal to the value of the keyword.
	Const *_const `json:"const,omitempty"`

	// The "definitions" keywords provides a standardized location for schema
	// authors to inline re-usable JSON Schemas into a more general schema. The
	// keyword does not directly affect the validation result.
	// This keyword's value MUST be an object. Each member value of this
	// object MUST be a valid JSON Schema.
	Definitions definitions `json:"definitions,omitempty"`

	// The value of "properties" MUST be an object. Each value of this object
	// MUST be a valid JSON Schema.
	// This keyword determines how child instances validate for objects, and
	// does not directly validate the immediate instance itself.
	// Validation succeeds if, for each name that appears in both the instance
	// and as a name within this keyword's value, the child instance for that
	// name successfully validates against the corresponding schema.
	Properties properties `json:"properties,omitempty"`

	// The value of "additionalProperties" MUST be a valid JSON Schema.
	// This keyword determines how child instances validate for objects,
	// and does not directly validate the immediate instance itself.
	// Validation with "additionalProperties" applies only to the child values
	// of instance names that do not match any names in "properties", and do
	// not match any regular expression in "patternProperties".
	// For all such properties, validation succeeds if the child instance
	// validates against the "additionalProperties" schema.
	AdditionalProperties *additionalProperties `json:"additionalProperties,omitempty"`

	// The value of this keyword MUST be an array. Elements of this array,
	// if any, MUST be strings, and MUST be unique.
	// An object instance is valid against this keyword if every item in the
	// array is the name of a property in the instance.
	Required required `json:"required,omitempty"`

	// The value of "propertyNames" MUST be a valid JSON Schema.
	// If the instance is an object, this keyword validates if every property
	// name in the instance validates against the provided schema. Note the
	// property name that the schema is testing will always be a string
	PropertyNames *propertyNames `json:"propertyNames,omitempty"`

	// This keyword specifies rules that are evaluated if the instance is an
	// object and contains a certain property.
	// This keyword's value MUST be an object. Each property specifies a
	// dependency. Each dependency value MUST be an array or a valid JSON
	// Schema.
	// If the dependency value is a subschema, and the dependency key is a
	// property in the instance, the entire instance must validate against the
	// dependency value.
	// If the dependency value is an array, each element in the array, if any,
	// MUST be a string, and MUST be unique. If the dependency key is a
	// property in the instance, each of the items in the dependency value
	// must be a property that exists in the instance.
	Dependencies dependencies `json:"dependencies,omitempty"`

	// The value of "patternProperties" MUST be an object. Each property name
	// of this object SHOULD be a valid regular expression, according to the
	// ECMA 262 regular expression dialect. Each property value of this object
	// MUST be a valid JSON Schema.
	// This keyword determines how child instances validate for objects, and
	// does not directly validate the immediate instance itself. Validation of
	// the primitive instance type against this keyword always succeeds.
	// Validation succeeds if, for each instance name that matches any regular
	// expressions that appear as a property name in this keyword's value, the
	// child instance for that name successfully validates against each schema
	// that corresponds to a matching regular expression.
	PatternProperties patternProperties `json:"patternProperties,omitempty"`

	// The value of "items" MUST be either a valid JSON Schema or an array of
	// valid JSON Schemas.
	// This keyword determines how child instances validate for arrays, and
	// does not directly validate the immediate instance itself.
	// If "items" is a schema, validation succeeds if all elements in the array
	// successfully validate against that schema.
	// If "items" is an array of schemas, validation succeeds if each element
	// of the instance validates against the schema at the same position,
	// if any.
	Items items `json:"items,omitempty"`

	// The value of this keyword MUST be a valid JSON Schema.
	// An array instance is valid against "contains" if at least one of its
	// elements is valid against the given schema. Note that when collecting
	// annotations, the subschema MUST be applied to every array element even
	// after the first match has been found. This is to ensure that all
	// possible annotations are collected.
	Contains *contains `json:"contains,omitempty"`

	// The value of "additionalItems" MUST be a valid JSON Schema.
	// This keyword determines how child instances validate for arrays, and
	// does not directly validate the immediate instance itself.
	// If "items" is an array of schemas, validation succeeds if every
	// instance element at a position greater than the size of "items"
	// validates against "additionalItems".
	// Otherwise, "additionalItems" MUST be ignored, as the "items" schema
	// (possibly the default value of an empty schema) is applied to all
	// elements.
	AdditionalItems *additionalItems `json:"additionalItems,omitempty"`

	// array limitations
	MinItems    *minItems    `json:"minItems,omitempty"`
	MaxItems    *maxItems    `json:"maxItems,omitempty"`
	UniqueItems *uniqueItems `json:"uniqueItems,omitempty"`

	// string limitations
	MinLength *minLength `json:"minLength,omitempty"`
	MaxLength *maxLength `json:"maxLength,omitempty"`
	Pattern   *pattern   `json:"pattern,omitempty"`
	Format    *format    `json:"format,omitempty"`

	// integer/number limitations
	MultipleOf       *multipleOf       `json:"multipleOf,omitempty"`
	Minimum          *minimum          `json:"minimum,omitempty"`
	Maximum          *maximum          `json:"maximum,omitempty"`
	ExclusiveMinimum *exclusiveMinimum `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *exclusiveMaximum `json:"exclusiveMaximum,omitempty"`

	// object size limitations
	MinProperties *minProperties `json:"minProperties,omitempty"`
	MaxProperties *maxProperties `json:"maxProperties,omitempty"`

	// The contentMediaType keyword specifies the MIME type of the contents
	// of a string.
	ContentMediaType *contentMediaType `json:"contentMediaType,omitempty"`

	// The contentEncoding keyword specifies the encoding used to store
	// the contents.
	ContentEncoding *contentEncoding `json:"contentEncoding,omitempty"`

	// Must be valid against any of the sub-schemas.
	AnyOf anyOf `json:"anyOf,omitempty"`

	// Must be valid against all of the sub-schemas.
	AllOf allOf `json:"allOf,omitempty"`

	// Must be valid against exactly one of the sub-schemas.
	OneOf oneOf `json:"oneOf,omitempty"`

	// Must not be valid against the given schema.
	Not *not `json:"not,omitempty"`

	// The if, then and else keywords allow the application of a sub-schema
	// based on the outcome of another schema.
	If   *_if   `json:"if,omitempty"`
	Then *_then `json:"then,omitempty"`
	Else *_else `json:"else,omitempty"`

	// If "readOnly" has a value of boolean true, it indicates that the value
	// of the instance is managed exclusively by the owning authority, and
	// attempts by an application to modify the value of this property are
	// expected to be ignored or rejected by that owning authority.
	ReadOnly *readOnly `json:"readOnly,omitempty"`

	// If "writeOnly" has a value of boolean true, it indicates that the value
	// is never present when the instance is retrieved from the owning
	// authority.
	// It can be present when sent to the owning authority to update or create
	// the document (or the resource it represents), but it will not be
	// included in any updated or newly created version of the instance.
	WriteOnly *writeOnly `json:"writeOnly,omitempty"`
}

// tempJsonSchema is an internal type that created because of the need of
// JsonSchema type to implement the Unmarshaler interface.
// An edge case of json schema, is the json schema "true" or "false".
// In those cases, json.Unmarshal will try to unmarshal a boolean into a
// JsonSchema instance which is a struct, and it will fail.8
// So I implemented the Unmarshaler and tried to unmarshal according to type
// of the raw data that represents the schema. But then I realized that I
// cannot use unmarshal data into a JsonSchema instance inside JsonSchema's
// UnmarshalJSON because it created an endless indirect recursive call
// between JsonSchema.UnmarshalJSON and json.Unmarshal, so I created a new type
// that has all of JsonSchema's field but does not inherit JsonSchema's methods
// (Particularly UnmarshalJSON) in order to be able to unmarshal a json schema
// without starting an endless loop of function calls.
type tempJsonSchema JsonSchema

// NewJsonSchema created a new JsonSchema instance, Unmarshals the byte array
// into the instance, and return the instance.
func NewJsonSchema(bytes []byte) (*JsonSchema, error) {
	var schema *JsonSchema

	// Check if the string s is a valid json.
	err := json.Unmarshal(bytes, &schema)
	if err != nil {
		return nil, err
	}

	err = schema.scanSchema("", "")
	if err != nil {
		fmt.Println("[JsonSchema DEBUG] connectRelatedKeywords() " +
			"failed: " + err.Error())
		return nil, err
	}

	return schema, nil
}

// scanSchema is a recursive function that connect the related
// keywords of the schema (as mentioned in the description of NewJsonSchema()).
// The function scans the schema in and it's sub-schemas and perform the
// required connections.
func (js *JsonSchema) scanSchema(schemaPath string, rootSchemaID string) error {
	js.connectRelatedKeywords()
	js.mapSubSchema(schemaPath, rootSchemaID)

	// Connect sub-schemas in "properties" field.
	for key := range js.Properties {
		err := js.Properties[key].scanSchema(schemaPath+"/properties/"+key, rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schema in "additionalProperties" field.
	if js.AdditionalProperties != nil {
		err := js.AdditionalProperties.scanSchema(schemaPath+"/additionalProperties", rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schema in "propertyNames" field.
	if js.PropertyNames != nil {
		err := js.PropertyNames.scanSchema(schemaPath+"/propertyNames", rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schemas in "dependencies" field.
	for key, value := range js.Dependencies {
		// Check if the dependency is a json schema or an array of properties.
		if v, ok := value.(map[string]interface{}); ok {
			subSchema := new(JsonSchema)
			// Marshal the dependency in order to Unmarshal it into JsonSchema struct.
			rawDependency, err := json.Marshal(v)
			if err != nil {
				return SchemaCompilationError{
					schemaPath + "/dependencies" + key,
					err.Error(),
				}
			}

			// Create a new JsonSchema instance.
			err = json.Unmarshal(rawDependency, subSchema)
			if err != nil {
				return SchemaCompilationError{
					schemaPath,
					err.Error(),
				}
			}

			err = subSchema.scanSchema(schemaPath+"/dependencies"+key, rootSchemaID)
			if err != nil {
				return err
			}

			// Save the new JsonSchema as the dependency itself.
			js.Dependencies[key] = subSchema
		}
	}

	// Connect sub-schemas in "patternProperties" field.
	for key := range js.PatternProperties {
		err := js.PatternProperties[key].scanSchema(schemaPath+"/patternProperties/"+key, rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schemas in "definitions" field.
	for key := range js.Definitions {
		err := js.Definitions[key].scanSchema(schemaPath+"/definitions/"+key, rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schemas in "items" field.
	if js.Items != nil {
		var items interface{}

		// Unmarshal the item to an empty interface variable in order
		// to check if the "items" is a single schema of a list of schemas.
		err := json.Unmarshal(js.Items, &items)
		if err != nil {
			return SchemaCompilationError{
				schemaPath,
				err.Error(),
			}
		}

		// Check the type of "items"
		switch v := items.(type) {
		// In this case, "items" is an object which means its a single schema.
		case map[string]interface{}, bool:
			{
				// Marshal the dependency in order to Unmarshal it into JsonSchema struct.
				rawSubSchema, err := json.Marshal(v)
				if err != nil {
					return SchemaCompilationError{
						schemaPath + "/items",
						err.Error(),
					}
				}

				subSchema := new(JsonSchema)

				// Create a new JsonSchema instance.
				err = json.Unmarshal(rawSubSchema, subSchema)
				if err != nil {
					return SchemaCompilationError{
						schemaPath + "/items",
						err.Error(),
					}
				}

				err = subSchema.scanSchema(schemaPath+"/items", rootSchemaID)
				if err != nil {
					return err
				}

				js.Items, err = json.Marshal(subSchema)
				if err != nil {
					return SchemaCompilationError{
						schemaPath + "/items",
						err.Error(),
					}
				}
			}
		// In this case "items" hold an array of schemas.
		case []interface{}:
			{
				// Iterate over each schema in "items".
				for index, value := range v {
					// Marshal the dependency in order to Unmarshal it into JsonSchema struct.
					rawSubSchema, err := json.Marshal(value)
					if err != nil {
						return SchemaCompilationError{
							schemaPath + "/items" + strconv.Itoa(index),
							err.Error(),
						}
					}

					subSchema := new(JsonSchema)

					// Create a new JsonSchema instance.
					err = json.Unmarshal(rawSubSchema, subSchema)
					if err != nil {
						return SchemaCompilationError{
							path: schemaPath + "/items" + strconv.Itoa(index),
							err:  "",
						}
					}

					err = subSchema.scanSchema(schemaPath+"/items"+strconv.Itoa(index), rootSchemaID)
					if err != nil {
						return nil
					}

					// Save the sub-schema in "items" array.
					v[index] = subSchema
				}

				// Marshal "items" back to a json.RawMessage and store it in the parent schema.
				js.Items, err = json.Marshal(v)
				if err != nil {
					return SchemaCompilationError{
						schemaPath + "/items",
						err.Error(),
					}
				}
			}
		}
	}

	// Connect sub-schema in "additionalItems" field.
	if js.AdditionalItems != nil {
		err := js.AdditionalItems.scanSchema(schemaPath+"/additionalItems", rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schema in "contains" field.
	if js.Contains != nil {
		err := js.Contains.scanSchema(schemaPath+"/contains", rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schemas in "anyOf" field.
	for index := range js.AnyOf {
		err := js.AnyOf[index].scanSchema(schemaPath+"/anyOf/"+strconv.Itoa(index), rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schemas in "allOf" field.
	for index := range js.AllOf {
		err := js.AllOf[index].scanSchema(schemaPath+"/allOf/"+strconv.Itoa(index), rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schemas in "oneOf" field.
	for index := range js.OneOf {
		err := js.OneOf[index].scanSchema(schemaPath+"/oneOf/"+strconv.Itoa(index), rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schema in "not" field.
	if js.Not != nil {
		err := js.Not.scanSchema(schemaPath+"/not", rootSchemaID)
		if err != nil {
			return err
		}
	}

	// Connect sub-schema in "if" field.
	if js.If != nil {
		err := js.If.scanSchema(schemaPath+"/if", rootSchemaID)
		if err != nil {
			return err
		}

		// Connect sub-schema in "then" field.
		if js.Then != nil {
			err := js.Then.scanSchema(schemaPath+"/then", rootSchemaID)
			if err != nil {
				return err
			}
		}

		// Connect sub-schema in "else" field.
		if js.Else != nil {
			err := js.Else.scanSchema(schemaPath+"/else", rootSchemaID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// connectRelatedKeywords is a receiver function that initialized references
// between keywordValidators that depend on each other:
// Schema.AdditionalProperties 	---> 	Schema.Properties
// Schema.AdditionalProperties 	---> 	Schema.PatternProperties
// JsonSchema.AdditionalItems 	---> 	JsonSchema.Items
// JsonSchema.If 				---> 	JsonSchema.Then
// JsonSchema.IF 				---> 	JsonSchema.Else
func (js *JsonSchema) connectRelatedKeywords() {
	// Connect sub-schema in "additionalProperties" field.
	if js.AdditionalProperties != nil {
		// If "properties" field exists in the schema, save the keywordValidator's
		// address in "AdditionalProperties".
		if js.Properties != nil {
			js.AdditionalProperties.siblingProperties = &js.Properties
		}

		// If "patternProperties" field exists in the schema, save the keywordValidator's
		// address in "AdditionalProperties".
		if js.PatternProperties != nil {
			js.AdditionalProperties.siblingPatternProperties = &js.PatternProperties
		}
	}

	// Connect sub-schema in "additionalItems" field.
	if js.AdditionalItems != nil {
		// If "items" field exists in the schema, save the keywordValidator's
		// address in "AdditionalItems".
		if js.Items != nil {
			js.AdditionalItems.siblingItems = &js.Items
		}
	}

	// Connect sub-schema in "if" field.
	if js.If != nil {
		// Connect sub-schema in "then" field.
		if js.Then != nil {

			// If "then" field exists in the schema, save the keywordValidator's
			// address in "If".
			js.If.siblingThen = js.Then
		}

		// Connect sub-schema in "else" field.
		if js.Else != nil {

			// If "else" field exists in the schema, save the keywordValidator's
			// address in "If".
			js.If.siblingElse = js.Else
		}
	}
}

func (js *JsonSchema) mapSubSchema(schemaPath, rootSchemaID string) {
	// If the schema path is not an empty string (means we are not in the root schema),
	// and the rootSchemaID is not an empty string (means the root schema contains
	// the "$id" field), map the current sub schema into the subSchemaMap of the rootSchema.
	if schemaPath != "" && rootSchemaID != "" {
		// If the rootSchema exists in the pool, add the sub schema to it.
		// Else, TODO: decide what to do.
		if rs, ok := rootSchemaPool[rootSchemaID]; ok && rs != nil {
			// If the root schema does not contain the sub schema already, add it to the
			// subSchemaMap.
			// Else, TODO: decide what to do.
			if _, ok := rs.subSchemaMap[schemaPath]; !ok {
				rs.subSchemaMap[schemaPath] = js
			}
		}
	}
}

// validateJsonData is a function that gets a byte array of data and validates
// it against the schema that encoded in the receiver's field.
func (js *JsonSchema) validateJsonData(jsonPath string, bytes []byte, rootSchemaId string) error {
	// If RejectAll field exists and true, reject the value.
	if js.RejectAll {
		return SchemaValidationError{
			jsonPath,
			"json schema \"false\" drops everything",
		}
	}

	// If the schema contains the $ref field, validate the data against the
	// referenced schema (and by the way ignore all the keywords of the current
	// schema).
	if js.Ref != nil {
		return js.Ref.validateByRef(jsonPath, bytes, rootSchemaId)
	}

	// Calculate the relative path in order to evaluate the data
	jsonTokens := strings.Split(jsonPath, "/")
	relativeJsonPath := "/" + jsonTokens[len(jsonTokens)-1]

	// Create a new JsonPointer.
	jsonPointer, err := jsonwalker.NewJsonPointer(relativeJsonPath)
	if err != nil {
		fmt.Println("[JsonSchema DEBUG] validateJsonData() " +
			"failed while trying to create JsonPointer " + jsonPath)
		return errors.Wrap(err, "JsonPointer creation failed")
	}

	// Get the piece of json that the current schema describes.
	value, err := jsonPointer.Evaluate(bytes)
	if err != nil {
		fmt.Println("[JsonSchema DEBUG] validateJsonData() " +
			"failed while trying to evaluate a JsonPointer " + jsonPath)
		return errors.Wrap(err, "JsonPointer evaluation failed")
	}

	// Marshal the evaluated value to a byte array.
	newBytes, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "data marshaling after JsonPointer evaluation failed")
	}

	// Create a new json data container
	jsonData := jsonData{
		newBytes,
		value,
	}

	// Get a slice of all of JsonSchema's field in order to iterate them
	// and call each of their validate() functions.
	keywordValidators := getNonNilKeywordsSlice(js)

	// Iterate over the keywords.
	for _, keyword := range keywordValidators {
		// Validate the value that we extracted from the jsonData at each
		// keyword.
		err := keyword.validate(jsonPath, jsonData, rootSchemaId)
		if err != nil {
			// If the error is a SchemaValidationError, it means it came from
			// a deeper call to this function, so we do not touch the error.
			if schemaValidationError, ok := err.(SchemaValidationError); ok {
				return schemaValidationError
			}

			// If the error is a KeywordValidationError, create a new
			// SchemaValidationError and return it.
			if keywordValidationError, ok := err.(KeywordValidationError); ok {
				return SchemaValidationError{
					jsonPath,
					keywordValidationError.Error(),
				}
			}

			return err
		}
	}

	return nil
}

// getNonNilKeywordsMap gets a reference to JsonSchema and returns a
// map of the schema's keywords that are not nil.
func getNonNilKeywordsSlice(js *JsonSchema) []keywordValidator {
	var slice []keywordValidator

	if js.Type != nil {
		slice = append(slice, js.Type)
	}

	if js.Const != nil {
		slice = append(slice, js.Const)
	}

	if js.Enum != nil {
		slice = append(slice, js.Enum)
	}

	if js.MinLength != nil {
		slice = append(slice, js.MinLength)
	}

	if js.MaxLength != nil {
		slice = append(slice, js.MaxLength)
	}

	if js.Pattern != nil {
		slice = append(slice, js.Pattern)
	}

	if js.Format != nil {
		slice = append(slice, js.Format)
	}

	if js.MultipleOf != nil {
		slice = append(slice, js.MultipleOf)
	}

	if js.Minimum != nil {
		slice = append(slice, js.Minimum)
	}

	if js.Maximum != nil {
		slice = append(slice, js.Maximum)
	}

	if js.ExclusiveMinimum != nil {
		slice = append(slice, js.ExclusiveMinimum)
	}

	if js.ExclusiveMaximum != nil {
		slice = append(slice, js.ExclusiveMaximum)
	}

	if js.Required != nil {
		slice = append(slice, js.Required)
	}

	if js.PropertyNames != nil {
		slice = append(slice, js.PropertyNames)
	}

	if js.Properties != nil {
		slice = append(slice, js.Properties)
	}

	if js.AdditionalProperties != nil {
		slice = append(slice, js.AdditionalProperties)
	}

	if js.PatternProperties != nil {
		slice = append(slice, js.PatternProperties)
	}

	if js.Dependencies != nil {
		slice = append(slice, js.Dependencies)
	}

	if js.MinProperties != nil {
		slice = append(slice, js.MinProperties)
	}

	if js.MaxProperties != nil {
		slice = append(slice, js.MaxProperties)
	}

	if js.Items != nil {
		slice = append(slice, js.Items)
	}

	if js.Contains != nil {
		slice = append(slice, js.Contains)
	}

	if js.AdditionalItems != nil {
		slice = append(slice, js.AdditionalItems)
	}

	if js.MinItems != nil {
		slice = append(slice, js.MinItems)
	}

	if js.MaxItems != nil {
		slice = append(slice, js.MaxItems)
	}

	if js.UniqueItems != nil {
		slice = append(slice, js.UniqueItems)
	}

	if js.AnyOf != nil {
		slice = append(slice, js.AnyOf)
	}

	if js.AllOf != nil {
		slice = append(slice, js.AllOf)
	}

	if js.OneOf != nil {
		slice = append(slice, js.OneOf)
	}

	if js.Not != nil {
		slice = append(slice, js.Not)
	}

	if js.If != nil {
		slice = append(slice, js.If)
	}

	// Return the map.
	return slice
}

func (js *JsonSchema) UnmarshalJSON(bytes []byte) error {
	// First, unmarshal the raw data into empty interface variable
	// in order to figure out its type.
	var value interface{}
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}

	// Create a new tempJsonSchema i the heap.
	tempSchema := new(tempJsonSchema)

	// Handle the data according to its type:
	// map[string]interface: An object json schema.
	// bool: A boolean json schema.
	switch schema := value.(type) {
	case map[string]interface{}:
		{
			// Marshal the raw data into an the temporary type that represents
			// a json schema.
			err = json.Unmarshal(bytes, tempSchema)
			if err != nil {
				return err
			}

			// Convert the temporary type to JsonSchema and assign its address
			// to the receiver.
			*js = JsonSchema(*tempSchema)
		}
	case bool:
		{
			// If the boolean schema is true, unmarshal an empty object into
			// the temporary schema (A valid json schema that accepts any
			// json value).
			// Else, unmarshal a json object with "rejectAll" flag (An internal
			// sign that represents a schema that rejects everything).
			if schema {
				err = json.Unmarshal([]byte("{}"), tempSchema)
				if err != nil {
					return err
				}
			} else {
				err = json.Unmarshal([]byte("{\"rejectAll\": true}"), tempSchema)
				if err != nil {
					return err
				}
			}

			// Convert the temporary type to JsonSchema and assign its address
			// to the receiver.
			*js = JsonSchema(*tempSchema)
		}
	default:
		{
			// If the data is not a json object or a json boolean, it is not a
			// valid schema.
			return errors.New("a valid json schema must be a json object" +
				" or a boolean")
		}
	}

	return nil
}
