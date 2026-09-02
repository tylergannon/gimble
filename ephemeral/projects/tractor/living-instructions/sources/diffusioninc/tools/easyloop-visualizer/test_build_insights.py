import importlib.util
import json
import os
import tempfile
import unittest


HERE = os.path.dirname(os.path.abspath(__file__))
SPEC = importlib.util.spec_from_file_location(
    "build_insights", os.path.join(HERE, "build_insights.py"))
bi = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(bi)


class ExternalTimelineRowsTest(unittest.TestCase):
    def write_json(self, root, name, value):
        path = os.path.join(root, name)
        with open(path, "w", encoding="utf-8") as f:
            json.dump(value, f)
        return path

    def row_doc(self, row_id="suite-a", start="2026-08-16T01:00:00Z",
                end="2026-08-16T01:00:10Z", status="pass"):
        return {
            "version": 1,
            "rows": [{
                "id": row_id,
                "label": "Suite A",
                "color": "#123abc",
                "description": "External suite intervals",
                "events": [{
                    "run": "20260815_223031",
                    "start": start,
                    "end": end,
                    "label": "coding/001 suite",
                    "status": status,
                    "details": "Observed in the raw log",
                    "source": "session.jsonl:L1-L2",
                }],
            }],
        }

    def test_load_render_and_combined_run_tag(self):
        with tempfile.TemporaryDirectory() as root:
            path = self.write_json(root, "rows.json", self.row_doc())
            rows = bi.load_external_timeline_rows([path])
        visits, lanes = bi.external_visits(rows, {"20260815_223031": "run4"})
        self.assertEqual(lanes, ["external-suite-a"])
        self.assertEqual(visits[0]["run"], "run4")
        self.assertEqual(visits[0]["verdict"], "PASS")
        rendered, vmeta, _, _ = bi.build_timeline(
            visits, [], "combined", True, combined=True, extra_lanes=lanes)
        self.assertIn("SUITE A", rendered)
        self.assertIn("external-suite-a", rendered)
        self.assertEqual(vmeta[0]["source"], "session.jsonl:L1-L2")
        self.assertEqual(vmeta[0]["run"], "run4")

    def test_repeated_row_merges_events_when_definition_agrees(self):
        with tempfile.TemporaryDirectory() as root:
            one = self.write_json(root, "one.json", self.row_doc("suite-merge"))
            doc = self.row_doc("suite-merge", "2026-08-16T02:00:00+00:00",
                               "2026-08-16T02:00:05+00:00", "fail")
            two = self.write_json(root, "two.json", doc)
            rows = bi.load_external_timeline_rows([one, two])
        self.assertEqual(len(rows), 1)
        self.assertEqual(len(rows[0]["events"]), 2)

    def test_overlapping_intervals_get_separate_tracks(self):
        doc = self.row_doc("suite-overlap")
        doc["rows"][0]["events"].append({
            "run": "20260815_223031",
            "start": "2026-08-16T01:00:05Z",
            "end": "2026-08-16T01:00:08Z",
            "status": "fail",
            "label": "overlap",
        })
        with tempfile.TemporaryDirectory() as root:
            path = self.write_json(root, "overlap.json", doc)
            rows = bi.load_external_timeline_rows([path])
        visits, _ = bi.external_visits(rows, {"20260815_223031": "run1"})
        self.assertEqual({v["track"] for v in visits}, {0, 1})
        self.assertEqual({v["tracks"] for v in visits}, {2})

    def test_offsetless_timestamp_is_rejected(self):
        doc = self.row_doc("suite-invalid", start="2026-08-16T01:00:00")
        with tempfile.TemporaryDirectory() as root:
            path = self.write_json(root, "bad.json", doc)
            with self.assertRaisesRegex(ValueError, "with Z or a UTC offset"):
                bi.load_external_timeline_rows([path])

    def test_cli_accepts_repeatable_files_without_consuming_runs(self):
        runs, files = bi.parse_cli([
            "20260815_010710", "--timeline-rows", "one.json",
            "20260815_223031", "--timeline-rows=two.json",
        ])
        self.assertEqual(runs, ["20260815_010710", "20260815_223031"])
        self.assertEqual(files, ["one.json", "two.json"])


class EasyLoopWorkflowTest(unittest.TestCase):
    def write_skill(self, root, name, agents):
        setup = os.path.join(root, "setup")
        os.makedirs(setup)
        path = os.path.join(setup, "SKILL.md")
        with open(path, "w", encoding="utf-8") as f:
            f.write("---\nname: %s\n---\n\n```yaml\nagents:\n%s\n```\n" %
                    (name, agents))

    def test_e2e_flow_includes_requirements_and_validation(self):
        agents = """  requirements:
    model: claude-opus-4-8
    reads:
      - user-response.md
    creates:
      - requirements.md
  validation:
    model: gpt-5.6-sol
    reads:
      - requirements.md
    creates:
      - validation.md"""
        with tempfile.TemporaryDirectory() as root:
            self.write_skill(root, "df-easy-loop-e2e", agents)
            flow = bi.parse_workflow(root)
            io_map = bi.parse_step_io(root)
            models = bi.parse_step_models(root)
        self.assertEqual(flow["kind"], "e2e")
        self.assertIn("requirements", io_map)
        self.assertIn("validation", io_map)
        self.assertEqual(models["validation"]["harness"], "Codex")

    def test_easy_flow_uses_only_its_defined_roles(self):
        agents = """  plan:
    model: claude-fable-5
    reads:
      - <spec_document>
    creates:
      - plan.md
  coding:
    model: gpt-5.6-sol
    reads:
      - updated-plan.md
      - <spec_document>
    creates:
      - coding-update.md"""
        with tempfile.TemporaryDirectory() as root:
            self.write_skill(root, "df-easy-loop-simple", agents)
            flow = bi.parse_workflow(root)
            io_map = bi.parse_step_io(root)
            models = bi.parse_step_models(root)
        self.assertEqual(flow["kind"], "easy")
        self.assertEqual(set(io_map), {"plan", "coding"})
        self.assertNotIn("requirements", io_map)
        self.assertNotIn("validation", io_map)
        self.assertEqual(models["plan"]["harness"], "Claude Code")


if __name__ == "__main__":
    unittest.main()
