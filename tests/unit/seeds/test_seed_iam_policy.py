"""TEST-SEED-IAM-001: the seed role may read only exact OFFER keys."""

from pathlib import Path
import unittest

ROOT = Path(__file__).resolve().parents[3]
POLICY_HEADER = """  MissionCatalogDeploymentPolicy:
    Type: AWS::IAM::Policy
    Condition: HasSeedDeploymentRole
    Properties:
      PolicyName: !Sub 'hbk-missions-catalog-${Environment}'
      Roles:
        - !Ref SeedDeploymentRoleName
      PolicyDocument:
        Version: '2012-10-17'
        Statement:
"""
QUERY_DEFINITIONS_SEASONS = """          - Effect: Allow
            Action:
              - dynamodb:Query
              - dynamodb:ConditionCheckItem
            Resource: !GetAtt MissionsTable.Arn
            Condition:
              ForAllValues:StringEquals:
                dynamodb:LeadingKeys: [DEFINITION, SEASON]"""
WRITE_DEFINITIONS_OFFERS = """          - Effect: Allow
            Action:
              - dynamodb:PutItem
              - dynamodb:UpdateItem
            Resource: !GetAtt MissionsTable.Arn
            Condition:
              ForAllValues:StringEquals:
                dynamodb:LeadingKeys: [DEFINITION, OFFER]"""
GET_EXACT_OFFER = """          - Effect: Allow
            Action:
              - dynamodb:GetItem
            Resource: !GetAtt MissionsTable.Arn
            Condition:
              ForAllValues:StringEquals:
                dynamodb:LeadingKeys: [OFFER]"""


def policy_statements(text):
    """Read the seed role's SAM policy verbatim, rejecting unexpected shape."""
    policy = text.split("  MissionCatalogDeploymentPolicy:\n", 1)
    if len(policy) != 2:
        raise ValueError("Seed deployment policy missing")
    body = ("  MissionCatalogDeploymentPolicy:\n" + policy[1]).split(
        "\n  MissionsTable:\n", 1)
    if len(body) != 2 or not body[0].startswith(POLICY_HEADER):
        raise ValueError("Seed policy role/resource boundary changed")
    lines = body[0][len(POLICY_HEADER):].splitlines()
    starts = [i for i, line in enumerate(lines) if line == "          - Effect: Allow"]
    if not starts or starts[0] != 0:
        raise ValueError("Seed policy Statement is malformed")
    return ["\n".join(lines[start:end]) for start, end in
            zip(starts, starts[1:] + [len(lines)])]


class SeedIAMPolicyTest(unittest.TestCase):
    def test_existing_reads_writes_and_role_stay_scoped(self):
        # Arrange
        template = (ROOT / "template.yaml").read_text(encoding="utf-8")

        # Act
        statements = policy_statements(template)

        # Assert
        self.assertEqual(statements[:2],
                         [QUERY_DEFINITIONS_SEASONS, WRITE_DEFINITIONS_OFFERS])

    def test_get_item_only_for_offers_on_existing_table(self):
        # Arrange
        template = (ROOT / "template.yaml").read_text(encoding="utf-8")

        # Act
        statements = policy_statements(template)

        # Assert
        self.assertEqual(statements,
                         [QUERY_DEFINITIONS_SEASONS, WRITE_DEFINITIONS_OFFERS,
                          GET_EXACT_OFFER])

    def test_escalated_read_or_write_statement_is_detected(self):
        # Arrange
        proposals = {
            "query offers": GET_EXACT_OFFER.replace("GetItem", "Query"),
            "scan offers": GET_EXACT_OFFER.replace("GetItem", "Scan"),
            "read contracts": GET_EXACT_OFFER.replace("[OFFER]", "[CONTRACT]"),
            "unrestricted read": GET_EXACT_OFFER.replace("[OFFER]", "[DEFINITION, OFFER]"),
            "extra write": GET_EXACT_OFFER.replace("GetItem", "GetItem\n              - dynamodb:DeleteItem"),
        }
        template = (ROOT / "template.yaml").read_text(encoding="utf-8")
        prefix = template.split("  MissionCatalogDeploymentPolicy:\n", 1)[0]
        suffix = "\n  MissionsTable:\n" + template.split("\n  MissionsTable:\n", 1)[1]
        for scope, proposal in proposals.items():
            with self.subTest(scope=scope):
                synthetic = prefix + POLICY_HEADER + "\n".join(
                    (QUERY_DEFINITIONS_SEASONS, WRITE_DEFINITIONS_OFFERS,
                     proposal)) + suffix

                # Act
                statements = policy_statements(synthetic)

                # Assert
                self.assertNotEqual(statements,
                                    [QUERY_DEFINITIONS_SEASONS,
                                     WRITE_DEFINITIONS_OFFERS, GET_EXACT_OFFER])


if __name__ == "__main__":
    unittest.main()
