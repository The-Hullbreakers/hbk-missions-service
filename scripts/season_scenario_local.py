"""Read and conditionally write fixed preview tables on DynamoDB Local."""

import time

from seed_values import av

KEYS = [{"AttributeName": "PK", "KeyType": "HASH"},
        {"AttributeName": "SK", "KeyType": "RANGE"}]
ATTRIBUTES = [{"AttributeName": field, "AttributeType": "S"}
              for field in ("PK", "SK", "GSI1PK", "GSI1SK", "GSI2PK", "GSI2SK")]
INDEXES = [{"IndexName": name,
            "KeySchema": [{"AttributeName": name + "PK", "KeyType": "HASH"},
                          {"AttributeName": name + "SK", "KeyType": "RANGE"}],
            "Projection": {"ProjectionType": "ALL"}}
           for name in ("GSI1", "GSI2")]


def table_exists(aws):
    request = {"Limit": 100}
    while True:
        response = aws.call("list-tables", request)
        if aws.table in response.get("TableNames", []):
            return True
        cursor = response.get("LastEvaluatedTableName")
        if not cursor:
            return False
        request["ExclusiveStartTableName"] = cursor


def scan_rows(aws, **extra):
    request = {"TableName": aws.table, "ConsistentRead": True, **extra}
    rows = []
    while True:
        response = aws.call("scan", request)
        rows.extend(response.get("Items", []))
        cursor = response.get("LastEvaluatedKey")
        if not cursor:
            return rows
        request["ExclusiveStartKey"] = cursor


def active_users(aws):
    rows = scan_rows(aws, FilterExpression="begins_with(#pk, :prefix) AND "
                     "begins_with(#sk, :meta) AND #status = :active",
                     ExpressionAttributeNames={"#pk": "PK", "#sk": "SK", "#status": "Status"},
                     ExpressionAttributeValues={":prefix": av("USER#"),
                                                ":meta": av("METADATA#"),
                                                ":active": av("active")},
                     ProjectionExpression="#pk,#sk,#status")
    return sorted({pk[5:] for row in rows
                   if (pk := row["PK"]["S"]).startswith("USER#")
                   and row["SK"]["S"] == "METADATA#" + pk[5:]})


def validate_schema(aws):
    table = aws.call("describe-table", {"TableName": aws.table})["Table"]
    indexes = table.get("GlobalSecondaryIndexes", [])
    if (table.get("TableStatus") != "ACTIVE"
            or table.get("KeySchema") != KEYS
            or sorted(table.get("AttributeDefinitions", []), key=lambda x: x["AttributeName"]) !=
            sorted(ATTRIBUTES, key=lambda x: x["AttributeName"])
            or len(indexes) != 2
            or any(next((index for index in indexes if index["IndexName"] == expected["IndexName"]),
                        {}).get("KeySchema") != expected["KeySchema"] or
                   next((index for index in indexes if index["IndexName"] == expected["IndexName"]),
                        {}).get("Projection", {}).get("ProjectionType") != "ALL"
                   for expected in INDEXES)):
        raise ValueError("Preview table schema differs from the missions table")


def create_preview_table(aws):
    aws.call("create-table", {"TableName": aws.table, "BillingMode": "PAY_PER_REQUEST",
                              "AttributeDefinitions": ATTRIBUTES, "KeySchema": KEYS,
                              "GlobalSecondaryIndexes": INDEXES})
    for _ in range(50):
        table = aws.call("describe-table", {"TableName": aws.table})["Table"]
        if table.get("TableStatus") == "ACTIVE":
            validate_schema(aws)
            return
        time.sleep(0.2)
    raise RuntimeError("Preview table did not become active")


def insert_missing(aws, rows):
    for start in range(0, len(rows), 100):
        actions = [{"Put": {"TableName": aws.table, "Item": row,
                            "ConditionExpression": "attribute_not_exists(PK) AND attribute_not_exists(SK)"}}
                   for row in rows[start:start + 100]]
        aws.call("transact-write-items", {"TransactItems": actions})
