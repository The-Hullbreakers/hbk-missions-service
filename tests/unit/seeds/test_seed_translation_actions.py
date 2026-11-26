"""TEST-SEED-LANG-002: seed actions protect old locales and require new English."""

import copy
import unittest

from translation_fixture import capture_error, one_pvp_mission, reviewed_record, seed


class SafeSeedActionTest(unittest.TestCase):
    def setUp(self):
        catalog = one_pvp_mission()
        self.spanish = seed.definitions(catalog, "staging")[0]
        self.english = reviewed_record(self.spanish)["English"]
        self.bilingual = {**copy.deepcopy(self.spanish), "ContentEn": seed.av(self.english)}

    def test_redeploy_old_catalog_preserves_enriched_current_and_version(self):
        # Arrange
        current = copy.deepcopy(self.bilingual)
        historical = seed.versioned(current)

        # Act
        actions = seed.definition_actions("table", self.spanish, current, historical)

        # Assert
        self.assertIn("ConditionCheck", actions[0])
        self.assertEqual(actions[1]["Put"]["Item"], current)
        self.assertEqual(historical["ContentEn"], current["ContentEn"])

    def test_translated_catalog_does_not_silently_enrich_existing_version(self):
        # Arrange
        current = copy.deepcopy(self.spanish)
        historical = seed.versioned(current)

        # Act
        actions = seed.definition_actions("table", self.bilingual, current, historical)

        # Assert
        self.assertIn("ConditionCheck", actions[0])
        self.assertNotIn("ContentEn", actions[1]["Put"]["Item"])
        self.assertNotIn("ContentEn", historical)

    def test_missing_english_rejects_first_or_new_version(self):
        # Arrange
        new_version = copy.deepcopy(self.spanish)
        new_version["Version"] = seed.av("5")
        cases = (("first publish", self.spanish, None), ("next version", new_version, self.spanish))

        for name, incoming, current in cases:
            with self.subTest(stage=name):
                # Act
                error = capture_error(lambda: seed.definition_actions("table", incoming, current, None))

                # Assert
                self.assertIsInstance(error, ValueError)

    def test_incomplete_english_rejects_new_version(self):
        # Arrange
        cases = {
            "empty title": {**self.english, "Title": ""},
            "missing rule": {**self.english, "Rules": self.english["Rules"][:-1]},
            "missing objective": {**self.english, "Objectives": []},
        }
        for name, english in cases.items():
            with self.subTest(missing=name):
                new_version = copy.deepcopy(self.spanish)
                new_version["Version"] = seed.av("5")
                new_version["ContentEn"] = seed.av(english)

                # Act
                error = capture_error(lambda: seed.definition_actions("table", new_version, self.spanish, None))

                # Assert
                self.assertIsInstance(error, ValueError)

    def test_new_version_writes_both_locales_in_definition_rows(self):
        # Arrange
        new_version = copy.deepcopy(self.bilingual)
        new_version["Version"] = seed.av("5")
        new_version["ContentEn"] = seed.av({**self.english, "Title": "Revised English title"})

        # Act
        actions = seed.definition_actions("table", new_version, self.spanish, None)

        # Assert
        self.assertEqual(actions[0]["Put"]["Item"]["ContentEn"], new_version["ContentEn"])
        self.assertEqual(actions[1]["Put"]["Item"]["ContentEn"], new_version["ContentEn"])

    def test_new_version_offer_has_same_bilingual_snapshot(self):
        # Arrange
        new_version = copy.deepcopy(self.bilingual)
        new_version["Version"] = seed.av("5")
        new_version["ContentEn"] = seed.av({**self.english, "Title": "Revised English title"})

        # Act
        offer = seed.offer_item(new_version, "season-1")

        # Assert
        self.assertEqual(offer["DefinitionVersion"], seed.av("5"))
        self.assertEqual(offer["ContentEn"], new_version["ContentEn"])

    def test_same_version_spanish_or_economics_drift_is_rejected(self):
        # Arrange
        title_drift = copy.deepcopy(self.bilingual)
        title_drift["Title"] = seed.av("Different Spanish title")
        merits_drift = copy.deepcopy(self.bilingual)
        merits_drift["Objectives"]["L"][0]["M"]["MeritPerUnit"] = seed.av(1000)
        historical = seed.versioned(self.bilingual)

        for incoming in (title_drift, merits_drift):
            with self.subTest(field="Title" if incoming is title_drift else "MeritPerUnit"):
                # Act
                error = capture_error(lambda: seed.definition_actions(
                    "table", incoming, self.bilingual, historical))

                # Assert
                self.assertIsInstance(error, ValueError)

    def test_same_version_conflicting_english_is_rejected(self):
        # Arrange
        incoming = copy.deepcopy(self.bilingual)
        incoming["ContentEn"] = seed.av({**self.english, "Title": "Conflicting English title"})

        # Act
        error = capture_error(lambda: seed.definition_actions(
            "table", incoming, self.bilingual, seed.versioned(self.bilingual)))

        # Assert
        self.assertIsInstance(error, ValueError)


if __name__ == "__main__":
    unittest.main()
