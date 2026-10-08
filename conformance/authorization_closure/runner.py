from __future__ import annotations

import json
import sys
from dataclasses import dataclass
from enum import Enum
from pathlib import Path
from typing import Any


class Verdict(str, Enum):
    CLOSED = "CLOSED"
    PARTIAL = "PARTIAL"
    UNKNOWN = "UNKNOWN"
    VIOLATION = "VIOLATION"


@dataclass(frozen=True)
class Evaluation:
    verdict: Verdict
    closed_sinks: tuple[str, ...]
    open_sinks: tuple[str, ...]
    unknown_sinks: tuple[str, ...]
    violating_sinks: tuple[str, ...]
    reason: str

    def to_dict(self) -> dict[str, Any]:
        return {
            "verdict": self.verdict.value,
            "closed_sinks": list(self.closed_sinks),
            "open_sinks": list(self.open_sinks),
            "unknown_sinks": list(self.unknown_sinks),
            "violating_sinks": list(self.violating_sinks),
            "reason": self.reason,
        }


def evaluate(trace: dict[str, Any]) -> Evaluation:
    sinks = set(trace.get("sinks") or [])
    events = sorted(trace.get("events") or [], key=lambda e: e["seq"])

    if not sinks:
        return Evaluation(
            Verdict.UNKNOWN, (), (), (), (),
            "no consequential sinks declared",
        )

    revoked_seq: int | None = None
    closed_at: dict[str, int] = {}
    evidence_unavailable: set[str] = set()
    effect_after_revocation: dict[str, int] = {}
    observed_sinks: set[str] = set()

    for event in events:
        etype = event.get("type")
        seq = int(event["seq"])
        sink = event.get("sink")

        if sink is not None:
            observed_sinks.add(sink)

        if etype == "revocation_recorded":
            if revoked_seq is None:
                revoked_seq = seq
            continue

        if etype == "sink_closed" and sink in sinks:
            closed_at[sink] = seq
            continue

        if etype == "evidence_unavailable" and sink in sinks:
            evidence_unavailable.add(sink)
            continue

        if etype == "effect_committed" and sink in sinks and revoked_seq is not None:
            close_seq = closed_at.get(sink)
            if seq > revoked_seq and (close_seq is None or seq < close_seq):
                effect_after_revocation[sink] = seq

    if revoked_seq is None:
        return Evaluation(
            Verdict.UNKNOWN,
            tuple(sorted(closed_at)),
            (),
            tuple(sorted(sinks)),
            (),
            "no authoritative revocation_recorded event",
        )

    violating = set(effect_after_revocation)
    if violating:
        return Evaluation(
            Verdict.VIOLATION,
            tuple(sorted(set(closed_at) - violating)),
            tuple(sorted(sinks - set(closed_at))),
            tuple(sorted(evidence_unavailable)),
            tuple(sorted(violating)),
            "consequential effect committed after revocation and before sink closure",
        )

    closed = set(closed_at)
    remaining = sinks - closed
    unknown = remaining & evidence_unavailable
    open_sinks = remaining - unknown

    if closed == sinks:
        return Evaluation(
            Verdict.CLOSED,
            tuple(sorted(closed)),
            (),
            (),
            (),
            "all declared consequential sinks have closure evidence",
        )

    if closed:
        return Evaluation(
            Verdict.PARTIAL,
            tuple(sorted(closed)),
            tuple(sorted(open_sinks)),
            tuple(sorted(unknown)),
            (),
            "some consequential paths are closed but closure is incomplete",
        )

    return Evaluation(
        Verdict.UNKNOWN,
        (),
        tuple(sorted(open_sinks)),
        tuple(sorted(unknown or (sinks - observed_sinks))),
        (),
        "no consequential sink has closure evidence",
    )


def load_trace(path: str | Path) -> dict[str, Any]:
    return json.loads(Path(path).read_text(encoding="utf-8"))


def main(argv: list[str] | None = None) -> int:
    argv = argv if argv is not None else sys.argv[1:]
    if len(argv) != 1:
        print("usage: runner.py TRACE.json", file=sys.stderr)
        return 2

    evaluation = evaluate(load_trace(argv[0]))
    print(json.dumps(evaluation.to_dict(), indent=2, sort_keys=True))
    return 1 if evaluation.verdict is Verdict.VIOLATION else 0


if __name__ == "__main__":
    raise SystemExit(main())
