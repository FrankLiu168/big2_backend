package helper

import (
	"encoding/json"
	"strings"
)

func ConvertToObject[T any](str string) (*T, error) {
	var result T
	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func ConvertToData[T any](item *T) (string, error) {
	o, err := json.Marshal(item)
	return string(o), err
}

func GetJsonPart(str string) string {
	s := strings.Index(str,"{")
	e := strings.Index(str,"}")
	return str[s:e+1]
}
