"""Atomic, fail-closed synchronization of a selected mission catalog."""

from seed_actions import definition_actions, offer_action
from seed_values import av, key, offer_item, versioned


def synchronize(aws, selected, apply=False, require_active=False):
    seasons = [row for row in aws.partition("SEASON") if row.get("State") == av("ACTIVE")]
    if len(seasons) > 1:
        raise ValueError("More than one active season; no catalog writes performed")
    if not seasons and require_active:
        raise ValueError("No active season; no catalog writes performed")
    season = seasons[0]["SK"]["S"] if seasons else None
    existing = {row["SK"]["S"]: row for row in aws.partition("DEFINITION")}
    actions = []
    if season:
        actions.append({"ConditionCheck": {
            "TableName": aws.table, "Key": key("SEASON", season),
            "ConditionExpression": "#state = :active",
            "ExpressionAttributeNames": {"#state": "State"},
            "ExpressionAttributeValues": {":active": av("ACTIVE")},
        }})
    for definition in selected:
        code = definition["SK"]["S"]
        actions.extend(definition_actions(aws.table, definition, existing.get(code),
                                         existing.get(versioned(definition)["SK"]["S"])))
        if season:
            candidate = offer_item(definition, season)
            stored_offer = aws.offer(candidate)
            previous = None
            if stored_offer and stored_offer.get("DefinitionVersion") != candidate["DefinitionVersion"]:
                previous_version = stored_offer.get("DefinitionVersion", {}).get("S")
                previous_definition = existing.get(f"{code}#{previous_version}")
                if previous_definition is None:
                    raise ValueError(f"Previous offer definition missing for {code}#{previous_version}")
                previous = offer_item(previous_definition, season)
            actions.append(offer_action(aws.table, candidate, stored_offer, previous))
    if len(actions) > 100:
        raise ValueError("Catalog exceeds 100 transaction operations; split into explicit catalogs")
    if apply:
        aws.call("transact-write-items", {"TransactItems": actions})
    return {"applied": apply, "table": aws.table, "season_id": season,
            "definitions": len(selected), "offers": len(selected) if season else 0,
            "status": "synchronized" if season else "definitions_only_no_active_season"}
