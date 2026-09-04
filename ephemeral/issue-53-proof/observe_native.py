#!/usr/bin/env python3
"""Transparent proof observer: retain model metadata, never prompts or secrets."""
import datetime
import json
import os
import pathlib
import subprocess
import sys
import threading

name = pathlib.Path(sys.argv[0]).name
binary = json.loads(os.environ["ISSUE53_NATIVE_BINARIES"])[name]
destination = os.environ["ISSUE53_NATIVE_RECEIPTS"]


def record(values):
    entry = {"time": datetime.datetime.now(datetime.timezone.utc).isoformat(),
             "harness": name, "pid": os.getpid(), **values}
    descriptor = os.open(destination, os.O_WRONLY | os.O_CREAT | os.O_APPEND, 0o600)
    try:
        os.write(descriptor, (json.dumps(entry) + "\n").encode())
    finally:
        os.close(descriptor)


args = sys.argv[1:]
selection = {}
for index, value in enumerate(args[:-1]):
    if value in ("--model", "--effort", "--session-id", "--conversation"):
        selection[value.removeprefix("--")] = args[index + 1]
record({"event": "native_launch", "selection": selection})

if name != "codex" or "app-server" not in args:
    os.execv(binary, [binary, *args])

# Codex receives its model and effort over JSON-RPC rather than CLI flags.
# Forward every byte unchanged; save only an allowlist of selection metadata.
child = subprocess.Popen([binary, *args], stdin=subprocess.PIPE,
                         stdout=subprocess.PIPE, stderr=None)


def forward_input():
    try:
        for line in sys.stdin.buffer:
            try:
                message = json.loads(line)
                if message.get("method") in ("thread/start", "thread/resume", "turn/start"):
                    parameters = message.get("params", {})
                    selected = {key: parameters[key] for key in
                                ("model", "modelProvider", "effort", "threadId")
                                if key in parameters}
                    record({"event": "native_request", "method": message["method"],
                            "selection": selected})
            except (ValueError, TypeError):
                pass
            child.stdin.write(line)
            child.stdin.flush()
    finally:
        child.stdin.close()


threading.Thread(target=forward_input, daemon=True).start()
for line in child.stdout:
    sys.stdout.buffer.write(line)
    sys.stdout.buffer.flush()
sys.exit(child.wait())
