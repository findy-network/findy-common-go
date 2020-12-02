package async

import (
	"context"

	"github.com/findy-network/findy-agent-api/grpc/agency"
	"github.com/findy-network/findy-grpc/agency/client"
)

type Pairwise struct {
	client.Pairwise
}

func NewPairwise(conn client.Conn, ID string) *Pairwise {
	return &Pairwise{Pairwise: client.Pairwise{Conn: conn, ID: ID}}
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
