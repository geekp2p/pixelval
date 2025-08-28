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
function connect() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(`${proto}://${location.host}/ws`);
  ws.onopen = () => { status.textContent = 'Connected'; btn.disabled = false; };
  ws.onclose = () => { status.textContent = 'Disconnected — retrying...'; btn.disabled = true; setTimeout(connect, 1500); };
  ws.onerror = () => { status.textContent = 'Error'; };
  ws.onmessage = ev => { try { addMsg(JSON.parse(ev.data)); } catch {} };
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
