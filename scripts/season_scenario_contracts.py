"""Preview contracts with frozen published terms and deterministic unit identities."""

import copy

from seed_values import av, key


def contract_for(offer, user, index):
    code = offer["DefinitionCode"]["S"]
    contract_id = f"preview-contract-{index + 1}-{code}"
    complete = index % 3 != 1
    objectives = offer["Objectives"]["L"]
    confirmed = []
    for entry in objectives:
        objective = entry["M"]
        count = int(objective["UnitLimit"]["N"]) if complete else 1
        for number in range(1, max(1, count) + 1):
            confirmed.append(f"{contract_id}~{objective['ObjectiveID']['S']}~unit-{number}~{user}")
    fields = ("Title", "Description", "Category", "Mode", "CompletionReward",
              "DefinitionCode", "DefinitionVersion", "ContentEn")
    contract = {**key("CONTRACT", contract_id),
                **{field: copy.deepcopy(offer[field]) for field in fields if field in offer}}
    contract.update({
        "OfferID": copy.deepcopy(offer["SK"]),
        "Rules": copy.deepcopy(offer.get("Rules", {"L": []})),
        "OfferSeasonID": copy.deepcopy(offer["SeasonID"]),
        "SeasonID": copy.deepcopy(offer["SeasonID"]),
        "Roster": av([user]), "Departed": av([]), "TeamID": av(""),
        "State": av("COMPLETED" if complete else "IN_PROGRESS"),
        "RosterFrozen": {"BOOL": True}, "Confirmed": av(confirmed),
        "Terms": copy.deepcopy(offer["Objectives"]), "CurrentPhase": av(0),
    })
    return contract
