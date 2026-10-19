const ICE_SERVERS = [{ urls: "stun:stun.l.google.com:19302" }];

export class RTCClient extends EventTarget {
  constructor(accessToken) {
    super();
    this.accessToken = accessToken;
    this.groupID = 0;
    this.socket = null;
    this.localStream = null;
    this.peers = new Map();
    this.joined = false;
  }

  async join(groupID) {
    if (this.socket) throw new Error("Клиент уже подключён к звонку");
    this.groupID = Number(groupID);
    this.localStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
    try {
      const { ticket } = await this.requestTicket();
      const wsProtocol = location.protocol === "https:" ? "wss:" : "ws:";
      const socket = new WebSocket(`${wsProtocol}//${location.host}/api/v1/ws?ticket=${encodeURIComponent(ticket)}`);
      this.socket = socket;
      socket.addEventListener("open", () => {
        this.send({ type: "call_join", group_id: this.groupID });
        this.emit("status", "Подключено к сигнализации, ожидаем участников…");
      });
      socket.addEventListener("message", event => this.onMessage(event.data));
      socket.addEventListener("error", () => this.emit("error", "Ошибка WebSocket-соединения"));
      socket.addEventListener("close", () => {
        if (this.socket === socket) this.cleanup(false);
        this.emit("closed");
      });
    } catch (error) {
      this.stopLocalMedia();
      throw error;
    }
  }

  async requestTicket() {
    const response = await fetch("/api/v1/ws-ticket", {
      method: "POST",
      headers: { Authorization: `Bearer ${this.accessToken}` },
    });
    const body = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(body.error?.message || "Не удалось получить билет WebSocket");
    return body;
  }

  async onMessage(raw) {
    let message;
    try { message = JSON.parse(raw); } catch { this.emit("error", "Сервер прислал некорректное сообщение"); return; }
    try {
      switch (message.type) {
        case "call_peers":
          this.joined = true;
          this.emit("status", "Вы в звонке");
          this.emit("joined", message.peers || []);
          for (const userID of message.peers || []) await this.createOffer(userID);
          break;
        case "call_peer_joined":
          this.emit("peer", { action: "joined", userID: message.user_id });
          break;
        case "call_offer": await this.acceptOffer(message.user_id, message.sdp); break;
        case "call_answer": await this.acceptAnswer(message.user_id, message.sdp); break;
        case "call_ice_candidate": await this.acceptCandidate(message.user_id, message.candidate); break;
        case "call_peer_left": this.removePeer(message.user_id); break;
        case "error":
          this.emit("error", message.message || "Ошибка звонка");
          if (!this.joined) this.cleanup(true);
          break;
      }
    } catch (error) {
      this.emit("error", error.message || "Ошибка установки WebRTC-соединения");
    }
  }

  async createOffer(userID) {
    const peer = this.getPeer(userID);
    const offer = await peer.pc.createOffer();
    await peer.pc.setLocalDescription(offer);
    this.send({ type: "call_offer", group_id: this.groupID, to_user_id: Number(userID), sdp: offer.sdp });
    peer.descriptionSent = true;
    this.flushLocalCandidates(peer);
  }

  async acceptOffer(userID, sdp) {
    const peer = this.getPeer(userID);
    await peer.pc.setRemoteDescription({ type: "offer", sdp });
    await this.flushRemoteCandidates(peer);
    const answer = await peer.pc.createAnswer();
    await peer.pc.setLocalDescription(answer);
    this.send({ type: "call_answer", group_id: this.groupID, to_user_id: Number(userID), sdp: answer.sdp });
    peer.descriptionSent = true;
    this.flushLocalCandidates(peer);
  }

  async acceptAnswer(userID, sdp) {
    const peer = this.getPeer(userID);
    await peer.pc.setRemoteDescription({ type: "answer", sdp });
    await this.flushRemoteCandidates(peer);
  }

  async acceptCandidate(userID, candidate) {
    const peer = this.getPeer(userID);
    if (peer.pc.remoteDescription) await peer.pc.addIceCandidate(candidate);
    else peer.remoteCandidates.push(candidate);
  }

  getPeer(userID) {
    const id = Number(userID);
    if (this.peers.has(id)) return this.peers.get(id);
    const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS });
    const peer = { pc, remoteCandidates: [], localCandidates: [], audio: null };
    this.peers.set(id, peer);
    for (const track of this.localStream.getTracks()) pc.addTrack(track, this.localStream);
    pc.addEventListener("icecandidate", event => {
      if (!event.candidate) return;
      if (!peer.descriptionSent) peer.localCandidates.push(event.candidate);
      else this.sendCandidate(id, event.candidate);
    });
    pc.addEventListener("track", event => {
      if (!peer.audio) {
        peer.audio = new Audio();
        peer.audio.autoplay = true;
        peer.audio.playsInline = true;
        peer.audio.srcObject = event.streams[0] || new MediaStream([event.track]);
        document.querySelector("#audio-container").append(peer.audio);
        peer.audio.play().catch(() => this.emit("status", "Нажмите на страницу, чтобы разрешить воспроизведение звука"));
      }
    });
    pc.addEventListener("connectionstatechange", () => {
      if (pc.connectionState === "connected") this.emit("peer", { action: "connected", userID: id });
      if (pc.connectionState === "failed") this.emit("error", `Не удалось соединиться с участником ${id}. Возможно, нужен TURN-сервер.`);
    });
    this.emit("peer", { action: "connecting", userID: id });
    return peer;
  }

  sendCandidate(userID, candidate) {
    this.send({ type: "call_ice_candidate", group_id: this.groupID, to_user_id: Number(userID), candidate: candidate.toJSON() });
  }

  flushLocalCandidates(peer) {
    for (const candidate of peer.localCandidates.splice(0)) this.sendCandidate([...this.peers].find(([, value]) => value === peer)[0], candidate);
  }

  async flushRemoteCandidates(peer) {
    for (const candidate of peer.remoteCandidates.splice(0)) await peer.pc.addIceCandidate(candidate);
  }

  setMuted(muted) {
    for (const track of this.localStream?.getAudioTracks() || []) track.enabled = !muted;
  }

  send(message) {
    if (this.socket?.readyState !== WebSocket.OPEN) throw new Error("WebSocket ещё не подключён");
    this.socket.send(JSON.stringify(message));
  }

  leave() {
    if (this.socket?.readyState === WebSocket.OPEN && this.joined) {
      try { this.send({ type: "call_leave", group_id: this.groupID }); } catch {}
    }
    this.cleanup(true);
  }

  cleanup(closeSocket) {
    const socket = this.socket;
    this.socket = null;
    this.joined = false;
    for (const id of [...this.peers.keys()]) this.removePeer(id);
    this.stopLocalMedia();
    if (closeSocket && socket && socket.readyState < WebSocket.CLOSING) socket.close();
  }

  stopLocalMedia() {
    for (const track of this.localStream?.getTracks() || []) track.stop();
    this.localStream = null;
  }

  removePeer(userID) {
    const id = Number(userID);
    const peer = this.peers.get(id);
    if (!peer) return;
    peer.pc.close();
    peer.audio?.remove();
    this.peers.delete(id);
    this.emit("peer", { action: "left", userID: id });
  }

  emit(type, detail) { this.dispatchEvent(new CustomEvent(type, { detail })); }
}
