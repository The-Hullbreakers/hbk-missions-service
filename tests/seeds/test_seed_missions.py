import copy
import importlib.util
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
spec = importlib.util.spec_from_file_location('seed_missions', ROOT / 'scripts/seed_missions.py')
seed = importlib.util.module_from_spec(spec)
spec.loader.exec_module(seed)


class CatalogTest(unittest.TestCase):
    def setUp(self):
        self.catalog = seed.load_catalog(ROOT / 'seeds/racing-missions.json')

    def test_catalog_counts_and_consecutive_lap_scoring(self):
        staging = seed.definitions(self.catalog, 'staging')
        production = seed.definitions(self.catalog, 'production')
        self.assertEqual((len(staging), len(production)), (24, 8))
        for d in staging:
            self.assertEqual(d['Mode'], {'S': 'SOLO'})
            self.assertEqual(d['Category'], {'S': 'RACING'})
            self.assertIn('cinco vueltas', d['Description']['S'])
            objectives = [x['M'] for x in d['Objectives']['L']]
            self.assertEqual(len(objectives), 1)
            self.assertTrue(all(x['UnitLimit'] == {'N': '5'} for x in objectives))
            total = (int(objectives[0]['MeritPerUnit']['N']) * 5
                     + int(d['CompletionReward']['N']))
            self.assertIn(total, [200, 400, 700, 1100])

    def test_duplicate_and_invalid_values_rejected(self):
        self.catalog['series'].append(copy.deepcopy(self.catalog['series'][0]))
        with self.assertRaises(ValueError):
            seed.definitions(self.catalog, 'staging')

    def test_changed_immutable_version_rejected(self):
        d = seed.definitions(self.catalog, 'staging')[0]
        old = copy.deepcopy(d)
        old['Title'] = {'S': 'Previously published'}
        with self.assertRaisesRegex(ValueError, 'version'):
            seed.definition_actions('table', d, old, old)

    def test_downgrade_rejected(self):
        d = seed.definitions(self.catalog, 'staging')[0]
        old = copy.deepcopy(d)
        old['Version'] = {'S': str(int(d['Version']['S']) + 1)}
        with self.assertRaisesRegex(ValueError, 'downgrade'):
            seed.definition_actions('table', d, old, None)

    def test_offer_carries_no_capacity_fields(self):
        d = seed.definitions(self.catalog, 'staging')[0]
        offer = seed.offer_item(d, 'season-1')
        self.assertNotIn('Capacity', offer)
        self.assertNotIn('remaining_capacity', offer)
        expression = seed.offer_action('table', offer)['Update']['UpdateExpression']
        self.assertNotIn('if_not_exists', expression)

    def test_offer_item_carries_frozen_objectives(self):
        d = seed.definitions(self.catalog, 'staging')[0]
        offer = seed.offer_item(d, 'season-1')
        self.assertEqual(offer['Objectives'], d['Objectives'])

    def test_render_has_three_items_per_mission(self):
        items = seed.render(seed.definitions(self.catalog, 'production'), 'season-1')
        self.assertEqual(len(items), 24)
        self.assertEqual(len({(x['PK']['S'], x['SK']['S']) for x in items}), 24)

    def test_pvp_live_catalog(self):
        catalog = seed.load_catalog(ROOT / 'seeds/pvp-missions.json')
        rows = seed.definitions(catalog, 'staging')
        self.assertEqual(len(rows), 12)
        self.assertEqual(seed.definitions(catalog, 'production'), rows)
        self.assertEqual(len(seed.render(rows, 'local-season')), 36)
        totals = []
        for row in rows:
            self.assertNotIn('Exclusivamente', row['Description']['S'])
            rules = [r['S'] for r in row['Rules']['L']]
            self.assertGreaterEqual(len(rules), 5)
            self.assertLessEqual(len(rules), 6)
            self.assertTrue(all(rule.strip() for rule in rules))
            self.assertTrue(any('Exclusivamente en el universo persistente PU/LIVE' in rule for rule in rules))
            self.assertTrue(any('video continuo' in rule for rule in rules))
            objectives = [o['M'] for o in row['Objectives']['L']]
            self.assertEqual(len(objectives), 1)
            per_unit = int(objectives[0]['MeritPerUnit']['N'])
            limit = int(objectives[0]['UnitLimit']['N'])
            totals.append(per_unit * limit + int(row['CompletionReward']['N']))
        self.assertEqual(totals, [300, 600, 1200, 3000, 500, 1000, 2000, 5000, 200, 600, 1000, 2000])
        catalog['missions'][0]['Objectives'].append(
            dict(catalog['missions'][0]['Objectives'][0]))
        with self.assertRaisesRegex(ValueError, 'duplicate objective'):
            seed.definitions(catalog, 'staging')


if __name__ == '__main__':
    unittest.main()
