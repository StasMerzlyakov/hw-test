package hw09structvalidator

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrCheckStringLength = errors.New("wong field length")
	ErrCheckStringRegexp = errors.New("value does not match regexp")
	ErrCheckStringEnum   = errors.New("value does not in expected values")
)

func stringValidationBuilder(tag string) (checker[string], error) {
	vldList := strings.Split(tag, "|")

	if len(vldList) == 1 {
		return createSimpleStringValidator(tag)
	}

	vldRetList := make([]checker[string], len(vldList))

	for id, vldDef := range vldList {
		var vld checker[string]
		var err error
		if vld, err = createSimpleStringValidator(vldDef); err != nil {
			return nil, err
		}
		vldRetList[id] = vld
	}

	return &stringValidatorList{
		lst: vldRetList,
	}, nil
}

func createSimpleStringValidator(tag string) (checker[string], error) {
	switch {
	case strings.HasPrefix(tag, "len"):
		tagArr := strings.Split(tag, ":")

		if len(tagArr) != 2 {
			return nil, fmt.Errorf("%w - %s: unexpected len validator definition", ErrTagParse, tag)
		}

		ln, err := strconv.Atoi(tagArr[1])
		if err != nil {
			return nil, fmt.Errorf("%w - %s: %s", ErrTagParse, tag, err.Error())
		}

		return &lenValidator{
			len: ln,
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
		inMap := make(map[string]bool, len(inList))
		for _, inStrVal := range inList {
			inMap[inStrVal] = true
		}

		return &enumStringValidator{
			enumMap: inMap,
		}, nil

	default:
		return nil, fmt.Errorf("%w - %s: unknown validator", ErrTagParse, tag)
	}
}

var _ checker[string] = (*lenValidator)(nil)

type lenValidator struct {
	len int
}

func (lv *lenValidator) Check(val string) error {
	if utf8.RuneCountInString(val) != lv.len {
		return ErrCheckStringLength
	}
	return nil
}

var _ checker[string] = (*regextValidator)(nil)

type regextValidator struct {
	reg *regexp.Regexp
}

func (re *regextValidator) Check(val string) error {
	if !re.reg.MatchString(val) {
		return ErrCheckStringRegexp
	}
	return nil
}

var _ checker[string] = (*enumStringValidator)(nil)

type enumStringValidator struct {
	enumMap map[string]bool
}

func (enum *enumStringValidator) Check(val string) error {
	if !enum.enumMap[val] {
		return ErrCheckStringEnum
	}
	return nil
}

var _ checker[string] = (*stringValidatorList)(nil)

type stringValidatorList struct {
	lst []checker[string]
}

func (sv *stringValidatorList) Check(val string) error {
	var errs []error

	for _, valid := range sv.lst {
		if err := valid.Check(val); err != nil {
			errs = append(errs, err)
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return ErrCheckErrorList(errs)
	}
}
