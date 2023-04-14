package utils

import (
	"encoding/json"
	"errors"
)

// Might not be the best solution. Copied from:
//
// https://stackoverflow.com/questions/34070369/removing-a-string-from-a-slice-in-go
func RemoveStringFromSlice(s []string, index int) ([]string, error) {
	if index >= len(s) {
		return nil, errors.New("index out of range")
	}
	return append(s[:index], s[index+1:]...), nil
}

func CheckStringSliceForDuplicates(s []string, str string) bool {
	for _, st := range s {
		if st == str {
			return true
		}
	}
	return false
}

func MarshalStruct(input interface{}) (string, error) {
	bytes, err := json.Marshal(input)
	return string(bytes), err
}
