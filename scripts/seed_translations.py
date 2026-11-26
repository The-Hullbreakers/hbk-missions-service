"""Source-bound English mission text; never alters Spanish or economics."""

import json
from pathlib import Path

from seed_catalog import definitions
from seed_values import av

CONTENT_FIELDS = {"Title", "Description", "Rules", "Objectives"}
OBJECTIVE_FIELDS = {"ObjectiveID", "Name", "Description"}


def source_spanish(row):
    return {
        "Title": row["Title"]["S"],
        "Description": row["Description"]["S"],
        "Rules": [rule["S"] for rule in row["Rules"]["L"]],
        "Objectives": [{
            field: objective["M"][field]["S"] for field in OBJECTIVE_FIELDS
        } for objective in row["Objectives"]["L"]],
    }


def validate_english(source, english):
    if not isinstance(english, dict) or set(english) != CONTENT_FIELDS:
        raise ValueError("English translation has invalid content fields")
    if any(not isinstance(english[field], str) or not english[field].strip()
           for field in ("Title", "Description")):
        raise ValueError("English title and description are required")
    rules, objectives = english["Rules"], english["Objectives"]
    if (not isinstance(rules, list) or len(rules) != len(source["Rules"])
            or any(not isinstance(rule, str) or not rule.strip() for rule in rules)):
        raise ValueError("English rules must match Spanish order and count")
    if not isinstance(objectives, list) or len(objectives) != len(source["Objectives"]):
        raise ValueError("English objectives must cover the Spanish objectives")
    for original, translated in zip(source["Objectives"], objectives):
        if (not isinstance(translated, dict) or set(translated) != OBJECTIVE_FIELDS
                or translated["ObjectiveID"] != original["ObjectiveID"]
                or not isinstance(translated["Name"], str) or not translated["Name"].strip()
                or not isinstance(translated["Description"], str)
                or (original["Description"].strip() and not translated["Description"].strip())):
            raise ValueError("English objective text or ID differs from the Spanish source")


def english_from_av(value):
    if not isinstance(value, dict) or len(value) != 1:
        raise ValueError("Invalid English attribute value")
    tag, content = next(iter(value.items()))
    if tag == "S" and isinstance(content, str):
        return content
    if tag == "M" and isinstance(content, dict):
        return {field: english_from_av(inner) for field, inner in content.items()}
    if tag == "L" and isinstance(content, list):
        return [english_from_av(inner) for inner in content]
    raise ValueError("Invalid English attribute value")


def require_complete_english(row):
    if "ContentEn" not in row:
        raise ValueError(f"English translation required for {row['SK']['S']}")
    source = source_spanish(row)
    english = english_from_av(row["ContentEn"])
    validate_english(source, english)


def translation_records(folder):
    records = {}
    for path in sorted(Path(folder).glob("*.json")):
        with path.open(encoding="utf-8") as source:
            payload = json.load(source)
        if not isinstance(payload, dict) or set(payload) != {"translations"}:
            raise ValueError(f"Invalid translation manifest: {path}")
        if not isinstance(payload["translations"], list):
            raise ValueError(f"Invalid translations in {path}")
        for record in payload["translations"]:
            if not isinstance(record, dict) or set(record) != {
                    "DefinitionCode", "DefinitionVersion", "SourceSpanish", "English"}:
                raise ValueError(f"Invalid translation record in {path}")
            code, version = record["DefinitionCode"], record["DefinitionVersion"]
            if not isinstance(code, str) or not isinstance(version, str):
                raise ValueError(f"Invalid translation identity in {path}")
            identity = (code, version)
            if identity in records:
                raise ValueError(f"Duplicate translation: {code}#{version}")
            records[identity] = record
    return records


def translated_definitions(catalog, environment, translations_dir):
    selected = definitions(catalog, environment)
    records = translation_records(translations_dir)
    result = []
    for row in selected:
        code, version = row["DefinitionCode"]["S"], row["Version"]["S"]
        record = records.get((code, version))
        if record is None:
            raise ValueError(f"Missing translation for {code}#{version}")
        if record["SourceSpanish"] != source_spanish(row):
            raise ValueError(f"Stale Spanish source for {code}#{version}")
        english = record["English"]
        validate_english(source_spanish(row), english)
        result.append({**row, "ContentEn": av(english)})
    return result
