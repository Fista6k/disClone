# Group call signaling protocol

The first voice-call version uses a peer-to-peer mesh. The backend does not receive or forward media; it validates group membership and relays signaling over the authenticated WebSocket at `/api/v1/ws`.

Each WebSocket JSON message has a `type`. A call also has a `group_id`.

## Join and leave

Join a group call:

```json
{"type":"call_join","group_id":42}
```

The server rejects users who are not members of the group and calls with more than four participants. The joining user receives `call_peers` with the other participant IDs. Existing participants receive `call_peer_joined`.

```json
{"type":"call_peers","group_id":42,"peers":[7,19]}
{"type":"call_peer_joined","group_id":42,"user_id":23}
```

Leave explicitly with `{"type":"call_leave","group_id":42}`. Disconnecting the WebSocket leaves all active calls. A participant removed from the group is removed from its call.

The client that just joined initiates an offer to each ID in `call_peers`; existing participants wait for those offers. This avoids both peers creating an offer at once.

## Offer, answer, and ICE

All signaling messages must target another participant in the same active call. The server adds `user_id` to the relayed message.

```json
{"type":"call_offer","group_id":42,"to_user_id":7,"sdp":"<session description>"}
{"type":"call_answer","group_id":42,"to_user_id":23,"sdp":"<session description>"}
{"type":"call_ice_candidate","group_id":42,"to_user_id":23,"candidate":{"candidate":"<candidate>","sdpMid":"0","sdpMLineIndex":0}}
```

The recipient receives the same message with `user_id` set to the sender. Clients should buffer incoming ICE candidates until the corresponding remote description has been applied.

The server broadcasts `call_peer_left` when a participant leaves or disconnects:

```json
{"type":"call_peer_left","group_id":42,"user_id":7}
```

This is signaling only. Clients still need to create and manage browser `RTCPeerConnection` objects, media tracks, permissions, and ICE configuration. For connections across restrictive NATs, configure STUN/TURN servers on the clients.
