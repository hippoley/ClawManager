import unittest

from conformance.authorization_closure.runner import Verdict, evaluate


class AuthorizationClosureTests(unittest.TestCase):
    def test_closed(self):
        result = evaluate({
            "operation_id": "op-1",
            "revocation_id": "rev-1",
            "sinks": ["rs", "worker"],
            "events": [
                {"seq": 1, "type": "authorization_accepted"},
                {"seq": 2, "type": "revocation_recorded"},
                {"seq": 3, "type": "sink_closed", "sink": "rs"},
                {"seq": 4, "type": "sink_closed", "sink": "worker"},
            ],
        })
        self.assertEqual(result.verdict, Verdict.CLOSED)

    def test_partial(self):
        result = evaluate({
            "operation_id": "op-2",
            "revocation_id": "rev-2",
            "sinks": ["rs", "worker"],
            "events": [
                {"seq": 1, "type": "revocation_recorded"},
                {"seq": 2, "type": "sink_closed", "sink": "rs"},
            ],
        })
        self.assertEqual(result.verdict, Verdict.PARTIAL)
        self.assertEqual(result.open_sinks, ("worker",))

    def test_unknown_without_revocation_evidence(self):
        result = evaluate({
            "operation_id": "op-3",
            "revocation_id": "rev-3",
            "sinks": ["rs"],
            "events": [
                {"seq": 1, "type": "authorization_accepted"},
            ],
        })
        self.assertEqual(result.verdict, Verdict.UNKNOWN)

    def test_violation_when_effect_commits_after_revocation(self):
        result = evaluate({
            "operation_id": "op-4",
            "revocation_id": "rev-4",
            "sinks": ["payment"],
            "events": [
                {"seq": 1, "type": "authorization_accepted"},
                {"seq": 2, "type": "revocation_recorded"},
                {"seq": 3, "type": "effect_committed", "sink": "payment"},
            ],
        })
        self.assertEqual(result.verdict, Verdict.VIOLATION)
        self.assertEqual(result.violating_sinks, ("payment",))

    def test_effect_before_revocation_is_not_a_violation(self):
        result = evaluate({
            "operation_id": "op-5",
            "revocation_id": "rev-5",
            "sinks": ["payment"],
            "events": [
                {"seq": 1, "type": "effect_committed", "sink": "payment"},
                {"seq": 2, "type": "revocation_recorded"},
                {"seq": 3, "type": "sink_closed", "sink": "payment"},
            ],
        })
        self.assertEqual(result.verdict, Verdict.CLOSED)


if __name__ == "__main__":
    unittest.main()
