"""TEST-SEED-LANG-002: the local bootstrap must seed both catalogs bilingually."""

import os
import subprocess
import sys
import unittest

from seed_translation_fixture import (
    ROOT, install_existing, new_local_table, seed, snapshot, spanish_and_bilingual,
)


class BilingualLocalBootstrapTest(unittest.TestCase):
    def test_local_bootstrap_writes_both_languages_without_changing_spanish(self):
        # Arrange
        aws = new_local_table()
        catalogs = ("racing-missions.json", "pvp-missions.json")
        spanish = {
            row["SK"]["S"]: row
            for name in catalogs
            for row in seed.definitions(seed.load_catalog(ROOT / "seeds" / name), "staging")
        }
        command = [sys.executable, "-B", str(ROOT / "scripts/seed_local_missions.py"),
                   "--endpoint-url", os.environ["DYNAMODB_SEED_TEST_ENDPOINT"],
                   "--table", aws.table, "--apply"]

        # Act
        result = subprocess.run(command, cwd=ROOT, capture_output=True, text=True)

        # Assert
        self.assertEqual(result.returncode, 0, result.stderr)
        stored = aws.partition("DEFINITION")
        offers = aws.partition("OFFER")
        self.assertEqual((len(stored), len(offers)), (72, 36))
        for row in stored:
            self.assertIn("ContentEn", row)
            if row["SK"]["S"] in spanish:
                self.assertEqual({key: value for key, value in row.items()
                                  if key != "ContentEn"}, spanish[row["SK"]["S"]])
        self.assertTrue(all("ContentEn" in offer for offer in offers))

    def test_local_bootstrap_preserves_spanish_only_legacy_rows_on_redeploy(self):
        # Arrange
        aws = new_local_table()
        spanish, _ = spanish_and_bilingual()
        install_existing(aws, spanish)
        before = snapshot(aws)
        command = [sys.executable, "-B", str(ROOT / "scripts/seed_local_missions.py"),
                   "--endpoint-url", os.environ["DYNAMODB_SEED_TEST_ENDPOINT"],
                   "--table", aws.table, "--apply"]

        # Act
        result = subprocess.run(command, cwd=ROOT, capture_output=True, text=True)

        # Assert
        self.assertEqual(result.returncode, 0, result.stderr)
        after = snapshot(aws)
        self.assertEqual((len(aws.partition("DEFINITION")), len(aws.partition("OFFER"))), (72, 36))
        for identity, original in before.items():
            self.assertEqual(after[identity], original)


if __name__ == "__main__":
    unittest.main()
