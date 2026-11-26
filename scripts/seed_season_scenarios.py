#!/usr/bin/env python3
"""Load isolated off-season, fresh-start or mid-season previews on DynamoDB Local."""

import argparse
from collections import Counter
import json
import os
from pathlib import Path
import sys

from seed_aws import AWS
from seed_catalog import load_catalog
from seed_translations import translated_definitions
from season_scenario_local import active_users, create_preview_table, insert_missing, scan_rows, table_exists, validate_schema
from season_scenario_model import PROFILES, build_scenario, missing_items, validate_target

ROOT = Path(__file__).resolve().parents[1]


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--scenario", choices=PROFILES, required=True)
    parser.add_argument("--endpoint-url", default="http://localhost:8000")
    parser.add_argument("--apply", action="store_true", help="Insert; otherwise preview only")
    args = parser.parse_args(argv)
    table = validate_target(args.endpoint_url, args.scenario)
    if table is None:
        parser.error("Only fixed preview tables on HTTP loopback DynamoDB Local are allowed")
    os.environ.update(AWS_ACCESS_KEY_ID="dummy", AWS_SECRET_ACCESS_KEY="dummy",
                      AWS_EC2_METADATA_DISABLED="true", AWS_DEFAULT_REGION="us-east-1")
    os.environ.pop("AWS_SESSION_TOKEN", None)
    os.environ.pop("AWS_PROFILE", None)
    definitions = []
    for name in ("racing-missions.json", "pvp-missions.json"):
        definitions.extend(translated_definitions(load_catalog(ROOT / "seeds" / name),
                                                  "staging", ROOT / "seeds/translations"))
    user_ids = active_users(AWS("hbk-users", "us-east-1", args.endpoint_url)) if args.scenario == "mid-season" else []
    desired = build_scenario(args.scenario, definitions, user_ids)
    if desired is None:
        raise ValueError("Scenario requires valid catalog and at least three active members")
    destination = AWS(table, "us-east-1", args.endpoint_url)
    exists = table_exists(destination)
    if exists:
        validate_schema(destination)
    existing = scan_rows(destination) if exists else []
    pending = missing_items(existing, desired)
    if pending is None:
        raise ValueError("Preview contains unexpected or divergent rows; no changes applied")
    if args.apply:
        if not exists:
            create_preview_table(destination)
        # An identity change during a load must fail before the first write.
        if args.scenario == "mid-season" and user_ids != active_users(AWS("hbk-users", "us-east-1", args.endpoint_url)):
            raise ValueError("Active members changed during preview generation")
        insert_missing(destination, pending)
        if missing_items(scan_rows(destination), desired) != []:
            raise RuntimeError("Preview verification failed")
    counts = dict(sorted(Counter(row["PK"]["S"] for row in desired).items()))
    print(json.dumps({"scenario": args.scenario, "table": table, "applied": args.apply,
                      "counts": counts, "pending": len(pending)}, sort_keys=True))


if __name__ == "__main__":
    try:
        main()
    except (ValueError, RuntimeError, OSError) as error:
        print("Local preview rejected; check endpoint, active members, schema, or conflicting rows",
              file=sys.stderr)
        sys.exit(1)
