from __future__ import annotations

from typing import Any


def map_openclaw_exec_launch_policy(observation: dict[str, Any]) -> dict[str, Any]:
    """Map OpenClaw launch-policy regression evidence into closure events.

    This adapter targets the evidence shape established by:
    src/agents/bash-tools.exec-launch-policy.integration.test.ts

    It does not claim OpenClaw emits this JSON natively.
    """

    route = observation["route"]
    timing = observation["timing"]
    native_launches = int(observation["nativeLaunches"])
    marker_exists = bool(observation["markerExists"])

    sink = {
        "Gateway child": "native-child-process",
        "Gateway PTY": "native-pty-construction",
        "Claude node": "native-claude-process",
    }[route]

    events: list[dict[str, Any]] = [
        {"seq": 1, "type": "authorization_accepted"},
    ]

    if timing == "before":
        events.extend(
            [
                {"seq": 2, "type": "revocation_recorded"},
                {"seq": 3, "type": "revocation_effective", "sink": sink},
            ]
        )

        if native_launches == 0 and not marker_exists:
            events.append({"seq": 4, "type": "sink_closed", "sink": sink})
        else:
            events.append({"seq": 4, "type": "effect_committed", "sink": sink})

    elif timing == "after":
        if native_launches:
            events.append({"seq": 2, "type": "effect_committed", "sink": sink})
        events.extend(
            [
                {"seq": 3, "type": "revocation_recorded"},
                {"seq": 4, "type": "revocation_effective", "sink": sink},
                {"seq": 5, "type": "sink_closed", "sink": sink},
            ]
        )
    else:
        raise ValueError(f"unsupported OpenClaw timing: {timing}")

    return {
        "operation_id": observation.get("operation_id", "openclaw-exec-launch-policy"),
        "revocation_id": observation.get("revocation_id", "openclaw-policy-revocation"),
        "sinks": [sink],
        "events": events,
        "evidence": {
            "repository": "openclaw/openclaw",
            "test": "src/agents/bash-tools.exec-launch-policy.integration.test.ts",
            "source_pr": 166901,
            "route": route,
            "timing": timing,
        },
    }
