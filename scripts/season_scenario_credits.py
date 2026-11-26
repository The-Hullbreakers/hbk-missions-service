"""Preview credits matching the existing per-unit and completion ledger keys."""

from seed_values import av, key


def _credit(contract, objective, unit, amount):
    contract_id = contract["SK"]["S"]
    recipient = contract["Roster"]["L"][0]["S"]
    season = contract["SeasonID"]["S"]
    category = contract["Category"]["S"]
    confirmation = f"{contract_id}~{objective}~{unit}~{recipient}"
    bucket = f"CREDIT#{season}#{category}#INDIVIDUAL"
    return {**key("MERIT_CREDIT", confirmation),
            "GSI1PK": av(f"{bucket}#{recipient}"), "GSI1SK": av(confirmation),
            "GSI2PK": av(bucket), "GSI2SK": av(f"RECIPIENT#{recipient}#{confirmation}"),
            "SeasonID": av(season), "Category": av(category), "Track": av("INDIVIDUAL"),
            "RecipientID": av(recipient), "Amount": av(amount),
            "ContractID": av(contract_id), "ObjectiveID": av(objective),
            "UnitID": av(unit), "ConfirmationID": av(confirmation)}


def credits_for(contract):
    objectives = {entry["M"]["ObjectiveID"]["S"]: entry["M"]
                  for entry in contract["Terms"]["L"]}
    credits = []
    for entry in contract["Confirmed"]["L"]:
        _, objective, unit, _ = entry["S"].split("~")
        terms = objectives[objective]
        amount = int(terms["MeritPerUnit"]["N"]) + int(terms["Bonus"]["N"])
        credits.append(_credit(contract, objective, unit, amount))
    if contract["State"]["S"] == "COMPLETED":
        reward = int(contract["CompletionReward"]["N"])
        if reward:
            contract_id = contract["SK"]["S"]
            credits.append(_credit(contract, "COMPLETION", contract_id, reward))
    return credits
