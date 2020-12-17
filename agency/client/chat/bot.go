package chat

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/findy-network/findy-agent-api/grpc/agency"
	"github.com/findy-network/findy-grpc/agency/client"
	"github.com/findy-network/findy-grpc/agency/client/chat/chat"
	"github.com/findy-network/findy-grpc/agency/fsm"
	"github.com/findy-network/findy-grpc/utils"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/lainio/err2"
)

type Bot struct {
	client.Conn
	fsm.MachineData
}

func LoadFSMMachineData(fName string, r io.Reader) (m fsm.MachineData, err error) {
	defer err2.Return(&err)
	data := err2.Bytes.Try(ioutil.ReadAll(r))
	return fsm.MachineData{FType: fName, Data: data}, nil
}

func LoadFSM(fName string, r io.Reader) (m *fsm.Machine, err error) {
	defer err2.Return(&err)
	data := err2.Bytes.Try(ioutil.ReadAll(r))
	m = loadFSMData(fName, data)
	err2.Check(m.Initialize())
	return m, nil
}

func loadFSMData(fName string, data []byte) *fsm.Machine {
	var machine fsm.Machine
	if filepath.Ext(fName) == ".json" {
		err2.Check(json.Unmarshal(data, &machine))
	} else {
		err2.Check(yaml.Unmarshal(data, &machine))
	}
	return &machine
}

func SaveFSM(m *fsm.Machine, fName string) (err error) {
	defer err2.Return(&err)
	data := marshalFSM(fName, m)
	err2.Check(ioutil.WriteFile(fName, data, 0644))
	return nil
}

func marshalFSM(fName string, fsm *fsm.Machine) []byte {
	var data []byte
	if filepath.Ext(fName) == ".json" {
		data = err2.Bytes.Try(json.MarshalIndent(fsm, "", "\t"))
	} else {
		data = err2.Bytes.Try(yaml.Marshal(fsm))
	}
	return data
}

func (b Bot) Run(intCh chan os.Signal) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // for server side stops, for proper cleanup

	ch, err := b.Conn.Listen(ctx, &agency.ClientID{Id: utils.UUID()})
	err2.Check(err)

	// this block is for development without fsm file
	//m := &EchoMachine //ReqProofMachine //EmailIssuerMachine //
	//err2.Check(SaveFSM(m, "echo-bot.yaml"))
	//url, err := fsm.GenerateURL("svg", m)
	//err2.Check(err)
	//println(url.String())
	//
	//m = &EmailIssuerMachine //
	//err2.Check(SaveFSM(m, "email-issuer-bot.yaml"))
	//url, err = fsm.GenerateURL("svg", m)
	//err2.Check(err)
	//println(url.String())
	//
	//m = &ReqProofMachine //EmailIssuerMachine //
	//err2.Check(SaveFSM(m, "email-verifier-bot.yaml"))
	//url, err = fsm.GenerateURL("svg", m)
	//err2.Check(err)
	//println(url.String())
	//
	//b.Machine = m

	//err2.Check(b.Machine.Initialize())
	//url, err := fsm.GenerateURL("svg", b.Machine)
	//err2.Check(err)
	//println(url.String())

	chat.Machine = b.MachineData

	go chat.Multiplexer(b.Conn)

loop:
	for {
		select {
		case status, ok := <-ch:
			if !ok {
				glog.V(2).Infoln("closed from server")
				break loop
			}
			glog.V(1).Infoln("listen status:",
				status.Notification.TypeId,
				status.Notification.Role,
				status.Notification.ProtocolId)
			chat.Status <- status
		case <-intCh:
			cancel()
			glog.V(2).Infoln("interrupted by user, cancel() called")
		}
	}
}
