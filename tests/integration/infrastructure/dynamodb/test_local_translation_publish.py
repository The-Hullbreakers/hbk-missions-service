"""TEST-SEED-LOCAL-002: local-only CLI publishes reviewed English safely."""

import copy
import json
import os
import subprocess
import unittest

from seed_translation_fixture import (
    ROOT, SEASON, install_existing, new_local_table, replace,
    seed, snapshot, spanish_and_bilingual,
)


def invoke(aws, mode, endpoint=None):
    command = [
        "go", "run", "./cmd/local-translation-backfill",
        "--endpoint-url", endpoint or os.environ["DYNAMODB_SEED_TEST_ENDPOINT"],
        "--table", aws.table, "--translations-dir", str(ROOT / "seeds/translations"),
        mode,
    ]
    return subprocess.run(command, cwd=ROOT, capture_output=True, text=True, timeout=180)


class LocalReviewedTranslationPublicationTest(unittest.TestCase):
    def test_preview_lists_eligible_rows_without_writing(self):
        # Arrange
        aws = new_local_table()
        spanish, _ = spanish_and_bilingual()
        install_existing(aws, spanish)
        before = snapshot(aws)
        identities = {
            ("DEFINITION", spanish["SK"]["S"]),
            ("DEFINITION", seed.versioned(spanish)["SK"]["S"]),
            ("OFFER", seed.offer_item(spanish, SEASON)["SK"]["S"]),
        }

        # Act
        result = invoke(aws, "--dry-run")

        # Assert
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertFalse(report["applied"])
        self.assertEqual({(row["partition"], row["id"]) for row in report["eligible"]}, identities)
        self.assertEqual(report["updated"], [])
        self.assertEqual(report["unresolved"], [])
        self.assertEqual(snapshot(aws), before)

    def test_apply_adds_only_english_to_three_owned_rows(self):
        # Arrange
        aws = new_local_table()
        spanish, bilingual = spanish_and_bilingual()
        install_existing(aws, spanish)
        before = snapshot(aws)

        # Act
        result = invoke(aws, "--apply")

        # Assert
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertTrue(report["applied"])
        self.assertEqual(len(report["updated"]), 3)
        self.assertEqual(report["unresolved"], [])
        after = snapshot(aws)
        self.assertEqual(set(after), set(before))
        for identity, original in before.items():
            self.assertEqual(after[identity]["ContentEn"], bilingual["ContentEn"])
            self.assertEqual({key: value for key, value in after[identity].items()
                              if key != "ContentEn"}, original)

    def test_unresolved_offer_blocks_all_updates_before_apply(self):
        # Arrange
        aws = new_local_table()
        spanish, _ = spanish_and_bilingual()
        install_existing(aws, spanish)
        divergent = copy.deepcopy(seed.offer_item(spanish, SEASON))
        divergent["Title"] = seed.av("Different Spanish snapshot")
        replace(aws, divergent)
        before = snapshot(aws)

        # Act
        result = invoke(aws, "--apply")

        # Assert
        self.assertNotEqual(result.returncode, 0)
        self.assertTrue(result.stdout.strip(), result.stderr)
        report = json.loads(result.stdout)
        self.assertFalse(report["applied"])
        self.assertEqual(report["updated"], [])
        self.assertIn(("OFFER", divergent["SK"]["S"]),
                      {(row["partition"], row["id"]) for row in report["unresolved"]})
        self.assertEqual(snapshot(aws), before)

    def test_repeating_apply_does_not_replace_existing_english(self):
        # Arrange
        aws = new_local_table()
        spanish, _ = spanish_and_bilingual()
        install_existing(aws, spanish)
        first = invoke(aws, "--apply")
        before = snapshot(aws)

        # Act
        result = invoke(aws, "--apply")

        # Assert
        self.assertEqual(first.returncode, 0, first.stderr)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(json.loads(result.stdout)["updated"], [])
        self.assertEqual(snapshot(aws), before)

    def test_cli_rejects_non_http_endpoint_before_access(self):
        # Arrange
        aws = new_local_table()
        before = snapshot(aws)

        # Act
        result = invoke(aws, "--dry-run", endpoint="https://localhost:18000")

        # Assert
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("loopback", result.stderr.lower())
        self.assertEqual(snapshot(aws), before)


if __name__ == "__main__":
    unittest.main()
