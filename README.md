# Recoding — Code Runner

A lightweight Go HTTP service that compiles and executes participant code for the [Recoding Exercise Platform](https://github.com/jvcByte/recoding).

Live at: **https://code-runner-cpqt.onrender.com**

---

## What it does

When a participant clicks **Run** in the coding editor, the platform sends their Go code to this service. The runner:

1. Validates the API key
2. Writes the code to a temporary directory
3. Runs `go run` with a 10-second timeout
4. Captures stdout, stderr, and exit code
5. Returns the result as JSON
6. Cleans up the temp directory

The banner file (`standard.txt`) is automatically injected as stdin by the main app for ASCII art exercises — participants read it via `os.Stdin` without needing file I/O.

---

## API

### `POST /run`

Execute Go code.

**Headers:**
```
Authorization: Bearer <RUNNER_API_KEY>
Content-Type: application/json
```

**Request:**
```json
{
  "code": "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"hello\")}",
  "language": "go",
  "stdin": ""
}
```

**Response:**
```json
{
  "stdout": "hello\n",
  "stderr": "",
  "exit_code": 0,
  "compile_output": ""
}
```

### `GET /health`

Returns `{"status":"ok"}`. Used by Render for health checks.

### `GET /`

Status page showing the server is online with endpoint documentation.

---

## Limits

| Limit | Value |
|-------|-------|
| Code size | 64 KB |
| Output size | 64 KB |
| Execution timeout | 10 seconds |
| Supported languages | Go only |

---

## Security

All requests to `/run` require a Bearer token:

```
Authorization: Bearer <RUNNER_API_KEY>
```

Set `RUNNER_API_KEY` as an environment variable on Render. The main app reads the same key from its `RUNNER_API_KEY` env var and sends it automatically.

Generate a key:
```bash
openssl rand -base64 32
```

---

## Deploy to Render

1. Go to [render.com](https://render.com) → New → Web Service
2. Connect `github.com/jvcByte/code-runner`
3. Runtime: **Docker**
4. Add environment variable: `RUNNER_API_KEY=<your-secret>`
5. Deploy

After deploy, set in the main app:
```env
RUNNER_URL=https://your-service.onrender.com
RUNNER_API_KEY=<same-secret>
```

---

## Local development

```bash
# Build and run
go build -o runner . && ./runner
# Server starts on :3001

# Test
curl -X POST http://localhost:3001/run \
  -H "Content-Type: application/json" \
  -d '{"code":"package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"hello\")}","language":"go","stdin":""}'
```

---

## Project structure

```
main.go                      — Entry point, server setup, route registration
internal/
  executor/
    executor.go              — Runs go code in temp dir with timeout and output limits
  handlers/
    run.go                   — POST /run — validates request, calls executor, returns JSON
    health.go                — GET /health — returns {"status":"ok"}
    index.go                 — GET / — HTML status page with server rack icon
  middleware/
    auth.go                  — Bearer token authentication (skipped if RUNNER_API_KEY unset)
    logger.go                — Structured JSON request logging (IP, agent, status, duration)
```

---

## Logging

All requests are logged as structured JSON to stdout (visible in Render's log viewer):

```json
{"time":"2026-04-14T10:00:00Z","level":"INFO","msg":"request","method":"POST","path":"/run","status":200,"ip":"1.2.3.4","agent":"...","duration":"523ms"}
```

Log levels:
- `INFO` — successful requests (2xx)
- `WARN` — client errors (4xx) and execution timeouts
- `ERROR` — server errors (5xx) and execution failures

---

## Notes

- Render free tier sleeps after 15 minutes of inactivity — first request after idle will be slow (~30s cold start)
- No sandboxing beyond the 10s timeout and output limits — code runs as the server process user
- Only Go is supported; other languages return an error response
