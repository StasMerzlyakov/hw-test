package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

var ErrTagParse = errors.New("parse tag error")

func (v ValidationErrors) Error() string {
	panic("implement me")
}

func Validate(v interface{}) error {
	vErr := ProcessField("", v, "nested")
	return ValidationErrors(vErr)
}

func ProcessField(path string, field any, tag string) []ValidationError {
	fmt.Printf("processing %s\n", path)
	var errors []ValidationError

	switch v := reflect.ValueOf(field); v.Kind() {
	case reflect.String:
		if tag != "" {
			if err := validateStringValue(v.String(), tag); err != nil {
				errors = append(errors, ValidationError{
					Field: path,
					Err:   err,
				})
			}
		}
	case reflect.Int:
		if tag != "" {
			if err := validateIntValue(v.Int(), tag); err != nil {
				errors = append(errors, ValidationError{
					Field: path,
					Err:   err,
				})
			}
		}
	case reflect.Struct:
		if tag == "nested" {
			for i := 0; i < v.NumField(); i++ {
				field := v.Type().Field(i)
				if field.IsExported() {
					value := v.Field(i).Interface()
					alias, _ := field.Tag.Lookup("validate")
					pErr := ProcessField(path+"/"+field.Name, value, alias)
					errors = append(errors, pErr...)
				}
			}
		}
	case reflect.Slice:
		sliceKind := reflect.TypeOf(field).Elem().Kind()
		switch sliceKind {
		case reflect.String:
			for i := 0; i < v.Len(); i++ {
				field := v.Index(i)
				value := field.String()
				if err := validateStringValue(value, tag); err != nil {
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
				if err := validateIntValue(value, tag); err != nil {
					errors = append(errors, ValidationError{
						Field: fmt.Sprintf("%s[%d]", path, i),
						Err:   err,
					})
				}
			}
		case reflect.Struct:
			if tag == "nested" {
				for i := 0; i < v.Len(); i++ {
					field := v.Index(i)
					value := field.Interface()
					elemPath := fmt.Sprintf("%s[%d]", path, i)
					vErrs := ProcessField(elemPath, value, tag)
					errors = append(errors, vErrs...)
				}
			}
		}
	}

	return errors
}

func validateIntValue(value int64, tag string) error {
	fmt.Printf("validate %d by %s\n", value, tag)
	validator, err := int64ValidationFn(tag)
	if err != nil {
		return err
	}
	return validator.Validate(value)
}

func validateStringValue(value string, tag string) error {
	fmt.Printf("validate %s by %s\n", value, tag)
	validator, err := stringValidationFn(tag)
	if err != nil {
		return err
	}
	return validator.Validate(value)
}

// validators block
type validationType interface {
	int64 | string
}

var int64ValidationFn = createValidationFn[int64](int64ValidationBuilder)
var stringValidationFn = createValidationFn[string](stringValidationBuilder)

func createValidationFn[V validationType](builder func(string) (validator[V], error)) func(string) (validator[V], error) {
	var validatorCache = make(map[string]validator[V])
	var validatorCacheMtx = &sync.Mutex{}
	return func(tag string) (validator[V], error) {
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

func int64ValidationBuilder(tag string) (validator[int64], error) {
	vldList := strings.Split(tag, "|")

	if len(vldList) == 1 {
		switch {
		case strings.HasPrefix(tag, "min"):
			tagArr := strings.Split(tag, ":")

			if len(tagArr) != 2 {
				return nil, fmt.Errorf("%w - %s: unexpected min validator definition", ErrTagParse, tag)
			}

			minVal, err := strconv.Atoi(tagArr[1])
			if err != nil {
				return nil, fmt.Errorf("%w - %s: %s", ErrTagParse, tag, err.Error())
			}

			return &minValidator{
				min: int64(minVal),
			}, nil
		case strings.HasPrefix(tag, "max"):
			tagArr := strings.Split(tag, ":")

			if len(tagArr) != 2 {
				return nil, fmt.Errorf("%w - %s: unexpected max validator definition", ErrTagParse, tag)
			}

			maxVal, err := strconv.Atoi(tagArr[1])
			if err != nil {
				return nil, fmt.Errorf("%w - %s: %s", ErrTagParse, tag, err.Error())
			}

			return &maxValidator{
				max: int64(maxVal),
			}, nil

		case strings.HasPrefix(tag, "in"):
			tagArr := strings.Split(tag, ":")

			if len(tagArr) != 2 {
				return nil, fmt.Errorf("%w - %s: unexpected in validator definition", ErrTagParse, tag)
			}

			inList := strings.Split(tagArr[1], ",")
			var inMap = make(map[int64]bool, len(inList))
			for _, inStrVal := range inList {
				val, err := strconv.Atoi(inStrVal)
				if err != nil {
					return nil, fmt.Errorf("%w - %s: %s", ErrTagParse, tag, err.Error())
				}
				inMap[int64(val)] = true
			}

			return &inInt64Validator{
				inMap: inMap,
			}, nil

		default:
			return nil, fmt.Errorf("%w - %s: unknown validator", ErrTagParse, tag)

		}
	}

	var vldRetList = make([]validator[int64], len(vldList))

	for id, vldDef := range vldList {
		if vld, err := int64ValidationBuilder(vldDef); err != nil {
			return nil, err
		} else {
			vldRetList[id] = vld
		}
	}

	return &int64ValidatorList{
		lst: vldRetList,
	}, nil
}

func stringValidationBuilder(tag string) (validator[string], error) {
	vldList := strings.Split(tag, "|")

	if len(vldList) == 1 {
		switch {
		case strings.HasPrefix(tag, "len"):
			tagArr := strings.Split(tag, ":")

			if len(tagArr) != 2 {
				return nil, fmt.Errorf("%w - %s: unexpected len validator definition", ErrTagParse, tag)
			}

			len, err := strconv.Atoi(tagArr[1])
			if err != nil {
				return nil, fmt.Errorf("%w - %s: %s", ErrTagParse, tag, err.Error())
			}

			return &lenValidator{
				len: len,
			}, nil
		case strings.HasPrefix(tag, "regexp"):
			tagArr := strings.Split(tag, ":")

			if len(tagArr) != 2 {
				return nil, fmt.Errorf("%w - %s: unexpected regexp validator definition", ErrTagParse, tag)
			}

			r, err := regexp.Compile(tagArr[1])
			if err != nil {
				return nil, fmt.Errorf("%w - %s: %s", ErrTagParse, tag, err.Error())
			}

			return &regextValidator{
				reg: r,
			}, nil

		case strings.HasPrefix(tag, "in"):
			tagArr := strings.Split(tag, ":")

			if len(tagArr) != 2 {
				return nil, fmt.Errorf("%w - %s: unexpected in validator definition", ErrTagParse, tag)
			}

			inList := strings.Split(tagArr[1], ",")
			var inMap = make(map[string]bool, len(inList))
			for _, inStrVal := range inList {
				inMap[inStrVal] = true
			}

			return &inStringValidator{
				inMap: inMap,
			}, nil

		default:
			return nil, fmt.Errorf("%w - %s: unknown validator", ErrTagParse, tag)

		}
	}

	var vldRetList = make([]validator[string], len(vldList))

	for id, vldDef := range vldList {
		if vld, err := stringValidationBuilder(vldDef); err != nil {
			return nil, err
		} else {
			vldRetList[id] = vld
		}
	}

	return &stringValidatorList{
		lst: vldRetList,
	}, nil
}

type validator[T validationType] interface {
	Validate(t T) error
}

// string validators
var _ validator[string] = (*lenValidator)(nil)

type lenValidator struct {
	len int
}

func (lv *lenValidator) Validate(val string) error {
	if len(val) != lv.len {
		return errors.New("wong field length")
	}
	return nil
}

var _ validator[string] = (*regextValidator)(nil)

type regextValidator struct {
	reg *regexp.Regexp
}

func (re *regextValidator) Validate(val string) error {
	if !re.reg.MatchString(val) {
		return errors.New("value does not match regexp")
	}
	return nil
}

var _ validator[string] = (*inStringValidator)(nil)

type inStringValidator struct {
	inMap map[string]bool
}

func (inV *inStringValidator) Validate(val string) error {
	if !inV.inMap[val] {
		return errors.New("value does not in expected values")
	}
	return nil
}

var _ validator[string] = (*stringValidatorList)(nil)

type stringValidatorList struct {
	lst []validator[string]
}

func (sv *stringValidatorList) Validate(val string) error {
	var errs strings.Builder

	for _, valid := range sv.lst {
		if err := valid.Validate(val); err != nil {
			errs.WriteString(fmt.Sprintf("%s;", err.Error()))
		}
	}

	if errs.Len() > 0 {
		return errors.New(errs.String())
	}
	return nil
}

// int validators
var _ validator[int64] = (*maxValidator)(nil)

type maxValidator struct {
	max int64
}

func (mx *maxValidator) Validate(val int64) error {
	if val > mx.max {
		return errors.New("value too big")
	}
	return nil
}

var _ validator[int64] = (*minValidator)(nil)

type minValidator struct {
	min int64
}

func (mn *minValidator) Validate(val int64) error {
	if val < mn.min {
		return errors.New("value too small")
	}
	return nil
}

var _ validator[int64] = (*inInt64Validator)(nil)

type inInt64Validator struct {
	inMap map[int64]bool
}

func (inV *inInt64Validator) Validate(val int64) error {
	if !inV.inMap[val] {
		return errors.New("value does not in expected values")
	}
	return nil
}

var _ validator[int64] = (*int64ValidatorList)(nil)

type int64ValidatorList struct {
	lst []validator[int64]
}

func (sv *int64ValidatorList) Validate(val int64) error {
	var errs strings.Builder

	for _, valid := range sv.lst {
		if err := valid.Validate(val); err != nil {
			errs.WriteString(fmt.Sprintf("%s;", err.Error()))
		}
	}

	if errs.Len() > 0 {
		return errors.New(errs.String())
	}
	return nil
}
