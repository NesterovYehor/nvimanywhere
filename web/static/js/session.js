// ============================================================
// Utils
// ============================================================

function debounce(fn, delay) {
  let t;
  return (...args) => {
    clearTimeout(t);
    t = setTimeout(() => fn(...args), delay);
  };
}

// ============================================================
// Terminal setup
// ============================================================

const term = new Terminal({
  cursorBlink: true,
  scrollback: 2000,

  eraseOnClear: true,
});

const fitAddon = new FitAddon.FitAddon();
term.loadAddon(fitAddon);

const container = document.getElementById('terminal');
if (!container) throw new Error('Missing #terminal');

term.open(container);
term.focus();

// ============================================================
// WebSocket
// ============================================================

const proto = location.protocol === 'https:' ? 'wss' : 'ws';
const ws = new WebSocket(`${proto}://${location.host}${location.pathname}`);
ws.binaryType = 'arraybuffer';

const enc = new TextEncoder();

// ============================================================
// Initial fit + resize
// ============================================================

function sendResize(cols, rows) {
  if (
    ws.readyState !== WebSocket.OPEN ||
    cols <= 0 ||
    rows <= 0
  ) {
    return;
  }

  ws.send(JSON.stringify({ type: 'resize', cols, rows }));
}

function fitAndResize() {
  fitAddon.fit();
  sendResize(term.cols, term.rows);
}

requestAnimationFrame(fitAndResize);

if (document.fonts?.ready) {
  document.fonts.ready.then(fitAndResize);
}

ws.addEventListener('open', fitAndResize, { once: true });

// ============================================================
// Input → server
// ============================================================

term.onData((data) => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(enc.encode(data));
  }
});

// ============================================================
// Server → terminal (RAW BYTES ONLY)
// ============================================================

ws.addEventListener('message', (ev) => {
  if (typeof ev.data === 'string') {
    try {
      const m = JSON.parse(ev.data);
      if (m?.type === 'exit') {
        term.write('\r\n\x1b[31m[disconnected]\x1b[0m\r\n');
      }
    } catch { }
    return;
  }

  term.write(new Uint8Array(ev.data));
});

// ============================================================
// Needed for signal to server that tab is closing so it can close session
// ============================================================
window.addEventListener("beforeunload", () => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: "disconnect" }))
  }
})


// ============================================================
// Resize handling
// ============================================================

const resizeDebounced = debounce(fitAndResize, 100);

window.addEventListener('resize', resizeDebounced);

if ('ResizeObserver' in window) {
  const ro = new ResizeObserver(resizeDebounced);
  ro.observe(container);
}

// ============================================================
// Cleanup
// ============================================================

ws.addEventListener('close', () => {
  window.location.href = '/';
});

