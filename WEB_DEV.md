# DEFLOT Web UI Development Guide

## Backend (Go)
The backend is ready. It serves the API and WebSockets at `localhost:8080`.

1. **Start Server**:
   ```bash
   go run main.go server
   ```
   *   Endpoint: `http://localhost:8080/api/health`
   *   WebSocket: `ws://localhost:8080/api/ws`

## Frontend (Your Mission)
You can now build the React frontend.

1.  **Initialize**:
    Create `internal/web/ui` using Vite (React).
2.  **Theme**:
    Use the "Void & Blood" aesthetic specified in `web_ui_design.md`.
3.  **Connection**:
    Connect to `ws://localhost:8080/api/ws` to receive log streams.

## Design Assets
*   **Colors**:
    *   Black: `#000000`
    *   Red: `#ff0000`
    *   Purple (Accent): `#8a2be2` (Burning Fire effect)
*   **Font**: `VT323` or `Share Tech Mono`
