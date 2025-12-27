#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
# Copyright Contributors to the OpenTimelineIO project
"""
Bridge for gotio to access Python OTIO adapters via stdin/stdout JSON.

This script runs as a long-lived subprocess, receiving JSON-RPC style commands
on stdin and returning results on stdout. This avoids the ~100-200ms Python
startup overhead per call.

Protocol:
- One JSON object per line on stdin (request)
- One JSON object per line on stdout (response)
- Request: {"id": N, "method": "...", "params": {...}}
- Response: {"id": N, "result": ..., "error": null} or {"id": N, "result": null, "error": "..."}
"""
import json
import sys
import traceback


def main():
    """Process commands from stdin, write responses to stdout."""
    # Line-buffered output
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue

        request_id = None
        try:
            request = json.loads(line)
            request_id = request.get("id")
            response = dispatch(request)
        except Exception as e:
            response = {
                "id": request_id,
                "result": None,
                "error": f"{type(e).__name__}: {e}\n{traceback.format_exc()}",
            }

        print(json.dumps(response, separators=(",", ":")), flush=True)


def dispatch(request):
    """Dispatch a request to the appropriate handler."""
    method = request.get("method", "")
    params = request.get("params", {})
    request_id = request.get("id")

    handlers = {
        "discover": handle_discover,
        "read_from_file": handle_read_from_file,
        "read_from_string": handle_read_from_string,
        "write_to_file": handle_write_to_file,
        "write_to_string": handle_write_to_string,
        "ping": handle_ping,
    }

    handler = handlers.get(method)
    if handler is None:
        return {
            "id": request_id,
            "result": None,
            "error": f"unknown method: {method}",
        }

    try:
        result = handler(params)
        return {"id": request_id, "result": result, "error": None}
    except Exception as e:
        return {
            "id": request_id,
            "result": None,
            "error": f"{type(e).__name__}: {e}",
        }


def handle_ping(params):
    """Simple ping to check if bridge is alive."""
    return "pong"


def handle_discover(params):
    """Return list of available adapters with their capabilities."""
    import opentimelineio as otio

    result = []
    for adapter in otio.plugins.ActiveManifest().adapters:
        info = {
            "name": adapter.name,
            "suffixes": list(adapter.suffixes) if adapter.suffixes else [],
            "features": {
                "read": adapter.has_feature("read_from_file")
                or adapter.has_feature("read_from_string"),
                "write": adapter.has_feature("write_to_file")
                or adapter.has_feature("write_to_string"),
            },
        }
        result.append(info)
    return result


def handle_read_from_file(params):
    """Read a file using an adapter and return OTIO JSON."""
    import opentimelineio as otio

    filepath = params["filepath"]
    adapter_name = params.get("adapter")
    args = params.get("args", {})

    obj = otio.adapters.read_from_file(filepath, adapter_name=adapter_name, **args)
    return otio.adapters.write_to_string(obj, "otio_json")


def handle_read_from_string(params):
    """Parse a string using an adapter and return OTIO JSON."""
    import opentimelineio as otio

    data = params["data"]
    adapter_name = params["adapter"]
    args = params.get("args", {})

    obj = otio.adapters.read_from_string(data, adapter_name=adapter_name, **args)
    return otio.adapters.write_to_string(obj, "otio_json")


def handle_write_to_file(params):
    """Write OTIO JSON to a file using an adapter."""
    import opentimelineio as otio

    data = params["data"]
    filepath = params["filepath"]
    adapter_name = params.get("adapter")
    args = params.get("args", {})

    obj = otio.adapters.read_from_string(data, "otio_json")
    otio.adapters.write_to_file(obj, filepath, adapter_name=adapter_name, **args)
    return True


def handle_write_to_string(params):
    """Serialize OTIO JSON to a string using an adapter."""
    import opentimelineio as otio

    data = params["data"]
    adapter_name = params["adapter"]
    args = params.get("args", {})

    obj = otio.adapters.read_from_string(data, "otio_json")
    return otio.adapters.write_to_string(obj, adapter_name=adapter_name, **args)


if __name__ == "__main__":
    main()
