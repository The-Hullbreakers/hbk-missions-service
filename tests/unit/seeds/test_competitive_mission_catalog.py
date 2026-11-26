"""TEST-SEED-LANG-004: competitive LIVE catalog renders bilingual three-row items."""

import json
import subprocess
import sys
import unittest

from translation_fixture import ROOT, seed

CATALOG = ROOT / "seeds/competitive-missions.json"

EXPECTED = {
    "mining-live-first-strike": ("MINING", 60, 20, 200),
    "mining-live-ore-sampler": ("MINING", 90, 30, 300),
    "arena-live-pirate-swarm": ("ARENA_COMMANDER", 150, 50, 500),
    "arena-live-duel": ("ARENA_COMMANDER", 120, 40, 400),
    "salvage-live-hull-to-hull": ("SALVAGE", 90, 30, 300),
    "hauling-live-service-route": ("HAULING", 90, 30, 300),
}


def render(environment):
    command = [sys.executable, "-B", str(ROOT / "scripts/seed_missions.py"),
               "--catalog", str(CATALOG), "--environment", environment,
               "--season-id", "test-season", "--render"]
    return subprocess.run(command, cwd=ROOT, capture_output=True, text=True)


class CompetitiveMissionCatalogTest(unittest.TestCase):
    def test_staging_render_is_bilingual_with_six_distinct_missions(self):
        # Arrange
        catalog = seed.load_catalog(CATALOG)
        spanish = seed.render(seed.definitions(catalog, "staging"), "test-season")

        # Act
        result = render("staging")

        # Assert
        self.assertEqual(result.returncode, 0, result.stderr)
        items = json.loads(result.stdout)["items"]
        self.assertEqual(len(items), 18)
        for offset in range(0, len(items), 3):
            versions = items[offset:offset + 3]
            self.assertEqual(versions[0]["ContentEn"], versions[1]["ContentEn"])
            self.assertEqual(versions[1]["ContentEn"], versions[2]["ContentEn"])
        for original, bilingual in zip(spanish, items):
            self.assertEqual({key: value for key, value in bilingual.items()
                              if key != "ContentEn"}, original)
            self.assertTrue(bilingual["ContentEn"]["M"]["Title"]["S"].strip())
            self.assertTrue(bilingual["ContentEn"]["M"]["Description"]["S"].strip())

    def test_six_missions_keep_metadata_merits_and_three_unit_cap(self):
        # Arrange
        catalog = seed.load_catalog(CATALOG)
        rows = seed.definitions(catalog, "staging")

        # Act
        codes = [row["DefinitionCode"]["S"] for row in rows]

        # Assert
        self.assertEqual(sorted(codes), sorted(EXPECTED))
        self.assertEqual(len(set(codes)), 6)
        for row in rows:
            code = row["DefinitionCode"]["S"]
            with self.subTest(code=code):
                category, per_unit, completion, total = EXPECTED[code]
                self.assertEqual(row["Category"]["S"], category)
                self.assertEqual(row["Mode"]["S"], "SOLO")
                self.assertEqual(row["State"]["S"], "PUBLISHED")
                self.assertEqual(row["Version"]["S"], "1")
                self.assertEqual(int(row["CompletionReward"]["N"]), completion)
                objectives = row["Objectives"]["L"]
                self.assertEqual(len(objectives), 1)
                objective = objectives[0]["M"]
                self.assertEqual(objective["Category"]["S"], category)
                self.assertEqual(int(objective["UnitLimit"]["N"]), 3)
                self.assertEqual(int(objective["Bonus"]["N"]), 0)
                self.assertEqual(int(objective["MeritPerUnit"]["N"]), per_unit)
                self.assertEqual(per_unit * 3 + completion, total)
                self.assertEqual(len(row["Rules"]["L"]), 5)
                first_rule = row["Rules"]["L"][0]["S"]
                self.assertIn("LIVE", first_rule)
                self.assertIn("PTU", first_rule)
                if category == "ARENA_COMMANDER":
                    self.assertIn("Arena Commander", first_rule)
                else:
                    self.assertIn("universo persistente", first_rule)

    def test_production_rejects_the_competitive_catalog(self):
        # Arrange/Act
        result = render("production")

        # Assert
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("not enabled", result.stderr)


if __name__ == "__main__":
    unittest.main()
