"""Deterministic, provider-free season preview model and safety guards."""

from urllib.parse import urlsplit

from seed_values import av, key, offer_item, versioned
from season_scenario_contracts import contract_for
from season_scenario_credits import credits_for

PROFILES = ("off-season", "fresh-start", "mid-season")


def validate_target(endpoint, scenario):
    if scenario not in PROFILES:
        return None
    try:
        url = urlsplit(endpoint)
        port = url.port
    except ValueError:
        return None
    if (url.scheme != "http" or url.hostname not in ("localhost", "127.0.0.1", "::1")
            or port is None or url.username is not None or url.password is not None or url.path or url.query or url.fragment):
        return None
    return f"hbk-missions-preview-{scenario}"


def missing_items(existing, desired):
    wanted = {}
    for row in desired:
        identity = (row["PK"]["S"], row["SK"]["S"])
        if identity in wanted:
            return None
        wanted[identity] = row
    found = {}
    for row in existing:
        identity = (row["PK"]["S"], row["SK"]["S"])
        if identity in found or identity not in wanted or wanted[identity] != row:
            return None
        found[identity] = row
    return [row for identity, row in wanted.items() if identity not in found]


def build_scenario(scenario, definitions, user_ids):
    if scenario not in PROFILES or not definitions:
        return None
    selected = sorted(definitions, key=lambda row: row["DefinitionCode"]["S"])
    rows = [row for definition in selected for row in (definition, versioned(definition))]
    if scenario == "off-season":
        return rows
    season_id = f"preview-{scenario}"
    rows.append({**key("SEASON", season_id), "State": av("ACTIVE"), "Official": {"BOOL": False}})
    offers = [offer_item(definition, season_id) for definition in selected]
    rows.extend(offers)
    if scenario == "fresh-start":
        return rows
    users = sorted(set(user_ids))
    if len(users) < 3 or any(not user or "~" in user for user in users):
        return None
    representatives = {}
    for offer in offers:
        representatives.setdefault(offer["Category"]["S"], offer)
    choices = [representatives[category] for category in sorted(representatives)]
    for index, user in enumerate(users[:12]):
        contract = contract_for(choices[index % len(choices)], user, index)
        rows.append(contract)
        rows.extend(credits_for(contract))
    return rows
