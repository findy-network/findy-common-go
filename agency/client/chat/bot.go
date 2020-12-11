package chat

import (
	"context"
	"encoding/json"
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
	fsm fsm.Machine
}

func (b *Bot) LoadFSM(fName string) (err error) {
	defer err2.Return(&err)
	data := err2.Bytes.Try(ioutil.ReadFile(fName))
	b.fsm = *loadFSMData(fName, data)
	err2.Check(b.fsm.Initialize())
	return nil
}

func loadFSMData(fName string, data []byte) *fsm.Machine {
	var fsm fsm.Machine
	if filepath.Ext(fName) == ".json" {
		err2.Check(json.Unmarshal(data, &fsm))
	} else {
		err2.Check(yaml.Unmarshal(data, &fsm))
	}
	return &fsm
}

func (b *Bot) SaveFSM(fName string) (err error) {
	defer err2.Return(&err)
	data := marshalFSM(fName, &b.fsm)
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
	//b.fsm = EchoMachine //ReqProofMachine //EmailIssuerMachine //
	b.fsm = EmailIssuerMachine //
	//b.fsm = ReqProofMachine //EmailIssuerMachine //
	err2.Check(b.fsm.Initialize())

	// this block is for testing file loading
	//err2.Check(b.SaveFSM("emailbot.json"))
	err2.Check(b.SaveFSM("emailbot.yaml"))
	chat.Machine = &b.fsm

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
