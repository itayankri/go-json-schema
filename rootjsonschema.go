package jsonvalidator

import (
	"encoding/json"
	"fmt"
)

// This is a package-level dictionary that contains all the reference-able
// root schema instances.
var rootSchemaPool = map[string]*RootJsonSchema{}

// RootJsonSchema is struct that contains a JsonSchema embedded into it
// (and therefore inherits all JsonSchema's methods) and a map of json path and
// a pointer to JsonSchema instance called subSchemaMap.
// subSchemaMap holds a record for each sub-schema that the root-schema contains.
type RootJsonSchema struct {
	JsonSchema
	subSchemaMap map[string]*JsonSchema
}

// NewJsonSchema creates a new RootJsonSchema instance, Unmarshals the byte array
// into the instance, and returns a pointer to the instance.
func NewRootJsonSchema(bytes []byte) (*RootJsonSchema, error) {
	var rootSchemaId string
	var rootSchema *RootJsonSchema

	// Check if the string s is a valid json.
	err := json.Unmarshal(bytes, &rootSchema)
	if err != nil {
		return nil, err
	}

	// Allocate space for the map in memory.
	rootSchema.subSchemaMap = make(map[string]*JsonSchema)

	// If the field $id in the rootSchema exists, add the rootSchema to the
	// rootSchemaPool
	if rootSchema.Id != nil {
		rootSchemaId = string(*rootSchema.Id)
	}
	//else {
	//	fmt.Println("[RootJsonSchema DEBUG] created a RootJsonSchema instance with no $id")
	//}

	if _, ok := rootSchemaPool[rootSchemaId]; !ok {
		rootSchemaPool[rootSchemaId] = rootSchema
	}

	err = rootSchema.scanSchema("", rootSchemaId)
	if err != nil {
		fmt.Println("[RootJsonSchema DEBUG] scanSchema() " +
			"failed: " + err.Error())
		return nil, err
	}

	return rootSchema, nil
}

// validate calls RootJsonSchema.validateJsonData() with an empty jsonPath
// (represents root), and the root-schema id if exists.
func (rs *RootJsonSchema) validateBytes(bytes []byte) error {
	var id string
	if rs.Id != nil {
		id = string(*rs.Id)
	} else {
		id = ""
	}

	return rs.validateJsonData("", bytes, id)
}
