"""REQ-0055-001/002: safe local seasonal previews."""

import sys
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "scripts"))

from seed_values import item
from season_scenario_model import build_scenario, missing_items, validate_target


class PreviewSafetyTest(unittest.TestCase):
    def test_rejects_remote_or_unapproved_destinations(self):
        for endpoint, scenario in [("https://localhost:8000", "mid-season"),
                                   ("http://localhost.evil.test:8000", "mid-season"),
                                   ("http://localhost:8000/path", "mid-season"),
                                   ("http://user:pass@localhost:8000", "mid-season"),
                                   ("http://@localhost:8000", "mid-season"),
                                   ("http://localhost:8000", "production")]:
            with self.subTest(endpoint=endpoint, scenario=scenario):
                # Arrange
                url, profile = endpoint, scenario
                # Act
                destination = validate_target(url, profile)
                # Assert
                self.assertIsNone(destination)

    def test_accepts_loopback_preview_destination(self):
        # Arrange
        endpoint, scenario = "http://127.0.0.1:8000", "mid-season"
        # Act
        destination = validate_target(endpoint, scenario)
        # Assert
        self.assertEqual(destination, "hbk-missions-preview-mid-season")

    def test_repeat_load_only_inserts_missing_rows(self):
        # Arrange
        season = item({"PK": "SEASON", "SK": "preview", "State": "ACTIVE"})
        contract = item({"PK": "CONTRACT", "SK": "preview-1", "State": "COMPLETED"})
        # Act
        pending = missing_items([season], [season, contract])
        # Assert
        self.assertEqual(pending, [contract])

    def test_repeat_load_does_not_duplicate_rows(self):
        # Arrange
        season = item({"PK": "SEASON", "SK": "preview", "State": "ACTIVE"})
        # Act
        pending = missing_items([season], [season])
        # Assert
        self.assertEqual(pending, [])

    def test_changed_row_is_not_overwritten(self):
        # Arrange
        old = item({"PK": "SEASON", "SK": "preview", "State": "FINALIZED"})
        desired = item({"PK": "SEASON", "SK": "preview", "State": "ACTIVE"})
        # Act
        pending = missing_items([old], [desired])
        # Assert
        self.assertIsNone(pending)

    def test_unknown_existing_row_is_not_overwritten(self):
        # Arrange
        old = item({"PK": "SEASON", "SK": "other", "State": "ACTIVE"})
        desired = item({"PK": "SEASON", "SK": "preview", "State": "ACTIVE"})
        # Act
        pending = missing_items([old], [desired])
        # Assert
        self.assertIsNone(pending)


class PreviewPhasesTest(unittest.TestCase):
    def setUp(self):
        self.definitions = [item({
            "PK": "DEFINITION", "SK": code, "DefinitionCode": code,
            "Version": "1", "State": "PUBLISHED", "Mode": "SOLO",
            "Title": code, "Category": category, "CompletionReward": 10,
            "Objectives": [{"ObjectiveID": "finish", "Category": category,
                            "MeritPerUnit": 40, "UnitLimit": 1, "Bonus": 7}],
        }) for code, category in (("race-a", "RACING"), ("duel-a", "PVP_DUEL"))]
        self.users = ["player-1", "player-2", "player-3"]

    def test_off_season_has_no_active_work(self):
        # Arrange
        definitions = self.definitions
        # Act
        rows = build_scenario("off-season", definitions, [])
        # Assert
        self.assertFalse(self._partition(rows, "SEASON"))
        self.assertFalse(self._partition(rows, "OFFER"))
        self.assertFalse(self._partition(rows, "MERIT_CREDIT"))

    def test_fresh_start_has_offers_without_merits(self):
        # Arrange
        definitions = self.definitions
        # Act
        rows = build_scenario("fresh-start", definitions, [])
        # Assert
        self.assertEqual(len(self._partition(rows, "SEASON")), 1)
        self.assertEqual(len(self._partition(rows, "OFFER")), 2)
        self.assertFalse(self._partition(rows, "MERIT_CREDIT"))

    def test_mid_season_has_connected_contracts_and_rankings(self):
        # Arrange
        definitions, users = self.definitions, self.users
        # Act
        rows = build_scenario("mid-season", definitions, users)
        # Assert
        contracts = self._partition(rows, "CONTRACT")
        credits = self._partition(rows, "MERIT_CREDIT")
        self.assertTrue({"IN_PROGRESS", "COMPLETED"} <=
                        {row["State"]["S"] for row in contracts})
        self.assertGreaterEqual(len(credits), 3)
        self.assertGreaterEqual(len({c["RecipientID"]["S"] for c in credits}), 3)
        self.assertTrue({"RACING", "PVP_DUEL"} <= {c["Category"]["S"] for c in credits})
        self.assertTrue(all(c["ContractID"]["S"] in {r["SK"]["S"] for r in contracts}
                            for c in credits))
        self.assertTrue(all("GSI1PK" in c and "GSI2PK" in c for c in credits))
        by_id = {r["SK"]["S"]: r for r in contracts}
        for credit in credits:
            contract = by_id[credit["ContractID"]["S"]]
            if credit["ObjectiveID"]["S"] == "COMPLETION":
                self.assertEqual(int(credit["Amount"]["N"]), int(contract["CompletionReward"]["N"]))
            else:
                self.assertEqual(int(credit["Amount"]["N"]), 47)
        self.assertTrue(any(c["ObjectiveID"]["S"] == "COMPLETION" for c in credits))
        keys = [(r["PK"]["S"], r["SK"]["S"]) for r in rows]
        self.assertEqual(len(keys), len(set(keys)))

    def test_mid_season_requires_three_active_users(self):
        # Arrange
        definitions, users = self.definitions, self.users[:2]
        # Act
        rows = build_scenario("mid-season", definitions, users)
        # Assert
        self.assertIsNone(rows)

    @staticmethod
    def _partition(rows, name):
        return [row for row in rows if row["PK"]["S"] == name]
