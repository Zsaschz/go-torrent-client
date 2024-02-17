package main

import (
	"reflect"
)

func MapToStruct(m map[string]interface{}, s interface{}) error {
	stValue := reflect.ValueOf(s).Elem()
	stType := stValue.Type()
	for i := 0; i < stType.NumField(); i++ {
		field := stType.Field(i)
		if value, ok := m[field.Name]; ok {
			stValue.Field(i).Set(reflect.ValueOf(value))
		}
	}
	return nil
}
