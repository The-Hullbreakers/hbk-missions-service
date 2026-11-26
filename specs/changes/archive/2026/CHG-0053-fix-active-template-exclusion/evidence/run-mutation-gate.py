"""Run isolated Go overlay mutants without modifying working-tree source."""

import json
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[5]
CASES = [
    (
        "roster-definition-conflict",
        "internal/missions/domain/team_participation.go",
        "decision.Reason == ReasonMissionDefinitionAlreadyActive",
        "decision.Reason == ReasonTeammateLimitReached",
        ["./tests/unit/missions/domain", "-run", "^TestDecideRosterAcceptance_RejectsMemberWithSameDefinition$"],
    ),
    (
        "board-exclusion",
        "internal/missions/application/list_offers.go",
        "if !hidden[definitionCode] {",
        "if hidden[definitionCode] {",
        ["./tests/unit/missions/application", "-run", "^TestListOffers_DepartedMemberSeesTemplateAgain$"],
    ),
    (
        "atomic-reservation",
        "internal/missions/infrastructure/dynamodb/contract_store_reservation.go",
        'ConditionExpression: aws.String("attribute_not_exists(PK)"),',
        'ConditionExpression: aws.String("attribute_exists(PK)"),',
        ["./tests/integration/infrastructure/dynamodb", "-run", "^TestCommitAcceptance_ConcurrentSameDefinitionSerializes$"],
    ),
    (
        "owner-scoped-release",
        "internal/missions/infrastructure/dynamodb/contract_store_reservation.go",
        'attribute_not_exists(PK) OR contract_id = :cid',
        'attribute_not_exists(PK) OR contract_id <> :cid',
        ["./tests/integration/infrastructure/dynamodb", "-run", "^TestContractStore_TerminalSaveReleasesOnlyActiveMembers$"],
    ),
    (
        "credit-conflict-classification",
        "internal/missions/infrastructure/dynamodb/contract_store_review.go",
        "i >= creditStart && i < creditEnd",
        "i < creditStart",
        ["./tests/integration/infrastructure/dynamodb", "-run", "^TestContractStore_SaveApprovedStaleReleaseIsNotDuplicate$"],
    ),
    (
        "departure-prior-members",
        "internal/missions/infrastructure/dynamodb/contract_store.go",
        'condition += fmt.Sprintf(" AND contains(#departed, %s)", key)',
        'condition += " AND attribute_exists(#departed)"',
        ["./tests/integration/infrastructure/dynamodb", "-run", "^TestContractStore_SaveDepartureRejectsSameCountDifferentDepartedSet$"],
    ),
    (
        "departure-legacy-empty",
        "internal/missions/infrastructure/dynamodb/contract_store.go",
        "if len(expected) == 0 {",
        "if len(expected) > 0 {",
        ["./tests/integration/infrastructure/dynamodb", "-run", "^TestContractStore_SaveDepartureRejectsMissingDepartedWithPriorCount$"],
    ),
    (
        "stale-save-membership-fence",
        "internal/missions/infrastructure/dynamodb/contract_store.go",
        'return "(attribute_not_exists(#departed) OR size(#departed) = :departedCount)"',
        'return "(attribute_not_exists(#departed) OR size(#departed) >= :departedCount)"',
        ["./tests/integration/infrastructure/dynamodb", "-run", "^TestContractStore_(StaleSaveDoesNotRestoreDepartedMember|StaleSaveApprovedDoesNotRestoreOrCredit)$"],
    ),
    (
        "last-member-departure",
        "internal/missions/infrastructure/dynamodb/contract_store_departure.go",
        "contract.State != domain.ContractAbandoned || len(contract.ActiveMembers()) != 0",
        "contract.State != domain.ContractAbandoned || len(contract.ActiveMembers()) == 0",
        ["./tests/integration/infrastructure/dynamodb", "-run", "^TestContractStore_LastMemberSaveDepartureReleasesGuards$"],
    ),
    (
        "departure-previous-state",
        "internal/missions/infrastructure/dynamodb/contract_store_departure.go",
        '":state": &types.AttributeValueMemberS{Value: string(previousState)}',
        '":state": &types.AttributeValueMemberS{Value: string(contract.State)}',
        ["./tests/integration/infrastructure/dynamodb", "-run", "^TestContractStore_LastMemberSaveDepartureReleasesGuards$"],
    ),
]


def main():
    results = []
    for name, source_name, original, mutant, test_args in CASES:
        source = ROOT / source_name
        text = source.read_text()
        if text.count(original) != 1:
            raise RuntimeError(f"{name}: expected one exact mutation site, found {text.count(original)}")
        with tempfile.TemporaryDirectory(prefix="chg0053-mutant-") as temp:
            temp_path = Path(temp)
            replacement = temp_path / "mutated.go"
            replacement.write_text(text.replace(original, mutant, 1))
            overlay = temp_path / "overlay.json"
            overlay.write_text(json.dumps({"Replace": {str(source): str(replacement)}}))
            command = ["go", "test", "-count=1", "-overlay", str(overlay), *test_args]
            run = subprocess.run(command, cwd=ROOT, capture_output=True, text=True, timeout=120)
        killed = run.returncode != 0 and "--- FAIL:" in run.stdout + run.stderr
        results.append(f"{name}: {'KILLED' if killed else 'SURVIVED/INVALID'}\n$ {' '.join(command)}\n{run.stdout}{run.stderr}")
    report = "\n".join(results)
    (Path(__file__).parent / "mutation-results.txt").write_text(report)
    print(report)
    if any("SURVIVED/INVALID" in result for result in results):
        raise SystemExit(1)


if __name__ == "__main__":
    main()
