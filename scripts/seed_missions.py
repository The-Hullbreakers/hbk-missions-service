#!/usr/bin/env python3
"""Render or atomically synchronize a versioned mission catalog using AWS CLI v2."""

import argparse
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(Path(__file__).resolve().parent))

from seed_actions import definition_actions, offer_action
from seed_aws import AWS
from seed_catalog import definitions, explicit_definitions, load_catalog, positive, rule_list
from seed_sync import synchronize
from seed_translations import translated_definitions
from seed_values import av, item, key, offer_item, render, unchanged, versioned


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--catalog", type=Path, default=ROOT / "seeds/racing-missions.json")
    parser.add_argument("--environment", choices=("staging", "production"), required=True)
    parser.add_argument("--table")
    parser.add_argument("--region", default=os.getenv("AWS_DEFAULT_REGION", "us-east-2"))
    parser.add_argument("--endpoint-url")
    parser.add_argument("--season-id", help="For offline rendering only; apply discovers the active season")
    parser.add_argument("--require-active-season", action="store_true")
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--render", action="store_true", help="Print AWS attribute-value JSON, without AWS calls")
    mode.add_argument("--dry-run", action="store_true", help="Read and validate remote state; do not write")
    mode.add_argument("--apply", action="store_true")
    args = parser.parse_args()
    selected = translated_definitions(load_catalog(args.catalog), args.environment,
                                      ROOT / "seeds/translations")
    if args.render:
        if not args.season_id:
            parser.error("--render requires --season-id")
        output = {"items": render(selected, args.season_id)}
    else:
        if not args.table or args.season_id:
            parser.error("--apply/--dry-run require --table and discover season automatically")
        output = synchronize(AWS(args.table, args.region, args.endpoint_url), selected,
                             args.apply, args.require_active_season)
        if output["season_id"] is None:
            print("WARNING: No active season. Offers remain pending; rerun after activating a season.", file=sys.stderr)
    print(json.dumps(output, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    try:
        main()
    except (ValueError, KeyError, RuntimeError, OSError, subprocess.TimeoutExpired) as error:
        print(f"Seed failed: {error}", file=sys.stderr)
        sys.exit(1)
