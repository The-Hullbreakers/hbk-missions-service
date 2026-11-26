"""AWS CLI transport for the offline mission catalog seed."""

import json
import os
from pathlib import Path
import subprocess
import tempfile

from seed_values import av


class AWS:
    def __init__(self, table, region, endpoint=None):
        self.table = table
        self.options = ["--region", region]
        if endpoint:
            self.options += ["--endpoint-url", endpoint]

    def call(self, operation, payload):
        # File input avoids shell interpolation and command-line size limits.
        with tempfile.TemporaryDirectory(prefix="hbk-missions-seed-") as directory:
            path = Path(directory) / "request.json"
            path.write_text(json.dumps(payload), encoding="utf-8")
            response = subprocess.run(
                ["aws", "dynamodb", operation, "--cli-input-json", f"file://{path}",
                 "--output", "json", "--no-cli-pager"] + self.options,
                capture_output=True, text=True, env={**os.environ, "AWS_PAGER": ""},
                timeout=120,
            )
        if response.returncode:
            raise RuntimeError(f"AWS {operation} failed: {response.stderr.strip()}")
        return json.loads(response.stdout or "{}")

    def partition(self, pk):
        request = {"TableName": self.table, "ConsistentRead": True,
                   "KeyConditionExpression": "PK = :pk", "ExpressionAttributeValues": {":pk": av(pk)}}
        rows = []
        while True:
            response = self.call("query", request)
            rows.extend(response.get("Items", []))
            cursor = response.get("LastEvaluatedKey")
            if not cursor:
                return rows
            request["ExclusiveStartKey"] = cursor

    def offer(self, offer):
        result = self.call("get-item", {
            "TableName": self.table, "ConsistentRead": True,
            "Key": {field: offer[field] for field in ("PK", "SK")},
        })
        return result.get("Item")
