"""Validate and render the canonical Spanish mission catalog."""

import json
import re

from seed_values import item


def load_catalog(path):
    with open(path, encoding="utf-8") as source:
        return json.load(source)


def positive(value, label):
    if type(value) is not int or value <= 0:
        raise ValueError(f"{label} must be a positive integer")
    return value


def rule_list(values, label="rules"):
    if (not isinstance(values, list) or not values
            or any(not isinstance(rule, str) or not rule.strip() for rule in values)):
        raise ValueError(f"{label} must be a non-empty list of non-blank strings")
    return values


def definitions(catalog, environment):
    if "missions" in catalog:
        return explicit_definitions(catalog, environment)
    version = str(positive(catalog["version"], "version"))
    rules = rule_list(catalog["rules"])
    result, codes = [], set()
    levels = catalog["levels"]
    if len(levels) != 4 or len({x["code"] for x in levels}) != 4:
        raise ValueError("Expected four distinct difficulty levels")
    for series in catalog["series"]:
        if series["ship"] not in ("Arrow", "Gladius"):
            raise ValueError("Only Arrow and Gladius are allowed")
        environments = series["environments"]
        if not environments or not set(environments) <= {"staging", "production"}:
            raise ValueError("Invalid environments")
        limits = series["limits_seconds"]
        if len(limits) != len(levels) or any(a <= b for a, b in zip(limits, limits[1:])):
            raise ValueError("Limits must strictly decrease with difficulty")
        for level, limit in zip(levels, limits):
            positive(limit, "limit")
            reward = positive(level["merits"], "merits")
            if reward % 50:
                raise ValueError("Merits must split evenly into five laps plus a completion bonus")
            code = f"racing-{series['code']}-{series['ship'].lower()}-{level['code']}"
            if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", code) or code in codes:
                raise ValueError(f"Invalid or duplicate definition code: {code}")
            codes.add(code)
            title = f"{series['title']} — {series['ship']} — {level['name']}"
            description = (
                f"Completa cinco vueltas lanzadas consecutivas en {series['track']} "
                f"con {series['ship']}, cada una en menos de {limit}.000 segundos."
            )
            mission_rules = list(rules)
            if series["calibration"] == "historical":
                mission_rules.append("Umbral provisional pendiente de calibración en el parche actual.")
            objectives = [
                {"ObjectiveID": "vueltas-verificadas", "Name": "Vueltas lanzadas verificadas",
                 "Description": "Cada una de las cinco vueltas lanzadas de la tanda, verificada en video.",
                 "Category": "RACING",
                 "MeritPerUnit": reward * 9 // 50, "Bonus": 0, "UnitLimit": 5},
            ]
            if environment in environments:
                result.append(item({
                    "PK": "DEFINITION", "SK": code, "DefinitionCode": code,
                    "Title": title, "Description": description, "Category": "RACING",
                    "Mode": "SOLO", "CompletionReward": reward // 10,
                    "State": "PUBLISHED", "Version": version, "Objectives": objectives,
                    "Rules": mission_rules,
                }))
    if not result:
        raise ValueError("Catalog has no enabled missions")
    return result


def explicit_definitions(catalog, environment):
    """Validate plain MissionDefinition rows before converting to AWS attributes."""
    fields = {"DefinitionCode", "Title", "Description", "Category", "Mode",
              "CompletionReward", "State", "Version", "Objectives", "Rules"}
    objective_fields = {"ObjectiveID", "Name", "Description", "Category", "MeritPerUnit",
                        "Bonus", "UnitLimit"}
    if environment not in catalog["environments"]:
        raise ValueError("Catalog is not enabled for this environment")
    result, codes = [], set()
    for mission in catalog["missions"]:
        if set(mission) != fields:
            raise ValueError("Invalid MissionDefinition fields")
        code = mission["DefinitionCode"]
        if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", code) or code in codes:
            raise ValueError(f"Invalid or duplicate definition code: {code}")
        codes.add(code)
        for field in ("Title", "Description", "Category"):
            if not isinstance(mission[field], str) or not mission[field].strip():
                raise ValueError(f"{field} is required")
        rule_list(mission["Rules"], "Rules")
        if type(mission["CompletionReward"]) is not int or mission["CompletionReward"] < 0:
            raise ValueError("CompletionReward must be a nonnegative integer")
        if mission["Mode"] not in ("SOLO", "TEAM", "FLEX") or mission["State"] != "PUBLISHED":
            raise ValueError("Invalid mode or state")
        if not isinstance(mission["Version"], str) or not re.fullmatch(r"[1-9][0-9]*", mission["Version"]):
            raise ValueError("Version must be a positive integer string")
        if not 1 <= len(mission["Objectives"]) <= 5:
            raise ValueError("Expected 1-5 objectives")
        ids = set()
        for objective in mission["Objectives"]:
            if set(objective) != objective_fields:
                raise ValueError("Invalid Objective fields")
            oid = objective["ObjectiveID"]
            if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", oid) or oid in ids:
                raise ValueError("Invalid or duplicate objective ID")
            ids.add(oid)
            if objective["Category"] != mission["Category"] or not objective["Name"] or not objective["Description"]:
                raise ValueError("Invalid objective content or category")
            positive(objective["MeritPerUnit"], "MeritPerUnit")
            for field in ("Bonus", "UnitLimit"):
                if type(objective[field]) is not int or objective[field] < 0:
                    raise ValueError(f"{field} must be a nonnegative integer")
        result.append(item({"PK": "DEFINITION", "SK": code, **mission}))
    if not result:
        raise ValueError("Catalog has no enabled missions")
    return result
