package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sbuf strings.Builder
	for _, err := range v {
		sbuf.WriteString(fmt.Sprintf("field %s validation error - %s", err.Field, err.Err.Error()))
	}
	return sbuf.String()
}

type ErrCheckErrorList []error

func (ecl ErrCheckErrorList) Error() string {
	var sbuf strings.Builder
	for _, err := range ecl {
		sbuf.WriteString(fmt.Sprintf("check error - %s;", err.Error()))
	}
	return sbuf.String()
}

var ErrTagParse = errors.New("parse tag error")

func Validate(v interface{}) error {
	vErr := processField("", v, "nested")
	if len(vErr) > 0 {
		return ValidationErrors(vErr)
	}
	return nil
}

func processField(path string, field any, tag string) []ValidationError {
	var errors []ValidationError

	switch v := reflect.ValueOf(field); v.Kind() { //nolint:exhaustive
	case reflect.String:
		if err := validateStringValue(v.String(), tag); err != nil {
			errors = append(errors, ValidationError{
				Field: path,
				Err:   err,
			})
		}

	case reflect.Int, reflect.Int16, reflect.Int8, reflect.Int32, reflect.Int64:
		if err := validateIntValue(v.Int(), tag); err != nil {
			errors = append(errors, ValidationError{
				Field: path,
				Err:   err,
			})
		}
	case reflect.Struct:
		if tag == "nested" {
			for i := 0; i < v.NumField(); i++ {
				field := v.Type().Field(i)
				if field.IsExported() {
					value := v.Field(i).Interface()
					alias, _ := field.Tag.Lookup("validate")
					pErr := processField(path+"/"+field.Name, value, alias)
					errors = append(errors, pErr...)
				}
			}
		}
	case reflect.Slice:
		if tag != "" {
			for i := 0; i < v.Len(); i++ {
				field := v.Index(i)
				fieldErr := processField(fmt.Sprintf("%s[%d]", path, i), field.Interface(), tag)
				errors = append(errors, fieldErr...)
			}
		}
	}

	return errors
}

func validateIntValue(value int64, tag string) error {
	if tag != "" {
		validator, err := int64ValidationFn(tag)
		if err != nil {
			return err
		}
		return validator.Check(value)
	}
	return nil
}

func validateStringValue(value string, tag string) error {
	if tag != "" {
		validator, err := stringValidationFn(tag)
		if err != nil {
			return err
		}
		return validator.Check(value)
	}
	return nil
}

type validationType interface {
	int64 | string
}

type checker[T validationType] interface {
	Check(t T) error
}

var (
	int64ValidationFn  = createValidationFn[int64](int64ValidationBuilder)
	stringValidationFn = createValidationFn[string](stringValidationBuilder)
)

func createValidationFn[V validationType](builder func(string) (checker[V], error)) func(string) (checker[V], error) {
	validatorCache := make(map[string]checker[V])
	validatorCacheMtx := &sync.Mutex{}
	return func(tag string) (checker[V], error) {
		validatorCacheMtx.Lock()
		defer validatorCacheMtx.Unlock()

		valid := validatorCache[tag]
		if valid != nil {
			return valid, nil
		}

		valid, err := builder(tag)
		if err != nil {
			return nil, err
		}
		validatorCache[tag] = valid
		return valid, nil
	}
}
