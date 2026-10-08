from __future__ import annotations

from dataclasses import dataclass, replace
from enum import Enum
from typing import Literal


class EffectTruth(str, Enum):
    NOT_DISPATCHED = "NOT_DISPATCHED"
    UNKNOWN = "UNKNOWN"
    ABSENT = "ABSENT"
    COMMITTED = "COMMITTED"


class ExecutionStatus(str, Enum):
    READY = "READY"
    IN_FLIGHT = "IN_FLIGHT"
    SUPERSEDED = "SUPERSEDED"
    TERMINAL = "TERMINAL"


@dataclass(frozen=True)
class DurableAuthorityState:
    logical_operation_id: str
    execution_generation: int = 1
    authority_revision: int = 1
    approval_revision: int | None = None
    approval_granted: bool = False
    effect_truth: EffectTruth = EffectTruth.NOT_DISPATCHED
    execution_status: ExecutionStatus = ExecutionStatus.READY
    current_result_generation: int | None = None
    obligation_owner: str | None = None

    def approve(self) -> "DurableAuthorityState":
        return replace(
            self,
            approval_granted=True,
            approval_revision=self.authority_revision,
        )

    def revoke_or_tighten_authority(self) -> "DurableAuthorityState":
        return replace(self, authority_revision=self.authority_revision + 1)

    def can_dispatch(self, *, require_current_authority: bool = True) -> bool:
        if self.execution_status in {ExecutionStatus.SUPERSEDED, ExecutionStatus.TERMINAL}:
            return False
        if self.effect_truth is EffectTruth.UNKNOWN:
            return False
        if self.effect_truth is EffectTruth.COMMITTED:
            return False
        if not self.approval_granted:
            return False
        if require_current_authority and self.approval_revision != self.authority_revision:
            return False
        return True

    def dispatch(self, *, require_current_authority: bool = True) -> "DurableAuthorityState":
        if not self.can_dispatch(require_current_authority=require_current_authority):
            raise PermissionError("dispatch is not currently authorized")
        return replace(
            self,
            execution_status=ExecutionStatus.IN_FLIGHT,
            effect_truth=EffectTruth.UNKNOWN,
        )

    def reconcile(self, truth: Literal["ABSENT", "COMMITTED"]) -> "DurableAuthorityState":
        return replace(self, effect_truth=EffectTruth(truth))

    def begin_successor_generation(self) -> "DurableAuthorityState":
        return replace(
            self,
            execution_generation=self.execution_generation + 1,
            execution_status=ExecutionStatus.READY,
            current_result_generation=None,
        )

    def supersede_current_generation(self) -> "DurableAuthorityState":
        return replace(self, execution_status=ExecutionStatus.SUPERSEDED)

    def accept_result(self, *, generation: int) -> "DurableAuthorityState":
        if generation != self.execution_generation:
            raise RuntimeError("stale generation result cannot mutate current state")
        if self.execution_status is ExecutionStatus.SUPERSEDED:
            raise RuntimeError("superseded generation result cannot mutate current state")
        return replace(
            self,
            current_result_generation=generation,
            execution_status=ExecutionStatus.TERMINAL,
        )

    def yield_to(self, continuation_owner: str) -> "DurableAuthorityState":
        return replace(
            self,
            obligation_owner=continuation_owner,
            execution_status=ExecutionStatus.READY,
        )

    def continuation_is_current(self, continuation_owner: str) -> bool:
        return self.obligation_owner == continuation_owner


def scenario_a_approval_survives_crash_when_authority_unchanged() -> None:
    s = DurableAuthorityState("op-a").approve()
    assert s.can_dispatch() is True


def scenario_b_historical_approval_does_not_override_revoked_authority() -> None:
    s = DurableAuthorityState("op-b").approve()
    s = s.revoke_or_tighten_authority()
    assert s.approval_granted is True
    assert s.can_dispatch() is False


def scenario_c_unknown_outcome_requires_reconciliation() -> None:
    s = DurableAuthorityState("op-c").approve().dispatch()
    assert s.effect_truth is EffectTruth.UNKNOWN
    assert s.can_dispatch() is False
    s = s.reconcile("COMMITTED")
    assert s.effect_truth is EffectTruth.COMMITTED
    assert s.can_dispatch() is False


def scenario_d_absent_effect_does_not_recreate_revoked_authority() -> None:
    s = DurableAuthorityState("op-d").approve().dispatch()
    s = s.reconcile("ABSENT")
    s = s.revoke_or_tighten_authority()
    assert s.effect_truth is EffectTruth.ABSENT
    assert s.can_dispatch() is False


def scenario_e_stale_completion_cannot_overwrite_successor() -> None:
    s = DurableAuthorityState("op-e").approve()
    old_generation = s.execution_generation
    s = s.begin_successor_generation()
    try:
        s.accept_result(generation=old_generation)
    except RuntimeError:
        pass
    else:
        raise AssertionError("stale result was accepted")


def scenario_f_yield_preserves_obligation_custody() -> None:
    s = DurableAuthorityState("op-f").yield_to("child-1")
    assert s.continuation_is_current("child-1") is True
    assert s.continuation_is_current("child-2") is False


def run_reference_scenarios() -> None:
    scenarios = [
        scenario_a_approval_survives_crash_when_authority_unchanged,
        scenario_b_historical_approval_does_not_override_revoked_authority,
        scenario_c_unknown_outcome_requires_reconciliation,
        scenario_d_absent_effect_does_not_recreate_revoked_authority,
        scenario_e_stale_completion_cannot_overwrite_successor,
        scenario_f_yield_preserves_obligation_custody,
    ]
    for scenario in scenarios:
        scenario()
        print(f"PASS {scenario.__name__}")


if __name__ == "__main__":
    run_reference_scenarios()
