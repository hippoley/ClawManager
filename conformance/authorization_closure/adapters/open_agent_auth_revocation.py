from __future__ import annotations

from typing import Any


def map_open_agent_auth_revocation_evidence(observation: dict[str, Any]) -> dict[str, Any]:
    """Map Open Agent Auth revocation test/source evidence into closure events.

    This adapter intentionally returns incomplete closure evidence when the
    authority-side revocation is proven but Resource Server enforcement is not.
    """

    if not observation.get("revocationServiceTestPassed"):
        raise ValueError("authority-side revocation evidence is required")

    sink = "open-agent-auth-resource-server"

    events: list[dict[str, Any]] = [
        {"seq": 1, "type": "revocation_recorded"},
    ]

    if observation.get("resourceServerObservedRevocation"):
        events.extend(
            [
                {"seq": 2, "type": "revocation_effective", "sink": sink},
                {"seq": 3, "type": "sink_closed", "sink": sink},
            ]
        )
    else:
        events.append({"seq": 2, "type": "evidence_unavailable", "sink": sink})

    return {
        "operation_id": observation.get(
            "operation_id", "open-agent-auth-aoat-revocation"
        ),
        "revocation_id": observation.get(
            "revocation_id", "open-agent-auth-token-revocation"
        ),
        "sinks": [sink],
        "events": events,
        "evidence": {
            "repository": "alibaba/open-agent-auth",
            "authority_test": (
                "open-agent-auth-core/src/test/java/com/alibaba/openagentauth/"
                "core/protocol/oauth2/token/revocation/"
                "InMemoryTokenRevocationServiceTest.java"
            ),
            "resource_server": (
                "open-agent-auth-framework/src/main/java/com/alibaba/"
                "openagentauth/framework/orchestration/DefaultResourceServer.java"
            ),
            "aoat_validator": (
                "open-agent-auth-core/src/main/java/com/alibaba/openagentauth/"
                "core/token/aoat/AoatValidator.java"
            ),
            "observation_kind": "source-and-upstream-test-evidence",
        },
    }
