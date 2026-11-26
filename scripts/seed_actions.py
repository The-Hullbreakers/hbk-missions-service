"""Conditional writes for versioned definitions and frozen offers."""

import copy

from seed_translations import require_complete_english
from seed_values import key, unchanged, versioned


def without_english(row):
    return {field: value for field, value in row.items() if field != "ContentEn"}


def compare_existing(incoming, stored, label):
    if without_english(incoming) != without_english(stored):
        raise ValueError(f"{label} differs from the stored Spanish or economics")
    if ("ContentEn" in incoming and "ContentEn" in stored
            and incoming["ContentEn"] != stored["ContentEn"]):
        raise ValueError(f"{label} has conflicting English")


def definition_actions(table, definition, current, historical):
    expected = versioned(definition)
    code = definition["SK"]["S"]
    same_version = current is not None and current["Version"] == definition["Version"]
    if current and int(current["Version"]["S"]) > int(definition["Version"]["S"]):
        raise ValueError(f"Refusing downgrade: {code}")
    if not same_version or "ContentEn" in definition:
        require_complete_english(definition)
    if historical:
        compare_existing(expected, historical, f"Immutable version {expected['SK']['S']}")
        if not same_version and historical.get("ContentEn") != expected.get("ContentEn"):
            raise ValueError(f"New version {expected['SK']['S']} lacks its English snapshot")
    if same_version:
        compare_existing(definition, current, f"Current version {code}")
    historical_action = (
        {"ConditionCheck": {"TableName": table, "Key": key("DEFINITION", expected["SK"]["S"]),
                            **unchanged(historical)}} if historical else
        {"Put": {"TableName": table, "Item": expected, **unchanged(None)}}
    )
    current_item = copy.deepcopy(current) if same_version else definition
    return [historical_action,
            {"Put": {"TableName": table, "Item": current_item, **unchanged(current)}}]


def offer_action(table, offer, stored=None, previous=None):
    if stored:
        if (stored.get("DefinitionCode") != offer["DefinitionCode"]
                or stored.get("SeasonID") != offer["SeasonID"]):
            raise ValueError("Offer identity or season differs")
        try:
            old_version = int(stored["DefinitionVersion"]["S"])
        except (KeyError, ValueError, TypeError) as error:
            raise ValueError("Offer version is invalid") from error
        next_version = int(offer["DefinitionVersion"]["S"])
        if old_version > next_version:
            raise ValueError("Refusing offer version downgrade")
        if old_version == next_version:
            for field, value in offer.items():
                if field != "ContentEn" and stored.get(field) != value:
                    raise ValueError(f"Offer {offer['SK']['S']} Spanish or economics differs")
            if ("ContentEn" in offer and "ContentEn" in stored
                    and offer["ContentEn"] != stored["ContentEn"]):
                raise ValueError(f"Offer {offer['SK']['S']} has conflicting English")
            desired = {field: value for field, value in offer.items()
                       if field not in ("PK", "SK", "ContentEn")}
        else:
            if not previous:
                raise ValueError("Cannot verify the previous offer version")
            for field, value in previous.items():
                if field != "ContentEn" and stored.get(field) != value:
                    raise ValueError(f"Previous offer {offer['SK']['S']} Spanish or economics differs")
            if ("ContentEn" in stored and
                    stored["ContentEn"] != previous.get("ContentEn")):
                raise ValueError(f"Previous offer {offer['SK']['S']} has conflicting English")
            desired = {field: value for field, value in offer.items()
                       if field not in ("PK", "SK")}
    else:
        desired = {field: value for field, value in offer.items()
                   if field not in ("PK", "SK")}
    names, values, assignments = {}, {}, []
    for index, (field, value) in enumerate(desired.items()):
        alias, parameter = f"#f{index}", f":f{index}"
        names[alias], values[parameter] = field, value
        assignments.append(f"{alias} = {parameter}")
    guard = unchanged(stored)
    names.update(guard.get("ExpressionAttributeNames", {}))
    values.update(guard.get("ExpressionAttributeValues", {}))
    return {"Update": {
        "TableName": table, "Key": {field: offer[field] for field in ("PK", "SK")},
        "UpdateExpression": "SET " + ", ".join(assignments),
        "ConditionExpression": guard["ConditionExpression"],
        "ExpressionAttributeNames": names, "ExpressionAttributeValues": values,
    }}
