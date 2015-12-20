package common

import (
	"fmt"
	"strconv"
)

func makeIntDesc(valName string, minValue, maxValue int64) string {
	return fmt.Sprintf("Parameter %s must be followed by an integer value in the range of [%d, %d]",
		valName, minValue, maxValue)
}

func ParseInt64Param(params []string, minValue, maxValue int64) ([]string, int64, *ErrorResponse) {
	valName := params[0]
	if len(params) < 2 {
		return nil, 0, InvalidRequest(makeIntDesc(valName, minValue, maxValue))
	}
	val, err := strconv.ParseInt(params[1], 10, 0)

	if err == nil && val >= minValue && val <= maxValue {
		return params[2:], val, nil
	}

	return nil, 0, InvalidRequest(makeIntDesc(valName, minValue, maxValue))
}

func makeStrDesc(valName string, minLen, maxLen int64) string {
	return fmt.Sprintf("Parameter %s must be followed by a string value with the min length %d and max length %d",
		minLen, maxLen)
}

func ParseStringParam(params []string, minLen, maxLen int64) ([]string, string, *ErrorResponse) {
	valName := params[0]
	if len(params) < 2 {
		return nil, "", InvalidRequest(makeStrDesc(valName, minLen, maxLen))
	}
	paramLen := int64(len(params[1]))
	if paramLen >= minLen && paramLen <= maxLen {
		return params[2:], params[1], nil
	}
	return nil, "", InvalidRequest(makeStrDesc(valName, minLen, maxLen))
}

// ParseUserItemId parses user provided item id that can not start with '_'.
func ParseUserItemId(params []string, minLen, maxLen int64) ([]string, string, *ErrorResponse) {
	valName := params[0]
	if len(params) >= 2 {
		itemLength := int64(len(params[1]))
		if itemLength >= minLen && itemLength <= maxLen {
			if ValidateUserItemId(params[1]) {
				return params[2:], params[1], nil
			} else {
				return nil, "", InvalidRequest("Only ^[a-zA-Z0-9][_a-zA-Z0-9]* symbols are allowed for id")
			}
		}
	}
	return nil, "", InvalidRequest(makeStrDesc(valName, minLen, maxLen))
}

// ParseItemId parses item id that can use all characters.
func ParseItemId(params []string, minLen, maxLen int64) ([]string, string, *ErrorResponse) {
	valName := params[0]
	if len(params) >= 2 {
		itemLength := int64(len(params[1]))
		if itemLength >= minLen && itemLength <= maxLen {
			if ValidateItemId(params[1]) {
				return params[2:], params[1], nil
			} else {
				return nil, "", InvalidRequest("Only [_a-zA-Z0-9]* symbols are allowed for id")
			}
		}
	}
	return nil, "", InvalidRequest(makeStrDesc(valName, minLen, maxLen))
}
