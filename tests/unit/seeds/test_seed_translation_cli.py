"""TEST-SEED-LANG-002: operator render includes reviewed-shape en snapshots."""

import json
import os
import subprocess
import sys
import unittest

from translation_fixture import ROOT, seed


class BilingualSeedRenderTest(unittest.TestCase):
    def test_racing_render_keeps_spanish_and_renders_english_on_all_three_rows(self):
        # Arrange
        catalog = seed.load_catalog(ROOT / "seeds/racing-missions.json")
        for environment, count in (("staging", 24), ("production", 8)):
            with self.subTest(environment=environment):
                spanish = seed.render(seed.definitions(catalog, environment), "test-season")
                command = [sys.executable, "-B", str(ROOT / "scripts/seed_missions.py"),
                           "--environment", environment, "--season-id", "test-season", "--render"]
                no_credentials = {key: value for key, value in os.environ.items()
                                  if not key.startswith("AWS_")}

                # Act
                result = subprocess.run(command, cwd=ROOT, env=no_credentials, capture_output=True, text=True)

                # Assert
                self.assertEqual(result.returncode, 0, result.stderr)
                items = json.loads(result.stdout)["items"]
                self.assertEqual(len(items), 3 * count)
                for offset in range(0, len(items), 3):
                    versions = items[offset:offset + 3]
                    self.assertEqual(versions[0]["ContentEn"], versions[1]["ContentEn"])
                    self.assertEqual(versions[1]["ContentEn"], versions[2]["ContentEn"])
                for original, bilingual in zip(spanish, items):
                    self.assertEqual({key: value for key, value in bilingual.items()
                                      if key != "ContentEn"}, original)
                    self.assertTrue(bilingual["ContentEn"]["M"]["Title"]["S"].strip())

    def test_explicit_pvp_render_is_bilingual_without_enrolling_it_in_default_seed(self):
        # Arrange
        catalog = seed.load_catalog(ROOT / "seeds/pvp-missions.json")
        spanish = seed.render(seed.definitions(catalog, "staging"), "test-season")
        command = [sys.executable, "-B", str(ROOT / "scripts/seed_missions.py"),
                   "--catalog", str(ROOT / "seeds/pvp-missions.json"),
                   "--environment", "staging", "--season-id", "test-season", "--render"]

        # Act
        result = subprocess.run(command, cwd=ROOT, capture_output=True, text=True)

        # Assert
        self.assertEqual(result.returncode, 0, result.stderr)
        items = json.loads(result.stdout)["items"]
        self.assertEqual(len(items), 36)
        for original, bilingual in zip(spanish, items):
            self.assertEqual({key: value for key, value in bilingual.items()
                              if key != "ContentEn"}, original)
            self.assertTrue(bilingual["ContentEn"]["M"]["Description"]["S"].strip())


if __name__ == "__main__":
    unittest.main()
