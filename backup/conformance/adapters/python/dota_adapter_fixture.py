#!/usr/bin/env python3
"""Language-neutral dota adapter used by the conformance suite."""

import base64
import json
import sys


def read_frame():
    length = None
    while True:
        line = sys.stdin.buffer.readline()
        if not line:
            return None
        if line in (b"\n", b"\r\n"):
            break
        name, value = line.decode("ascii").split(":", 1)
        if name.lower() == "content-length":
            length = int(value.strip())
    if length is None or length < 0 or length > 32 * 1024 * 1024:
        raise ValueError("invalid Content-Length")
    return json.loads(sys.stdin.buffer.read(length))


def write_frame(message):
    payload = json.dumps(message, separators=(",", ":")).encode("utf-8")
    sys.stdout.buffer.write(f"Content-Length: {len(payload)}\r\n\r\n".encode("ascii"))
    sys.stdout.buffer.write(payload)
    sys.stdout.buffer.flush()


def result(request, value):
    return {"jsonrpc": "2.0", "id": request.get("id"), "result": value}


def operation(method):
    if method == "exportPlan":
        files = [{
            "path": ".fixture-adapter.txt",
            "encoding": "base64",
            "content": base64.b64encode(b"dota protocol fixture\n").decode("ascii"),
        }]
    elif method == "importPlan":
        files = [{
            "path": ".agents/extensions/org.open-dot-agents.fixture/source.txt",
            "encoding": "base64",
            "content": base64.b64encode(b"fixture provenance\n").decode("ascii"),
        }]
    else:
        return {"diagnostics": [], "losses": []}
    return {"diagnostics": [], "losses": [], "plan": {"files": files}}


def main():
    while True:
        request = read_frame()
        if request is None:
            return
        method = request.get("method")
        if method == "initialize":
            response = result(request, {"protocolVersion": "1.0"})
        elif method == "describe":
            response = result(request, {
                "id": "org.open-dot-agents.fixture",
                "name": "Python conformance adapter",
                "version": "1.0.0",
                "protocolVersion": "1.0",
                "target": "fixture",
                "capabilities": ["validate", "import", "export"],
                "categoryStatuses": {
                    "instructions": "mapped",
                    "rules": "unsupported",
                    "agents": "unsupported",
                    "skills": "unsupported",
                    "tools": "unsupported",
                },
                "inputPatterns": [".agents/**"],
                "maxSnapshotBytes": 16777216,
            })
        elif method in ("validate", "exportPlan", "importPlan"):
            response = result(request, operation(method))
        elif method == "shutdown":
            write_frame(result(request, None))
            return
        else:
            response = {
                "jsonrpc": "2.0",
                "id": request.get("id"),
                "error": {"code": -32601, "message": "method not found"},
            }
        write_frame(response)


if __name__ == "__main__":
    main()
