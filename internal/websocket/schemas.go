package websocket

import "encoding/json"

const (
	CallJoinType         = "call_join"
	CallLeaveType        = "call_leave"
	CallOfferType        = "call_offer"
	CallAnswerType       = "call_answer"
	CallICECandidateType = "call_ice_candidate"
	CallPeersType        = "call_peers"
	CallPeerJoinedType   = "call_peer_joined"
	CallPeerLeftType     = "call_peer_left"
)

type WSMessage struct {
	Type        string          `json:"type"`
	Content     string          `json:"content"`
	RecipientID int64           `json:"recipient_id,omitempty"`
	GroupID     int64           `json:"group_id,omitempty"`
	ToUserID    int64           `json:"to_user_id,omitempty"`
	SDP         string          `json:"sdp,omitempty"`
	Candidate   json.RawMessage `json:"candidate,omitempty"`
}

type CallMessage struct {
	Type      string          `json:"type"`
	GroupID   int64           `json:"group_id"`
	UserID    int64           `json:"user_id,omitempty"`
	ToUserID  int64           `json:"to_user_id,omitempty"`
	Peers     []int64         `json:"peers,omitempty"`
	SDP       string          `json:"sdp,omitempty"`
	Candidate json.RawMessage `json:"candidate,omitempty"`
}

type WSError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
