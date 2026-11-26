"""DynamoDB attribute values and stable catalog item identities."""

import copy
import hashlib


def av(value):
    if isinstance(value, str):
        return {"S": value}
    if type(value) is int:
        return {"N": str(value)}
    if isinstance(value, list):
        return {"L": [av(x) for x in value]}
    if isinstance(value, dict):
        return {"M": {k: av(v) for k, v in value.items()}}
    raise ValueError(f"Unsupported value: {value!r}")


def item(values):
    return {k: av(v) for k, v in values.items()}


def key(pk, sk):
    return item({"PK": pk, "SK": sk})


def versioned(definition):
    result = copy.deepcopy(definition)
    result["SK"] = av(definition["DefinitionCode"]["S"] + "#" + definition["Version"]["S"])
    return result


def offer_item(definition, season):
    # Stable across definition versions: redeploying updates the same offer row.
    season_key = hashlib.sha256(season.encode()).hexdigest()
    result = {k: copy.deepcopy(v) for k, v in definition.items()
              if k not in ("PK", "SK", "State", "Version")}
    result.update(key("OFFER", f"seed-{definition['DefinitionCode']['S']}-{season_key}"))
    result.update(item({"SeasonID": season, "DefinitionVersion": definition["Version"]["S"]}))
    return result


def render(selected, season):
    return [row for d in selected for row in (versioned(d), d, offer_item(d, season))]


def unchanged(old):
    if old is None:
        return {"ConditionExpression": "attribute_not_exists(PK)"}
    fields = [field for field in old if field not in ("PK", "SK")]
    names = {f"#c{i}": field for i, field in enumerate(fields)}
    values = {f":c{i}": old[field] for i, field in enumerate(fields)}
    predicates = [f"#c{i} = :c{i}" for i in range(len(fields))]
    if "ContentEn" not in old:
        names["#absent_en"] = "ContentEn"
        predicates.append("attribute_not_exists(#absent_en)")
    return {
        "ConditionExpression": " AND ".join(predicates),
        "ExpressionAttributeNames": names,
        "ExpressionAttributeValues": values,
    }
