package async

import (
	"context"

	"github.com/findy-network/findy-common-go/agency/client"
	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	didexchange "github.com/findy-network/findy-common-go/std/didexchange/invitation"
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
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_BASIC_MESSAGE,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_BasicMessage{
			BasicMessage: &agency.Protocol_BasicMessageMsg{
				Content: content,
			},
		},
	}
	return pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
}

func (pw Pairwise) Issue(ctx context.Context, credDefID, attrsJSON string) (pid *agency.ProtocolID, err error) {
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
	return pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
}

func (pw Pairwise) IssueWithAttrs(
	ctx context.Context,
	credDefID string,
	attrs *agency.Protocol_IssuingAttributes,
) (
	pid *agency.ProtocolID,
	err error,
) {
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
	return pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
}

func (pw Pairwise) ProposeIssue(ctx context.Context, credDefID, attrsJSON string) (pid *agency.ProtocolID, err error) {
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
	return pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
}

func (pw Pairwise) ProposeIssueWithAttrs(
	ctx context.Context,
	credDefID string,
	attrs *agency.Protocol_IssuingAttributes,
) (
	pid *agency.ProtocolID,
	err error,
) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_ISSUE_CREDENTIAL,
		Role:         agency.Protocol_ADDRESSEE,
		StartMsg: &agency.Protocol_IssueCredential{IssueCredential: &agency.Protocol_IssueCredentialMsg{
			CredDefID: credDefID,
			AttrFmt: &agency.Protocol_IssueCredentialMsg_Attributes{
				Attributes: attrs,
			},
		}},
	}
	return pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
}

func (pw Pairwise) ReqProof(ctx context.Context, proofAttrs string) (pid *agency.ProtocolID, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_PRESENT_PROOF,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_PresentProof{
			PresentProof: &agency.Protocol_PresentProofMsg{
				AttrFmt: &agency.Protocol_PresentProofMsg_AttributesJSON{
					AttributesJSON: proofAttrs}}},
	}
	return pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
}

func (pw Pairwise) ReqProofWithAttrs(ctx context.Context, proofAttrs *agency.Protocol_Proof) (pid *agency.ProtocolID, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_PRESENT_PROOF,
		Role:         agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_PresentProof{
			PresentProof: &agency.Protocol_PresentProofMsg{
				AttrFmt: &agency.Protocol_PresentProofMsg_Attributes{
					Attributes: proofAttrs}}},
	}
	return pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
}

func (pw Pairwise) ProposeProof(ctx context.Context, proofAttrs string) (pid *agency.ProtocolID, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_PRESENT_PROOF,
		Role:         agency.Protocol_ADDRESSEE,
		StartMsg: &agency.Protocol_PresentProof{
			PresentProof: &agency.Protocol_PresentProofMsg{
				AttrFmt: &agency.Protocol_PresentProofMsg_AttributesJSON{
					AttributesJSON: proofAttrs}}},
	}
	return pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
}

func (pw Pairwise) ProposeProofWithAttrs(ctx context.Context, proofAttrs *agency.Protocol_Proof) (pid *agency.ProtocolID, err error) {
	protocol := &agency.Protocol{
		ConnectionID: pw.ID,
		TypeID:       agency.Protocol_PRESENT_PROOF,
		Role:         agency.Protocol_ADDRESSEE,
		StartMsg: &agency.Protocol_PresentProof{
			PresentProof: &agency.Protocol_PresentProofMsg{
				AttrFmt: &agency.Protocol_PresentProofMsg_Attributes{
					Attributes: proofAttrs}}},
	}
	return pw.Conn.DoStart(ctx, protocol, pw.cOpts...)
}

// Connection is a helper wrapper to start a connection protocol in the agency.
// The invitationStr accepts both JSON and URL formated invitations. The agency
// does the same.
func (pw *Pairwise) Connection(ctx context.Context, invitationStr string) (pid *agency.ProtocolID, err error) {
	defer err2.Return(&err)

	// assert that invitation is OK, and we need to return the connection ID
	// because it's the task id as well
	invitation, err := didexchange.Translate(invitationStr)
	err2.Check(err)

	protocol := &agency.Protocol{
		TypeID: agency.Protocol_DIDEXCHANGE,
		Role:   agency.Protocol_INITIATOR,
		StartMsg: &agency.Protocol_DIDExchange{DIDExchange: &agency.Protocol_DIDExchangeMsg{
			Label:          pw.Label,
			InvitationJSON: invitationStr,
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
) (
	pid *agency.ProtocolID, err error,
) {
	state := &agency.ProtocolState{
		ProtocolID: &agency.ProtocolID{
			TypeID: protocol,
			Role:   agency.Protocol_RESUMER,
			ID:     id,
		},
		State: protocolState,
	}
	return pw.Conn.DoResume(ctx, state, pw.cOpts...)
}
