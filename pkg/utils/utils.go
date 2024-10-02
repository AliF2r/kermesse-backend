package utils

import (
	"fmt"
)

func ConvertToInt(input map[string]interface{}, key string) (int, error) {
	value, ok := input[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("%s dosn't exist", key)
	}
	floatValue, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("%s is not a valid number", key)
	}
	return int(floatValue), nil
}
