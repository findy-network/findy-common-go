// Package dto is for data transfer objects.
package dto

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
)

// ToJSON is a helper to convert dto to JSON string.
func ToJSON(dto interface{}) string {
	return string(ToJSONBytes(dto))
}

// FromJSONStr is a helper to convert object from a JSON string.
func FromJSONStr(str string, dto interface{}) {
	FromJSON([]byte(str), dto)
}

// FromJSON is a helper to convert byte JSON to a data object.
func FromJSON(bytes []byte, dto interface{}) {
	err := json.Unmarshal(bytes, dto)
	if err != nil {
		glog.Errorf("%s: from JSON:\n%s\n", err.Error(), string(bytes))
		panic(err)
	}
}

// ToJSONBytes is a helper to convert dto to JSON byte data.
func ToJSONBytes(dto interface{}) []byte {
	output, err := json.Marshal(dto)
	if err != nil {
		fmt.Println("err marshaling to JSON:", err)
		return nil
	}
	return output
}

// JSONArray returns a JSON array in string from strings given.
func JSONArray(strs ...string) string {
	return DoJSONArray(strs)
}

// DoJSONArray returns a JSON array in string from strings given.
func DoJSONArray(strs []string) string {
	var b strings.Builder
	b.WriteRune('[')

	for i, s := range strs {
		if i > 0 {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(s)
		b.WriteRune('"')
	}
	b.WriteRune(']')
	return b.String()
}

// ToGOB returns bytes of the object in GOB format.
func ToGOB(dto interface{}) []byte {
	var buf bytes.Buffer
	// Create an encoder and send a value.
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(dto)
	if err != nil {
		glog.Error("encode:", err)
		panic(err)
	}
	return buf.Bytes()
}

// FromGOB reads object from bytes. Remember pass a pointer to preallocated
// object of the right type.
//
//	p := &PSM{}
//	dto.FromGOB(d, p)
//	return p
func FromGOB(data []byte, dto interface{}) {
	network := bytes.NewReader(data)
	dec := gob.NewDecoder(network)
	err := dec.Decode(dto)
	if err != nil {
		glog.Error("decode:", err)
		panic(err)
	}
}
