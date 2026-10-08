import unittest

from conformance.authorization_closure.adapters.open_agent_auth_revocation import (
    map_open_agent_auth_revocation_evidence,
)
from conformance.authorization_closure.runner import Verdict, evaluate


class OpenAgentAuthRevocationAdapterTests(unittest.TestCase):
    def test_authority_revoked_without_resource_server_closure_is_unknown(self):
        observation = {
            "revocationServiceTestPassed": True,
            "resourceServerObservedRevocation": False,
        }

        trace = map_open_agent_auth_revocation_evidence(observation)
        result = evaluate(trace)

        self.assertEqual(result.verdict, Verdict.UNKNOWN)
        self.assertEqual(
            result.unknown_sinks,
            ("open-agent-auth-resource-server",),
        )

    def test_observed_resource_server_revocation_can_close_sink(self):
        observation = {
            "revocationServiceTestPassed": True,
            "resourceServerObservedRevocation": True,
        }

        trace = map_open_agent_auth_revocation_evidence(observation)
        result = evaluate(trace)

        self.assertEqual(result.verdict, Verdict.CLOSED)


if __name__ == "__main__":
    unittest.main()
