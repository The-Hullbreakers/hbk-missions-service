"""TEST-SEED-LANG-002: concurrent authored English must not be overwritten."""

import copy
import unittest

from seed_translation_fixture import (
    SEASON, capture_error, install_existing, new_local_table,
    seed, snapshot, spanish_and_bilingual,
)


class ConcurrentTranslationSeedTest(unittest.TestCase):
    def test_intervening_offer_translation_blocks_redeploy_atomically(self):
        # Arrange
        aws = new_local_table()
        _, bilingual = spanish_and_bilingual()
        install_existing(aws, bilingual)
        before = snapshot(aws)
        rival_offer = copy.deepcopy(seed.offer_item(bilingual, SEASON))
        rival_offer["ContentEn"]["M"]["Title"] = seed.av("Newly authored English")
        expected = {**before, ("OFFER", rival_offer["SK"]["S"]): rival_offer}
        original_call = aws.call

        def concurrent_write(operation, payload):
            if operation == "transact-write-items":
                original_call("put-item", {
                    "TableName": aws.table, "Item": rival_offer,
                    "ConditionExpression": "attribute_exists(PK)",
                })
            return original_call(operation, payload)

        aws.call = concurrent_write
        try:
            # Act
            error = capture_error(lambda: seed.synchronize(aws, [bilingual], apply=True))
        finally:
            aws.call = original_call

        # Assert
        self.assertIsInstance(error, (ValueError, RuntimeError))
        self.assertEqual(snapshot(aws), expected)


if __name__ == "__main__":
    unittest.main()
