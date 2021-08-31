package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	ops "github.com/findy-network/findy-common-go/grpc/ops/v1"
	"github.com/findy-network/findy-common-go/jwt"
	"github.com/findy-network/findy-common-go/rpc"
	didexchange "github.com/findy-network/findy-common-go/std/didexchange/invitation"
	"github.com/findy-network/findy-common-go/tools/sleeper"
	"github.com/findy-network/findy-common-go/utils"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
)

const sleeperFloor = 10 * time.Second

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
		PKI:  rpc.LoadPKIWithServerName(tlsPath, fullAddr),
		JWT:  "",
		Addr: fullAddr,
		Opts: opts,
	}
	return cfg
}

func BuildClientConnBase(
	tlsPath, addr string,
	port int,
	opts []grpc.DialOption,
) *rpc.ClientCfg {

	cfg := &rpc.ClientCfg{
		PKI:  rpc.LoadPKIWithServerName(tlsPath, addr),
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

func (pw Pairwise) Issue(
	ctx context.Context,
	credDefID, attrsJSON string,
) (
	ch chan *agency.ProtocolState,
	err error,
) {
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
		StartMsg: &agency.Protocol_IssueCredential{
			IssueCredential: &agency.Protocol_IssueCredentialMsg{
				CredDefID: credDefID,
				AttrFmt: &agency.Protocol_IssueCredentialMsg_Attributes{
					Attributes: attrs,
				},
			},
		},
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (pw Pairwise) ProposeIssue(
	ctx context.Context,
	credDefID, attrsJSON string,
) (
	ch chan *agency.ProtocolState,
	err error,
) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_ISSUE_CREDENTIAL,
		Role:         agency.Protocol_ADDRESSEE,
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

func (pw Pairwise) ProposeIssueWithAttrs(
	ctx context.Context,
	credDefID string,
	attrs *agency.Protocol_IssuingAttributes,
) (
	ch chan *agency.ProtocolState, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_ISSUE_CREDENTIAL,
		Role:         agency.Protocol_ADDRESSEE,
		StartMsg: &agency.Protocol_IssueCredential{
			IssueCredential: &agency.Protocol_IssueCredentialMsg{
				CredDefID: credDefID,
				AttrFmt: &agency.Protocol_IssueCredentialMsg_Attributes{
					Attributes: attrs,
				},
			},
		},
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (pw *Pairwise) Connection(
	ctx context.Context,
	invitationJSON string,
) (
	connID string,
	ch chan *agency.ProtocolState,
	err error,
) {
	defer err2.Return(&err)

	// assert that invitation is OK, and we need to return the connection ID
	// because it's the task id as well
	var invitation didexchange.Invitation
	err2.Check(json.Unmarshal([]byte(invitationJSON), &invitation))

	protocol := &agency.Protocol{
		TypeID: agency.Protocol_DIDEXCHANGE,
		Role:   agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_DIDExchange{
			DIDExchange: &agency.Protocol_DIDExchangeMsg{
				Label:          pw.Label,
				InvitationJSON: invitationJSON,
			},
		},
	}
	ch, err = pw.Conn.doRun(ctx, protocol)
	err2.Check(err)
	connID = invitation.ID
	pw.ID = connID
	return connID, ch, err
}

func (pw Pairwise) Ping(
	ctx context.Context,
) (
	ch chan *agency.ProtocolState,
	err error,
) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_TRUST_PING,
		Role:         agency.Protocol_INITIATOR,
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (pw Pairwise) BasicMessage(
	ctx context.Context,
	content string,
) (
	ch chan *agency.ProtocolState,
	err error,
) {
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

func (pw Pairwise) ReqProof(
	ctx context.Context,
	proofAttrs string,
) (
	ch chan *agency.ProtocolState,
	err error,
) {
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

func (pw Pairwise) ReqProofWithAttrs(
	ctx context.Context,
	proofAttrs *agency.Protocol_Proof,
) (
	ch chan *agency.ProtocolState,
	err error,
) {
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

func (pw Pairwise) ProposeProof(
	ctx context.Context,
	proofAttrs string,
) (
	ch chan *agency.ProtocolState,
	err error,
) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_PRESENT_PROOF,
		Role:         agency.Protocol_ADDRESSEE,
		StartMsg: &agency.Protocol_PresentProof{
			PresentProof: &agency.Protocol_PresentProofMsg{
				AttrFmt: &agency.Protocol_PresentProofMsg_AttributesJSON{
					AttributesJSON: proofAttrs}}},
	}
	return pw.Conn.doRun(ctx, protocol)
}

func (pw Pairwise) ProposeProofWithAttrs(
	ctx context.Context,
	proofAttrs *agency.Protocol_Proof,
) (
	ch chan *agency.ProtocolState,
	err error,
) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_PRESENT_PROOF,
		Role:         agency.Protocol_ADDRESSEE,
		StartMsg: &agency.Protocol_PresentProof{
			PresentProof: &agency.Protocol_PresentProofMsg{
				AttrFmt: &agency.Protocol_PresentProofMsg_Attributes{
					Attributes: proofAttrs}}},
	}
	return pw.Conn.doRun(ctx, protocol)
}

// ListenAndRetry listens both status notifications and agent questions. Client
// is specified by clientID. Both notifications and questions arriwe from same
// channel which is Question type. NOTE! This function handles connection errors
// by itself i.e. it tries to reopen error connnections until its terminated by
// ctx.Cancel. The actual retry logic is implemented in ListenStatusAndRetry and
// WaitAndRetry functions.
func (conn Conn) ListenAndRetry(
	ctx context.Context,
	client *agency.ClientID,
	cOpts ...grpc.CallOption,
) (
	ch chan *agency.Question,
) {
	listenStatusCh := conn.ListenStatusAndRetry(ctx, client, cOpts...)
	waitQuestionCh := conn.WaitAndRetry(ctx, client, cOpts...)
	glog.V(3).Infoln("successful start of general listen id:", client.ID)
	ch = make(chan *agency.Question)

	go merge(listenStatusCh, waitQuestionCh, ch)

	return ch
}

// Listen listens both status notifications and agent questions. Client is
// specified by clientID. Both notifications and questions arriwe from same
// channel which is Question type.
func (conn Conn) Listen(
	ctx context.Context,
	client *agency.ClientID,
	cOpts ...grpc.CallOption,
) (
	ch chan *agency.Question,
	err error,
) {
	defer err2.Return(&err)

	listenStatusCh, err := conn.ListenStatus(ctx, client, cOpts...)
	err2.Check(err)
	waitQuestionCh, err := conn.Wait(ctx, client, cOpts...)
	err2.Check(err)
	glog.V(3).Infoln("successful start of general listen id:", client.ID)
	ch = make(chan *agency.Question)

	go merge(listenStatusCh, waitQuestionCh, ch)

	return ch, nil
}

func merge(
	listenStatusCh <-chan *agency.AgentStatus,
	waitQuestionCh <-chan *agency.Question,
	ch chan<- *agency.Question,
) {
	defer close(ch)
	exitCount := 0
loop:
	for {
		select {
		case status, ok := <-listenStatusCh:
			if !ok {
				if exitCount == 1 {
					break loop
				} else {
					exitCount++
					continue loop
				}
			}
			q := &agency.Question{
				Status: status,
			}
			ch <- q
		case question, ok := <-waitQuestionCh:
			if !ok {
				if exitCount == 1 {
					break loop
				} else {
					exitCount++
					continue loop
				}
			}
			ch <- question
		}
	}
	glog.V(3).Infoln("general listen return")
}

// ListenStatus listens agent notification statuses. Client connection is
// identifed by ClientID. NOTE! The function filters KEEPALIVE messages.
func (conn Conn) ListenStatus(
	ctx context.Context,
	client *agency.ClientID,
	cOpts ...grpc.CallOption,
) (
	ch chan *agency.AgentStatus,
	err error,
) {
	defer err2.Return(&err)

	c := agency.NewAgentServiceClient(conn)
	statusCh := make(chan *agency.AgentStatus)

	stream, err := c.Listen(ctx, client, cOpts...)
	err2.Check(err)
	glog.V(3).Infoln("successful start of ListenStatus id:", client.ID)
	go transportStatus(stream, statusCh, nil)
	return statusCh, nil
}

// ListenStatusAndRetry listens agent notification statuses. Client connection
// is identifed by ClientID. NOTE! This function handles connection errors by
// itself i.e. it tries to reopen error connnections until its terminated by
// ctx.Cancel. NOTE! The function filters KEEPALIVE messages.
func (conn Conn) ListenStatusAndRetry( // nolint:dupl
	ctx context.Context,
	client *agency.ClientID,
	cOpts ...grpc.CallOption,
) (
	sch chan *agency.AgentStatus,
) {
	sch = make(chan *agency.AgentStatus)
	go func() {
		// catch all because worker
		defer err2.CatchTrace(func(err error) {
			glog.Warning(err)
		})

		sleeper := sleeper.New(sleeperFloor)
		var statusCh chan *agency.AgentStatus
		var errCh chan error
		var err error

	loop:
		statusCh, errCh, err = conn.ListenStatusErr(ctx, client, cOpts...)
		if err != nil {
			glog.V(1).Infoln("error:", err, "waiting...")
			sleeper.Sleep(time.Sleep)
			glog.V(1).Infoln("retry")
			goto loop
		}

		for {
			select {
			case <-ctx.Done():
				glog.V(1).Infoln("DONE called closing channel")
				close(sch)
				return
			case chErr := <-errCh:
				glog.V(1).Infoln("error:", chErr, "waiting ..")
				sleeper.Sleep(time.Sleep)
				glog.V(1).Infoln(".. retry")
				goto loop
			case status, ok := <-statusCh:
				if !ok {
					glog.V(1).Infoln("channel closed from other end")
					close(sch)
					return
				}
				sch <- status
			}
		}
	}()
	return sch
}

// ListenStatusErr listens agent notification statuses. It terminates on an
// error and transports it to the caller thru the error channel. Client
// connection is identifed by ClientID. NOTE! The function filters KEEPALIVE
// messages.
func (conn Conn) ListenStatusErr(
	ctx context.Context,
	client *agency.ClientID,
	cOpts ...grpc.CallOption,
) (
	ch chan *agency.AgentStatus,
	errCh chan error,
	err error,
) {
	defer err2.Return(&err)

	c := agency.NewAgentServiceClient(conn)
	statusCh := make(chan *agency.AgentStatus)
	errCh = make(chan error)

	stream, err := c.Listen(ctx, client, cOpts...)
	err2.Check(err)
	glog.V(3).Infoln("successful start of listenStatusErr id:", client.ID)

	go transportStatus(stream, statusCh, errCh)
	return statusCh, errCh, nil
}

// WaitAndRetry listens agent notification questions. Client connection is
// identifed by ClientID. NOTE! This function handles connection errors by
// itself i.e. it tries to reopen error connnections until its terminated by
// ctx.Cancel. NOTE! The function filters KEEPALIVE messages.
func (conn Conn) WaitAndRetry( // nolint:dupl
	ctx context.Context,
	client *agency.ClientID,
	cOpts ...grpc.CallOption,
) (
	ch chan *agency.Question,
) {
	ch = make(chan *agency.Question)
	go func() {
		// Catch all, panics as well because this is worker
		defer err2.CatchTrace(func(err error) {
			glog.Warning(err)
		})

		sleeper := sleeper.New(sleeperFloor)
		var questionCh chan *agency.Question
		var errCh chan error
		var err error

	loop:
		questionCh, errCh, err = conn.WaitErr(ctx, client, cOpts...)
		if err != nil {
			glog.V(1).Infoln("error:", err, "waiting...")
			sleeper.Sleep(time.Sleep)
			glog.V(1).Infoln("retry")
			goto loop
		}

		for {
			select {
			case <-ctx.Done():
				glog.V(1).Infoln("DONE called")
				close(ch)
				return
			case chErr := <-errCh:
				glog.V(1).Infoln("error:", chErr, "waiting...")
				sleeper.Sleep(time.Sleep)
				glog.V(1).Infoln("retry")
				goto loop
			case question, ok := <-questionCh:
				if !ok {
					glog.V(1).Infoln("closed from other end")
					close(ch)
					return
				}
				ch <- question
			}
		}
	}()
	return ch
}

// WaitErr listens agent notification questions. It terminates on an error and
// transports it to the caller thru an error channel. Client connection is
// identifed by ClientID. NOTE! The function filters KEEPALIVE messages.
func (conn Conn) WaitErr(
	ctx context.Context,
	client *agency.ClientID,
	cOpts ...grpc.CallOption,
) (
	ch chan *agency.Question,
	errCh chan error,
	err error,
) {
	defer err2.Return(&err)

	c := agency.NewAgentServiceClient(conn)
	statusCh := make(chan *agency.Question)
	errCh = make(chan error)

	stream, err := c.Wait(ctx, client, cOpts...)
	err2.Check(err)
	glog.V(3).Infoln("successful start of waitErr id:", client.ID)

	go transportWait(stream, statusCh, errCh)
	return statusCh, errCh, nil
}

// Wait listens agent notification questions. It terminates on error and closes
// the returned listening channel. Client connection is identifed by ClientID.
// NOTE! The function filters KEEPALIVE messages.
func (conn Conn) Wait(
	ctx context.Context,
	client *agency.ClientID,
	cOpts ...grpc.CallOption,
) (
	ch chan *agency.Question,
	err error,
) {
	defer err2.Return(&err)

	c := agency.NewAgentServiceClient(conn)
	statusCh := make(chan *agency.Question)

	stream, err := c.Wait(ctx, client, cOpts...)
	err2.Check(err)
	glog.V(3).Infoln("successful start of Wait id:", client.ID)

	go transportWait(stream, statusCh, nil /* errCh */)
	return statusCh, nil
}

func transportStatus(
	stream agency.AgentService_ListenClient,
	statusCh chan<- *agency.AgentStatus,
	errCh chan<- error,
) {
	defer err2.CatchTrace(func(err error) {
		glog.V(1).Infoln("WARNING: error when reading response:", err)
		if errCh != nil {
			errCh <- err
		} else {
			close(statusCh)
		}
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
}

func transportWait(
	stream agency.AgentService_WaitClient,
	statusCh chan<- *agency.Question,
	errCh chan<- error,
) {
	defer err2.CatchTrace(func(err error) {
		glog.V(1).Infoln("WARNING: error when reading response:", err)
		if errCh != nil {
			errCh <- err
		} else {
			close(statusCh)
		}
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
