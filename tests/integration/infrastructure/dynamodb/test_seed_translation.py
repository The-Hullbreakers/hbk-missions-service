"""TEST-SEED-LANG-002: safe seed against an owned DynamoDB Local table."""

import copy
import unittest

from seed_translation_fixture import (
    SEASON, capture_error, get, install_existing, new_local_table,
    put_new, replace, seed, snapshot, spanish_and_bilingual,
)


class SafeTranslationSeedTest(unittest.TestCase):
    def setUp(self):
        self.aws = new_local_table()
        self.spanish, self.bilingual = spanish_and_bilingual()

    def test_legacy_version_is_not_silently_enriched_on_redeploy(self):
        # Arrange
        install_existing(self.aws, self.spanish)
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [self.bilingual], apply=True))

        # Assert
        self.assertIsNone(error)
        self.assertEqual(snapshot(self.aws), before)
        self.assertNotIn("ContentEn", get(self.aws, self.spanish))

    def test_old_catalog_cannot_erase_english_enriched_after_backfill(self):
        # Arrange
        install_existing(self.aws, self.bilingual)
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [self.spanish], apply=True))

        # Assert
        self.assertIsNone(error)
        self.assertEqual(snapshot(self.aws), before)

    def test_redeploy_bilingual_row_keeps_offer_and_contract_frozen(self):
        # Arrange
        install_existing(self.aws, self.bilingual)
        contract = seed.item({"PK": "CONTRACT", "SK": "contract-sentinel",
                              "Title": "Frozen terms", "ContentEn": {"Title": "Frozen English"}})
        put_new(self.aws, contract)
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [self.bilingual], apply=True))

        # Assert
        self.assertIsNone(error)
        self.assertEqual(snapshot(self.aws), before)

    def test_same_version_offer_spanish_drift_is_rejected_without_writes(self):
        # Arrange
        install_existing(self.aws, self.spanish)
        offer = seed.offer_item(self.spanish, SEASON)
        drifted = copy.deepcopy(offer)
        drifted["Title"] = seed.av("Different Spanish offer title")
        replace(self.aws, drifted)
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [self.spanish], apply=True))

        # Assert
        self.assertIsInstance(error, ValueError)
        self.assertEqual(snapshot(self.aws), before)

    def test_same_version_offer_english_conflict_is_rejected_without_writes(self):
        # Arrange
        install_existing(self.aws, self.bilingual)
        offer = seed.offer_item(self.bilingual, SEASON)
        drifted = copy.deepcopy(offer)
        drifted["ContentEn"]["M"]["Title"] = seed.av("Another translation")
        replace(self.aws, drifted)
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [self.bilingual], apply=True))

        # Assert
        self.assertIsInstance(error, ValueError)
        self.assertEqual(snapshot(self.aws), before)

    def test_new_version_without_english_does_not_change_any_row(self):
        # Arrange
        install_existing(self.aws, self.bilingual)
        new_version = copy.deepcopy(self.spanish)
        new_version["Version"] = seed.av("5")
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [new_version], apply=True))

        # Assert
        self.assertIsInstance(error, ValueError)
        self.assertEqual(snapshot(self.aws), before)

    def test_new_version_with_incomplete_english_does_not_change_any_row(self):
        # Arrange
        install_existing(self.aws, self.bilingual)
        new_version = copy.deepcopy(self.bilingual)
        new_version["Version"] = seed.av("5")
        new_version["ContentEn"]["M"]["Rules"]["L"] = []
        before = snapshot(self.aws)

        # Act
        error = capture_error(lambda: seed.synchronize(self.aws, [new_version], apply=True))

        # Assert
        self.assertIsInstance(error, ValueError)
        self.assertEqual(snapshot(self.aws), before)

    def test_new_version_advances_offer_and_preserves_old_snapshots(self):
        # Arrange
        install_existing(self.aws, self.bilingual)
        contract = seed.item({"PK": "CONTRACT", "SK": "contract-sentinel",
                              "Title": "Frozen terms", "ContentEn": {"Title": "Frozen English"}})
        put_new(self.aws, contract)
        old_version = seed.versioned(self.bilingual)
        new_version = copy.deepcopy(self.bilingual)
        new_version["Version"] = seed.av("5")
        new_version["ContentEn"]["M"]["Title"] = seed.av("Revised English title")
        new_history = seed.versioned(new_version)
        new_offer = seed.offer_item(new_version, SEASON)

        # Act
        result = seed.synchronize(self.aws, [new_version], apply=True)

        # Assert
        self.assertTrue(result["applied"])
        self.assertEqual(get(self.aws, old_version), old_version)
        self.assertEqual(get(self.aws, new_version)["ContentEn"], new_version["ContentEn"])
        self.assertEqual(get(self.aws, new_history), new_history)
        self.assertEqual(get(self.aws, new_offer)["ContentEn"], new_version["ContentEn"])
        self.assertEqual(get(self.aws, contract), contract)


if __name__ == "__main__":
    unittest.main()
