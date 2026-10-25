"""Tiny HTTP listener so the worker can run as a free Render web service.

Render only keeps a web service alive while it answers HTTP on $PORT, and wakes a sleeping
one on an inbound request. The API pings this endpoint when it enqueues a job.
"""
from __future__ import annotations

import threading
from http.server import BaseHTTPRequestHandler, HTTPServer


class _Handler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:  # noqa: N802
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"ok")

    do_HEAD = do_GET

    def log_message(self, *args: object) -> None:
        pass


def start_health_server(port: int) -> None:
    server = HTTPServer(("0.0.0.0", port), _Handler)
    threading.Thread(target=server.serve_forever, daemon=True).start()
