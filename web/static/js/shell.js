/* ============================================================
 * Command Metadata
 * ------------------------------------------------------------
 * Static command descriptions and command → handler mapping.
 * This is declarative configuration, not behavior.
 * ============================================================
 */

const help = {
  nvim: 'nvim  <Url>()     Creates new nvim session and upload repo by url if you provided it',
  clear: 'clear            Clear shell screen',
  help: 'help [cmd]        Show help',
};

const commands = {
  help: handleHelp,
  nvim: handleNvim,
  clear: clearHistory,
};


/* ============================================================
 * DOM Utilities & Element References
 * ------------------------------------------------------------
 * Low-level DOM access helpers and required UI elements.
 * All direct DOM lookups live here.
 * ============================================================
 */

function $(elid) {
  return document.getElementById(elid);
}

const source = $('source');
const view = $('view');
const history = $('history');

if (!source || !view || !history) {
  console.error('Missing #source, #view, or #terminal in the DOM');
}


/* ============================================================
 * Event Wiring
 * ------------------------------------------------------------
 * Keyboard and input event bindings.
 * These translate raw browser events into semantic actions.
 * ============================================================
 */

source.addEventListener('keydown', onKeyDown);
source.addEventListener('input', writing);

function onKeyDown(e) {
  if (e.key === "Enter") {
    e.preventDefault();
    handler();
  } else if (e.key === "l" && e.ctrlKey) {
    e.preventDefault();
    clearTerminal();
  } else if (e.key === "Escape") {
    e.preventDefault();
    clearCmdLine();
  }
}


/* ============================================================
 * Command Dispatch & Parsing
 * ------------------------------------------------------------
 * Takes raw input, parses it, resolves a command,
 * and dispatches to the appropriate handler.
 * ============================================================
 */

async function handler() {
  const input = source.value.trim();
  if (!input) return;

  await addLine(input, 'cmd', 200);
  clearCmdLine();

  const parsed = parseCmd(input);
  if (!parsed) return;

  const { name, fn, args } = parsed;

  if (fn) fn(args);
  else addLine(`Command ${name} is not found`, "color2", 30);
}

function parseCmd(input) {
  const parts = input.trim().split(/\s+/);
  if (parts.length === 0) return null;

  return {
    name: parts[0],
    fn: commands[parts[0]],
    args: parts[1] ?? null,
  };
}

function commitCmdLine(text) {
  const p = document.createElement('p');
  p.textContent = text;
  p.className = 'cmd';
  history.appendChild(p);
}


/* ============================================================
 * Built-in Command Handlers
 * ------------------------------------------------------------
 * High-level command behavior (help, clear, nvim).
 * Each handler represents a user-visible command.
 * ============================================================
 */

async function handleHelp(cmdName) {
  if (!cmdName) {
    for (let text of Object.values(help)) {
      await addLine(text, "color2", 500);
    };
  } else {
    await addLine(help[cmdName], "color2", 500);
  }
  clearCmdLine();
}


/* ============================================================
 * Async Step System (Generic Progress Mechanism)
 * ------------------------------------------------------------
 * This is the generic abstraction for sequential terminal steps.
 * Steps describe *when they complete*, not *how they are run*.
 * ============================================================
 */

function wait(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

function step(text, waitFor) {
  return { text, waitFor };
}

async function runSteps(steps) {
  let lastResult = null;

  for (const step of steps) {
    const line = await addLine(step.text, "color2 loading", 0);

    try {
      lastResult = await step.waitFor();
      line.classList.remove("loading");
      line.classList.add("success");
    } catch (err) {
      line.classList.remove("loading");
      line.classList.add("error");
      throw err;
    }
  }

  return lastResult;
}


/* ============================================================
 * Nvim Command (Domain Logic)
 * ------------------------------------------------------------
 * Orchestrates request startup and step execution.
 * Uses the generic step runner without special cases.
 * ============================================================
 */

async function handleNvim() {
  const body = getBody();
  const requestPromise = fetch('/sessions/new', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    },
    body: JSON.stringify(body)
  }).then(res => {
    if (!res.ok) throw new Error("request failed");
    return res.json();
  });

  const steps = [
    step('Starting session…', () => wait(700)),
    step('  Creating session…', () => wait(700)),
    step('  Preparing workspace…', () => wait(700)),
    step('  Preparing environment…', () => wait(700)),
    step('  Starting editor…', () => requestPromise),
  ];

  const data = await runSteps(steps);

  window.location = window.location + data.endpoint;
}


/* ============================================================
 * Input / View Synchronization
 * ------------------------------------------------------------
 * Keeps the mirrored command preview in sync with input.
 * ============================================================
 */

function writing() {
  view.textContent = source.value;
}


/* ============================================================
 * Request Body Construction
 * ------------------------------------------------------------
 * Extracts structured data (e.g. repo URL) from the command line.
 * ============================================================
 */

function getUrl() {
  const cmd = source.value;
  const parts = cmd.trim().split(/\s+/);
  const i = parts.indexOf("-r");
  if (i === -1 || i === parts.length - 1) return null;

  const candidate = parts[i + 1];
  return (candidate) ? candidate : null;
}

function isValidUrl(s) {
  try {
    new URL(s);
    return true;
  } catch {
    return false;
  }
}

function getBody() {
  const url = getUrl();
  return url ? { Repo: url } : {};
}


/* ============================================================
 * Terminal Rendering Utilities
 * ------------------------------------------------------------
 * Low-level helpers for mutating terminal state and UI.
 * These functions do not know about commands or steps.
 * ============================================================
 */

function clearHistory() {
  history.replaceChildren();
}

function clearCmdLine() {
  source.value = "";
  view.textContent = "";
}

function addLine(text, style, time) {
  let t = "";

  for (let i = 0; i < text.length; i++) {
    if (text.charAt(i) === " " && text.charAt(i + 1) === " ") {
      t += "&nbsp;&nbsp;";
      i++;
    } else {
      t += text.charAt(i);
    }
  }

  return new Promise((resolve) => {
    setTimeout(() => {
      const next = document.createElement("p");
      next.innerHTML = t;
      next.className = style;
      history.appendChild(next);
      window.scrollTo(0, document.body.offsetHeight);
      resolve(next);
    }, time);
  });
}

