package models

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// constants for describing possible field types
const (
	String            Kind = 1
	Integer           Kind = 2
	Float             Kind = 3
	Instant           Kind = 4
	Duration          Kind = 5
	Url               Kind = 6
	WorkitemReference Kind = 7
	User              Kind = 8
	Enum              Kind = 9
	List              Kind = 10
)

// Kind is the kind of field type
type Kind byte

/*
FieldType describes the possible values of a FieldDefinition
*/
type FieldType interface {
	GetKind() Kind
	/*
	   ConvertToModel converts a field value for use in the REST API
	*/
	ConvertToModel(value interface{}) (interface{}, error)
	/*
		ConvertToModel converts a field value for storage in the db
	*/
	ConvertFromModel(value interface{}) (interface{}, error)
}

/*
FieldDefintion describes type & other restrictions of a field
*/
type FieldDefinition struct {
	Required bool
	Type     FieldType
}

/*
 Convert a field value for storage as json. As the system matures, add more checks (for example whether a user is in the system, etc.)
*/
func (field FieldDefinition) ConvertToModel(name string, value interface{}) (interface{}, error) {
	if field.Required && value == nil {
		return nil, fmt.Errorf("Value %s is required", name)
	}
	return field.Type.ConvertToModel(value)
}

/*
 Convert from json storage to API form.
*/
func (field FieldDefinition) ConvertFromModel(name string, value interface{}) (interface{}, error) {
	if field.Required && value == nil {
		return nil, fmt.Errorf("Value %s is required", name)
	}
	return field.Type.ConvertFromModel(value)
}

type rawFieldDef struct {
	Type     rawFieldType
	Required bool
}

type rawEnumType struct {
	BaseType SimpleType
	Values   []interface{}
}

type rawFieldType struct {
	Kind  Kind
	Extra *json.RawMessage
}
// UnmarshalJSON implements encoding/json.Unmarshaler
func (self *FieldDefinition) UnmarshalJSON(bytes []byte) error {

	temp := rawFieldDef{}

	err := json.Unmarshal(bytes, &temp)
	if err != nil {
		return err
	}

	switch temp.Type.Kind {
	case List:
		var baseType SimpleType
		err = json.Unmarshal(*temp.Type.Extra, &baseType)
		if err != nil {
			return err
		}
		theType := ListType{SimpleType: SimpleType{Kind: temp.Type.Kind}, ComponentType: baseType}
		*self = FieldDefinition{Type: theType, Required: temp.Required}
	case Enum:
		var extraInfo rawEnumType
		err = json.Unmarshal(*temp.Type.Extra, &extraInfo)
		if err != nil {
			return err
		}
		theType := EnumType{SimpleType: SimpleType{Kind: temp.Type.Kind}, BaseType: extraInfo.BaseType, Values: extraInfo.Values}
		*self = FieldDefinition{Type: theType, Required: temp.Required}
	default:
		*self = FieldDefinition{Type: SimpleType{Kind: temp.Type.Kind}, Required: temp.Required}
	}
	return nil
}

// MarshalJSON implements encoding/json.Marshaler
func (self FieldDefinition) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{ \"type\": {")
	buf.WriteString(fmt.Sprintf("\"kind\": %d", self.Type.GetKind()))
	switch complexType := self.Type.(type) {
	case ListType:
		buf.WriteString(", \"extra\": ")
		v, err := json.Marshal(complexType.ComponentType)
		if err != nil {
			return nil, err
		}
		buf.Write(v)
	case EnumType:
		buf.WriteString(", \"extra\": ")
		r := rawEnumType{
			BaseType: complexType.BaseType,
			Values:   complexType.Values,
		}
		v, err := json.Marshal(r)
		if err != nil {
			return nil, err
		}
		buf.Write(v)
	}
	buf.WriteString("}")

	buf.WriteString(", \"required\": ")
	if self.Required {
		buf.WriteString("true")
	} else {
		buf.WriteString("false")
	}

	buf.WriteString(" }")
	return buf.Bytes(), nil
}
