package utils

import (
	"fmt"
	"net/http"
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

func GetParams(r *http.Request) map[string]interface{} {
	queryParams := r.URL.Query()
	params := make(map[string]interface{})
	for key, values := range queryParams {
		if len(values) == 1 {
			params[key] = values[0]
		} else {
			params[key] = values
		}
	}
	return params
}
