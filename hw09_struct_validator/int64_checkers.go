package hw09structvalidator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrCheckInt64Max  = errors.New("value too big")
	ErrCheckInt64Min  = errors.New("value too small")
	ErrCheckInt64Enum = errors.New("value does not in expected values")
)

func int64ValidationBuilder(tag string) (checker[int64], error) {
	vldList := strings.Split(tag, "|")

	if len(vldList) == 1 {
		return createInt64StringValidator(tag)
	}

	vldRetList := make([]checker[int64], len(vldList))

	for id, vldDef := range vldList {
		var vld checker[int64]
		var err error
		if vld, err = createInt64StringValidator(vldDef); err != nil {
			return nil, err
		}
		vldRetList[id] = vld
	}

	return &int64ValidatorList{
		lst: vldRetList,
	}, nil
}

func createInt64StringValidator(tag string) (checker[int64], error) {
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
		inMap := make(map[int64]bool, len(inList))
		for _, inStrVal := range inList {
			val, err := strconv.Atoi(inStrVal)
			if err != nil {
				return nil, fmt.Errorf("%w - %s: %s", ErrTagParse, tag, err.Error())
			}
			inMap[int64(val)] = true
		}

		return &enumInt64Validator{
			emumMap: inMap,
		}, nil

	default:
		return nil, fmt.Errorf("%w - %s: unknown validator", ErrTagParse, tag)
	}
}

var _ checker[int64] = (*maxValidator)(nil)

type maxValidator struct {
	max int64
}

func (mx *maxValidator) Check(val int64) error {
	if val > mx.max {
		return ErrCheckInt64Max
	}
	return nil
}

var _ checker[int64] = (*minValidator)(nil)

type minValidator struct {
	min int64
}

func (mn *minValidator) Check(val int64) error {
	if val < mn.min {
		return ErrCheckInt64Min
	}
	return nil
}

var _ checker[int64] = (*enumInt64Validator)(nil)

type enumInt64Validator struct {
	emumMap map[int64]bool
}

func (inV *enumInt64Validator) Check(val int64) error {
	if !inV.emumMap[val] {
		return ErrCheckInt64Enum
	}
	return nil
}

var _ checker[int64] = (*int64ValidatorList)(nil)

type int64ValidatorList struct {
	lst []checker[int64]
}

func (sv *int64ValidatorList) Check(val int64) error {
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
