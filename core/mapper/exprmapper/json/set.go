package json

import (
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/json/field"
)

type JSONData struct {
	container *Container
	rw        sync.RWMutex
}

func SetStringValue(data interface{}, jsonData string, mappingField *field.MappingField) (interface{}, error) {
	jsonParsed, err := ParseJSON([]byte(jsonData))
	if err != nil {
		return nil, err

	}
	container := &JSONData{container: jsonParsed, rw: sync.RWMutex{}}
	err = handleSetValue(data, container, mappingField.Getfields())
	return container.container.object, err
}

func SetFieldValue(data interface{}, jsonData interface{}, mappingField *field.MappingField) (interface{}, error) {
	switch t := jsonData.(type) {
	case string:
		return SetStringValue(data, t, mappingField)
	default:
		jsonParsed, err := Consume(jsonData)
		if err != nil {
			return nil, err
		}
		container := &JSONData{container: jsonParsed, rw: sync.RWMutex{}}
		err = handleSetValue(data, container, mappingField.Getfields())
		if err != nil {
			return nil, err
		}
		return container.container.object, nil
	}

}
