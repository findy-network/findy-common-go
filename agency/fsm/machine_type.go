package fsm

import (
	"encoding/json"

	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

type MachineType int

const (
	MachineTypeNone         = 0
	MachineTypeConversation = 1 + iota
	MachineTypeBackend
)

var machineTypeNames = map[MachineType]string{
	MachineTypeNone:         "MachineTypeNone",
	MachineTypeConversation: "MachineTypeConversation",
	MachineTypeBackend:      "MachineTypeBackend",
}

var machineTypeValues = map[string]MachineType{
	"MachineTypeNone":         MachineTypeNone,
	"MachineTypeConversation": MachineTypeConversation,
	"MachineTypeBackend":      MachineTypeBackend,
}

func (mt MachineType) String() string {
	return machineTypeNames[mt]
}

func ParseMachineType(s string) (mt MachineType, err error) {
	defer err2.Handle(&err)
	mt = assert.MKeyExists(machineTypeValues, s)
	return mt, nil
}

func (mt *MachineType) MarshalJSON() ([]byte, error) {
	return json.Marshal(mt.String())
}

func (mt *MachineType) UnmarshalJSON(data []byte) (err error) {
	defer err2.Handle(&err)

	var machineType string
	try.To(json.Unmarshal(data, &machineType))
	*mt = try.To1(ParseMachineType(machineType))
	return nil
}
