"""In-memory DynamoDB transport contract for local-only scenario loads."""

import sys
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "scripts"))
from season_scenario_local import (ATTRIBUTES, INDEXES, KEYS, create_preview_table,
                                   insert_missing, scan_rows, table_exists)
from seed_values import item


class LocalAWSFake:
    def __init__(self, responses):
        self.table = "hbk-missions-preview-mid-season"
        self.responses = iter(responses)
        self.calls = []

    def call(self, operation, payload):
        self.calls.append((operation, payload))
        return next(self.responses)


class ScenarioTransportTest(unittest.TestCase):
    def test_existing_table_is_found_across_pages(self):
        # Arrange
        aws = LocalAWSFake([{"TableNames": ["unrelated"], "LastEvaluatedTableName": "unrelated"},
                            {"TableNames": ["hbk-missions-preview-mid-season"]}])
        # Act
        found = table_exists(aws)
        # Assert
        self.assertTrue(found)
        self.assertEqual(aws.calls[1][1]["ExclusiveStartTableName"], "unrelated")

    def test_scan_reads_every_page(self):
        # Arrange
        first = item({"PK": "SEASON", "SK": "first"})
        second = item({"PK": "SEASON", "SK": "second"})
        aws = LocalAWSFake([{"Items": [first], "LastEvaluatedKey": {"PK": first["PK"], "SK": first["SK"]}},
                            {"Items": [second]}])
        # Act
        rows = scan_rows(aws)
        # Assert
        self.assertEqual(rows, [first, second])
        self.assertEqual(aws.calls[1][1]["ExclusiveStartKey"], {"PK": first["PK"], "SK": first["SK"]})

    def test_create_table_uses_readers_schema_and_waits_until_active(self):
        # Arrange
        table = {"TableStatus": "ACTIVE", "KeySchema": KEYS,
                 "AttributeDefinitions": ATTRIBUTES, "GlobalSecondaryIndexes": INDEXES}
        aws = LocalAWSFake([{}, {"Table": table}, {"Table": table}])
        # Act
        create_preview_table(aws)
        # Assert
        self.assertEqual([name for name, _ in aws.calls], ["create-table", "describe-table", "describe-table"])
        self.assertEqual(aws.calls[0][1]["GlobalSecondaryIndexes"], INDEXES)
        self.assertEqual(aws.calls[0][1]["BillingMode"], "PAY_PER_REQUEST")

    def test_insert_batches_conditional_puts_without_replacement(self):
        # Arrange
        aws = LocalAWSFake([{}, {}])
        rows = [item({"PK": "OFFER", "SK": f"preview-{index}"}) for index in range(101)]
        # Act
        insert_missing(aws, rows)
        # Assert
        self.assertEqual([len(call[1]["TransactItems"]) for call in aws.calls], [100, 1])
        self.assertTrue(all(action["Put"]["ConditionExpression"] ==
                            "attribute_not_exists(PK) AND attribute_not_exists(SK)"
                            for _, call in aws.calls for action in call["TransactItems"]))
