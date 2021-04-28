package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	ops "github.com/findy-network/findy-common-go/grpc/ops/v1"
	"github.com/findy-network/findy-common-go/jwt"
	"github.com/findy-network/findy-common-go/rpc"
	didexchange "github.com/findy-network/findy-common-go/std/didexchange/invitation"
	"github.com/findy-network/findy-common-go/utils"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
)

type Conn struct {
	*grpc.ClientConn
	cfg *rpc.ClientCfg
}

type Pairwise struct {
	Conn
	ID    string
	Label string
}

// BuildConnBase builds the rpc.ClientCfg from tls path and full service address
// including the port e.g. localhost:50051.
func BuildConnBase(tlsPath, fullAddr string, opts []grpc.DialOption) *rpc.ClientCfg {
	cfg := &rpc.ClientCfg{
		PKI:  rpc.LoadPKI(tlsPath),
		JWT:  "",
		Addr: fullAddr,
		Opts: opts,
	}
	return cfg
}

func BuildClientConnBase(tlsPath, addr string, port int, opts []grpc.DialOption) *rpc.ClientCfg {
	cfg := &rpc.ClientCfg{
		PKI:  rpc.LoadPKI(tlsPath),
		JWT:  "",
		Addr: fmt.Sprintf("%s:%d", addr, port),
		Opts: opts,
	}
	return cfg
}

func TryAuthOpen(jwtToken string, conf *rpc.ClientCfg) (c Conn) {
	if conf == nil {
		panic(errors.New("conf cannot be nil"))
	}
	conf.JWT = jwtToken
	conn, err := rpc.ClientConn(*conf)
	err2.Check(err)
	return Conn{ClientConn: conn, cfg: conf}
}

func TryOpen(user string, conf *rpc.ClientCfg) (c Conn) {
	glog.V(3).Infof("client with user \"%s\"", user)
	return TryAuthOpen(jwt.BuildJWT(user), conf)
}

func OkStatus(s *agency.ProtocolState) bool {
	return s.State == agency.ProtocolState_OK
}

