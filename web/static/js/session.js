// --- small utility: debounce ---
function debounce(fn, delay) {
  let t;
  return function(...args) {
    const ctx = this;
    clearTimeout(t);
    t = setTimeout(() => fn.apply(ctx, args), delay);
  };
}

// 1) Create the terminal instance
const term = new Terminal({
  cursorBlink: true,
  convertEol: true,
  fontFamily:
    'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace',
  fontSize: 14,
  theme: {
    background: '#0d1226',
    foreground: '#e6e8ee',
    cursor: '#e6e8ee',
  },
  scrollback: 2000,
});

const fitAddon = new FitAddon.FitAddon();
term.loadAddon(fitAddon);
// 2) Mount it into the page
const container = document.getElementById('terminal');
if (!container) {
  console.error('Missing <div id="terminal"> in the HTML');
} else {
  term.open(container);
}

const token = window.NVIM_ANYWHERE?.token;
if (!token) {
  console.error('Missing session token');
}

const proto = location.protocol === 'https:' ? 'wss' : 'ws';
const ws = new WebSocket(
  `${proto}://${location.host}/pty?token=${encodeURIComponent(token || '')}`,
);
ws.binaryType = 'arraybuffer';

// Reuse encoders/decoders to avoid per-keystroke allocations
const enc = new TextEncoder();
const dec = new TextDecoder();

if (document.fonts?.ready) {
  document.fonts.ready.then(() => {
    fitAddon.fit();
    sendResize(term.cols, term.rows);
  });
} else {
  fitAddon.fit();
  sendResize(term.cols, term.rows);
}

// Send initial resize once the socket opens
ws.addEventListener(
  'open',
  () => {
    sendResize(term.cols, term.rows);
  },
  { once: true },
);


// Tiny banner (optional)
term.write('NvimAnywhere terminal\r\n');
term.write('---------------------------------------\r\n\r\n$ ');

// Keystrokes → server (binary)
term.onData((data) => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(enc.encode(data));
  }
});

// Server → terminal
ws.addEventListener('message', (ev) => {
  if (typeof ev.data === 'string') {
    try {
      const m = JSON.parse(ev.data);
      if (m && m.type === 'exit') {
        term.write('\r\n\x1b[31m[disconnected]\x1b[0m\r\n');
      }
    } catch {
      // ignore unknown control frames
    }
    return;
  }
  // binary PTY data
  const text = dec.decode(new Uint8Array(ev.data));
  term.write(text);
});

// Control: send resize (JSON text)
function sendResize(cols, rows) {
  if (typeof cols !== 'number' || typeof rows !== 'number') return;
  if (cols <= 0 || rows <= 0) return;
  if (ws.readyState === WebSocket.OPEN) {
    console.log(JSON.stringify({ type: 'resize', cols: cols, rows: rows }))
    ws.send(JSON.stringify({ type: 'resize', cols: cols, rows: rows }));
  }
}

// Debounced window resize → send one resize after layout settles
const sendResizeDebounced = debounce(() => {
  sendResize(term.cols, term.rows);
}, 100);

window.addEventListener('resize', sendResizeDebounced);

// When xterm recalculates its own size (e.g., container change)
term.onResize(({ cols, rows }) => sendResize(cols, rows));

// (Better) Observe container size changes too; debounce to avoid bursts
if ('ResizeObserver' in window && container) {
  const ro = new ResizeObserver(
    debounce(() => {
      // if you use xterm fit addon, call fit() here first
      sendResize(term.cols, term.rows);
    }, 80),
  );
  ro.observe(container);
}

