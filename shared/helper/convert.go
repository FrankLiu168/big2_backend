package helper

import (
	"big2backend/shared/data"
	"encoding/json"
	"strings"
)

func ConvertToBasePayload(str string) *data.BasePayload {
	basePayload, _ := ConvertToObject[data.BasePayload](str)
	return basePayload
}

func ConvertToPayload[T1 any](basePayload *data.BasePayload) *T1 {
	payload, _ := ConvertToObject[T1](basePayload.Data)
	return payload
}

func PackPayload[T1 any](commandAction data.CommandAction, target string, payload *T1) string {
	basePayload := data.BasePayload{
		CommandAction: commandAction,
		Target:        target,
	}
	cmdPayload, _ := ConvertToData(payload)
	basePayload.Data = cmdPayload
	str, _ := ConvertToData(&basePayload)
	return str
}

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
	s := strings.Index(str, "{")
	e := strings.Index(str, "}")
	return str[s : e+1]
}
