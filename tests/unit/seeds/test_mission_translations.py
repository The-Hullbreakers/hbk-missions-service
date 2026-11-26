"""TEST-SEED-LANG-001: every published catalog text has one source-bound en."""

import copy
from pathlib import Path
import tempfile
import unittest

from translation_fixture import (
    ROOT, capture_error, one_pvp_mission, reviewed_record, seed, write_manifest,
)


class TranslatedCatalogTest(unittest.TestCase):
    def test_all_36_source_versions_have_complete_english_without_economic_changes(self):
        # Arrange
        catalogs = (("racing-missions.json", 24), ("pvp-missions.json", 12))

        for filename, count in catalogs:
            with self.subTest(catalog=filename):
                catalog = seed.load_catalog(ROOT / "seeds" / filename)
                spanish = seed.definitions(catalog, "staging")

                # Act
                translated = seed.translated_definitions(catalog, "staging", ROOT / "seeds/translations")

                # Assert
                self.assertEqual(len(spanish), count)
                self.assertEqual(len(translated), count)
                self.assertEqual({row["DefinitionCode"]["S"] for row in translated},
                                 {row["DefinitionCode"]["S"] for row in spanish})
                for original, bilingual in zip(spanish, translated):
                    self.assertEqual({k: v for k, v in bilingual.items() if k != "ContentEn"}, original)
                    english = bilingual["ContentEn"]["M"]
                    self.assertTrue(english["Title"]["S"].strip())
                    self.assertTrue(english["Description"]["S"].strip())
                    self.assertEqual(len(english["Rules"]["L"]), len(original["Rules"]["L"]))
                    self.assertTrue(all(rule["S"].strip() for rule in english["Rules"]["L"]))
                    objectives = [o["M"] for o in english["Objectives"]["L"]]
                    self.assertEqual({o["ObjectiveID"]["S"] for o in objectives},
                                     {o["M"]["ObjectiveID"]["S"] for o in original["Objectives"]["L"]})
                    self.assertTrue(all(o["Name"]["S"].strip() and o["Description"]["S"].strip()
                                        for o in objectives))

    def test_production_racing_keeps_its_eight_existing_mission_ids(self):
        # Arrange
        catalog = seed.load_catalog(ROOT / "seeds/racing-missions.json")
        spanish = seed.definitions(catalog, "production")

        # Act
        translated = seed.translated_definitions(catalog, "production", ROOT / "seeds/translations")

        # Assert
        self.assertEqual(len(spanish), 8)
        self.assertEqual(len(translated), 8)
        self.assertEqual([r["DefinitionCode"] for r in translated],
                         [r["DefinitionCode"] for r in spanish])

    def test_invalid_source_identity_or_incomplete_english_never_renders(self):
        # Arrange
        catalog = one_pvp_mission()
        row = seed.definitions(catalog, "staging")[0]
        valid = reviewed_record(row)
        cases = {
            "missing record": [],
            "duplicate record": [valid, copy.deepcopy(valid)],
            "wrong version": [{**valid, "DefinitionVersion": "5"}],
            "stale Spanish": [{**valid, "SourceSpanish": {
                **valid["SourceSpanish"], "Title": "Otro titulo"}}],
            "missing English objective": [{**valid, "English": {
                **valid["English"], "Objectives": []}}],
            "wrong objective ID": [{**valid, "English": {
                **valid["English"], "Objectives": [{
                    **valid["English"]["Objectives"][0], "ObjectiveID": "different-id"}]}}],
            "incomplete English rule": [{**valid, "English": {
                **valid["English"], "Rules": [""] + valid["English"]["Rules"][1:]}}],
        }
        for name, records in cases.items():
            with self.subTest(reason=name), tempfile.TemporaryDirectory() as tmp:
                translations = Path(tmp)
                write_manifest(translations, records)

                # Act
                error = capture_error(lambda: seed.translated_definitions(catalog, "staging", translations))

                # Assert
                self.assertIsInstance(error, ValueError, name)


if __name__ == "__main__":
    unittest.main()
