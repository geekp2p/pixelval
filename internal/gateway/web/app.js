const logEl = document.getElementById('log');
const txt = document.getElementById('txt');
const btn = document.getElementById('send');
const status = document.getElementById('status');
const nick = document.getElementById('nick');
const room = document.getElementById('room');
const apply = document.getElementById('apply');

function addMsg({from, id, text, ts}) {
  const div = document.createElement('div');
  div.className = 'msg';
  const when = new Date((ts||Date.now()/1000)*1000).toLocaleString();
  const shortId = id ? ` (${id.slice(-8)})` : '';
  div.innerHTML = `<div class="meta">${from||'anon'}${shortId} • ${when}</div><div>${text}</div>`;
  logEl.appendChild(div);
  logEl.scrollTop = logEl.scrollHeight;
}

let ws;
let reconnectTimer;

function connect() {
  // 1) ปิดตัวเก่าถ้ายังค้างอยู่ (OPEN/CONNECTING)
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
    try { ws.onopen = ws.onclose = ws.onmessage = ws.onerror = null; } catch {}
    try { ws.close(); } catch {}
  }

  // 2) เคลียร์รีไทม์เมอร์ก่อนตั้งใหม่
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }

  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(`${proto}://${location.host}/ws`);

  ws.onopen = () => {
    status.textContent = 'Connected';
    btn.disabled = false;
  };

  ws.onmessage = ev => {
    try { addMsg(JSON.parse(ev.data)); } catch {}
  };

  ws.onerror = () => {
    status.textContent = 'Error';
  };

  ws.onclose = () => {
    status.textContent = 'Disconnected — retrying...';
    btn.disabled = true;
    // 3) ตั้ง reconnect แบบไม่ซ้อน
    reconnectTimer = setTimeout(() => {
      // เปิดใหม่เฉพาะกรณียังปิดอยู่จริง ๆ
      if (!ws || ws.readyState === WebSocket.CLOSED) connect();
    }, 1500);
  };
}

connect();

function loadConfig() {
  fetch('/config').then(r => r.json()).then(cfg => {
    nick.value = cfg.nick || '';
    room.value = cfg.room || '';
    const idText = cfg.id ? ` • ID: ${cfg.id.slice(-8)}` : '';
    status.textContent = `Room: ${cfg.room} • Nick: ${cfg.nick}${idText}`;
  }).catch(()=>{});
}
loadConfig();

function send() {
  const t = txt.value.trim();
  if (!t || !ws || ws.readyState !== 1) return;
  ws.send(t);
  txt.value = '';
}
btn.onclick = send;
txt.onkeydown = e => { if (e.key === 'Enter') send(); };

apply.onclick = () => {
  const body = {nick: nick.value.trim(), room: room.value.trim()};
  fetch('/config', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(body)})
    .then(() => { loadConfig(); logEl.innerHTML=''; });
};
