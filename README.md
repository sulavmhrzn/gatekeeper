# Gatekeeper 🛡️
A lightweight, high-performance API Gateway and Reverse Proxy built in Go.

## What it does
Gatekeeper sits in front of your web applications (like Node.js, Python, or Go backends) and acts as a security and routing layer. It ensures that only authorized requests reach your actual servers.

## Features
- Security: Rejects any request that doesn't provide a valid X-Gatekeeper-Key.
- Intelligent Routing: Automatically routes traffic to different backends based on the URL path.
- Path Stripping: Cleans prefixes (like /api) so backends receive clean, simple URLs.
- Crash Protection: Uses recovery middleware to stay online even if a specific request causes a panic.
- Performance Tracking: Detailed logging of request duration and HTTP status codes.

## How to Run
1. Configure your routes in config.yaml:

```yaml
server:
  port: 8000
  secret_key: "my-secret"
routes:
  - path: "/api"
    target: "http://localhost:9000"
  - path: "/docs"
    target: "http://localhost:7000"
```
2. Start the Gateway:
```sh
go run main.go
```

3. Access your app:
```sh
# This will be blocked (401 Unauthorized)
curl http://localhost:8000/api/users

# This will pass through to your backend
curl -H "X-Gatekeeper-Key: my-secret" http://localhost:8000/api/users
```