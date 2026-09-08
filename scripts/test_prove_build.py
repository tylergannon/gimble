"""Token-free regression checks for the build proof's preflight boundaries."""
from pathlib import Path
import runpy
import subprocess
import tempfile
import unittest
from unittest.mock import patch

PROOF = runpy.run_path(str(Path(__file__).with_name("prove-build.py")))


class BuildProofTests(unittest.TestCase):
    def test_binary_requires_matching_clean_vcs_metadata(self):
        verify = PROOF["verify_build"]
        info = "\tbuild\tvcs.revision=abc123\n\tbuild\tvcs.modified=false\n"
        verify(info, "abc123")
        for invalid in (info.replace("abc123", "other"),
                        info.replace("false", "true"),
                        "binary built with -buildvcs=false"):
            with self.subTest(metadata=invalid), self.assertRaises(RuntimeError):
                verify(invalid, "abc123")

    def test_fixture_ignores_generated_and_untracked_files(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            fixture = root / "fixture"
            fixture.mkdir()
            (fixture / ".gitignore").write_text("__pycache__/\nevidence-*.json\n")
            (fixture / "quote.py").write_text("print('seed')\n")
            (fixture / "README.md").write_text("Instructions, not fixture input.\n")
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            subprocess.run(["git", "add", "."], cwd=root, check=True)
            (fixture / "evidence-standard.json").write_text("stale evidence\n")
            (fixture / "__pycache__").mkdir()
            (fixture / "__pycache__/quote.pyc").write_bytes(b"cache")
            (fixture / "untracked.txt").write_text("not a canonical input\n")

            def git(*args):
                return subprocess.check_output(args, cwd=root, text=True).strip()

            files = PROOF["fixture_files"]
            with patch.dict(files.__globals__, REPO=root, FIXTURE=fixture, run=git):
                self.assertEqual({p.name for p in files()}, {".gitignore", "quote.py"})


if __name__ == "__main__":
    unittest.main()
