#!/usr/bin/env python3
"""Seed racing and PU/LIVE PvP missions into the local application's database."""

import argparse
import json
import os
import sys
from urllib.parse import urlparse

from seed_missions import (
    AWS, ROOT, av, key, load_catalog, offer_item, synchronize,
    translated_definitions, versioned,
)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--endpoint-url', default='http://localhost:8000')
    parser.add_argument('--table', default='hbk-missions-local')
    parser.add_argument('--region', default='us-east-1')
    parser.add_argument('--apply', action='store_true', help='Write; default only validates remote state')
    args = parser.parse_args()
    if urlparse(args.endpoint_url).hostname not in ('localhost', '127.0.0.1', '::1'):
        parser.error('This bootstrap script only supports loopback DynamoDB Local')
    # DynamoDB Local authentication only; never use cloud credentials here.
    os.environ.update(AWS_ACCESS_KEY_ID='dummy', AWS_SECRET_ACCESS_KEY='dummy',
                      AWS_EC2_METADATA_DISABLED='true')
    os.environ.pop('AWS_SESSION_TOKEN', None)
    aws = AWS(args.table, args.region, args.endpoint_url)
    catalogs = [(name, translated_definitions(load_catalog(ROOT / 'seeds' / name),
                                              'staging', ROOT / 'seeds/translations'))
                for name in ('racing-missions.json', 'pvp-missions.json')]
    # Validate both catalogs before bootstrap or catalog writes.
    for _, selected in catalogs:
        synchronize(aws, selected, apply=False)
    seasons = aws.partition('SEASON')
    active = [s for s in seasons if s.get('State') == av('ACTIVE')]
    if not active and seasons:
        raise ValueError('Existing inactive seasons found; activate a season explicitly before seeding')
    if not active and args.apply:
        season = {**key('SEASON', 'local-training'), 'State': av('ACTIVE'), 'Official': {'BOOL': False}}
        aws.call('put-item', {'TableName': args.table, 'Item': season,
                             'ConditionExpression': 'attribute_not_exists(PK)'})
    season_id = active[0]['SK']['S'] if active else 'local-training' if args.apply else None
    before_definitions = {d['SK']['S']: d for d in aws.partition('DEFINITION')} if args.apply else {}
    before_offers = {}
    if args.apply:
        for _, selected in catalogs:
            for row in selected:
                candidate = offer_item(row, season_id)
                before_offers[candidate['SK']['S']] = aws.offer(candidate)
    results = []
    for name, selected in catalogs:
        result = synchronize(aws, selected, apply=args.apply, require_active=args.apply)
        result['catalog'] = name
        results.append(result)
    if args.apply:
        stored = {d['SK']['S']: d for d in aws.partition('DEFINITION')}
        for (_, selected), result in zip(catalogs, results):
            for row in selected:
                prior = before_definitions.get(row['SK']['S'])
                expected_current = prior if prior and prior['Version'] == row['Version'] else row
                if stored.get(row['SK']['S']) != expected_current:
                    raise RuntimeError(f'Definition verification failed: {row["SK"]}')
                historical = versioned(row)
                expected_history = before_definitions.get(historical['SK']['S'], historical)
                if stored.get(historical['SK']['S']) != expected_history:
                    raise RuntimeError('Version verification failed')
                expected = offer_item(row, result['season_id'])
                actual = aws.offer(expected) or {}
                prior_offer = before_offers.get(expected['SK']['S'])
                if prior_offer and prior_offer['DefinitionVersion'] == expected['DefinitionVersion']:
                    if actual != prior_offer:
                        raise RuntimeError(f'Offer verification failed: {expected["SK"]}')
                elif any(actual.get(field) != value for field, value in expected.items()):
                    raise RuntimeError(f'Offer verification failed: {expected["SK"]}')
    print(json.dumps({'applied': args.apply, 'endpoint': args.endpoint_url,
                      'table': args.table, 'catalogs': results}, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    try:
        main()
    except (ValueError, RuntimeError) as error:
        print(f'Local seed failed: {error}', file=sys.stderr)
        sys.exit(1)
