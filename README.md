# Go Code Runner

A lightweight HTTP service that compiles and runs Go code. Designed to be deployed on Render (free tier).

## API

### POST /run

```json
{
  "code": "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"hello\")}",
  "language": "go",
  "stdin": ""
}
```

Response:
```json
{
  "stdout": "hello\n",
  "stderr": "",
  "exit_code": 0,
  "compile_output": ""
}
```

### GET /health

Returns `{"status":"ok"}`.

## Limits

- Code: 64KB max
- Output: 64KB max  
- Timeout: 10 seconds

## Deploy to Render

1. Push this repo to GitHub
2. Go to [render.com](https://render.com) → New → Web Service
3. Connect your repo, set **Root Directory** to `runner`
4. Runtime: **Docker**
5. Deploy

After deploy, set `RUNNER_URL=https://your-runner.onrender.com` in your main app's environment variables.
