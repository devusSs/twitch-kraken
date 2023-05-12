package utils

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

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

// Helper function to generate random string
func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// Generate a random string of len x.
func RandomString(length int) string {
	return stringWithCharset(length, charset)
}
