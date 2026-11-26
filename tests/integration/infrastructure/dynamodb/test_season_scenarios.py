"""REQ-0055-001/002: local preview CLI against DynamoDB Local."""

import json
import os
from pathlib import Path
import subprocess
import sys
import unittest
from unittest import mock

ROOT = Path(__file__).resolve().parents[4]
sys.path.insert(0, str(ROOT / "scripts"))
from seed_aws import AWS

ENDPOINT = "http://localhost:8000"
ENV = {**os.environ, "AWS_ACCESS_KEY_ID": "dummy", "AWS_SECRET_ACCESS_KEY": "dummy",
       "AWS_EC2_METADATA_DISABLED": "true", "AWS_DEFAULT_REGION": "us-east-1"}
ENV.pop("AWS_SESSION_TOKEN", None)


class LocalScenarioTest(unittest.TestCase):
    def setUp(self):
        self._environment = mock.patch.dict(os.environ, ENV, clear=True)
        self._environment.start()
        self.addCleanup(self._environment.stop)

    def test_fresh_start_preserves_existing_table(self):
        # Arrange
        original = AWS("hbk-missions-local", "us-east-1", ENDPOINT)
        partitions = ("SEASON", "DEFINITION", "OFFER", "MERIT_CREDIT")
        before = {name: original.partition(name) for name in partitions}
        command = [sys.executable, "-B", str(ROOT / "scripts/seed_season_scenarios.py"),
                   "--scenario", "fresh-start", "--endpoint-url", ENDPOINT, "--apply"]
        # Act
        process = subprocess.run(command, capture_output=True, text=True, env=ENV, timeout=180)
        # Assert
        self.assertEqual(process.returncode, 0, process.stderr)
        output = json.loads(process.stdout)
        self.assertEqual(output["table"], "hbk-missions-preview-fresh-start")
        self.assertEqual(output["scenario"], "fresh-start")
        self.assertTrue(output["applied"])
        self.assertEqual(before, {name: original.partition(name) for name in partitions})
        preview = AWS(output["table"], "us-east-1", ENDPOINT)
        self.assertEqual(len(preview.partition("SEASON")), 1)
        self.assertEqual(len(preview.partition("OFFER")), 36)
        self.assertFalse(preview.partition("MERIT_CREDIT"))

    def test_remote_endpoint_is_rejected_without_contact(self):
        # Arrange
        command = [sys.executable, "-B", str(ROOT / "scripts/seed_season_scenarios.py"),
                   "--scenario", "fresh-start", "--endpoint-url",
                   "http://localhost.evil.invalid:8000", "--apply"]
        # Act
        process = subprocess.run(command, capture_output=True, text=True, env=ENV, timeout=30)
        # Assert
        self.assertNotEqual(process.returncode, 0)
        self.assertNotIn("AWS_ACCESS_KEY", process.stderr)
