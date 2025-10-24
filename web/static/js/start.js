const form = document.getElementById('startForm');
const btn  = document.getElementById('startBtn');
const msg  = document.getElementById('msg');
const err  = document.getElementById('err');

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  err.textContent = ""; msg.textContent = "";
  btn.disabled = true; btn.textContent = "Starting…";

  const repo = document.getElementById('repo').value.trim();

  try {
    const res = await fetch('/sessions', {
      method: 'POST',
      headers: {'Content-Type':'application/json'},
      body: JSON.stringify({ repo })
    });
    if (!res.ok) {
      btn.disabled = false; btn.textContent = "Start session";
      const text = await res.text();
      err.textContent = res.status === 429
        ? "Too many starts. Please wait a moment."
        : (text || "Failed to start session.");
      return;
    }
    const data = await res.json();
    msg.textContent = "Attaching terminal…";
    window.location = data.url; // e.g. /s/<token>
  } catch {
    btn.disabled = false; btn.textContent = "Start session";
    err.textContent = "Network error. Please try again.";
  }
});

