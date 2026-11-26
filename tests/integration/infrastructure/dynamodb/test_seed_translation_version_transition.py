"""TEST-SEED-LANG-002: version advances cannot hide prior offer drift."""

import copy
import unittest

from seed_translation_fixture import (
    SEASON, capture_error, install_existing, new_local_table,
    replace, seed, snapshot, spanish_and_bilingual,
)


class VersionTransitionSafetyTest(unittest.TestCase):
    def test_dry_run_rejects_corrupt_prior_offer_before_version_advance(self):
        # Arrange
        _, published = spanish_and_bilingual()
        for field in ("Spanish title", "English title", "MeritPerUnit"):
            with self.subTest(field=field):
                aws = new_local_table()
                install_existing(aws, published)
                offer = copy.deepcopy(seed.offer_item(published, SEASON))
                if field == "Spanish title":
                    offer["Title"] = seed.av("Corrupted Spanish")
                elif field == "English title":
                    offer["ContentEn"]["M"]["Title"] = seed.av("Corrupted English")
                else:
                    offer["Objectives"]["L"][0]["M"]["MeritPerUnit"] = seed.av(9999)
                replace(aws, offer)
                before = snapshot(aws)
                next_version = copy.deepcopy(published)
                next_version["Version"] = seed.av("5")

                # Act
                error = capture_error(lambda: seed.synchronize(aws, [next_version], apply=False))

                # Assert
                self.assertIsInstance(error, ValueError)
                self.assertEqual(snapshot(aws), before)

    def test_apply_rejects_corrupt_prior_offer_without_partial_writes(self):
        # Arrange
        aws = new_local_table()
        _, published = spanish_and_bilingual()
        install_existing(aws, published)
        offer = copy.deepcopy(seed.offer_item(published, SEASON))
        offer["Title"] = seed.av("Corrupted Spanish")
        replace(aws, offer)
        before = snapshot(aws)
        next_version = copy.deepcopy(published)
        next_version["Version"] = seed.av("5")

        # Act
        error = capture_error(lambda: seed.synchronize(aws, [next_version], apply=True))

        # Assert
        self.assertIsInstance(error, ValueError)
        self.assertEqual(snapshot(aws), before)


if __name__ == "__main__":
    unittest.main()
