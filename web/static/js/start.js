
const help = [
  'nvim           Creates new nvim session',
  'nvim -r<Url>   Creates new nvim session and upload repo by url you provided',
];

// tiny id helper
function $(elid) {
  return document.getElementById(elid);
}

// cache DOM refs once
const source = $('source');   // <textarea id="source">
const view = $('view');     // <span id="view">
const output = $('terminal'); // <div id="terminal">

// guard (helps debug missing ids)
if (!source || !view || !output) {
  console.error('Missing #source, #view, or #terminal in the DOM');
}

// mirror textarea → span (safe)
function writing() {
  view.textContent = source.value; // white-space: pre-wrap in CSS if you want newlines
}

function handler(e) {
  if (e.key !== 'Enter') return;
  e.preventDefault();

  const cmd = source.value.trim();
  if (!cmd) return;

  const [bin, ...rest] = cmd.split(/\s+/);
  switch (bin) {
    case 'help': {
      for (const lineText of help) {
        const line = document.createElement('p');
        line.textContent = lineText;
        output.appendChild(line);
      }
      break;
    }
    case 'nvim': {
      start()
    }
    default: {
      const line = document.createElement('p');
      line.textContent = 'command not found';
      output.appendChild(line);
    }
  }

  // scroll latest into view, then reset input
  output.scrollTop = output.scrollHeight;
  source.value = '';
  source.focus();
  view.textContent = '';
}

// events
source.addEventListener('input', writing);
source.addEventListener('keydown', handler);

// optional: initial sync
writing();


function getUrl() {
  const cmd = source.value;
  const parts = cmd.trim().split(/\s+/);
  const i = parts.indexOf("-r");
  if (i === -1 || i === parts.length - 1) return null;
  const candidate = parts[i + 1]
  return isValidUrl(candidate) ? candidate : null;
}

function isValidUrl(s) {
  try { new URL(s); return true; } catch { return false; }
}

function getBody() {
  const url = getUrl();
  return url ? { Repo: url } : {};
}

async function start() {
  try {
    const body = getBody();
    const res = await fetch('/sessions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
      body: JSON.stringify(body)
    });
    if (!res.ok) {
      const text = await res.text();
      console.log(text);
      return;
    }
    const data = await res.json();
    window.location = window.location + data.endpoint;
  } catch (err) {
    console.error(err);
  }
}
