package handlers

import (
	"fmt"
	"net/http"
)

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Recoding — Code Runner</title>
  <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32' fill='none'%3E%3Crect x='4' y='4' width='24' height='10' rx='2' stroke='%23818cf8' stroke-width='2' fill='%230c0f14'/%3E%3Crect x='4' y='18' width='24' height='10' rx='2' stroke='%23818cf8' stroke-width='2' fill='%230c0f14'/%3E%3Ccircle cx='9' cy='9' r='1.5' fill='%2310b981'/%3E%3Ccircle cx='9' cy='23' r='1.5' fill='%2310b981'/%3E%3Cline x1='14' y1='9' x2='24' y2='9' stroke='%23475569' stroke-width='1.5' stroke-linecap='round'/%3E%3Cline x1='14' y1='23' x2='24' y2='23' stroke='%23475569' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E">
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: #050709;
      color: #e2e8f0;
      font-family: 'Inter', system-ui, sans-serif;
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
    }
    .card {
      text-align: center;
      padding: 3rem 2.5rem;
      border: 1px solid rgba(255,255,255,0.08);
      border-radius: 16px;
      background: #0c0f14;
      max-width: 420px;
      width: 90%;
      box-shadow: 0 0 60px rgba(99,102,241,0.08);
    }
    .logo {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 10px;
      margin-bottom: 1.5rem;
    }
    svg { display: block; }
    h1 { font-size: 1.1rem; font-weight: 700; color: #f1f5f9; letter-spacing: -0.02em; }
    .sub { font-size: 13px; color: #64748b; margin-top: 0.3rem; }
    .status {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      margin-top: 1.5rem;
      padding: 0.4rem 0.9rem;
      background: rgba(16,185,129,0.1);
      border: 1px solid rgba(16,185,129,0.25);
      border-radius: 20px;
      font-size: 12px;
      font-weight: 600;
      color: #10b981;
      letter-spacing: 0.04em;
      text-transform: uppercase;
    }
    .dot {
      width: 7px; height: 7px;
      border-radius: 50%;
      background: #10b981;
      box-shadow: 0 0 6px #10b981;
      animation: pulse 1.5s ease-in-out infinite;
    }
    @keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.4} }
    .endpoint {
      margin-top: 1.5rem;
      padding: 0.75rem 1rem;
      background: #111520;
      border-radius: 8px;
      font-family: monospace;
      font-size: 12px;
      color: #818cf8;
      text-align: left;
      line-height: 1.8;
    }
    .endpoint span { color: #64748b; }
  </style>
</head>
<body>
  <div class="card">
    <div class="logo">
      <svg width="28" height="28" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
        <polygon points="16,2 28,9 28,23 16,30 4,23 4,9"
          stroke="#818cf8" stroke-width="2" stroke-linejoin="round" fill="none"/>
      </svg>
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#475569" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <rect x="2" y="2" width="20" height="8" rx="2"/>
        <rect x="2" y="14" width="20" height="8" rx="2"/>
        <line x1="6" y1="6" x2="6.01" y2="6"/>
        <line x1="6" y1="18" x2="6.01" y2="18"/>
      </svg>
    </div>
    <h1>Recoding Code Runner</h1>
    <p class="sub">Lightweight Go execution service</p>
    <div class="status">
      <span class="dot"></span>
      Online
    </div>
    <div class="endpoint">
      <span>POST</span> /run<br>
      <span>GET &nbsp;</span> /health
    </div>
  </div>
</body>
</html>`

func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}
