import unittest

from conformance.authorization_closure.adapters.openclaw_exec_launch_policy import (
    map_openclaw_exec_launch_policy,
)
from conformance.authorization_closure.runner import Verdict, evaluate


class OpenClawExecLaunchPolicyAdapterTests(unittest.TestCase):
    def test_gateway_pty_revoked_before_native_launch_maps_to_closed(self):
        observation = {
            "route": "Gateway PTY",
            "timing": "before",
            "nativeLaunches": 0,
            "markerExists": False,
            "operation_id": "openclaw-166901-gateway-pty-before",
            "revocation_id": "openclaw-166901-policy-revocation",
        }

        trace = map_openclaw_exec_launch_policy(observation)
        result = evaluate(trace)

        self.assertEqual(trace["sinks"], ["native-pty-construction"])
        self.assertEqual(result.verdict, Verdict.CLOSED)
        self.assertEqual(result.closed_sinks, ("native-pty-construction",))

    def test_native_launch_after_revocation_effective_maps_to_violation(self):
        observation = {
            "route": "Gateway PTY",
            "timing": "before",
            "nativeLaunches": 1,
            "markerExists": True,
        }

        trace = map_openclaw_exec_launch_policy(observation)
        result = evaluate(trace)

        self.assertEqual(result.verdict, Verdict.VIOLATION)


if __name__ == "__main__":
    unittest.main()
