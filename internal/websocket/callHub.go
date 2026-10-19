package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/Fista6k/disClone/internal/domain"
	"github.com/Fista6k/disClone/internal/groups"
)

const MaxCallParticipants = 4

type CallHub struct {
	join         chan callJoin
	signal       chan callSignal
	leave        chan callLeave
	leaveAll     chan *Client
	removeMember chan GroupMembershipChange
	rooms        map[int64]map[int64]*Client
	groupService *groups.GroupService
}

type callJoin struct {
	client  *Client
	groupID int64
	result  chan error
}

type callSignal struct {
	client  *Client
	message WSMessage
	result  chan error
}

type callLeave struct {
	client  *Client
	groupID int64
}

func NewCallHub(groupService *groups.GroupService) *CallHub {
	return &CallHub{
		join:         make(chan callJoin),
		signal:       make(chan callSignal),
		leave:        make(chan callLeave),
		leaveAll:     make(chan *Client),
		removeMember: make(chan GroupMembershipChange),
		rooms:        make(map[int64]map[int64]*Client),
		groupService: groupService,
	}
}

func (h *CallHub) Run() {
	for {
		select {
		case request := <-h.join:
			h.handleJoin(request)
		case request := <-h.signal:
			h.handleSignal(request)
		case request := <-h.leave:
			h.leaveClient(request.client, request.groupID)
		case client := <-h.leaveAll:
			for groupID := range h.roomsForClient(client) {
				h.leaveClient(client, groupID)
			}
		case membership := <-h.removeMember:
			if room := h.rooms[membership.GroupID]; room != nil {
				if client := room[membership.UserID]; client != nil {
					h.leaveClient(client, membership.GroupID)
				}
			}
		}
	}
}

func (h *CallHub) Join(ctx context.Context, client *Client, groupID int64) error {
	isMember, err := h.groupService.IsMember(ctx, groupID, client.UserID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrNotGroupMember
	}

	result := make(chan error, 1)
	request := callJoin{client: client, groupID: groupID, result: result}
	select {
	case h.join <- request:
	case <-client.done:
		return errors.New("websocket connection is closed")
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-result:
		return err
	case <-client.done:
		return errors.New("websocket connection is closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *CallHub) Signal(ctx context.Context, client *Client, message WSMessage) error {
	if message.GroupID <= 0 || message.ToUserID <= 0 || message.ToUserID == client.UserID {
		return errors.New("invalid call signaling target")
	}
	if (message.Type == CallOfferType || message.Type == CallAnswerType) && message.SDP == "" {
		return errors.New("sdp is required")
	}
	if message.Type == CallICECandidateType && (len(message.Candidate) == 0 || !json.Valid(message.Candidate)) {
		return errors.New("valid candidate is required")
	}

	result := make(chan error, 1)
	request := callSignal{client: client, message: message, result: result}
	select {
	case h.signal <- request:
	case <-client.done:
		return errors.New("websocket connection is closed")
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-result:
		return err
	case <-client.done:
		return errors.New("websocket connection is closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *CallHub) Leave(client *Client, groupID int64) {
	if groupID <= 0 {
		client.sendError("group_id is required")
		return
	}
	select {
	case h.leave <- callLeave{client: client, groupID: groupID}:
	case <-client.done:
	}
}

func (h *CallHub) LeaveAll(client *Client) {
	h.leaveAll <- client
}

func (h *CallHub) RemoveClientFromGroup(userID, groupID int64) {
	h.removeMember <- GroupMembershipChange{UserID: userID, GroupID: groupID}
}

func (h *CallHub) handleJoin(request callJoin) {
	if request.client.IsClosed() {
		request.result <- errors.New("websocket connection is closed")
		return
	}

	room := h.rooms[request.groupID]
	if room == nil {
		room = make(map[int64]*Client)
		h.rooms[request.groupID] = room
	}
	if _, alreadyJoined := room[request.client.UserID]; !alreadyJoined && len(room) >= MaxCallParticipants {
		request.result <- errors.New("call has reached its participant limit")
		return
	}
	if oldClient := room[request.client.UserID]; oldClient != nil && oldClient != request.client {
		oldClient.Close()
	}

	peers := make([]int64, 0, len(room))
	for userID := range room {
		if userID != request.client.UserID {
			peers = append(peers, userID)
		}
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i] < peers[j] })

	wasJoined := room[request.client.UserID] == request.client
	room[request.client.UserID] = request.client
	if !wasJoined {
		for userID, peer := range room {
			if userID != request.client.UserID {
				peer.enqueueCallMessage(CallMessage{
					Type:    CallPeerJoinedType,
					GroupID: request.groupID,
					UserID:  request.client.UserID,
				})
			}
		}
	}
	request.client.enqueueCallMessage(CallMessage{Type: CallPeersType, GroupID: request.groupID, Peers: peers})
	request.result <- nil
}

func (h *CallHub) handleSignal(request callSignal) {
	message := request.message
	room := h.rooms[message.GroupID]
	if room == nil || room[request.client.UserID] != request.client {
		request.result <- errors.New("join the call before sending signaling messages")
		return
	}
	peer := room[message.ToUserID]
	if peer == nil || peer.IsClosed() {
		request.result <- errors.New("target user is not in this call")
		return
	}

	peer.enqueueCallMessage(CallMessage{
		Type:      message.Type,
		GroupID:   message.GroupID,
		UserID:    request.client.UserID,
		ToUserID:  message.ToUserID,
		SDP:       message.SDP,
		Candidate: message.Candidate,
	})
	request.result <- nil
}

func (h *CallHub) leaveClient(client *Client, groupID int64) {
	room := h.rooms[groupID]
	if room == nil || room[client.UserID] != client {
		return
	}
	delete(room, client.UserID)
	for _, peer := range room {
		peer.enqueueCallMessage(CallMessage{Type: CallPeerLeftType, GroupID: groupID, UserID: client.UserID})
	}
	if len(room) == 0 {
		delete(h.rooms, groupID)
	}
}

func (h *CallHub) roomsForClient(client *Client) map[int64]struct{} {
	groupIDs := make(map[int64]struct{})
	for groupID, room := range h.rooms {
		if room[client.UserID] == client {
			groupIDs[groupID] = struct{}{}
		}
	}
	return groupIDs
}