func (pw Pairwise) Issue(ctx context.Context, credDefID, attrsJSON string) (ch chan *agency.ProtocolState, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_ISSUE_CREDENTIAL,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_IssueCredential{
			IssueCredential: &agency.Protocol_IssueCredentialMsg{
				CredDefID: credDefID,
				AttrFmt: &agency.Protocol_IssueCredentialMsg_AttributesJSON{
					AttributesJSON: attrsJSON,
				},
			},
		},
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (pw Pairwise) IssueWithAttrs(
	ctx context.Context,
	credDefID string,
	attrs *agency.Protocol_IssuingAttributes,
) (
	ch chan *agency.ProtocolState, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_ISSUE_CREDENTIAL,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_IssueCredential{IssueCredential: &agency.Protocol_IssueCredentialMsg{
			CredDefID: credDefID,
			AttrFmt: &agency.Protocol_IssueCredentialMsg_Attributes{
				Attributes: attrs,
			},
		}},
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (pw *Pairwise) Connection(ctx context.Context, invitationJSON string) (connID string, ch chan *agency.ProtocolState, err error) {
	defer err2.Return(&err)

	// assert that invitation is OK, and we need to return the connection ID
	// because it's the task id as well
	var invitation didexchange.Invitation
	err2.Check(json.Unmarshal([]byte(invitationJSON), &invitation))

	protocol := &agency.Protocol{
		TypeID: agency.Protocol_DIDEXCHANGE,
		Role:   agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_DIDExchange{DIDExchange: &agency.Protocol_DIDExchangeMsg{
			Label:          pw.Label,
			InvitationJSON: invitationJSON,
		}},
	}
	ch, err = pw.Conn.doRun(ctx, protocol)
	err2.Check(err)
	connID = invitation.ID
	pw.ID = connID
	return connID, ch, err
}

func (pw Pairwise) Ping(ctx context.Context) (ch chan *agency.ProtocolState, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_TRUST_PING,
		Role:         agency.Protocol_INITIATOR,
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (pw Pairwise) BasicMessage(ctx context.Context, content string) (ch chan *agency.ProtocolState, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_BASIC_MESSAGE,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_BasicMessage{
			BasicMessage: &agency.Protocol_BasicMessageMsg{
				Content: content,
			},
		},
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (pw Pairwise) ReqProof(ctx context.Context, proofAttrs string) (ch chan *agency.ProtocolState, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_PRESENT_PROOF,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_PresentProof{
			PresentProof: &agency.Protocol_PresentProofMsg{
				AttrFmt: &agency.Protocol_PresentProofMsg_AttributesJSON{
					AttributesJSON: proofAttrs}}},
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (pw Pairwise) ReqProofWithAttrs(ctx context.Context, proofAttrs *agency.Protocol_Proof) (ch chan *agency.ProtocolState, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_PRESENT_PROOF,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_PresentProof{
			PresentProof: &agency.Protocol_PresentProofMsg{
				AttrFmt: &agency.Protocol_PresentProofMsg_Attributes{
					Attributes: proofAttrs}}},
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (conn Conn) Listen(ctx context.Context, client *agency.ClientID, cOpts ...grpc.CallOption) (ch chan *agency.Question, err error) {
	defer err2.Return(&err)

	listenStatusCh, err := conn.ListenStatus(ctx, client, cOpts...)
	err2.Check(err)
	waitQuestionCh, err := conn.Wait(ctx, client, cOpts...)
	err2.Check(err)
	glog.V(3).Infoln("successful start of general listen id:", client.ID)
	ch = make(chan *agency.Question)

	go func() {
		defer close(ch)
	loop:
		for {
			select {
			case status, ok := <-listenStatusCh:
				if !ok {
					break loop
				}
				q := &agency.Question{
					Status: status,
				}
				ch <- q
			case question, ok := <-waitQuestionCh:
				if !ok {
					break loop
				}
				ch <- question
			}
		}
		glog.V(3).Infoln("general listen return")
	}()
	return ch, nil
}

func (conn Conn) ListenStatus(ctx context.Context, protocol *agency.ClientID, cOpts ...grpc.CallOption) (ch chan *agency.AgentStatus, err error) {
	defer err2.Return(&err)

	c := agency.NewAgentServiceClient(conn)
	statusCh := make(chan *agency.AgentStatus)

	stream, err := c.Listen(ctx, protocol, cOpts...)
	err2.Check(err)
	glog.V(3).Infoln("successful start of ListenStatus id:", protocol.ID)
	go func() {
		defer err2.CatchTrace(func(err error) {
			glog.V(1).Infoln("WARNING: error when reading response:", err)
			close(statusCh)
		})
		for {
			status, err := stream.Recv()
			if err == io.EOF {
				glog.V(3).Infoln("status stream end")
				close(statusCh)
				break
			}
			err2.Check(err)
			if status.Notification.TypeID == agency.Notification_KEEPALIVE {
				glog.V(5).Infoln("keepalive, no forward to client")
				continue
			}
			statusCh <- status
		}
	}()
	return statusCh, nil
}

func (conn Conn) Wait(ctx context.Context, protocol *agency.ClientID, cOpts ...grpc.CallOption) (ch chan *agency.Question, err error) {
	defer err2.Return(&err)

	c := agency.NewAgentServiceClient(conn)
	statusCh := make(chan *agency.Question)

	stream, err := c.Wait(ctx, protocol, cOpts...)
	err2.Check(err)
	glog.V(3).Infoln("successful start of Wait id:", protocol.ID)
	go func() {
		defer err2.CatchTrace(func(err error) {
			glog.V(1).Infoln("WARNING: error when reading response:", err)
			close(statusCh)
		})
		for {
			status, err := stream.Recv()
			if err == io.EOF {
				glog.V(3).Infoln("status stream end")
				close(statusCh)
				break
			}
			err2.Check(err)
			if status.TypeID == agency.Question_KEEPALIVE {
				glog.V(5).Infoln("keepalive, no forward to client")
				continue
			}
			statusCh <- status
		}
	}()
	return statusCh, nil
}

func (conn Conn) PSMHook(ctx context.Context, cOpts ...grpc.CallOption) (ch chan *ops.AgencyStatus, err error) {
	defer err2.Return(&err)

	opsClient := ops.NewAgencyServiceClient(conn)
	statusCh := make(chan *ops.AgencyStatus)

	stream, err := opsClient.PSMHook(ctx, &ops.DataHook{ID: utils.UUID()}, cOpts...)
	err2.Check(err)
	glog.V(3).Infoln("successful start of listen PSM hook id:")
	go func() {
		defer err2.CatchTrace(func(err error) {
			glog.V(1).Infoln("WARNING: error when reading response:", err)
			close(statusCh)
		})
		for {
			status, err := stream.Recv()
			if err == io.EOF {
				glog.V(3).Infoln("status stream end")
				close(statusCh)
				break
			}
			err2.Check(err)
			statusCh <- status
		}
	}()
	return statusCh, nil
}

func (conn Conn) doRun(ctx context.Context, protocol *agency.Protocol) (ch chan *agency.ProtocolState, err error) {
	defer err2.Return(&err)

	c := agency.NewProtocolServiceClient(conn)
	statusCh := make(chan *agency.ProtocolState)

	stream, err := c.Run(ctx, protocol)
	err2.Check(err)
	glog.V(3).Infoln("successful start of:", protocol.TypeID)
	go func() {
		defer err2.CatchTrace(func(err error) {
			glog.V(3).Infoln("err when reading response", err)
			close(statusCh)
		})
		for {
			status, err := stream.Recv()
			if err == io.EOF {
				glog.V(3).Infoln("status stream end")
				close(statusCh)
				break
			}
			err2.Check(err)
			statusCh <- status
		}
	}()
	return statusCh, nil
}

func (conn Conn) DoStart(ctx context.Context, protocol *agency.Protocol, cOpts ...grpc.CallOption) (pid *agency.ProtocolID, err error) {
	defer err2.Return(&err)

	c := agency.NewProtocolServiceClient(conn)
	pid, err = c.Start(ctx, protocol, cOpts...)
	err2.Check(err)

	glog.V(3).Infoln("successful start of:", protocol.TypeID)
	return pid, nil
}

func (conn Conn) DoResume(ctx context.Context, state *agency.ProtocolState, cOpts ...grpc.CallOption) (pid *agency.ProtocolID, err error) {
	defer err2.Return(&err)

	c := agency.NewProtocolServiceClient(conn)
	pid, err = c.Resume(ctx, state, cOpts...)
	err2.Check(err)

	glog.V(3).Infoln("successful resume of:", state.ProtocolID.TypeID)
	return pid, nil
}

func (conn Conn) DoRelease(ctx context.Context, id *agency.ProtocolID, cOpts ...grpc.CallOption) (pid *agency.ProtocolID, err error) {
	defer err2.Return(&err)

	c := agency.NewProtocolServiceClient(conn)
	pid, err = c.Release(ctx, id, cOpts...)
	err2.Check(err)

	glog.V(3).Infoln("successful release of:", id.TypeID)
	return pid, nil
}

func (conn Conn) DoStatus(ctx context.Context, id *agency.ProtocolID, cOpts ...grpc.CallOption) (status *agency.ProtocolStatus, err error) {
	defer err2.Return(&err)

	c := agency.NewProtocolServiceClient(conn)
	status, err = c.Status(ctx, id, cOpts...)
	err2.Check(err)

	glog.V(3).Infoln("successful status of:", id.TypeID)
	return status, nil
}
