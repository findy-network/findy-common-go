// Package chat implments state-machine based chat bots
package chat

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/findy-network/findy-common-go/agency/client"
	"github.com/findy-network/findy-common-go/agency/client/chat/chat"
	"github.com/findy-network/findy-common-go/agency/fsm"
	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/findy-network/findy-common-go/utils"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type Bot struct {
	client.Conn
	fsm.MachineData
}

func LoadFSMMachineData(fName string, r io.Reader) (m fsm.MachineData, err error) {
	defer err2.Handle(&err)
	data := try.To1(io.ReadAll(r))
	return fsm.MachineData{FType: fName, Data: data}, nil
}

func LoadFSM(fName string, r io.Reader) (m *fsm.Machine, err error) {
	defer err2.Handle(&err)
	data := try.To1(io.ReadAll(r))
	m = loadFSMData(fName, data)
	try.To(m.Initialize())
	return m, nil
}

func loadFSMData(fName string, data []byte) *fsm.Machine {
	var machine fsm.Machine
	if filepath.Ext(fName) == ".json" {
		try.To(json.Unmarshal(data, &machine))
	} else {
		try.To(yaml.Unmarshal(data, &machine))
	}
	return &machine
}

func SaveFSM(m *fsm.Machine, fName string) (err error) {
	defer err2.Handle(&err)
	data := marshalFSM(fName, m)
	try.To(os.WriteFile(fName, data, 0644))
	return nil
}

func marshalFSM(fName string, fsm *fsm.Machine) []byte {
	var data []byte
	if filepath.Ext(fName) == ".json" {
		data = try.To1(json.MarshalIndent(fsm, "", "\t"))
	} else {
		data = try.To1(yaml.Marshal(fsm))
	}
	return data
}

// Run starts to run a chatbot and its FSM instances.
func (b Bot) Run(intCh chan os.Signal) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // for server side stops, for proper cleanup

	client := &agency.ClientID{ID: utils.UUID()}
	ch := try.To1(b.Conn.ListenStatus(ctx, client))
	questionCh := try.To1(b.Conn.Wait(ctx, client))

	chat.Machine = b.MachineData

	go chat.Multiplexer(b.Conn, intCh)

loop:
	for {
		select {
		case status, ok := <-ch:
			if !ok {
				glog.V(20).Infoln("closed from server")
				break loop
			}
			glog.V(5).Infoln("listen status:",
				status.Notification.TypeID,
				status.Notification.Role,
				status.Notification.ProtocolID)
			chat.Status <- status
		case question, ok := <-questionCh:
			if !ok {
				glog.V(20).Infoln("closed from server")
				break loop
			}
			glog.V(5).Infoln("listen question status:",
				question.TypeID,
				question.Status.Notification.Role,
				question.Status.Notification.ProtocolID)
			chat.Question <- question
		case <-intCh:
			cancel()
		}
	}
}
