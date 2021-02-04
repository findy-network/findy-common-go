package async

import (
	"context"
	"encoding/json"

	"github.com/findy-network/findy-agent-api/grpc/agency"
	didexchange "github.com/findy-network/findy-agent/std/didexchange/invitation"
	"github.com/findy-network/findy-grpc/agency/client"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
)

type Pairwise struct {
	client.Pairwise
	cOpts []grpc.CallOption
}

func NewPairwise(conn client.Conn, ID string, cOpts ...grpc.CallOption) *Pairwise {
	return &Pairwise{Pairwise: client.Pairwise{Conn: conn, ID: ID}, cOpts: cOpts}
}

func (pw Pairwise) BasicMessage(ctx context.Context, content string) (pid *agency.ProtocolID, err error) {
	protocol := &agency.Protocol{
		ConnectionId: pw.ID,
		TypeId:       agency.Protocol_BASIC_MESSAGE,
		Role:         agency.Protocol_INITIATOR,
		StartMsg:     &agency.Protocol_BasicMessage{BasicMessage: content},
	}
	return pw.Conn.DoStart(ctx, protocol)
}

func (pw Pairwise) Issue(ctx context.Context, credDefID, attrsJSON string) (pid *agency.ProtocolID, err error) {
	protocol := &agency.Protocol{
		ConnectionId: pw.ID,
		TypeId:       agency.Protocol_ISSUE,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_CredDef{CredDef: &agency.Protocol_Issuing{
			CredDefId: credDefID,
			Attrs:     &agency.Protocol_Issuing_AttributesJson{AttributesJson: attrsJSON},
		}},
	}
	return pw.Conn.DoStart(ctx, protocol)
}

func (pw Pairwise) IssueWithAttrs(ctx context.Context, credDefID string, attrs *agency.Protocol_Attrs) (pid *agency.ProtocolID, err error) {
	protocol := &agency.Protocol{
		ConnectionId: pw.ID,
		TypeId:       agency.Protocol_ISSUE,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_CredDef{CredDef: &agency.Protocol_Issuing{
			CredDefId: credDefID,
			Attrs:     &agency.Protocol_Issuing_Attrs_{Attrs_: attrs},
		}},
	}
	return pw.Conn.DoStart(ctx, protocol)
}

func (pw Pairwise) ReqProof(ctx context.Context, proofAttrs string) (pid *agency.ProtocolID, err error) {
	protocol := &agency.Protocol{
		ConnectionId: pw.ID,
		TypeId:       agency.Protocol_PROOF,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_ProofReq{
			ProofReq: &agency.Protocol_ProofRequest{
				AttrFmt: &agency.Protocol_ProofRequest_AttributesJson{
					AttributesJson: proofAttrs}}},
	}
	return pw.Conn.DoStart(ctx, protocol)
}

func (pw Pairwise) ReqProofWithAttrs(ctx context.Context, proofAttrs *agency.Protocol_Proof) (pid *agency.ProtocolID, err error) {
	protocol := &agency.Protocol{
		ConnectionId: pw.ID,
		TypeId:       agency.Protocol_PROOF,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_ProofReq{
			ProofReq: &agency.Protocol_ProofRequest{
				AttrFmt: &agency.Protocol_ProofRequest_Attrs{
					Attrs: proofAttrs}}},
	}
	return pw.Conn.DoStart(ctx, protocol)
}

func (pw *Pairwise) Connection(ctx context.Context, invitationJSON string) (pid *agency.ProtocolID, err error) {
	defer err2.Return(&err)

	// assert that invitation is OK, and we need to return the connection ID
	// because it's the task id as well
	var invitation didexchange.Invitation
	err2.Check(json.Unmarshal([]byte(invitationJSON), &invitation))

	protocol := &agency.Protocol{
		TypeId: agency.Protocol_CONNECT,
		Role:   agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_ConnAttr{ConnAttr: &agency.Protocol_Connection{
			Label:          pw.Label,
			InvitationJson: invitationJSON,
		}},
	}
	pid, err = pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
	err2.Check(err)
	pw.ID = invitation.ID
	return pid, err
}

func (pw *Pairwise) Resume(
	ctx context.Context,
	id string,
	protocol agency.Protocol_Type,
	protocolState agency.ProtocolState_State,
) (pid *agency.ProtocolID, err error) {
	state := &agency.ProtocolState{
		ProtocolId: &agency.ProtocolID{
			TypeId: protocol,
			Role:   agency.Protocol_RESUME,
			Id:     id,
		},
		State: protocolState,
	}
	return pw.Conn.DoResume(ctx, state, pw.cOpts...)
}

func (pw *Pairwise) Release(ctx context.Context, id string) (pid *agency.ProtocolID, err error) {
	protocolID := &agency.ProtocolID{
		Id: id,
	}
	return pw.Conn.DoRelease(ctx, protocolID, pw.cOpts...)
}

func (pw *Pairwise) Status(ctx context.Context, id string) (pid *agency.ProtocolStatus, err error) {
	protocolID := &agency.ProtocolID{
		Id: id,
	}
	return pw.Conn.DoStatus(ctx, protocolID, pw.cOpts...)
}
