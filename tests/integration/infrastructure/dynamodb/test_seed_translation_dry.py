"""TEST-SEED-LANG-002: dry-run must detect unsafe redeploy without writes."""

import copy
import unittest

from seed_translation_fixture import (
    SEASON, capture_error, install_existing, new_local_table,
    replace, seed, snapshot, spanish_and_bilingual,
)


class SafeTranslationDryRunTest(unittest.TestCase):
    def setUp(self):
        self.aws = new_local_table()
        self.spanish, self.bilingual = spanish_and_bilingual()

    def test_legacy_es_only_rows_remain_unchanged_in_translated_dry_run(self):
        # Arrange
        install_existing(self.aws, self.spanish)
        before = snapshot(self.aws)

        # Act
        result = seed.synchronize(self.aws, [self.bilingual], apply=False)

        # Assert
        self.assertFalse(result["applied"])
        self.assertEqual(snapshot(self.aws), before)

    def test_dry_run_rejects_offer_spanish_drift(self):
        # Arrange
        install_existing(self.aws, self.spanish)
        offer = seed.offer_item(self.spanish, SEASON)
        drifted = copy.deepcopy(offer)
        drifted["Title"] = seed.av("Different Spanish offer title")
        replace(self.aws, drifted)
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [self.spanish], apply=False))

        # Assert
        self.assertIsInstance(error, ValueError)
        self.assertEqual(snapshot(self.aws), before)

    def test_dry_run_rejects_offer_english_conflict(self):
        # Arrange
        install_existing(self.aws, self.bilingual)
        offer = seed.offer_item(self.bilingual, SEASON)
        drifted = copy.deepcopy(offer)
        drifted["ContentEn"]["M"]["Title"] = seed.av("Incorrect English")
        replace(self.aws, drifted)
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [self.bilingual], apply=False))

        # Assert
        self.assertIsInstance(error, ValueError)
        self.assertEqual(snapshot(self.aws), before)

    def test_dry_run_rejects_offer_merit_drift(self):
        # Arrange
        install_existing(self.aws, self.spanish)
        offer = seed.offer_item(self.spanish, SEASON)
        drifted = copy.deepcopy(offer)
        drifted["Objectives"]["L"][0]["M"]["MeritPerUnit"] = seed.av(1000)
        replace(self.aws, drifted)
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [self.spanish], apply=False))

        # Assert
        self.assertIsInstance(error, ValueError)
        self.assertEqual(snapshot(self.aws), before)


if __name__ == "__main__":
    unittest.main()
