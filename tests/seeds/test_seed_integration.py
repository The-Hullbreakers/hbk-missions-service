"""Run against a new isolated DynamoDB Local table; leave it in place."""

import copy
import os
import unittest
from urllib.parse import urlparse
import uuid

from test_seed_missions import ROOT, seed


def with_test_english(catalog, previous):
    translations = {row["DefinitionCode"]["S"]: row["ContentEn"] for row in previous}
    return [{**row, "ContentEn": copy.deepcopy(translations[row["DefinitionCode"]["S"]])}
            for row in seed.definitions(catalog, "staging")]


@unittest.skipUnless(os.getenv("DYNAMODB_SEED_TEST_ENDPOINT"), "DynamoDB Local endpoint not configured")
class SeedIntegrationTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        endpoint = os.environ["DYNAMODB_SEED_TEST_ENDPOINT"]
        parsed = urlparse(endpoint)
        if (parsed.scheme != "http"
                or parsed.hostname not in ("localhost", "127.0.0.1", "::1")
                or parsed.port != 18000):
            raise ValueError("Seed integration tests require the dedicated loopback port 18000")
        os.environ.update(AWS_ACCESS_KEY_ID="dummy", AWS_SECRET_ACCESS_KEY="dummy",
                          AWS_EC2_METADATA_DISABLED="true")
        os.environ.pop("AWS_SESSION_TOKEN", None)
        cls.aws = seed.AWS("hbk-seed-test-" + uuid.uuid4().hex, "us-east-1", endpoint)
        cls.aws.call("create-table", {
            "TableName": cls.aws.table, "BillingMode": "PAY_PER_REQUEST",
            "AttributeDefinitions": [{"AttributeName": k, "AttributeType": "S"} for k in ("PK", "SK")],
            "KeySchema": [{"AttributeName": "PK", "KeyType": "HASH"},
                          {"AttributeName": "SK", "KeyType": "RANGE"}],
        })

    def put(self, row):
        self.aws.call("put-item", {"TableName": self.aws.table, "Item": row})

    def get(self, row):
        return self.aws.call("get-item", {"TableName": self.aws.table, "ConsistentRead": True,
                             "Key": {k: row[k] for k in ("PK", "SK")}}).get("Item")

    def test_deployment_lifecycle(self):
        catalog = seed.load_catalog(ROOT / "seeds/racing-missions.json")
        selected = seed.translated_definitions(catalog, "staging", ROOT / "seeds/translations")

        # A clean installation has definitions but no invented season/offers.
        result = seed.synchronize(self.aws, selected, apply=True)
        self.assertEqual(result["status"], "definitions_only_no_active_season")
        self.assertEqual(len(self.aws.partition("DEFINITION")), 48)
        self.assertEqual(self.aws.partition("OFFER"), [])
        with self.assertRaisesRegex(ValueError, "No active season"):
            seed.synchronize(self.aws, selected, apply=True, require_active=True)

        season = seed.item({"PK": "SEASON", "SK": "test-season-1", "State": "ACTIVE", "Official": "test"})
        self.put(season)
        seed.synchronize(self.aws, selected, apply=True)
        self.assertEqual(len(self.aws.partition("OFFER")), 24)
        offer = seed.offer_item(selected[0], "test-season-1")
        seed.synchronize(self.aws, selected, apply=True)
        self.assertEqual(self.get(offer), offer)

        # A manual contract sentinel must survive every seed operation untouched.
        contract = seed.item({"PK": "CONTRACT", "SK": "test-contract", "DefinitionVersion": "1"})
        self.put(contract)
        changed = copy.deepcopy(catalog)
        changed["levels"][0]["merits"] = 300
        with self.assertRaisesRegex(ValueError, "version"):
            seed.synchronize(self.aws, seed.definitions(changed, "staging"), apply=True)
        changed["version"] = catalog["version"] + 1
        new_selected = with_test_english(changed, selected)
        seed.synchronize(self.aws, new_selected, apply=False)
        self.assertEqual(self.get(offer)["DefinitionVersion"], seed.av(str(catalog["version"])))
        seed.synchronize(self.aws, new_selected, apply=True)
        updated = self.get(offer)
        self.assertEqual(updated["DefinitionVersion"], seed.av(str(changed["version"])))
        self.assertNotIn("Capacity", updated)
        self.assertNotIn("remaining_capacity", updated)
        self.assertEqual(self.get(seed.versioned(selected[0])), seed.versioned(selected[0]))
        self.assertEqual(self.get(contract), contract)
        with self.assertRaisesRegex(ValueError, "downgrade"):
            seed.synchronize(self.aws, selected, apply=True)

        # New seasons get new offers; prior seasons keep their own offer rows.
        season["State"] = seed.av("FINALIZED")
        self.put(season)
        next_season = seed.item({"PK": "SEASON", "SK": "test-season-2", "State": "ACTIVE"})
        self.put(next_season)
        seed.synchronize(self.aws, new_selected, apply=True)
        self.assertEqual(len(self.aws.partition("OFFER")), 48)
        self.assertEqual(self.get(offer), updated)
        fresh = self.get(seed.offer_item(new_selected[0], "test-season-2"))
        self.assertEqual(fresh, seed.offer_item(new_selected[0], "test-season-2"))

        # Lifecycle transitions between read and write cancel the entire transaction.
        real_call = self.aws.call

        def freeze_before_transaction(operation, payload):
            if operation == "transact-write-items":
                next_season["State"] = seed.av("FROZEN")
                real_call("put-item", {"TableName": self.aws.table, "Item": next_season})
            return real_call(operation, payload)

        self.aws.call = freeze_before_transaction
        changed["version"] += 1
        frozen_selected = with_test_english(changed, new_selected)
        try:
            with self.assertRaisesRegex(RuntimeError, "TransactionCanceledException"):
                seed.synchronize(self.aws, frozen_selected, apply=True)
        finally:
            self.aws.call = real_call
        self.assertIsNone(self.get(seed.versioned(frozen_selected[0])))
        self.assertEqual(self.get(new_selected[0]), new_selected[0])

        next_season["State"] = seed.av("ACTIVE")
        season["State"] = seed.av("ACTIVE")
        self.put(next_season)
        self.put(season)
        with self.assertRaisesRegex(ValueError, "More than one active season"):
            seed.synchronize(self.aws, new_selected, apply=True)


if __name__ == "__main__":
    unittest.main()
