import { RTCClient } from "./rtc-client.js";

const loginForm = document.querySelector("#login-form");
const callForm = document.querySelector("#call-form");
const status = document.querySelector("#status");
const peersList = document.querySelector("#peers");
const joinButton = document.querySelector("#join-button");
const muteButton = document.querySelector("#mute-button");
const leaveButton = document.querySelector("#leave-button");
let accessToken = "";
let client = null;
let muted = false;

function setStatus(text) { status.textContent = text; }

loginForm.addEventListener("submit", async event => {
  event.preventDefault();
  const data = Object.fromEntries(new FormData(loginForm));
  try {
    const response = await fetch("/api/v1/auth/login", {
      method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(data),
    });
    const body = await response.json();
    if (!response.ok) throw new Error(body.error?.message || "Не удалось войти");
    accessToken = body.access_token;
    await loadGroups();
    loginForm.hidden = true;
    callForm.hidden = false;
    setStatus("Вход выполнен. Можно подключаться к звонку.");
  } catch (error) { setStatus(error.message); }
});

async function loadGroups() {
  const response = await fetch("/api/v1/groups", { headers: { Authorization: `Bearer ${accessToken}` } });
  const groups = await response.json();
  if (!response.ok) throw new Error(groups.error?.message || "Не удалось загрузить список групп");
  const select = callForm.elements.group_id;
  select.replaceChildren();
  for (const group of groups) {
    const option = document.createElement("option");
    option.value = group.id;
    option.textContent = `${group.name} (ID: ${group.id})`;
    select.append(option);
  }
  if (groups.length === 0) throw new Error("У аккаунта пока нет групп для звонка");
}

callForm.addEventListener("submit", async event => {
  event.preventDefault();
  joinButton.disabled = true;
  setStatus("Запрашиваем доступ к микрофону…");
  peersList.replaceChildren();
  try {
    client = new RTCClient(accessToken);
    client.addEventListener("status", event => setStatus(event.detail));
    client.addEventListener("error", event => setStatus(event.detail));
    client.addEventListener("closed", () => {
      client = null;
      resetCallUI("Соединение закрыто");
    });
    client.addEventListener("joined", event => {
      muteButton.disabled = false;
      leaveButton.disabled = false;
      for (const id of event.detail) addPeer(id, "Соединяемся");
    });
    client.addEventListener("peer", event => {
      const { action, userID } = event.detail;
      if (action === "left") document.querySelector(`#peer-${userID}`)?.remove();
      else addPeer(userID, action === "connected" ? "На связи" : action === "connecting" ? "Соединяемся" : "Подключился");
    });
    await client.join(new FormData(callForm).get("group_id"));
    leaveButton.disabled = false;
  } catch (error) {
    setStatus(error.message);
    client?.leave();
    client = null;
    joinButton.disabled = false;
  }
});

function addPeer(id, label) {
  let item = document.querySelector(`#peer-${id}`);
  if (!item) {
    item = document.createElement("li");
    item.id = `peer-${id}`;
    peersList.append(item);
  }
  item.textContent = `Участник ${id} — ${label}`;
}

muteButton.addEventListener("click", () => {
  muted = !muted;
  client?.setMuted(muted);
  muteButton.textContent = muted ? "Включить микрофон" : "Выключить микрофон";
});

leaveButton.addEventListener("click", () => {
  client?.leave();
  client = null;
  resetCallUI("Вы вышли из звонка");
});

function resetCallUI(message) {
  joinButton.disabled = false;
  muteButton.disabled = true;
  leaveButton.disabled = true;
  muteButton.textContent = "Выключить микрофон";
  muted = false;
  peersList.replaceChildren();
  document.querySelector("#audio-container").replaceChildren();
  setStatus(message);
}
