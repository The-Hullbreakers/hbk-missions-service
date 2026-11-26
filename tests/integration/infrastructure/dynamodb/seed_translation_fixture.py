"""Private loopback DynamoDB Local fixture for TEST-SEED-LANG-002."""

import copy
import json
import os
from pathlib import Path
import sys
from urllib.parse import urlparse
import uuid

ROOT = Path(__file__).resolve().parents[4]
sys.path.insert(0, str(ROOT / "scripts"))
import seed_missions as seed  # noqa: E402

SEASON = "translation-test-season"


def new_local_table():
    endpoint = os.environ["DYNAMODB_SEED_TEST_ENDPOINT"]
    parsed = urlparse(endpoint)
    if (parsed.scheme != "http"
            or parsed.hostname not in ("localhost", "127.0.0.1", "::1")
            or parsed.port != 18000 or parsed.path not in ("", "/")
            or parsed.username or parsed.password):
        raise ValueError("Seed tests require the dedicated HTTP loopback DynamoDB Local port 18000")
    os.environ.update(AWS_ACCESS_KEY_ID="dummy", AWS_SECRET_ACCESS_KEY="dummy",
                      AWS_EC2_METADATA_DISABLED="true")
    os.environ.pop("AWS_SESSION_TOKEN", None)
    aws = seed.AWS("hbk-seed-lang-" + uuid.uuid4().hex, "us-east-1", endpoint)
    aws.call("create-table", {
        "TableName": aws.table, "BillingMode": "PAY_PER_REQUEST",
        "AttributeDefinitions": [{"AttributeName": field, "AttributeType": "S"}
                                 for field in ("PK", "SK")],
        "KeySchema": [{"AttributeName": "PK", "KeyType": "HASH"},
                      {"AttributeName": "SK", "KeyType": "RANGE"}],
    })
    put_new(aws, seed.item({"PK": "SEASON", "SK": SEASON, "State": "ACTIVE"}))
    return aws


def spanish_and_bilingual():
    catalog = copy.deepcopy(seed.load_catalog(ROOT / "seeds/pvp-missions.json"))
    catalog["missions"] = catalog["missions"][:1]
    spanish = seed.definitions(catalog, "staging")[0]
    records = json.loads((ROOT / "seeds/translations/pvp-fps-v4.en.json").read_text(
        encoding="utf-8"))["translations"]
    english = next(record["English"] for record in records
                   if record["DefinitionCode"] == spanish["DefinitionCode"]["S"])
    return spanish, {**copy.deepcopy(spanish), "ContentEn": seed.av(english)}


def put_new(aws, row):
    aws.call("put-item", {"TableName": aws.table, "Item": row,
                          "ConditionExpression": "attribute_not_exists(PK)"})


def replace(aws, row):
    aws.call("put-item", {"TableName": aws.table, "Item": row,
                          "ConditionExpression": "attribute_exists(PK)"})


def get(aws, row):
    result = aws.call("get-item", {
        "TableName": aws.table, "ConsistentRead": True,
        "Key": {field: row[field] for field in ("PK", "SK")},
    })
    return result.get("Item")


def install_existing(aws, definition):
    for row in (definition, seed.versioned(definition), seed.offer_item(definition, SEASON)):
        put_new(aws, row)


def snapshot(aws):
    return {
        (row["PK"]["S"], row["SK"]["S"]): row
        for name in ("DEFINITION", "OFFER", "CONTRACT")
        for row in aws.partition(name)
    }


def capture_error(operation):
    try:
        operation()
    except (ValueError, RuntimeError) as error:
        return error
    return None
