package hw09structvalidator

import (
	"fmt"
	"reflect"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	panic("implement me")
}

func Validate(v interface{}) error {
	_ = ProcessField("", v, "")
	return nil
}

func ProcessField(path string, field any, tag string) []ValidationError {
	fmt.Printf("processing %s\n", path)
	var errors []ValidationError

	switch v := reflect.ValueOf(field); v.Kind() {
	case reflect.String:
		if tag != "" {
			if err := ValidateStringValue(v.String(), tag); err != nil {
				errors = append(errors, ValidationError{
					Field: path,
					Err:   err,
				})
			}
		}
	case reflect.Int:
		if tag != "" {
			if err := ValidateIntValue(v.Int(), tag); err != nil {
				errors = append(errors, ValidationError{
					Field: path,
					Err:   err,
				})
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if field.IsExported() {
				value := v.Field(i).Interface()
				alias, _ := field.Tag.Lookup("validate")
				pErr := ProcessField(path+"/"+field.Name, value, alias)
				errors = append(errors, pErr...)
			}
		}
	case reflect.Slice:
		sliceKind := reflect.TypeOf(field).Elem().Kind()
		switch sliceKind {
		case reflect.String:
			for i := 0; i < v.Len(); i++ {
				field := v.Index(i)
				value := field.String()
				if err := ValidateStringValue(value, tag); err != nil {
					errors = append(errors, ValidationError{
						Field: fmt.Sprintf("%s[%d]", path, i),
						Err:   err,
					})
				}
			}
		case reflect.Int:
			for i := 0; i < v.Len(); i++ {
				field := v.Index(i)
				value := field.Int()
				if err := ValidateIntValue(value, tag); err != nil {
					errors = append(errors, ValidationError{
						Field: fmt.Sprintf("%s[%d]", path, i),
						Err:   err,
					})
				}
			}
		case reflect.Struct:
			for i := 0; i < v.Len(); i++ {
				field := v.Index(i)
				value := field.Interface()
				elemPath := fmt.Sprintf("%s[%d]", path, i)
				vErrs := ProcessField(elemPath, value, tag)
				errors = append(errors, vErrs...)
			}
		}
	}

	return errors
}

func ValidateIntValue(value int64, tag string) error {
	fmt.Printf("validate %d by %s\n", value, tag)
	return nil
}

func ValidateStringValue(value string, tag string) error {
	fmt.Printf("validate %s by %s\n", value, tag)
	return nil
}
