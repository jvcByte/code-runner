# Recoding — Code Runner

A lightweight Go HTTP service that compiles and executes participant code for the [Recoding Exercise Platform](https://github.com/jvcByte/recoding).

---

## What it does

When a participant clicks **Run** in the coding editor, the platform sends their Go code to this service. The runner:

1. Writes the code to a temporary directory
2. Runs `go run` with a 10-second timeout
3. Captures stdout, stderr, and exit code
4. Returns the result as JSON
5. Cleans up the temp directory

The banner file (`standard.txt`) is automatically injected as stdin by the main app for ASCII art exercises.

---

## API

### `POST /run`

Execute Go code.

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

Status page showing the server is online.

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

Set `RUNNER_API_KEY` to a random secret. All requests to `/run` must include:

```
Authorization: Bearer <your-key>
```

The main app reads `RUNNER_API_KEY` from its environment and sends it automatically.

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

After deploy, set `RUNNER_URL=https://your-service.onrender.com` and `RUNNER_API_KEY=<same-secret>` in the main app's environment.

---

## Local development

```bash
go run .
# Server starts on :3001

# Test it
curl -X POST http://localhost:3001/run \
  -H "Content-Type: application/json" \
  -d '{"code":"package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"hello\")}","language":"go","stdin":""}'
```

---

## Project structure

```
main.go                      — Entry point, server setup
internal/
  executor/executor.go       — Runs go code in temp dir with timeout
  handlers/
    run.go                   — POST /run handler
    health.go                — GET /health handler
    index.go                 — GET / status page
  middleware/
    auth.go                  — Bearer token authentication
    logger.go                — Structured JSON request logging
```

---

## Logging

All requests are logged as structured JSON to stdout:

```json
{"time":"...","level":"INFO","msg":"request","method":"POST","path":"/run","status":200,"ip":"1.2.3.4","agent":"...","duration":"523ms"}
```

Log levels: `INFO` (2xx), `WARN` (4xx + timeouts), `ERROR` (5xx + execution errors).
