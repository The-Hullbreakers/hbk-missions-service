"""Independent catalog fixtures for TEST-SEED-LANG-001."""

import copy
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "scripts"))
import seed_missions as seed  # noqa: E402


def one_pvp_mission():
    catalog = copy.deepcopy(seed.load_catalog(ROOT / "seeds/pvp-missions.json"))
    catalog["missions"] = catalog["missions"][:1]
    return catalog


def reviewed_record(row):
    source = {
        "Title": row["Title"]["S"],
        "Description": row["Description"]["S"],
        "Rules": [rule["S"] for rule in row["Rules"]["L"]],
        "Objectives": [{
            key: objective["M"][key]["S"]
            for key in ("ObjectiveID", "Name", "Description")
        } for objective in row["Objectives"]["L"]],
    }
    english = {
        "Title": "English title",
        "Description": "English description",
        "Rules": [f"English rule {i}" for i in range(len(source["Rules"]))],
        "Objectives": [{
            "ObjectiveID": objective["ObjectiveID"],
            "Name": f"English objective {i}",
            "Description": f"English objective description {i}",
        } for i, objective in enumerate(source["Objectives"])],
    }
    return {
        "DefinitionCode": row["DefinitionCode"]["S"],
        "DefinitionVersion": row["Version"]["S"],
        "SourceSpanish": source,
        "English": english,
    }


def write_manifest(folder, records):
    folder.mkdir(parents=True, exist_ok=True)
    (folder / "sample.json").write_text(
        json.dumps({"translations": records}, ensure_ascii=False),
        encoding="utf-8",
    )


def capture_error(operation):
    try:
        operation()
    except Exception as error:
        return error
    return None
