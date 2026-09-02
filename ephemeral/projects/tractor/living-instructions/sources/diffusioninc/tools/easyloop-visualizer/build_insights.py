#!/usr/bin/env python3
"""
Build a self-contained HTML insights page for ANY EasyLoop run(s).

Usage:
  python3 build_insights.py                       # list runs, build most-recent 2
  python3 build_insights.py RUN [RUN ...]         # build exactly these runs (in order)
  python3 build_insights.py RUN --timeline-rows FILE [--timeline-rows FILE ...]
  python3 build_insights.py --list               # print available runs and exit

Each RUN is either a full path to a run directory OR a bare run id under
~/.easyloop/runs (e.g. 20260824_180612).

Everything is discovered per-run at build time — no run id, worktree, session
home, driver session or role reads/creates table is hardcoded. An EasyLoop run
lives at ~/.easyloop/runs/<RUN_ID>/ (RUN_ID = YYYYMMDD_HHMMSS = run start time):
  setup/SKILL.md          this run's own workflow spec (per-role reads:/creates:)
  agents/<role>/<NNN>/    one dir per visit (session-id.txt, outcome.yaml, artifacts)
The orchestrator (top-level driver) is NOT recorded anywhere; it is auto-found
as the session transcript that references the most of the run's sub-agent
session ids, and the working_dir is recovered from that transcript's CLI
--cd/--add-dir tokens.

Output: easyloop-workflow.html  (dark, inline CSS/JS, no network)

External timeline rows are loaded from versioned JSON files. Each event names
the run id it belongs to, so it is drawn in that individual run and in the
Combined view. Repeat --timeline-rows to load more than one file.

The centerpiece is THE PIPELINE: every visit in chronological order
(requirements -> plan -> plan-critique -> plan-update -> coding -> validation ->
review -> coding -> validation -> review ...), each showing its verdict, the
commits it produced, and the plan bullets it knocked off.
"""
import os, re, subprocess, json, html, sys
from datetime import datetime

HOME = os.path.expanduser("~")
RUNS = os.path.join(HOME, ".easyloop", "runs")
OUT  = os.path.join(os.path.dirname(os.path.abspath(__file__)), "easyloop-workflow.html")

# Candidate session homes searched when resolving a sub-agent session id or the
# orchestrator transcript. Codex sessions live under <home>/sessions/YYYY/MM/DD/
# rollout-*-<id>.jsonl; Claude sessions under <home>/projects/<slug>/<id>.jsonl.
CODEX_HOMES  = [os.path.join(HOME, h) for h in (".codex", ".codexDiffusion")]
CLAUDE_HOMES = [os.path.join(HOME, h) for h in (".claudeDiffusion", ".claude", ".claudeHome")]

CB      = re.compile(r'^\s*- \[[ xX]\]', re.M)
CBDONE  = re.compile(r'^\s*- \[[xX]\]\s*(.+)$', re.M)
STEPS   = ["requirements", "plan", "plan-critique", "plan-update", "coding", "review", "validation"]
PROPOSED = re.compile(r'Exact checklist item:\s*\*\*(.+?)\*\*', re.S)

# primary content file per step
MAINFILE = {
    "requirements": "requirements.md",
    "plan": "plan.md",
    "plan-critique": "plan-critique.md",
    "plan-update": "updated-plan.md",
    "coding": "coding-update.md",
    "review": "review.md",
    "validation": "validation.md",
}
STEPLABEL = {
    "requirements": "Requirements", "plan": "Plan", "plan-critique": "Plan Critique",
    "plan-update": "Plan Update", "coding": "Coding", "review": "Review", "validation": "Validation",
    "driver": "Driver",
}
STEPCOLOR = {
    "requirements": "#7f8da3", "plan": "#65a6ff", "plan-critique": "#b98cff",
    "plan-update": "#5cc2ff", "coding": "#e0982e", "review": "#3fb6a8", "validation": "#8b6fc7",
}

def read(p):
    try:
        with open(p, encoding="utf-8", errors="replace") as f:
            return f.read()
    except OSError:
        return ""

def mtime(p):
    try:
        return os.path.getmtime(p)
    except OSError:
        return 0

def next_of(vd):
    m = re.search(r'next:\s*`?([a-z-]+)', read(os.path.join(vd, "outcome.yaml")))
    return m.group(1).strip() if m else ""

# ---- text helpers -------------------------------------------------------
def strip_md(s):
    s = re.sub(r'`([^`]*)`', r'\1', s)
    s = re.sub(r'\*\*([^*]*)\*\*', r'\1', s)
    s = re.sub(r'\*([^*]*)\*', r'\1', s)
    s = re.sub(r'_([^_]*)_', r'\1', s)
    s = re.sub(r'\s+', ' ', s)
    return s.strip()

def _is_boilerplate(p):
    # italic scaffold notes like "_...rebuilt on every visit..._"
    return p.startswith('_') or p.startswith('#') or p.startswith('>') \
        or "rebuilt on every visit" in p.lower()

def first_para(txt, skip_prefixes=()):
    """Return the first substantive paragraph (blank-line separated)."""
    paras, cur = [], []
    for ln in txt.splitlines():
        if ln.strip():
            cur.append(ln.strip())
        elif cur:
            paras.append(" ".join(cur)); cur = []
    if cur:
        paras.append(" ".join(cur))
    for p in paras:
        if _is_boilerplate(p):
            continue
        if any(p.lower().startswith(pre) for pre in skip_prefixes):
            continue
        return strip_md(p)
    return ""

def clip(s, n):
    return s if len(s) <= n else s[:n - 1].rstrip() + "…"

def summarize(step, vd):
    t = read(os.path.join(vd, MAINFILE.get(step, "")))
    if step == "review":
        # drop the bold decision line; keep the reasoning
        body = re.sub(r'^\s*\*\*Decision.*?$', '', t, flags=re.M)
        body = re.sub(r'^\s*## Decision\s*$', '', body, flags=re.M)
        return clip(first_para(body), 260)
    if step == "validation":
        body = re.sub(r'^\s*## Outcome\s*$', '', t, flags=re.M)
        return clip(first_para(body), 260)
    return clip(first_para(t), 260)

def verdict_of(step, vd):
    """Return (label, kind) or (None, None)."""
    if step == "validation":
        t = read(os.path.join(vd, "validation.md"))
        if re.search(r'\*\*\s*FAIL', t, re.I):
            return ("FAIL", "fail")
        if re.search(r'\*\*\s*PASS', t, re.I):
            return ("PASS", "pass")
    if step == "review":
        head = read(os.path.join(vd, "review.md"))[:500]
        if re.search(r'INCOMPLETE', head, re.I):
            return ("INCOMPLETE", "incomplete")
        if re.search(r'\bCOMPLETE\b', head, re.I):
            return ("COMPLETE", "complete")
    return (None, None)

def norm_item(s):
    return clip(strip_md(s), 150)

def _match_checked(item, checkedset):
    """Does a proposed checklist item match one the review actually checked off?
    Proposed + checked strings both pass through clip(strip_md(...),150) so an
    exact hit is common; fall back to a case-insensitive substring test (either
    direction) to absorb minor wording/truncation drift."""
    if not item:
        return False
    if item in checkedset:
        return True
    il = item.lower().rstrip("… ").strip()
    if len(il) < 8:
        return False
    for c in checkedset:
        cl = c.lower().rstrip("… ").strip()
        if not cl:
            continue
        if il in cl or cl in il:
            return True
    return False

def checked_items(txt):
    return [norm_item(m) for m in CBDONE.findall(txt)]

def proposed_checks(vd):
    t = read(os.path.join(vd, "proposed-plan-checks.md"))
    if not t:
        return []
    # run2 format: "- Exact checklist item: **<text>**"
    exact = [clip(strip_md(m), 150) for m in PROPOSED.findall(t)]
    if exact:
        return exact
    # run1 format: top-level "- <text>" bullets (skip indented "- Evidence:" sub-bullets)
    items = []
    for ln in t.splitlines():
        if re.match(r'^- ', ln):
            txt = strip_md(ln[2:])
            if txt.lower().startswith("evidence"):
                continue
            items.append(clip(txt, 150))
    return items

# ---- scan ---------------------------------------------------------------
def scan_run(root, baseline=None):
    out = {"root": root, "steps": {}, "visits": []}
    for step in STEPS:
        sdir = os.path.join(root, "agents", step)
        visits = []
        if os.path.isdir(sdir):
            for name in sorted(os.listdir(sdir)):
                vd = os.path.join(sdir, name)
                if not os.path.isdir(vd):
                    continue
                anchor = os.path.join(vd, "outcome.yaml")
                if not os.path.exists(anchor):
                    anchor = os.path.join(vd, "session-id.txt")
                if not os.path.exists(anchor):
                    anchor = vd
                sid = read(os.path.join(vd, "session-id.txt")).strip()
                vlabel, vkind = verdict_of(step, vd)
                # timeline start/end anchors: session-id.txt written at start,
                # outcome.yaml at end. outcome.yaml is the AUTHORITATIVE end.
                # When a round is resumed/repaired, its session-id.txt gets
                # re-stamped to a time AFTER outcome.yaml was written, making the
                # session-id mtime a bogus start anchor. The old code blindly
                # swapped start/end in that case, which stretched every repaired
                # bar out to the re-stamp time and made sequential visits look
                # concurrent. Instead: trust outcome.yaml as the end, DROP the
                # re-stamped start anchor (leave start=None), and backfill start
                # from the prior visit's end in the sequential-chain pass below.
                sid_m = mtime(os.path.join(vd, "session-id.txt"))
                out_m = mtime(os.path.join(vd, "outcome.yaml"))
                if out_m:
                    end = out_m
                else:
                    # In-progress or interrupted visit: no outcome.yaml yet.
                    # Using the anchor (session-id.txt) mtime collapses the bar
                    # to a zero-width sliver at start time. Instead take the
                    # LATEST activity in the visit dir (e.g. a still-growing
                    # cli-output.jsonl) so a running step shows its real elapsed
                    # time. Falls back to the anchor mtime for an empty dir.
                    act = [mtime(os.path.join(vd, f)) for f in os.listdir(vd)]
                    act = [a for a in act if a]
                    end = max(act) if act else mtime(anchor)
                if out_m and sid_m and sid_m > out_m:
                    start = None            # re-stamped anchor; backfilled later
                else:
                    start = sid_m or mtime(anchor)
                    if start > end:         # last-resort guard for odd data
                        start = end
                v = {
                    "step": step, "idx": name, "mtime": mtime(anchor),
                    "start": start, "end": end,
                    "next": next_of(vd), "sid": sid,
                    "summary": summarize(step, vd),
                    "verdict": vlabel, "vkind": vkind,
                    "commits": [],
                    "vd": vd,
                }
                # bullets knocked off
                if step == "coding":
                    prop = proposed_checks(vd)
                    v["checks"] = prop
                    v["checks_kind"] = "proposed"
                    v["proposed"] = prop
                    v["suggested"] = len(prop)
                # plan snapshot for progress + confirmed-done delta
                plan = os.path.join(vd, "updated-plan.md")
                if os.path.exists(plan):
                    txt = read(plan)
                    v["total"] = len(CB.findall(txt))
                    v["_checked"] = checked_items(txt)
                    v["done"] = len(v["_checked"])
                visits.append(v)
                out["visits"].append(v)
        out["steps"][step] = visits
    out["visits"].sort(key=lambda x: x["mtime"])
    # sequential-chain start backfill: any visit whose start anchor was a
    # re-stamped session-id.txt (start left None above) inherits its start from
    # the end of the immediately preceding visit in wall-clock order. EasyLoop
    # steps execute sequentially, so the prior visit's end is the honest start —
    # this yields back-to-back bars instead of the bogus overlap the re-stamp
    # caused. Visits are already sorted by mtime (= outcome.yaml end time).
    prev_end = None
    for v in out["visits"]:
        if v.get("start") is None:
            v["start"] = prev_end if (prev_end is not None and prev_end <= v["end"]) else v["end"]
        prev_end = v["end"]
    # confirmed-done delta across plan-bearing visits (chronological).
    # baseline = boxes already checked when this run started (e.g. inherited
    # from a prior run) so we don't count inherited boxes as newly confirmed.
    seen = set(baseline or [])
    for v in out["visits"]:
        if "_checked" in v:
            newly = [c for c in v["_checked"] if c not in seen]
            for c in v["_checked"]:
                seen.add(c)
            v["confirmed"] = len(newly)
            if v["step"] in ("review", "plan-update") and newly:
                v["checks"] = newly
                v["checks_kind"] = "confirmed"
    # for each coding visit: how many plan boxes the FOLLOWING review(s) confirmed
    # done before the next coding round (the honest "checked" counterpart to suggested)
    order = out["visits"]
    for i, v in enumerate(order):
        if v["step"] == "coding":
            conf = 0
            reviewed = False
            checkedset = set()
            for w in order[i + 1:]:
                if w["step"] == "coding":
                    break
                if w["step"] in ("review", "plan-update"):
                    conf += w.get("confirmed", 0)
                    reviewed = True
                    for c in (w.get("_checked") or []):
                        checkedset.add(c)
            v["review_confirmed"] = conf
            # Per proposed box: accepted (a following review checked it off),
            # rejected (a review happened but did NOT check it), or pending
            # (no review has run for this coding round yet).
            status = []
            for pt in (v.get("proposed") or []):
                if not reviewed:
                    s = "pending"
                elif _match_checked(pt, checkedset):
                    s = "accepted"
                else:
                    s = "rejected"
                status.append({"t": pt, "s": s})
            v["proposed_status"] = status
    # final checked-box set (for feeding a continuation run's baseline)
    fc = []
    for v in out["visits"]:
        if "_checked" in v:
            fc = v["_checked"]
    out["final_checked"] = set(fc)
    return out

# ---- git ----------------------------------------------------------------
# `wt` is the per-run worktree (working_dir). It may be None/missing, in which
# case every git helper degrades to empty output (git-log / worktree panels just
# show "unavailable"). `since` is a "YYYY-MM-DD 00:00" cutoff derived from the
# earliest run's start date.
def git_log(wt, since):
    if not wt or not os.path.isdir(wt):
        return []
    try:
        raw = subprocess.check_output(
            ["git", "-C", wt, "log", "--since=%s" % since,
             "--pretty=format:%h|%ct|%aI|%s"], text=True, stderr=subprocess.DEVNULL)
    except Exception:
        return []
    rows = []
    for line in raw.splitlines():
        parts = line.split("|", 3)
        if len(parts) == 4:
            rows.append({"h": parts[0], "ct": int(parts[1]), "iso": parts[2], "msg": parts[3]})
    return rows

def git_numstat(wt, since):
    """Return {shorthash: (insertions, deletions, files)}."""
    if not wt or not os.path.isdir(wt):
        return {}
    try:
        raw = subprocess.check_output(
            ["git", "-C", wt, "log", "--since=%s" % since,
             "--pretty=format:C|%h", "--numstat"], text=True, stderr=subprocess.DEVNULL)
    except Exception:
        return {}
    sizes = {}
    cur = None; ins = dele = files = 0
    def flush():
        if cur:
            sizes[cur] = (ins, dele, files)
    for line in raw.splitlines():
        if line.startswith("C|"):
            flush()
            cur = line[2:].strip(); ins = dele = files = 0
        elif line.strip():
            p = line.split("\t")
            if len(p) == 3:
                a, d, _ = p
                files += 1
                ins += int(a) if a.isdigit() else 0
                dele += int(d) if d.isdigit() else 0
    flush()
    return sizes

def git_log_full(wt):
    """FULL branch history back to the root commit (no --since filter)."""
    if not wt or not os.path.isdir(wt):
        return []
    try:
        raw = subprocess.check_output(
            ["git", "-C", wt, "log", "--pretty=format:%h|%ct|%aI|%s"], text=True, stderr=subprocess.DEVNULL)
    except Exception:
        return []
    rows = []
    for line in raw.splitlines():
        parts = line.split("|", 3)
        if len(parts) == 4:
            rows.append({"h": parts[0], "ct": int(parts[1]), "iso": parts[2], "msg": parts[3]})
    return rows

def git_numstat_full(wt):
    """{shorthash: (insertions, deletions, files)} across the FULL history."""
    if not wt or not os.path.isdir(wt):
        return {}
    try:
        raw = subprocess.check_output(
            ["git", "-C", wt, "log", "--pretty=format:C|%h", "--numstat"], text=True, stderr=subprocess.DEVNULL)
    except Exception:
        return {}
    sizes = {}
    cur = None; ins = dele = files = 0
    def flush():
        if cur:
            sizes[cur] = (ins, dele, files)
    for line in raw.splitlines():
        if line.startswith("C|"):
            flush()
            cur = line[2:].strip(); ins = dele = files = 0
        elif line.strip():
            p = line.split("\t")
            if len(p) == 3:
                a, d, _ = p
                files += 1
                ins += int(a) if a.isdigit() else 0
                dele += int(d) if d.isdigit() else 0
    flush()
    return sizes

INFRA = ["isolation", "storage", "migration", "cloud sql", "postgres", "provision",
         "database", "least privilege", "advisory-lock", "traceability", "ci ", "ci and",
         "encrypted", "connector", "compatibility", "manifest storage", "drift", "role",
         "bootstrap", "cloud build", "container", "readiness", "healthz", "readyz", "probe",
         "git replacement"]
PRODUCT = ["edit graph", "edit-graph", "node management", "node ", "nodes", "render",
           "workspace", "artifact", "voice", "pairing", "web foundation", "react", "canvas",
           "run browser", "session", "crash-safe", "directory reconciliation", "username menu"]
SETUP = ["wip", "agents.md", "scaffold", "styleguide", "agent guide"]

def categorize(msg):
    m = msg.lower()
    for k in PRODUCT:
        if k in m: return "product"
    for k in SETUP:
        if k in m: return "setup"
    for k in INFRA:
        if k in m: return "infra"
    return "other"

def fmt_time(iso):
    try:
        return datetime.fromisoformat(iso).strftime("%H:%M")
    except Exception:
        return iso[11:16] if len(iso) > 16 else iso

# ---- driver + compaction events ----------------------------------------
# Compaction events are sourced exactly the way the ORIGINAL viz sourced them:
# Codex rollout JSONL records with top-level  type == "compacted"  (each such
# record is one auto-compaction of that session's context window). The EasyLoop
# SUB-AGENT visit sessions and the OUTER-LOOP driver session can live in ANY of
# the candidate session homes (CODEX_HOMES / CLAUDE_HOMES); we determine which by
# WHERE a session id resolves.
# Claude-hosted sub-agent sessions (the requirements/plan/plan-update/review
# roles) DO record compaction too — as a  type=="system" & subtype=="compact_boundary"
# line in their Claude Code transcript (under <claude_home>/projects/...).
# We scan those as well, so the Claude role lanes get compaction glyphs.
_UUID_RE = re.compile(r'([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})')

def iso_to_epoch(s):
    if not s:
        return 0
    try:
        if s.endswith("Z"):
            s = s[:-1] + "+00:00"
        return datetime.fromisoformat(s).timestamp()
    except Exception:
        return 0

# A compacted record leads with "timestamp" and carries "type":"compacted" in
# its head (possibly with an "ordinal" field in between, as the driver rollout
# does). We read the timestamp off the line PREFIX and confirm the type marker
# in the same bounded head slice — never json.loads the (potentially multi-MB)
# replacement_history the record also carries. Session JSONLs are HUGE — every
# scanner here is line-by-line and only ever inspects a bounded slice per line.
_TS_RE = re.compile(r'^\s*\{\s*"timestamp"\s*:\s*"([^"]+)"')

def scan_compactions(path):
    """Line-by-line scan for Codex  type=="compacted"  events. Returns
    [{ct, iso, n}] with n a per-session sequential number (1,2,3...). Only the
    line prefix is inspected; the giant replacement_history is never parsed."""
    out = []
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            for ln in f:
                head = ln[:200]
                if '"type":"compacted"' not in head and '"type": "compacted"' not in head:
                    continue
                m = _TS_RE.match(head)
                if not m:
                    continue
                out.append({"ct": iso_to_epoch(m.group(1)), "iso": m.group(1)})
    except OSError:
        return []
    for i, c in enumerate(out):
        c["n"] = i + 1
    return out

# ---- session resolution (generic across homes/dates) ------------------
# Codex sessions live under <codex_home>/sessions/YYYY/MM/DD/rollout-*-<id>.jsonl;
# the id is embedded in the filename. Claude sessions live under
# <claude_home>/projects/<slug>/<id>.jsonl. We build one index of each keyed by
# session id -> path. The Codex index is scanned lazily per date (sessions are
# huge and there can be many days), the Claude index once (uuids are unique).
_codex_idx_cache = {}     # date-key "YYYY/MM/DD" -> {uuid: path}

def _codex_index_for_dates(dates):
    """{session_uuid: rollout_path} across CODEX_HOMES for the given
    ("YYYY","MM","DD") date tuples (cached per date)."""
    idx = {}
    for y, mo, d in dates:
        key = "%s/%s/%s" % (y, mo, d)
        sub = _codex_idx_cache.get(key)
        if sub is None:
            sub = {}
            for home in CODEX_HOMES:
                day = os.path.join(home, "sessions", y, mo, d)
                if not os.path.isdir(day):
                    continue
                for fn in os.listdir(day):
                    if not fn.endswith(".jsonl"):
                        continue
                    m = _UUID_RE.search(fn)
                    if m:
                        sub.setdefault(m.group(1), os.path.join(day, fn))
            _codex_idx_cache[key] = sub
        for k, v in sub.items():
            idx.setdefault(k, v)
    return idx

def _claude_session_index():
    """{session_uuid: transcript_path} across the Claude config homes' projects/."""
    idx = {}
    uu = re.compile(r'([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\.jsonl$')
    for home in CLAUDE_HOMES:
        proj = os.path.join(home, "projects")
        if not os.path.isdir(proj):
            continue
        for root, _dirs, files in os.walk(proj):
            for fn in files:
                m = uu.search(fn)
                if m:
                    idx.setdefault(m.group(1), os.path.join(root, fn))
    return idx

_CLAUDE_SESS_IDX = _claude_session_index()

def resolve_session(sid, dates):
    """Resolve a session/thread id to (path, engine). Codex first (searched only
    in the run's candidate date dirs), then Claude. Returns (None, None) if the
    id resolves nowhere (degrade gracefully)."""
    if not sid:
        return (None, None)
    cx = _codex_index_for_dates(dates).get(sid)
    if cx:
        return (cx, "codex")
    cl = _CLAUDE_SESS_IDX.get(sid)
    if cl:
        return (cl, "claude")
    return (None, None)

# ---- Claude Code compaction support -----------------------------------
# A Claude compaction is a single line with type=="system" &
# subtype=="compact_boundary", carrying a top-level ISO `timestamp` and
# compactMetadata.trigger ("auto"|"manual"). Unlike the Codex record, the
# timestamp is NOT at the line prefix, so we bounded-json.loads only the (small)
# boundary lines the substring prefilter matches — never a big line.
def scan_claude_compactions(path):
    """Line-by-line scan for Claude Code  subtype=="compact_boundary"  events.
    Returns [{ct, iso, n, trigger}] with n a per-session sequential number. A
    cheap substring prefilter avoids json.loads on every line; the boundary
    record itself is small (no giant history payload)."""
    out = []
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            for ln in f:
                if '"compact_boundary"' not in ln:
                    continue
                if len(ln) > 200000:
                    continue
                try:
                    o = json.loads(ln)
                except Exception:
                    continue
                if o.get("type") != "system" or o.get("subtype") != "compact_boundary":
                    continue
                iso = o.get("timestamp", "")
                meta = o.get("compactMetadata") or {}
                out.append({"ct": iso_to_epoch(iso), "iso": iso,
                            "trigger": meta.get("trigger")})
    except OSError:
        return []
    out.sort(key=lambda c: c["ct"])
    for i, c in enumerate(out):
        c["n"] = i + 1
    return out

def gather_subagent_compactions(rec):
    """For every distinct sub-agent session used by this run's visits, scan its
    transcript for compaction events (Codex type=="compacted" / Claude
    compact_boundary). Placed on the visit's role lane. `rec` is a run record;
    sessions are resolved with the run's own candidate date dirs."""
    run_tag = rec["tag"]
    dates = rec["dates"]
    sid_steps = {}
    for v in rec["run"]["visits"]:
        sid = v.get("sid")
        if sid:
            sid_steps.setdefault(sid, set()).add(v["step"])
    events = []
    for sid, steps in sid_steps.items():
        path, engine = resolve_session(sid, dates)
        if not path:
            continue
        if engine == "codex":
            comps = scan_compactions(path)
        else:
            comps = scan_claude_compactions(path)
        lane = "coding" if "coding" in steps else sorted(steps)[0]
        for c in comps:
            events.append({"ct": c["ct"], "iso": c["iso"], "lane": lane,
                           "kind": "compaction", "n": c["n"], "sid": sid[:13],
                           "src": "sub-agent", "engine": engine,
                           "trigger": c.get("trigger"), "run": run_tag})
    return events

_SYS_MARK = ("<environment_context>", "<user_instructions>", "# agents.md",
             "agents.md instructions for", "<instructions>")

def _is_system_msg(t):
    low = t.lower().lstrip()
    return any(mk in low for mk in _SYS_MARK)

def scan_user_messages(path):
    """Human turns in a Codex rollout: response_item / payload.message / role
    user. Skips pathologically long lines (tool dumps) and, later, the system
    scaffolding — leaving John's genuine steering turns."""
    out = []
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            for ln in f:
                if len(ln) > 300000:
                    continue
                if '"role":"user"' not in ln and '"role": "user"' not in ln:
                    continue
                if '"message"' not in ln:
                    continue
                try:
                    o = json.loads(ln)
                except Exception:
                    continue
                if o.get("type") != "response_item":
                    continue
                p = o.get("payload") if isinstance(o.get("payload"), dict) else {}
                if p.get("type") != "message" or p.get("role") != "user":
                    continue
                txt = "".join(it.get("text", "") for it in p.get("content", [])
                              if isinstance(it, dict))
                out.append({"ct": iso_to_epoch(o.get("timestamp", "")),
                            "iso": o.get("timestamp", ""), "text": txt})
    except OSError:
        return []
    return out

def scan_first_user_codex(path):
    """First genuine (non-system-scaffolding) human turn in a Codex rollout —
    i.e. the task prompt the driver sent this agent. '' if none found."""
    for m in scan_user_messages(path):
        t = m.get("text", "").strip()
        if t and not _is_system_msg(t):
            return m["text"]
    return ""

def scan_first_user_claude(path):
    """First genuine user turn in a Claude Code transcript = the task prompt sent
    to the agent. Skips tool_result turns and system scaffolding. '' if none."""
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            for ln in f:
                if '"type":"user"' not in ln and '"type": "user"' not in ln:
                    continue
                if len(ln) > 300000:
                    continue
                try:
                    o = json.loads(ln)
                except Exception:
                    continue
                if o.get("type") != "user":
                    continue
                msg = o.get("message") if isinstance(o.get("message"), dict) else {}
                if msg.get("role") != "user":
                    continue
                c = msg.get("content")
                if isinstance(c, list):
                    # skip tool_result-only turns (no genuine text block)
                    txt = "".join(b.get("text", "") for b in c
                                  if isinstance(b, dict) and b.get("type") == "text")
                elif isinstance(c, str):
                    txt = c
                else:
                    txt = ""
                if not txt.strip() or _is_system_msg(txt):
                    continue
                return txt
    except OSError:
        return ""
    return ""

# ---- idle-gap detection (intra-session dead time) ---------------------
# A resumed session (the outer loop re-attaching to the SAME coding session
# after e.g. a gcloud reauth stall) appends its transcript with no explicit
# idle marker, so a multi-hour stall folds into the visit's continuous bar and
# becomes invisible. To surface it, scan the session's event stream for the
# wall-clock gap between adjacent events; any gap longer than IDLE_THRESHOLD is
# real dead time. We read only each line's timestamp (cheap) — never the full
# record. Codex rollout lines lead with the timestamp (_TS_RE); Claude lines
# carry it as a top-level field somewhere in the line (_TS_ANY_RE).
IDLE_THRESHOLD = 600          # seconds (10 min) — John-chosen threshold
_TS_ANY_RE = re.compile(r'"timestamp"\s*:\s*"([^"]+)"')

def scan_event_times(path, engine):
    """Sorted list of event epoch timestamps in a session transcript. Line-by-
    line, bounded slice per line — never parses the giant payloads."""
    ts = []
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            for ln in f:
                if engine == "codex":
                    m = _TS_RE.match(ln[:200])
                else:
                    if '"timestamp"' not in ln:
                        continue
                    m = _TS_ANY_RE.search(ln if len(ln) <= 100000 else ln[:100000])
                if m:
                    e = iso_to_epoch(m.group(1))
                    if e:
                        ts.append(e)
    except OSError:
        return []
    ts.sort()
    return ts

_idle_times_cache = {}        # sess_path -> sorted [epoch] (scan each session once)

def idle_segments(path, engine, start, end, threshold=IDLE_THRESHOLD):
    """Idle spans (adjacent-event gaps > threshold) that fall INSIDE a visit's
    [start,end] window. A shared session (coding/review reuse one across rounds)
    is scanned once and each visit takes only the events in its own window.
    Segment ends are clipped to the window. Returns [{a,b,dur}] (epochs)."""
    if not path or not start or not end or end <= start:
        return []
    times = _idle_times_cache.get(path)
    if times is None:
        times = scan_event_times(path, engine)
        _idle_times_cache[path] = times
    pts = [t for t in times if start - 1 <= t <= end + 1]
    segs = []
    prev = None
    for t in pts:
        if prev is not None and (t - prev) > threshold:
            a = max(prev, start); b = min(t, end)
            if b - a > threshold:
                segs.append({"a": a, "b": b, "dur": b - a})
        prev = t
    return segs

def resolve_idle(run_visits):
    """Attach v['idle'] = intra-visit idle segments, using the session path/engine
    already resolved onto each visit."""
    for v in run_visits:
        v["idle"] = idle_segments(v.get("sess_path"), v.get("engine"),
                                  v.get("start"), v.get("end"))

def gather_driver_events(path, run_assign):
    """Driver-lane events: John's steering messages + the driver session's own
    compactions. Run assignment via the same window logic used for commits."""
    if not os.path.exists(path):
        return []
    ev, seen = [], set()
    for m in scan_user_messages(path):
        t = m["text"].strip()
        if not t or _is_system_msg(t):
            continue
        key = t[:80]                       # collapse near-duplicate re-sends
        if key in seen:
            continue
        seen.add(key)
        ev.append({"ct": m["ct"], "iso": m["iso"], "lane": "driver",
                   "kind": "msg", "text": t, "src": "driver"})
    for c in scan_compactions(path):
        ev.append({"ct": c["ct"], "iso": c["iso"], "lane": "driver",
                   "kind": "compaction", "n": c["n"], "src": "driver"})
    for e in ev:
        e["run"] = run_assign(e["ct"])
    return ev

# ---- per-run windows + commit assignment (generic over N runs) ---------
def run_window(run):
    ts = [v["mtime"] for v in run["visits"] if v["mtime"]]
    return (min(ts) if ts else 0, max(ts) if ts else 0)

def make_assign_run(records):
    """Return assign(ct) -> run tag. A commit inside a run's window (+30min
    trailing slack) belongs to it; otherwise it goes to the run whose window is
    nearest in time. Reduces to the original 2-run logic for two runs."""
    wins = [(r["tag"], r["window"]) for r in records]
    def assign(ct):
        for tag, (a, b) in wins:
            if a <= ct <= b + 1800:
                return tag
        best, bd = wins[0][0], None
        for tag, (a, b) in wins:
            d = 0 if a <= ct <= b else min(abs(ct - a), abs(ct - b))
            if bd is None or d < bd:
                bd = d; best = tag
        return best
    return assign

def make_assign_run_all(records, assign):
    """assign_run for FULL history: any commit before the earliest run's start
    is 'pre' (pre-run history); otherwise the ordinary window assignment."""
    starts = [r["window"][0] for r in records if r["window"][0]]
    t_first = min(starts) if starts else 0
    def assign_all(ct):
        if t_first and ct < t_first:
            return "pre"
        return assign(ct)
    return assign_all

def attribute_commits(run, run_commits):
    """Attach each commit to the coding visit it happened during."""
    coding = [v for v in run["visits"] if v["step"] == "coding"]
    coding.sort(key=lambda v: v["mtime"])
    for c in sorted(run_commits, key=lambda x: x["ct"]):
        target = None
        for v in coding:
            if v["mtime"] >= c["ct"]:      # first coding visit that closed at/after the commit
                target = v
                break
        if target is None and coding:
            target = coding[-1]            # trailing commit -> last coding visit
        if target is not None:
            target["commits"].append(c)

def cat_counts(cs):
    d = {"infra": 0, "product": 0, "setup": 0, "other": 0}
    for c in cs: d[c["cat"]] += 1
    return d

def plan_progress(run):
    return [v for v in run["visits"] if v["step"] in ("review", "plan-update") and "done" in v]

def last_plan(run):
    revs = [v for v in run["visits"] if "done" in v]
    if revs:
        r = revs[-1]
        return r.get("total", 0), r.get("done", 0)
    return 0, 0

# ================= HTML ==================================================
def esc(s): return html.escape(str(s))
CATCOLOR = {"infra": "#e0982e", "product": "#3fb6a8", "setup": "#7f8da3", "other": "#8b6fc7"}
VCOLOR = {"pass": "#3fb6a8", "fail": "#ff5f6d", "complete": "#3fb6a8",
          "incomplete": "#e0982e", None: "#7f8da3"}
# run identity (used to distinguish runs in the Combined view). Populated per
# build by set_run_identity(); keyed by each run's tag (run1, run2, ... runN)
# plus the synthetic "pre" tag for pre-run commits.
RUN_PALETTE = ["#65a6ff", "#3fb6a8", "#e0982e", "#b98cff", "#5cc2ff",
               "#ff9f43", "#27e1c1", "#d880ff"]
RUNCOL = {"pre": "#7f8da3"}
RUNLABEL = {"pre": "pre-run"}

def set_run_identity(records):
    RUNCOL.clear(); RUNLABEL.clear()
    RUNCOL["pre"] = "#7f8da3"; RUNLABEL["pre"] = "pre-run"
    for i, r in enumerate(records):
        RUNCOL[r["tag"]] = RUN_PALETTE[i % len(RUN_PALETTE)]
        RUNLABEL[r["tag"]] = r["label_short"]

def counts(run):
    return {s: len(run["steps"].get(s, [])) for s in STEPS}

def build_meta(rec):
    """Assemble the run-card / stats meta dict for one discovered run record."""
    r = rec["run"]; ci = rec["commits"]
    total, done = last_plan(r)
    w = rec.get("window") or (None, None)
    dur = human_dur((w[1] or 0) - (w[0] or 0)) if (w[0] and w[1]) else "—"
    span = (f"{hhmm(w[0])} → {hhmm(w[1])}" if (w[0] and w[1]) else "")
    return dict(id=rec["id"], label=rec["label"], sub=rec["sub"],
                counts=counts(r), commits=ci, catc=cat_counts(ci),
                total=total, done=done, status=rec["status"],
                progress=plan_progress(r),
                dur=dur, span=span,
                dur_secs=int((w[1] or 0) - (w[0] or 0)) if (w[0] and w[1]) else 0)

# ---------- the pipeline (centerpiece) ----------
def commit_chip(c):
    col = CATCOLOR[c["cat"]]
    return (f'<span class="pcommit" title="{esc(c["cat"])} · {esc(c["t"])}">'
            f'<span class="pc-hash" style="color:{col}">{esc(c["h"])}</span>'
            f'<span class="pc-msg">{esc(c["msg"])}</span></span>')

def visit_card(v, i, show_run=False):
    col = STEPCOLOR[v["step"]]
    vk = v.get("vkind")
    run_badge = ""
    if show_run and v.get("run"):
        rc = RUNCOL.get(v["run"], "#7f8da3")
        run_badge = (f'<span class="prun" style="color:{rc};border-color:{rc}55;'
                     f'background:{rc}14">{esc(RUNLABEL.get(v["run"], v["run"]))}</span>')
    verdict_html = ""
    if v.get("verdict"):
        verdict_html = (f'<span class="vbadge" style="color:{VCOLOR[vk]};'
                        f'border-color:{VCOLOR[vk]}55;background:{VCOLOR[vk]}14">'
                        f'{esc(v["verdict"])}</span>')
    prog_html = ""
    if "done" in v:
        prog_html = f'<span class="vprog">{v["done"]}/{v["total"]} boxes</span>'
    nxt = v.get("next")
    next_html = (f'<span class="vnext">→ {esc(nxt)}</span>' if nxt else '')
    # per-visit metrics row: commits (count + size) and checkboxes (suggested vs checked)
    ins = sum(c["ins"] for c in v["commits"])
    dele = sum(c["del"] for c in v["commits"])
    nfiles = sum(c["files"] for c in v["commits"])
    metrics = []
    if v["commits"]:
        metrics.append(
            f'<span class="vm">◆ {len(v["commits"])} commit{"s" if len(v["commits"])!=1 else ""} '
            f'· <b style="color:#3fb6a8">+{ins}</b> <b style="color:#ff5f6d">−{dele}</b> '
            f'· {nfiles} file{"s" if nfiles!=1 else ""}</span>')
    if v["step"] == "coding":
        metrics.append(
            f'<span class="vm">☐ {v.get("suggested",0)} boxes suggested '
            f'· <b style="color:#3fb6a8">☑ {v.get("review_confirmed",0)} confirmed by review</b></span>')
    elif v["step"] in ("review", "plan-update") and v.get("confirmed"):
        metrics.append(f'<span class="vm">☑ {v["confirmed"]} newly confirmed done</span>')
    metrics_html = f'<div class="vmetrics">{"".join(metrics)}</div>' if metrics else ""
    # commits
    commits_html = ""
    if v["commits"]:
        rows = "".join(commit_chip(c) for c in sorted(v["commits"], key=lambda x: x["ct"]))
        commits_html = (f'<div class="vsub"><span class="vsub-t">committed '
                        f'({len(v["commits"])} · +{ins} −{dele} · {nfiles}f)</span>'
                        f'<div class="pcommits">{rows}</div></div>')
    # knocked-off bullets
    checks_html = ""
    checks = v.get("checks") or []
    if checks:
        kind = v.get("checks_kind", "")
        lab = "knocked off (proposed)" if kind == "proposed" else "plan boxes confirmed done"
        items = "".join(f'<li>{esc(c)}</li>' for c in checks)
        checks_html = (f'<div class="vsub"><span class="vsub-t" '
                       f'style="color:#3fb6a8">{lab} ({len(checks)})</span>'
                       f'<ul class="pchecks">{items}</ul></div>')
    summ = f'<div class="vsummary">{esc(v["summary"])}</div>' if v.get("summary") else ""
    return f"""
    <div class="pvisit" data-step="{esc(v["step"])}">
      <div class="pdot" style="background:{col};box-shadow:0 0 0 4px {col}22"></div>
      <div class="pcard" style="border-left:3px solid {col}">
        <div class="pcard-head">
          <span class="pstage" style="color:{col};border-color:{col}55">{esc(STEPLABEL[v["step"]])}</span>
          <span class="pidx">{esc(v["idx"])}</span>
          {run_badge}{verdict_html}{prog_html}
          <span class="ptime">{esc(fmt_time(v["mtime"] and datetime.fromtimestamp(v["mtime"]).isoformat() or ""))}</span>
          {next_html}
        </div>
        {metrics_html}
        {summ}
        {checks_html}
        {commits_html}
      </div>
    </div>"""

def pipeline(run, meta, active, show_run=False):
    cards = "".join(visit_card(v, i, show_run) for i, v in enumerate(run["visits"]))
    cls = "on" if active else ""
    return (f'<div class="pipe {cls}" data-run="{esc(meta["id"])}">'
            f'<div class="pline"></div>{cards}</div>')


# ---------- horizontal role-swimlane timeline (old-viz style) ----------
# Role lane order + colors retain the original visualizer's timeline language.
# "driver" is a synthetic TOP lane for the outer-loop agent (steering messages
# + its own compactions); the seven below are the graph roles.
GORDER = ["driver", "requirements", "plan", "plan-critique", "plan-update", "coding", "validation", "review"]
GCOLOR = {"driver": "#9db4ff", "requirements": "#27e1c1", "plan": "#65a6ff", "plan-critique": "#ffbe5c",
          "plan-update": "#9478ff", "coding": "#32d7ff", "validation": "#ff9f43", "review": "#d880ff"}
# marker glyph colors: driver steering message vs (any) compaction event
MSG_COLOR = "#5ad1ff"
COMPACT_COLOR = "#ff9f43"

EXTERNAL_STATUS = {
    "pass": ("PASS", "pass"),
    "fail": ("FAIL", "fail"),
    "interrupted": ("INTERRUPTED", "incomplete"),
    "running": ("RUNNING", None),
    "info": ("INFO", None),
}
_EXTERNAL_LANES = set()

def _external_time(value, where):
    """Parse an unambiguous ISO-8601 timestamp from an external row file."""
    if not isinstance(value, str) or not re.search(r'(?:Z|[+-]\d\d:\d\d)$', value):
        raise ValueError(f"{where} must be an ISO-8601 timestamp with Z or a UTC offset")
    try:
        return datetime.fromisoformat(value[:-1] + "+00:00" if value.endswith("Z") else value).timestamp()
    except ValueError as exc:
        raise ValueError(f"{where} is not a valid ISO-8601 timestamp: {value!r}") from exc

def load_external_timeline_rows(paths):
    """Load and merge version-1 external timeline-row JSON files.

    Files may repeat a row id to append events, but its label/color must agree.
    Unknown run ids are retained and simply do not render unless that run is one
    of the runs selected on the command line.
    """
    for lane in _EXTERNAL_LANES:
        STEPLABEL.pop(lane, None)
        GCOLOR.pop(lane, None)
    _EXTERNAL_LANES.clear()
    merged = {}
    order = []
    for path in paths:
        try:
            with open(path, encoding="utf-8") as f:
                doc = json.load(f)
        except (OSError, json.JSONDecodeError) as exc:
            raise ValueError(f"cannot read timeline rows {path!r}: {exc}") from exc
        if not isinstance(doc, dict) or doc.get("version") != 1:
            raise ValueError(f"{path}: expected an object with version: 1")
        rows = doc.get("rows")
        if not isinstance(rows, list):
            raise ValueError(f"{path}: rows must be an array")
        for ri, raw in enumerate(rows):
            where = f"{path}: rows[{ri}]"
            if not isinstance(raw, dict):
                raise ValueError(f"{where} must be an object")
            rid = raw.get("id")
            if not isinstance(rid, str) or not re.fullmatch(r'[A-Za-z][A-Za-z0-9_-]*', rid):
                raise ValueError(f"{where}.id must match [A-Za-z][A-Za-z0-9_-]*")
            label = raw.get("label")
            if not isinstance(label, str) or not label.strip():
                raise ValueError(f"{where}.label must be a non-empty string")
            color = raw.get("color", "#7f8da3")
            if not isinstance(color, str) or not re.fullmatch(r'#[0-9A-Fa-f]{6}', color):
                raise ValueError(f"{where}.color must be a six-digit hex color")
            events = raw.get("events")
            if not isinstance(events, list):
                raise ValueError(f"{where}.events must be an array")
            lane = "external-" + rid
            if lane in GORDER:
                raise ValueError(f"{where}.id collides with an existing timeline lane")
            if rid not in merged:
                merged[rid] = {
                    "id": rid, "lane": lane, "label": label.strip(), "color": color,
                    "description": str(raw.get("description") or "").strip(),
                    "events": [], "files": [],
                }
                order.append(rid)
            row = merged[rid]
            if row["label"] != label.strip() or row["color"].lower() != color.lower():
                raise ValueError(f"{where}: repeated row {rid!r} must keep the same label and color")
            row["files"].append(os.path.abspath(path))
            for ei, event in enumerate(events):
                ewhere = f"{where}.events[{ei}]"
                if not isinstance(event, dict):
                    raise ValueError(f"{ewhere} must be an object")
                run_id = event.get("run")
                if not isinstance(run_id, str) or not run_id.strip():
                    raise ValueError(f"{ewhere}.run must be a non-empty run id")
                start = _external_time(event.get("start"), ewhere + ".start")
                end = _external_time(event.get("end"), ewhere + ".end")
                if end < start:
                    raise ValueError(f"{ewhere}.end must not precede start")
                status = str(event.get("status") or "info").lower()
                if status not in EXTERNAL_STATUS:
                    raise ValueError(f"{ewhere}.status must be one of {', '.join(EXTERNAL_STATUS)}")
                row["events"].append({
                    "run_id": os.path.basename(run_id.rstrip(os.sep)),
                    "start": start, "end": end, "status": status,
                    "label": str(event.get("label") or row["label"]).strip(),
                    "details": str(event.get("details") or "").strip(),
                    "source": str(event.get("source") or os.path.abspath(path)).strip(),
                })
    out = [merged[rid] for rid in order]
    for row in out:
        row["events"].sort(key=lambda e: (e["start"], e["end"], e["label"]))
        STEPLABEL[row["lane"]] = row["label"]
        GCOLOR[row["lane"]] = row["color"]
        _EXTERNAL_LANES.add(row["lane"])
    return out

def external_visits(rows, run_tags):
    """Return timeline-only visit-shaped events and the occupied external lanes."""
    visits, lanes = [], []
    for row in rows:
        lane_events = []
        for event in row["events"]:
            tag = run_tags.get(event["run_id"])
            if not tag:
                continue
            verdict, vkind = EXTERNAL_STATUS[event["status"]]
            lane_events.append({
                "step": row["lane"], "idx": verdict, "label": event["label"],
                "start": event["start"], "end": event["end"], "mtime": event["end"],
                "next": "", "sid": "", "summary": event["details"],
                "verdict": verdict, "vkind": vkind, "commits": [], "idle": [],
                "checks": [], "external": True, "source": event["source"],
                "run": tag, "status": event["status"],
            })
        if not lane_events:
            continue
        # Imported intervals may genuinely overlap. Allocate compact tracks in
        # the one row rather than applying graph-role bars' sequential nudge.
        track_ends = []
        for event in sorted(lane_events, key=lambda e: (e["start"], e["end"])):
            track = next((i for i, end in enumerate(track_ends) if end <= event["start"]), None)
            if track is None:
                track = len(track_ends)
                track_ends.append(event["end"])
            else:
                track_ends[track] = event["end"]
            event["track"] = track
        ntracks = len(track_ends)
        for event in lane_events:
            event["tracks"] = ntracks
        visits.extend(lane_events)
        lanes.append(row["lane"])
    visits.sort(key=lambda e: (e["start"], e["end"], e["step"]))
    return visits, lanes

def human_dur(secs):
    secs = max(0, int(secs))
    m, s = divmod(secs, 60)
    if m >= 60:
        h, m = divmod(m, 60)
        return f"{h}h {m}m"
    return f"{m}m {s}s" if m else f"{s}s"

def hhmm(ts):
    return fmt_time(datetime.fromtimestamp(ts).isoformat()) if ts else ""

def visit_json(v, show_run=False):
    d = {
        "lab": v.get("label") or f'{STEPLABEL[v["step"]]} {v["idx"]}',
        "step": v["step"], "idx": v["idx"],
        "t0": hhmm(v.get("start")), "t1": hhmm(v.get("end")),
        "dur": human_dur((v.get("end") or 0) - (v.get("start") or 0)),
        "verdict": v.get("verdict"), "vkind": v.get("vkind"), "next": v.get("next", ""),
        "nCommit": len(v["commits"]),
        "ins": sum(c["ins"] for c in v["commits"]),
        "del": sum(c["del"] for c in v["commits"]),
        "files": sum(c["files"] for c in v["commits"]),
        "hashes": [{"h": c["h"], "msg": c["msg"], "ins": c["ins"], "del": c["del"]}
                   for c in sorted(v["commits"], key=lambda x: x["ct"])],
        "suggested": v.get("suggested"), "review_confirmed": v.get("review_confirmed"),
        "confirmed": v.get("confirmed"), "done": v.get("done"), "total": v.get("total"),
        "summary": v.get("summary", ""), "checks": v.get("checks") or [],
        "checks_kind": v.get("checks_kind", ""),
        "proposed_status": v.get("proposed_status") or [],
        "reads": v.get("reads_meta") or [],
        "creates": v.get("creates_meta") or [],
        "model": v.get("model", ""), "harness": v.get("harness", ""),
        "prompt_blob": v.get("prompt_blob", ""),
        "idle": [{"t0": hhmm(s["a"]), "t1": hhmm(s["b"]), "dur": human_dur(s["dur"])}
                 for s in (v.get("idle") or [])],
        "idle_total": human_dur(sum(s["dur"] for s in v["idle"])) if v.get("idle") else "",
        "external": bool(v.get("external")), "source": v.get("source", ""),
        "status": v.get("status", ""),
    }
    if show_run:
        d["run"] = v.get("run")
    return d

# ---- collapse-empty-gaps support --------------------------------------
# A gap is a stretch of wall-clock time with NOTHING to show (no visit bar
# active, no commit diamond). Gaps STRICTLY LONGER than this are collapsible.
GAP_THRESHOLD = 3600          # seconds (1 hour)
BREAK_PX = 50                 # fixed pixel width a collapsed gap is replaced with
COMMIT_PAD = 60              # +/- seconds of "occupied" padding around a commit

def find_gaps(visits, commits, t0, t1, threshold=GAP_THRESHOLD, extra_pts=None):
    """Union the visit spans + padded commit instants (+ padded marker instants,
    so driver/compaction markers are never swallowed by a collapsed gap), then
    return the list of (gap_start, gap_end) holes between them exceeding `threshold`."""
    iv = []
    for v in visits:
        s, e = v.get("start"), v.get("end")
        if s and e:
            iv.append((min(s, e), max(s, e)))
    for c in commits:
        iv.append((c["ct"] - COMMIT_PAD, c["ct"] + COMMIT_PAD))
    for t in (extra_pts or []):
        iv.append((t - COMMIT_PAD, t + COMMIT_PAD))
    if not iv:
        return []
    iv.sort()
    merged = [list(iv[0])]
    for s, e in iv[1:]:
        if s <= merged[-1][1]:
            merged[-1][1] = max(merged[-1][1], e)
        else:
            merged.append([s, e])
    gaps = []
    for i in range(len(merged) - 1):
        gs = max(merged[i][1], t0)
        ge = min(merged[i + 1][0], t1)
        if ge - gs > threshold:
            gaps.append((gs, ge))
    return gaps

def make_scale(t0, t1, gaps, LEFT, RIGHT, W):
    """Piecewise time->x mapping. Each collapsed gap is subtracted from the
    time axis and replaced with a fixed BREAK_PX-wide break marker.
    Returns (X, breaks) where breaks = [(x_left, x_right, gap_start, gap_end)]."""
    kept, cur = [], t0
    for gs, ge in gaps:
        if gs > cur:
            kept.append((cur, gs))
        cur = ge
    kept.append((cur, t1))
    total = sum(e - s for s, e in kept) or 1
    avail = (W - LEFT - RIGHT) - len(gaps) * BREAK_PX
    if avail < 60:
        avail = 60
    pps = avail / total                       # pixels per second in busy regions
    seg, px = [], float(LEFT)
    for i, (s, e) in enumerate(kept):
        seg.append((s, e, px))
        px += (e - s) * pps
        if i < len(gaps):                     # a break follows every kept seg but the last
            px += BREAK_PX
    def X(t):
        if t <= seg[0][0]:
            return LEFT
        for s, e, p0 in seg:
            if t <= e:
                return p0 if t < s else p0 + (t - s) * pps
        return LEFT + (W - LEFT - RIGHT)
    breaks = []
    for i, (gs, ge) in enumerate(gaps):
        s, e, p0 = seg[i]
        xl = p0 + (e - s) * pps
        breaks.append((xl, xl + BREAK_PX, gs, ge))
    return X, breaks

# ---- off-run-hidden geometry support ----------------------------------
# When off-run commits are HIDDEN, the wall-clock spans they occupied ALONE
# (no visit active, no run-active commit) must be removed so the busy content
# closes up — not merely blanked. That is the same idea as Collapse-gaps, but
# triggered by the off-run-hide state, with NO minimum-1h threshold and NO
# break marker (the space just disappears). Geometry therefore depends on TWO
# independent booleans (gaps collapsed? off-run hidden?), so each view can emit
# up to four x-mappings; make_scale_gen builds any of them from a single list
# of "removed" spans (each either a zero-width cut or a BREAK_PX marker).
def merge_iv(iv):
    """Merge a list of (a,b) intervals into sorted non-overlapping spans."""
    iv = sorted((min(a, b), max(a, b)) for a, b in iv)
    if not iv:
        return []
    m = [list(iv[0])]
    for s, e in iv[1:]:
        if s <= m[-1][1]:
            m[-1][1] = max(m[-1][1], e)
        else:
            m.append([s, e])
    return [(s, e) for s, e in m]

def complement(occ, t0, t1):
    """Return the holes within [t0,t1] not covered by merged intervals `occ`."""
    holes, cur = [], t0
    for s, e in occ:
        s = max(s, t0); e = min(e, t1)
        if s > cur:
            holes.append((cur, s))
        cur = max(cur, e)
    if cur < t1:
        holes.append((cur, t1))
    return holes

def make_scale_gen(t0, t1, removed, LEFT, RIGHT, W):
    """Generalized piecewise time->x mapping. `removed` is a sorted, non-
    overlapping list of (start, end, replace_px, draw_marker): each removed
    span is collapsed to `replace_px` pixels (0 => fully closed up) and, if
    draw_marker, contributes a break band/label. Kept regions share one
    pixels-per-second rate. Returns (X, breaks, dom0) where breaks =
    [(xl,xr,gs,ge)] for markered spans and dom0 = first visible (kept) time."""
    pieces, cur = [], t0
    for gs, ge, px, mk in removed:
        gs = max(gs, t0); ge = min(ge, t1)
        if ge <= gs:
            continue
        if gs > cur:
            pieces.append(("k", cur, gs))
        pieces.append(("r", gs, ge, px, mk))
        cur = max(cur, ge)
    if cur < t1:
        pieces.append(("k", cur, t1))
    if not pieces:
        pieces = [("k", t0, t1)]
    kept_total = sum(p[2] - p[1] for p in pieces if p[0] == "k") or 1
    reserved = sum(p[3] for p in pieces if p[0] == "r")
    avail = (W - LEFT - RIGHT) - reserved
    if avail < 60:
        avail = 60
    pps = avail / kept_total
    seg, breaks, x, dom0 = [], [], float(LEFT), None
    for p in pieces:
        if p[0] == "k":
            s, e = p[1], p[2]
            if dom0 is None:
                dom0 = s
            seg.append(("k", s, e, x))
            x += (e - s) * pps
        else:
            _, s, e, px, mk = p
            xl = x
            x += px
            seg.append(("r", s, e, xl, x))
            if mk:
                breaks.append((xl, x, s, e))
    if dom0 is None:
        dom0 = t0
    right = x
    def X(t):
        if t <= seg[0][1]:
            return LEFT if seg[0][0] == "k" else seg[0][3]
        for entry in seg:
            if t <= entry[2]:
                if entry[0] == "k":
                    return entry[3] + max(0.0, t - entry[1]) * pps
                return entry[4]            # collapsed span -> its right edge
        return right
    return X, breaks, dom0

def skip_label(secs):
    secs = int(max(0, secs))
    h, rem = divmod(secs, 3600)
    m = rem // 60
    if h and m >= 5:
        return f"~{h}h {m}m skipped"
    if h:
        return f"~{h}h skipped"
    return f"~{m}m skipped"

def _render_svg(visits, commits, tid, combined, LEFT, TOP, RIGHT, LANE, BOT,
                lanes, W, H, X, collapsed, gaps, breaks, uid, drop_offrun=False,
                markers=None):
    """Render ONE timeline SVG for the given time->x mapping. `collapsed`
    controls tick strategy + whether break markers are drawn. When
    `drop_offrun` is set, off-run ("pre") commit diamonds are NOT emitted at
    all — used by the off-run-hidden geometry variants whose x-mapping has
    already reclaimed the space those commits occupied. The commit enumerator
    still advances past them so `data-ci` indices stay aligned with CMETA."""
    hid = f"hatch_{uid}"
    iid = f"idle_{uid}"
    p = [f'<svg viewBox="0 0 {W} {H}" width="{W}" height="{H}" class="tlsvg">']
    p.append(f'<defs><pattern id="{hid}" width="6" height="6" patternTransform="rotate(45)" '
             'patternUnits="userSpaceOnUse"><line x1="0" y1="0" x2="0" y2="6" '
             'stroke="#ff5d70" stroke-opacity="0.45" stroke-width="2"/></pattern>'
             # idle/dead-time hatch: bold slate-grey so it reads clearly as
             # "nothing happened" (distinct from the red failure hatch above)
             f'<pattern id="{iid}" width="6" height="6" patternTransform="rotate(45)" '
             'patternUnits="userSpaceOnUse"><rect width="6" height="6" fill="#5c6980" '
             'fill-opacity="0.30"/>'
             '<line x1="0" y1="0" x2="0" y2="6" stroke="#d7deea" '
             'stroke-opacity="0.95" stroke-width="3"/></pattern></defs>')
    for i, step in enumerate(lanes):
        y = TOP + i * LANE
        fill = "#0b1018" if i % 2 == 0 else "#0e131c"
        lanecls = "tl-lane-a" if i % 2 == 0 else "tl-lane-b"
        p.append(f'<rect class="{lanecls}" x="0" y="{y}" width="{W}" height="{LANE}" '
                 f'fill="{fill}" fill-opacity="0.6"/>')
        # lane heading = hover/click target for that role's model + harness
        p.append(f'<g class="lanehit" data-run="{tid}" data-lane="{step}">'
                 f'<rect x="0" y="{y}" width="{LEFT}" height="{LANE}" fill="#000" '
                 f'fill-opacity="0" pointer-events="all"/>'
                 f'<text class="role-{step}" x="12" y="{y + LANE/2 + 3:.0f}" fill="{GCOLOR[step]}" '
                 f'font-size="9" font-family="monospace" pointer-events="none">'
                 f'{STEPLABEL[step].upper()}</text></g>')
    # break bands drawn first so bars/diamonds sit on top of them
    if collapsed:
        for xl, xr, gs, ge in breaks:
            midx = (xl + xr) / 2
            p.append(f'<rect class="tlbreak tl-breakband" x="{xl:.1f}" y="{TOP}" width="{xr-xl:.1f}" '
                     f'height="{H-BOT-TOP}" fill="#e0982e" fill-opacity="0.06"/>')
            for ex in (xl, xr):
                p.append(f'<line class="tl-breakline" x1="{ex:.1f}" y1="{TOP}" x2="{ex:.1f}" y2="{H-BOT}" '
                         f'stroke="#e0982e" stroke-opacity="0.55" stroke-width="1" stroke-dasharray="3 3"/>')
            zy = TOP + (H - BOT - TOP) / 2
            p.append(f'<path d="M {midx-5:.1f} {zy-9:.1f} L {midx+3:.1f} {zy-3:.1f} '
                     f'L {midx-3:.1f} {zy+3:.1f} L {midx+5:.1f} {zy+9:.1f}" '
                     f'stroke="#f0b45a" stroke-width="2" fill="none" pointer-events="none"/>')
            p.append(f'<text class="tlbreaklab tl-breaklab" x="{midx:.1f}" y="{H-BOT+14:.0f}" fill="#e0982e" '
                     f'font-size="9" font-family="monospace" text-anchor="middle">'
                     f'{esc(skip_label(ge-gs))}</text>')
    # time ticks
    if collapsed:
        # piecewise ticks: axis start, both edges of every break, axis end
        tick_times = [X_t0(X)]
        for xl, xr, gs, ge in breaks:
            tick_times += [gs, ge]
        tick_times.append(X_t1(X))
        seen = set()
        for t in tick_times:
            x = X(t)
            key = round(x)
            if key in seen:
                continue
            seen.add(key)
            p.append(f'<line class="tl-tickline" x1="{x:.1f}" y1="{TOP}" x2="{x:.1f}" y2="{H - BOT}" '
                     f'stroke="#263143" stroke-opacity="0.5"/>')
            p.append(f'<text class="tl-ticktext" x="{x:.1f}" y="{TOP - 9}" fill="#718198" font-size="9" '
                     f'font-family="monospace" text-anchor="middle">{hhmm(t)}</text>')
    else:
        ticks = 10
        for k in range(ticks + 1):
            t = X_t0(X) + (X_t1(X) - X_t0(X)) * k / ticks
            x = X(t)
            p.append(f'<line class="tl-tickline" x1="{x:.1f}" y1="{TOP}" x2="{x:.1f}" y2="{H - BOT}" '
                     f'stroke="#263143" stroke-opacity="0.5"/>')
            p.append(f'<text class="tl-ticktext" x="{x:.1f}" y="{TOP - 9}" fill="#718198" font-size="9" '
                     f'font-family="monospace" text-anchor="middle">{hhmm(t)}</text>')
    # run-start markers: in the Combined view, a vertical guide line at the
    # instant each run began (its earliest visit start) so it's obvious where one
    # EasyLoop run ends and the next picks up. Drawn under the bars, non-interactive.
    if combined:
        run_starts = {}
        for v in visits:
            r = v.get("run")
            if r and r != "pre" and v.get("start"):
                if r not in run_starts or v["start"] < run_starts[r]:
                    run_starts[r] = v["start"]
        for r, st in sorted(run_starts.items(), key=lambda kv: kv[1]):
            rx = X(st)
            rc = RUNCOL.get(r, "#9db4ff")
            p.append(f'<line class="tl-runstart" x1="{rx:.1f}" y1="{TOP}" x2="{rx:.1f}" '
                     f'y2="{H-BOT:.0f}" stroke="{rc}" stroke-opacity="0.85" stroke-width="1.5" '
                     f'stroke-dasharray="2 3" pointer-events="none"/>')
            p.append(f'<text class="tl-runstartlab" x="{rx+3:.1f}" y="14" fill="{rc}" '
                     f'font-size="8.5" font-family="monospace" pointer-events="none">'
                     f'▶ {esc(RUNLABEL.get(r, r))}</text>')
    # per-lane running right edge, so a min-width bar never renders on top of
    # the previous bar in the same lane. On a compressed axis (esp. the Combined
    # view, whose span covers every run + the full commit history) short but
    # honest sequential bars — coding -> validation -> review within one round —
    # collapse to the 6px floor and their true start-x land within a few pixels
    # of each other, producing false-looking overlap. EasyLoop steps in a lane
    # are strictly sequential, so nudging a clamped bar right to sit flush after
    # its predecessor is faithful. Only fires on actual overlap, so wider bars
    # (individual run views) stay pixel-identical. Visits arrive start-sorted, so
    # per-lane x1 is non-decreasing.
    lane_x2 = {}
    for i, v in enumerate(visits):
        if not v.get("start") or v["step"] not in lanes:
            continue
        lane_y = TOP + lanes.index(v["step"]) * LANE
        if v.get("external"):
            tracks = max(1, int(v.get("tracks") or 1))
            slot = (LANE - 8) / tracks
            y = lane_y + 4 + int(v.get("track") or 0) * slot
            h = max(6, slot - 3)
        else:
            y = lane_y + 8
            h = LANE - 16
        x1 = X(v["start"]); x2 = X(v["end"]); w = max(6, x2 - x1)
        if not v.get("external"):
            last = lane_x2.get(v["step"])
            if last is not None and x1 < last:
                x1 = last                   # sit flush after the prior bar
            lane_x2[v["step"]] = x1 + w
        col = GCOLOR[v["step"]]
        failed = v.get("vkind") in ("fail", "incomplete")
        if failed:
            stroke = "#e0982e" if v.get("status") == "interrupted" else "#ff5d70"
        elif v.get("external") and v.get("status") == "pass":
            stroke = "#3fb6a8"
        elif combined:
            stroke = RUNCOL.get(v.get("run"), col)
        else:
            stroke = col
        p.append(f'<rect class="hit role-{v["step"]}" data-run="{tid}" data-i="{i}" x="{x1:.1f}" y="{y}" '
                 f'width="{w:.1f}" height="{h}" rx="5" fill="{col}" '
                 f'fill-opacity="{0.46 if failed else 0.72}" stroke="{stroke}" '
                 f'stroke-width="{1.8 if combined and not failed else 1.2}"/>')
        if failed:
            p.append(f'<rect x="{x1:.1f}" y="{y}" width="{w:.1f}" height="{h}" rx="5" '
                     f'fill="url(#{hid})" pointer-events="none"/>')
        # idle/dead-time segments: gaps in the session's event stream longer than
        # IDLE_THRESHOLD, drawn as a grey hatched block INSIDE the bar with a
        # duration label, so a resumed-session stall (e.g. the gcloud reauth gap)
        # no longer folds silently into elapsed time. Map idle epochs through the
        # same X() and shift by any lane-nudge (x1 - X(start)); clip to the bar.
        idle_hides_idx = False
        shift = x1 - X(v["start"])
        for seg in (v.get("idle") or []):
            ia = max(x1, min(X(seg["a"]) + shift, x1 + w))
            ib = max(x1, min(X(seg["b"]) + shift, x1 + w))
            iw = ib - ia
            if iw < 1.5:
                continue
            # draw the idle block TALLER than the bar so it stands proud of it
            iy = y - 5
            ih = h + 10
            p.append(f'<rect x="{ia:.1f}" y="{iy}" width="{iw:.1f}" height="{ih}" rx="3" '
                     f'fill="url(#{iid})" stroke="#d7deea" stroke-opacity="0.85" '
                     f'stroke-width="1.1" pointer-events="none"/>')
            if iw > 34:
                idle_hides_idx = True
                p.append(f'<text x="{(ia + ib) / 2:.1f}" y="{y + h/2 + 3:.0f}" fill="#f2f6fc" '
                         f'font-size="7.5" font-family="monospace" text-anchor="middle" '
                         f'pointer-events="none">{esc(human_dur(seg["dur"]))} idle</text>')
        if w > 22 and v["step"] != "review" and not idle_hides_idx:
            p.append(f'<text class="tl-bartext" x="{x1 + w/2:.1f}" y="{y + h/2 + 3:.0f}" fill="#eef5ff" font-size="8" '
                     f'font-family="monospace" text-anchor="middle" pointer-events="none">{v["idx"]}</text>')
        if v["step"] == "review":
            cx = x1 + w / 2; cy = y + h / 2
            p.append(f'<circle cx="{cx:.1f}" cy="{cy:.1f}" r="8" fill="#150d1d" '
                     f'stroke="#e29aff" pointer-events="none"/>')
            p.append(f'<text x="{cx:.1f}" y="{cy + 2.5:.0f}" fill="#f6e4ff" font-size="8" '
                     f'font-family="monospace" text-anchor="middle" pointer-events="none">{v.get("confirmed",0)}</text>')
    # "off-run" = a commit that landed while NO EasyLoop run was active: it is not
    # covered by any run's active window and so is not attributed to a run
    # visit. In the Combined view these are exactly the commits the run
    # classification already tags "pre" (before Run 1 started, or otherwise
    # outside any run) and draws with the gray pre-run marker. Reuse that
    # classification rather than recomputing occupancy. Per-run (Run 1 / Run 2)
    # timelines only carry run1/run2-tagged commits, so nothing is tagged
    # off-run there and the toggle is a no-op for them.
    cy = TOP + lanes.index("coding") * LANE + LANE - 5
    for ci, c in enumerate(sorted(commits, key=lambda x: x["ct"])):
        if drop_offrun and c.get("run") == "pre":
            continue                       # off-run hidden: space already reclaimed
        x = X(c["ct"]); r = 4 + min(7, (c["ins"] + c["del"]) / 350.0)
        col = CATCOLOR[c["cat"]]
        cstroke = RUNCOL.get(c.get("run"), "#07090d") if combined else "#07090d"
        offcls = " offrun" if c.get("run") == "pre" else ""
        p.append(f'<path class="hit cat-{c["cat"]}{offcls}" data-run="{tid}" data-ci="{ci}" '
                 f'd="M {x:.1f} {cy - r:.1f} L {x + r:.1f} {cy:.1f} L {x:.1f} {cy + r:.1f} '
                 f'L {x - r:.1f} {cy:.1f} Z" fill="{col}" stroke="{cstroke}" '
                 f'stroke-width="{1.4 if combined else 1}"/>')
    # driver steering messages + compaction events (old-viz compaction glyph:
    # a dark-filled ring with an orange crosshair). Markers are index-aligned to
    # MMETA via data-mi and drawn on top of bars/diamonds so they stay clickable.
    for mi, mk in enumerate(markers or []):
        lane = mk.get("lane")
        if lane not in lanes:
            continue
        mx = X(mk["ct"])
        myc = TOP + lanes.index(lane) * LANE + LANE / 2
        if mk["kind"] == "compaction":
            col = COMPACT_COLOR
            p.append(f'<g class="hit" data-run="{tid}" data-mi="{mi}">'
                     f'<circle cx="{mx:.1f}" cy="{myc:.1f}" r="7" fill="#07090d" '
                     f'stroke="{col}" stroke-width="2"/>'
                     f'<path d="M{mx-3.5:.1f},{myc:.1f}h7M{mx:.1f},{myc-3.5:.1f}v7" '
                     f'stroke="{col}" stroke-width="1.5" pointer-events="none"/></g>')
        else:
            col = MSG_COLOR
            s = 5.5
            p.append(f'<g class="hit" data-run="{tid}" data-mi="{mi}">'
                     f'<rect x="{mx-s:.1f}" y="{myc-s:.1f}" width="{2*s:.1f}" height="{2*s:.1f}" '
                     f'rx="2" fill="{col}" fill-opacity="0.9" stroke="#07131c" stroke-width="1"/>'
                     f'<path d="M{mx-2:.1f},{myc:.1f}h4M{mx:.1f},{myc-2:.1f}v4" '
                     f'stroke="#07131c" stroke-width="1" pointer-events="none" opacity="0.55"/></g>')
    p.append('</svg>')
    return "".join(p)

# X carries its own domain endpoints so tick code can reach them
def X_t0(X): return X.__dict__["t0"]
def X_t1(X): return X.__dict__["t1"]

def build_timeline(visits, commits, tid, active, combined=False, markers=None,
                   extra_lanes=None, base_lanes=None):
    markers = markers or []
    mcts = [m["ct"] for m in markers if m.get("ct")]
    starts = [v["start"] for v in visits if v.get("start")]
    ends = [v["end"] for v in visits if v.get("end")]
    cts = [c["ct"] for c in commits]
    t0 = min(starts + cts + mcts) if (starts or cts or mcts) else 0
    t1 = max(ends + cts + mcts) if (ends or cts or mcts) else t0 + 60
    if t1 <= t0:
        t1 = t0 + 60
    LEFT, TOP, RIGHT, LANE, BOT = 128, 46, 26, 40, 30
    lanes = list(base_lanes or GORDER) + list(extra_lanes or [])
    H = TOP + len(lanes) * LANE + BOT

    # ----- normal (true-to-scale) mapping -----
    mins = (t1 - t0) / 60.0
    W = int(min(2600, max(1040, LEFT + RIGHT + mins * 3.6)))
    def Xn(t): return LEFT + (t - t0) / (t1 - t0) * (W - LEFT - RIGHT)
    Xn.__dict__["t0"] = t0; Xn.__dict__["t1"] = t1

    # ----- collapsed mapping (gaps > 1h removed) -----
    gaps = find_gaps(visits, commits, t0, t1, extra_pts=mcts)
    kept_secs = (t1 - t0) - sum(ge - gs for gs, ge in gaps)
    Wc = int(min(2600, max(1040, LEFT + RIGHT + (kept_secs / 60.0) * 3.6 + len(gaps) * BREAK_PX)))
    Xc, breaks = make_scale(t0, t1, gaps, LEFT, RIGHT, Wc)
    Xc.__dict__["t0"] = t0; Xc.__dict__["t1"] = t1

    # ----- data (positions are NOT baked into this; index-aligned to bars) ----
    vmeta = []
    for v in visits:
        if not v.get("start") or v["step"] not in lanes:
            vmeta.append(None); continue
        vmeta.append(visit_json(v, show_run=combined))
    cmeta = []
    for c in sorted(commits, key=lambda x: x["ct"]):
        cm = {"h": c["h"], "msg": c["msg"], "t": c["t"], "ins": c["ins"],
              "del": c["del"], "files": c["files"], "cat": c["cat"]}
        if combined:
            cm["run"] = c.get("run")
        cmeta.append(cm)
    mmeta = []
    for m in markers:
        if m["kind"] == "compaction":
            mmeta.append({"kind": "compaction", "src": m.get("src"), "n": m.get("n"),
                          "t": hhmm(m["ct"]), "lane": m.get("lane"), "sid": m.get("sid"),
                          "engine": m.get("engine"), "trigger": m.get("trigger"),
                          "run": m.get("run") if combined else None})
        else:
            mmeta.append({"kind": "msg", "src": "driver", "t": hhmm(m["ct"]),
                          "text": m.get("text", ""),
                          "run": m.get("run") if combined else None})

    # ----- the two OFF-RUN-SHOWN variants (unchanged geometry) -----
    #   fs = gaps-full   + off-run shown   (the default view)
    #   cs = gaps-collapsed + off-run shown
    svg_fs = _render_svg(visits, commits, tid, combined, LEFT, TOP, RIGHT, LANE, BOT,
                         lanes, W, H, Xn, False, [], [], tid + "_fs", markers=markers)
    svg_cs = _render_svg(visits, commits, tid, combined, LEFT, TOP, RIGHT, LANE, BOT,
                         lanes, Wc, H, Xc, True, gaps, breaks, tid + "_cs", markers=markers)

    # ----- the OFF-RUN-HIDDEN variants -----
    # Only views that actually CARRY off-run ("pre") commits need distinct
    # hidden geometry; Run 1 / Run 2 have zero off-run commits, so their hidden
    # geometry is identical to the shown geometry — reuse the same SVGs (alias
    # via CSS classes) instead of emitting redundant duplicates.
    has_off = any(c.get("run") == "pre" for c in commits)
    if has_off:
        # occupancy of everything that is NOT off-run (visit spans + padded
        # run-active commit instants). Holes in it that contain an off-run
        # commit are the spans occupied ONLY by off-run commits -> removable.
        vis_iv = [(v["start"], v["end"]) for v in visits
                  if v.get("start") and v.get("end")]
        keep_iv = vis_iv + [(c["ct"] - COMMIT_PAD, c["ct"] + COMMIT_PAD)
                            for c in commits if c.get("run") != "pre"]
        # protect marker instants so an off-run cut never swallows a driver/
        # compaction marker (they are not off-run commits).
        keep_iv += [(t - COMMIT_PAD, t + COMMIT_PAD) for t in mcts]
        keep_dead = complement(merge_iv(keep_iv), t0, t1)
        pre_cts = [c["ct"] for c in commits if c.get("run") == "pre"]
        def _has_pre(s, e):
            return any(s <= t <= e for t in pre_cts)
        offrun_cuts = [(s, e) for (s, e) in keep_dead if _has_pre(s, e)]
        # gaps of the NON-off-run timeline that exceed the 1h threshold and hold
        # no off-run commit -> ordinary collapse-gaps breaks (only used when the
        # gaps toggle is also collapsed).
        keep_break_gaps = [(s, e) for (s, e) in keep_dead
                           if not _has_pre(s, e) and (e - s) > GAP_THRESHOLD]
        cut_secs = sum(e - s for s, e in offrun_cuts)

        # fh = gaps-full + off-run hidden : normal scale, off-run spans cut out.
        kept2 = (t1 - t0) - cut_secs
        W2 = int(min(2600, max(1040, LEFT + RIGHT + (kept2 / 60.0) * 3.6)))
        removed2 = sorted((s, e, 0.0, False) for s, e in offrun_cuts)
        X2, _b2, dom2 = make_scale_gen(t0, t1, removed2, LEFT, RIGHT, W2)
        X2.__dict__["t0"] = dom2; X2.__dict__["t1"] = t1
        svg_fh = _render_svg(visits, commits, tid, combined, LEFT, TOP, RIGHT, LANE, BOT,
                             lanes, W2, H, X2, False, [], [], tid + "_fh", drop_offrun=True,
                             markers=markers)

        # ch = gaps-collapsed + off-run hidden : off-run spans cut (0px, no
        # marker) AND the ordinary >1h gaps collapsed to break markers.
        brk_secs = sum(e - s for s, e in keep_break_gaps)
        kept4 = (t1 - t0) - cut_secs - brk_secs
        W4 = int(min(2600, max(1040, LEFT + RIGHT + (kept4 / 60.0) * 3.6
                                     + len(keep_break_gaps) * BREAK_PX)))
        removed4 = sorted([(s, e, 0.0, False) for s, e in offrun_cuts]
                          + [(s, e, float(BREAK_PX), True) for s, e in keep_break_gaps])
        X4, breaks4, dom4 = make_scale_gen(t0, t1, removed4, LEFT, RIGHT, W4)
        X4.__dict__["t0"] = dom4; X4.__dict__["t1"] = t1
        svg_ch = _render_svg(visits, commits, tid, combined, LEFT, TOP, RIGHT, LANE, BOT,
                             lanes, W4, H, X4, True, keep_break_gaps, breaks4,
                             tid + "_ch", drop_offrun=True, markers=markers)
        variants = (f'<div class="tlvar tlvar-fs">{svg_fs}</div>'
                    f'<div class="tlvar tlvar-fh">{svg_fh}</div>'
                    f'<div class="tlvar tlvar-cs">{svg_cs}</div>'
                    f'<div class="tlvar tlvar-ch">{svg_ch}</div>')
    else:
        # no off-run commits -> hidden == shown; alias to avoid duplicate SVGs.
        variants = (f'<div class="tlvar tlvar-fs tlvar-fh">{svg_fs}</div>'
                    f'<div class="tlvar tlvar-cs tlvar-ch">{svg_cs}</div>')

    cls = "on" if active else ""
    html_ = (f'<div class="tl {cls}" data-run="{tid}" data-gaps="{len(gaps)}">'
             f'<div class="tlscroll">{variants}</div></div>')
    return html_, vmeta, cmeta, mmeta

# ---- Task A: per-run off-run commit re-tagging -------------------------
# The per-run timelines (Run 1 / Run 2) draw the commits in c1 / c2, which are
# the SAME list objects RUN1_META["commits"] / RUN2_META["commits"] reference
# and which build_timeline receives below. assign_run() only ever tags those
# "run1"/"run2" — never "pre" — so build_timeline saw has_off=False and the
# "hide off-run commits" toggle had no geometry to reclaim (a visual no-op).
# Re-tag any commit NOT covered by one of that run's OWN visit spans (padded by
# ±COMMIT_PAD) as off-run ("pre"), matching the combined view's off-run notion.
# Only c1/c2 are mutated; the Combined view builds off the separate
# all_commits/combined_commits list (assign_run_all already yields "pre") and is
# untouched. cat_counts/run_card key off c["cat"], not c["run"], so run cards and
# the commit->coding-visit attribution done above are unaffected.
def retag_offrun(run, run_commits):
    spans = [(v["start"], v["end"]) for v in run["visits"]
             if v.get("start") and v.get("end")]
    n = 0
    for c in run_commits:
        covered = any(s - COMMIT_PAD <= c["ct"] <= e + COMMIT_PAD for s, e in spans)
        if not covered:
            c["run"] = "pre"
            n += 1
    return n

# ---- Task B: consumed/produced artifact resolution ---------------------
# Per graph role: the bare input filenames it reads and the output filenames it
# creates. Parsed PER RUN from that run's own setup/SKILL.md `agents:`
# block (see parse_step_io) rather than a hardcoded table. This built-in table is
# used only as a fallback when a run has no usable setup/SKILL.md.
STEP_IO_FALLBACK = {
    "requirements":  {"reads": ["user-response.md"],
                      "creates": ["requirements.md", "validation-holdout.md"]},
    "plan":          {"reads": ["requirements.md"],
                      "creates": ["plan.md"]},
    "plan-critique": {"reads": ["plan.md", "requirements.md"],
                      "creates": ["plan-critique.md"]},
    "plan-update":   {"reads": ["requirements.md", "plan.md", "plan-critique.md"],
                      "creates": ["updated-plan.md"]},
    "coding":        {"reads": ["updated-plan.md", "requirements.md", "review.md"],
                      "creates": ["coding-update.md", "proposed-plan-checks.md"]},
    "validation":    {"reads": ["requirements.md", "validation-holdout.md",
                                "updated-plan.md", "coding-update.md"],
                      "creates": ["validation.md"]},
    "review":        {"reads": ["requirements.md", "updated-plan.md", "coding-update.md",
                                "proposed-plan-checks.md", "validation.md"],
                      "creates": ["review.md", "updated-plan.md"]},
}

# A bare artifact filename token: strip trailing "# ..." comments and "(...)"
# notes ("- coding-update.md (all prior coding visits)" -> "coding-update.md").
_ART_RE = re.compile(r'^([A-Za-z0-9._/-]+\.[A-Za-z0-9]+)')

def parse_step_io(run_dir):
    """Parse per-role reads:/creates: from a run's setup/SKILL.md `agents:`
    block. Falls back to STEP_IO_FALLBACK if the file is missing/unparseable.
    Simple indentation-aware scan (stdlib only, no YAML dependency)."""
    yml = os.path.join(run_dir, "setup", "SKILL.md")
    txt = read(yml)
    if not txt:
        return dict(STEP_IO_FALLBACK)
    lines = txt.splitlines()
    # locate top-level `agents:` (no leading whitespace)
    start = None
    for i, ln in enumerate(lines):
        if re.match(r'^agents:\s*$', ln):
            start = i + 1
            break
    if start is None:
        return dict(STEP_IO_FALLBACK)
    io = {}
    role = None; bucket = None; role_indent = None
    for ln in lines[start:]:
        if not ln.strip() or ln.lstrip().startswith("#"):
            continue
        indent = len(ln) - len(ln.lstrip())
        stripped = ln.strip()
        # a new top-level section (no indent, ends with ':') ends the agents block
        if indent == 0 and stripped.endswith(":"):
            break
        m_role = re.match(r'^( {2,4})([a-z][a-z-]*):\s*$', ln)
        if m_role and (role_indent is None or len(m_role.group(1)) <= role_indent):
            role = m_role.group(2)
            role_indent = len(m_role.group(1))
            io.setdefault(role, {"reads": [], "creates": []})
            bucket = None
            continue
        if role is None:
            continue
        if re.match(r'^\s+reads:\s*$', ln):
            bucket = "reads"; continue
        if re.match(r'^\s+creates:\s*$', ln):
            bucket = "creates"; continue
        m_item = re.match(r'^\s+-\s+(.*)$', ln)
        if m_item and bucket:
            am = _ART_RE.match(m_item.group(1).strip())
            if am:
                io[role][bucket].append(am.group(1))
            continue
        # any other key under the role (model:, next:, prompt_construction: ...)
        # closes the current list bucket
        if re.match(r'^\s+[a-z_]+:', ln):
            bucket = None
    # keep only roles we actually render, and drop empties gracefully
    return {r: io[r] for r in io} or dict(STEP_IO_FALLBACK)

# Per-role model + coding-harness, parsed PER RUN from that run's own
# setup/SKILL.md `agents:` block (same block parse_step_io reads). Harness
# is derived from the model-name prefix exactly as the yaml's own rule states
# (claude-* -> Claude Code, gpt-* -> Codex). Fallback mirrors STEP_IO_FALLBACK's
# built-in table, used only when the yaml is missing/unparseable.
STEP_MODEL_FALLBACK = {
    "requirements":  {"model": "claude-opus-4-8", "harness": "Claude Code"},
    "plan":          {"model": "claude-fable-5",  "harness": "Claude Code"},
    "plan-critique": {"model": "gpt-5.6-sol",     "harness": "Codex"},
    "plan-update":   {"model": "claude-opus-4-8", "harness": "Claude Code"},
    "coding":        {"model": "gpt-5.6-sol",     "harness": "Codex"},
    "validation":    {"model": "gpt-5.6-sol",     "harness": "Codex"},
    "review":        {"model": "claude-opus-4-8", "harness": "Claude Code"},
}

def harness_for(model):
    """Coding harness derived from the model-name prefix (the yaml's own rule)."""
    if model.startswith("claude-"):
        return "Claude Code"
    if model.startswith("gpt-"):
        return "Codex"
    return ""

def parse_step_models(run_dir):
    """Parse per-role `model:` from a run's setup/SKILL.md `agents:` block.
    Returns {role: {"model","harness"}}. Falls back to STEP_MODEL_FALLBACK if the
    file is missing/unparseable (mirrors parse_step_io's indentation-aware scan)."""
    yml = os.path.join(run_dir, "setup", "SKILL.md")
    txt = read(yml)
    if not txt:
        return dict(STEP_MODEL_FALLBACK)
    lines = txt.splitlines()
    start = None
    for i, ln in enumerate(lines):
        if re.match(r'^agents:\s*$', ln):
            start = i + 1
            break
    if start is None:
        return dict(STEP_MODEL_FALLBACK)
    out = {}
    role = None; role_indent = None
    for ln in lines[start:]:
        if not ln.strip() or ln.lstrip().startswith("#"):
            continue
        indent = len(ln) - len(ln.lstrip())
        stripped = ln.strip()
        if indent == 0 and stripped.endswith(":"):
            break
        m_role = re.match(r'^( {2,4})([a-z][a-z-]*):\s*$', ln)
        if m_role and (role_indent is None or len(m_role.group(1)) <= role_indent):
            role = m_role.group(2)
            role_indent = len(m_role.group(1))
            continue
        if role is None:
            continue
        m_model = re.match(r'^\s+model:\s*([^\s#]+)', ln)
        if m_model:
            model = m_model.group(1).strip()
            out[role] = {"model": model, "harness": harness_for(model)}
    return out or dict(STEP_MODEL_FALLBACK)

def parse_workflow(run_dir):
    """Identify which EasyLoop flow produced a run from setup/SKILL.md."""
    txt = read(os.path.join(run_dir, "setup", "SKILL.md"))
    match = re.search(r'^name:\s*([^\s#]+)', txt, re.M)
    name = match.group(1) if match else ""
    if name == "df-easy-loop-e2e":
        return {"name": name, "kind": "e2e", "label": "E2E"}
    if name == "df-easy-loop-simple":
        return {"name": name, "kind": "easy", "label": "Easy"}
    return {"name": name, "kind": "custom", "label": name or "Custom"}

# Best-effort model of the top-level orchestrator (the driver is NOT in the
# agents block). Harness comes from the discovered engine; the model is read from
# the orchestrator transcript's first bounded `"model":"..."` token (Codex
# turn_context / Claude record). Never blocks the build — returns "" on failure.
_MODEL_RE = re.compile(r'"model"\s*:\s*"([^"]+)"')

def driver_model(orch):
    """Return {"model","harness"} for the top-level orchestrator lane."""
    eng = orch.get("engine") if orch else None
    harness = "Codex" if eng == "codex" else "Claude Code" if eng == "claude" else ""
    model = ""
    path = orch.get("path") if orch else None
    if path and os.path.exists(path):
        try:
            with open(path, encoding="utf-8", errors="replace") as f:
                for ln in f:
                    if '"model"' not in ln or len(ln) > 500000:
                        continue
                    m = _MODEL_RE.search(ln)
                    if m:
                        model = m.group(1); break
        except OSError:
            pass
    return {"model": model, "harness": harness}

# De-dup artifact contents into a module-level BLOBS dict keyed by a stable id
# ("b0","b1",...) assigned by first-seen content (clipped BEFORE hashing).
BLOB_CAP = 40000
BLOBS = {}
_blob_ids = {}   # clipped content -> id
def _blob(content):
    if len(content) > BLOB_CAP:
        content = content[:BLOB_CAP] + "\n… [truncated]"
    bid = _blob_ids.get(content)
    if bid is None:
        bid = f"b{len(_blob_ids)}"
        _blob_ids[content] = bid
        BLOBS[bid] = content
    return bid

def resolve_artifacts(run_visits, io_map):
    """Attach v["reads_meta"] / v["creates_meta"] (lists of {name, from, blob}).
    PRODUCED: each `creates` file read from the visit's OWN vd if non-empty.
    CONSUMED: each `reads` file resolved to the most recent prior visit of the
    SAME run whose vd holds that file non-empty (skip if none). `io_map` is the
    per-run reads/creates table parsed from that run's setup/SKILL.md."""
    ordered = sorted(run_visits,
                     key=lambda x: (x.get("start") or x.get("mtime") or 0))
    for i, v in enumerate(ordered):
        io = io_map.get(v["step"], {})
        vd = v.get("vd", "")
        creates = []
        for fn in io.get("creates", []):
            content = read(os.path.join(vd, fn))
            if content.strip():
                creates.append({"name": fn, "from": "produced here",
                                "blob": _blob(content)})
        reads = []
        for fn in io.get("reads", []):
            for w in reversed(ordered[:i]):
                wc = read(os.path.join(w.get("vd", ""), fn))
                if wc.strip():
                    reads.append({"name": fn,
                                  "from": f'{STEPLABEL[w["step"]]} {w["idx"]}',
                                  "blob": _blob(wc)})
                    break
        v["reads_meta"] = reads
        v["creates_meta"] = creates

def resolve_prompts(run_visits):
    """Attach v["prompt_blob"] = the full user message that was sent to that agent.
    Source order: (1) the visit's own prompt.md if present (the exact persisted
    invocation prompt); else (2) the first genuine user turn in the agent's
    transcript. A single sub-agent session can span several visits (coding/review
    reuse one session across rounds), so the transcript's first user turn is only
    unambiguous for the EARLIEST visit on that session — later visits without a
    prompt.md are left blank rather than mislabeled with round 1's prompt."""
    sid_first = {}
    for v in sorted(run_visits, key=lambda x: (x.get("start") or x.get("mtime") or 0)):
        sid = v.get("sid")
        if sid and sid not in sid_first:
            sid_first[sid] = id(v)
    for v in run_visits:
        text = ""
        pm = os.path.join(v.get("vd", ""), "prompt.md")
        if os.path.exists(pm):
            text = read(pm)
        if not text.strip():
            path, engine = v.get("sess_path"), v.get("engine")
            # transcript fallback only for the first visit on a shared session
            if path and sid_first.get(v.get("sid")) == id(v):
                if engine == "codex":
                    text = scan_first_user_codex(path)
                elif engine == "claude":
                    text = scan_first_user_claude(path)
        if text.strip():
            v["prompt_blob"] = _blob(text)

def js_json(o):
    return json.dumps(o).replace("</", "<\\/")

# ---------- summary run cards (kept, below the pipeline) ----------
def cat_bar(catc, total):
    if total == 0:
        return '<div class="catbar"><span class="cseg" style="width:100%;background:var(--spark-line)"></span></div>'
    segs = []
    for k in ["infra", "product", "setup", "other"]:
        n = catc.get(k, 0)
        if n == 0: continue
        pct = 100.0 * n / total
        segs.append(f'<span class="cseg" title="{k}: {n}" style="width:{pct:.1f}%;background:{CATCOLOR[k]}"></span>')
    return '<div class="catbar">' + "".join(segs) + '</div>'

def progress_spark(progress, total):
    if not progress:
        return '<div class="muted small">no review checkpoints yet</div>'
    total = total or max((p.get("total", 0) for p in progress), default=185)
    W, H = 520, 90
    n = len(progress)
    xs = lambda i: 8 + (W - 16) * (i / max(1, n - 1))
    ys = lambda d: H - 10 - (H - 20) * (d / max(1, total))
    pts = " ".join(f"{xs(i):.1f},{ys(p['done']):.1f}" for i, p in enumerate(progress))
    dots = "".join(
        f'<circle cx="{xs(i):.1f}" cy="{ys(p["done"]):.1f}" r="3" style="fill:var(--product)">'
        f'<title>{p["step"]} {p["idx"]}: {p["done"]}/{p.get("total",total)} done</title></circle>'
        for i, p in enumerate(progress))
    last = progress[-1]
    return (f'<svg class="spark" viewBox="0 0 {W} {H}" preserveAspectRatio="none">'
            f'<line x1="8" y1="{ys(0):.1f}" x2="{W-8}" y2="{ys(0):.1f}" style="stroke:var(--spark-line)"/>'
            f'<polyline points="{pts}" fill="none" style="stroke:var(--product)" stroke-width="2"/>'
            f'{dots}</svg>'
            f'<div class="small muted">checked plan boxes across {n} checkpoints · '
            f'ends at <b style="color:var(--text)">{last["done"]}/{last.get("total",total)}</b></div>')

def run_card(m):
    cc = m["counts"]; ncommit = len(m["commits"]); catc = m["catc"]
    infra_pct = (100 * catc["infra"] // ncommit) if ncommit else 0
    prod_pct = (100 * catc["product"] // ncommit) if ncommit else 0
    status_pill = ('<span class="pill live">● live</span>' if m["status"] == "live"
                   else '<span class="pill stopped">■ stopped</span>')
    return f"""
    <section class="runcard">
      <div class="rc-head">
        <div>
          <div class="rc-title">{esc(m["label"])} {status_pill}</div>
          <div class="rc-sub">{esc(m["sub"])}</div>
          <div class="rc-id">{esc(m["id"])}</div>
        </div>
        <div class="rc-prog">
          <div class="big">{m["done"]}<span>/{m["total"] or "?"}</span></div>
          <div class="small muted">plan boxes done</div>
          <div class="rc-dur" title="wall-clock from first activity to last">⏱ {esc(m["dur"])}{(' · ' + esc(m["span"])) if m["span"] else ''}</div>
        </div>
      </div>
      <div class="rc-stats">
        <div class="st"><b>{cc["coding"]}</b><span>coding</span></div>
        <div class="st"><b>{cc["review"]}</b><span>review</span></div>
        <div class="st"><b>{cc["validation"]}</b><span>validation</span></div>
        <div class="st"><b>{ncommit}</b><span>commits</span></div>
        <div class="st"><b style="color:#e0982e">{infra_pct}%</b><span>infra</span></div>
        <div class="st"><b style="color:#3fb6a8">{prod_pct}%</b><span>product</span></div>
      </div>
      <div class="rc-block"><div class="section-title">Build mix</div>{cat_bar(catc, ncommit)}</div>
      <div class="rc-block"><div class="section-title">Plan progress</div>{progress_spark(m["progress"], m["total"])}</div>
    </section>"""

def render_doc(records, metas, tabs_html, timelines_html, comblegend_html,
               pipes_html, cards_html, stats_html, insight_html, title,
               subtitle_html, footer_ids, VMETA, CMETA, MMETA, LANEMETA, gen_time):
  HTMLDOC = f"""<!doctype html>
<html lang="en" data-theme="dark-original"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{esc(title)}</title>
<script>/* apply saved theme before paint to avoid a flash */
(function(){{try{{var t=localStorage.getItem('gri_theme');
if(t==='dark-original'||t==='dark-diffusion'||t==='light-diffusion')
document.documentElement.setAttribute('data-theme',t);}}catch(e){{}}}})();</script>
<style>
:root,:root[data-theme="dark-original"]{{
--mono:ui-monospace,"SF Mono",Menlo,monospace;
--bg:#07090d;--panel:#0d1119;--panel2:#111722;--line:#263143;--line2:#182031;
--muted:#7f8da3;--text:#e6edf6;--text-soft:#c5cfdb;--text-soft2:#cdd6e2;--text-soft3:#d6dfea;
--text-check:#d6e8e2;--text-metric:#aeb9c8;--text-dim:#556;--text-dim2:#8fa1b8;
--infra:#e0982e;--product:#3fb6a8;--setup:#7f8da3;--other:#8b6fc7;--accent:#3fb6a8;
--insight-a:#12181f;--insight-b:#0d1119;--spark-line:#1a2230;
--tl-scrollbg:#080b11;--tl-lane-a:#0b1018;--tl-lane-b:#0e131c;--tl-tick:#263143;
--tl-ticktext:#718198;--tl-bar-text:#eef5ff;--tl-break:#e0982e;
--btn-on-bg:#141c28;--btn-on-border:#3a4a63;--tip-bg:#090d14;--tip-border:#40516c;
--role-driver:#9db4ff;--role-requirements:#27e1c1;--role-plan:#65a6ff;--role-plan-critique:#ffbe5c;
--role-plan-update:#9478ff;--role-coding:#32d7ff;--role-validation:#ff9f43;--role-review:#d880ff;
--cat-infra:#e0982e;--cat-product:#3fb6a8;--cat-setup:#7f8da3;--cat-other:#8b6fc7;
color-scheme:dark;}}
/* Dark Diffusion — derived from styleguide-visual diffusion-app.css [data-dx-theme="dark"]
   (bg=midnight #000, surface=ink #0e0f0f, surface-raised=ink-2 #1a1b1b, text=white,
   text-soft/faint rgba white .68/.56, border rgba white .14) + brand palette
   (charge #c8ff00, resolve #00e5a0, coherence #4da6ff, ignition #ff6835,
   corona #ffef5b, verve #b869ff, liftoff #ff5b7c). */
:root[data-theme="dark-diffusion"]{{
--bg:#000000;--panel:#0e0f0f;--panel2:#1a1b1b;--line:rgba(255,255,255,0.14);--line2:rgba(255,255,255,0.08);
--muted:rgba(255,255,255,0.56);--text:#ffffff;--text-soft:rgba(255,255,255,0.68);--text-soft2:rgba(255,255,255,0.68);
--text-soft3:rgba(255,255,255,0.74);--text-check:#00e5a0;--text-metric:rgba(255,255,255,0.56);
--text-dim:rgba(255,255,255,0.42);--text-dim2:rgba(255,255,255,0.56);
--infra:#ff6835;--product:#00e5a0;--setup:#9a9d9d;--other:#b869ff;--accent:#c8ff00;
--insight-a:#1a1b1b;--insight-b:#0e0f0f;--spark-line:rgba(255,255,255,0.12);
--tl-scrollbg:#0e0f0f;--tl-lane-a:#121313;--tl-lane-b:#171818;--tl-tick:rgba(255,255,255,0.14);
--tl-ticktext:rgba(255,255,255,0.56);--tl-bar-text:#0e0f0f;--tl-break:#ff6835;
--btn-on-bg:#1a1b1b;--btn-on-border:rgba(255,255,255,0.30);--tip-bg:#0e0f0f;--tip-border:rgba(255,255,255,0.22);
--role-driver:#4da6ff;--role-requirements:#00e5a0;--role-plan:#7cc4ff;--role-plan-critique:#ffef5b;
--role-plan-update:#b869ff;--role-coding:#c8ff00;--role-validation:#ff6835;--role-review:#ff5b7c;
--cat-infra:#ff6835;--cat-product:#00e5a0;--cat-setup:#9a9d9d;--cat-other:#b869ff;
color-scheme:dark;}}
/* Light Diffusion — derived from diffusion-app.css [data-dx-theme="light"]
   (bg=snow #f4f5f6, surface=white, surface-inset #e9ebed, text=ink #0e0f0f,
   text-soft #555b60, text-faint #6b7176, border rgba black .14, success-ink
   #047857, warning-ink #a83b0b, error-ink #c81e3c, info-ink #1d4ed8). Accents are
   saturated brand-derived tones chosen to stay legible on the light surface. */
:root[data-theme="light-diffusion"]{{
--bg:#f4f5f6;--panel:#ffffff;--panel2:#ffffff;--line:rgba(0,0,0,0.14);--line2:rgba(0,0,0,0.08);
--muted:#6b7176;--text:#0e0f0f;--text-soft:#555b60;--text-soft2:#555b60;--text-soft3:#3a3f43;
--text-check:#047857;--text-metric:#555b60;--text-dim:#6b7176;--text-dim2:#555b60;
--infra:#c85a15;--product:#047857;--setup:#6b7176;--other:#8b3fd6;--accent:#047857;
--insight-a:#ffffff;--insight-b:#f4f5f6;--spark-line:rgba(0,0,0,0.10);
--tl-scrollbg:#eef0f2;--tl-lane-a:#f0f1f3;--tl-lane-b:#e9ebed;--tl-tick:rgba(0,0,0,0.12);
--tl-ticktext:#6b7176;--tl-bar-text:#0e0f0f;--tl-break:#c85a15;
--btn-on-bg:#e6e8ea;--btn-on-border:rgba(0,0,0,0.30);--tip-bg:#ffffff;--tip-border:rgba(0,0,0,0.18);
--role-driver:#1d6fd0;--role-requirements:#05966b;--role-plan:#3a7bd0;--role-plan-critique:#b26a00;
--role-plan-update:#8b3fd6;--role-coding:#3f8f00;--role-validation:#d1500f;--role-review:#c81e5c;
--cat-infra:#c85a15;--cat-product:#047857;--cat-setup:#6b7176;--cat-other:#8b3fd6;
color-scheme:light;}}
*{{box-sizing:border-box}}html,body{{margin:0;background:var(--bg);color:var(--text);
font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}}
.wrap{{max-width:1180px;margin:auto;padding:26px}}
.eyebrow{{font:10px var(--mono);letter-spacing:.22em;text-transform:uppercase;color:var(--muted)}}
h1{{font-size:30px;margin:6px 0 2px}}h1 span{{color:var(--product)}}
h2{{font-size:19px;margin:34px 0 4px}}
.lede{{color:var(--muted);max-width:820px}}
.stats{{display:grid;grid-template-columns:repeat(6,1fr);gap:8px;margin:20px 0}}
.tile{{background:var(--panel);border:1px solid var(--line);border-radius:10px;padding:14px}}
.tile b{{display:block;font-size:26px;line-height:1}}.tile span{{font:10px var(--mono);
letter-spacing:.12em;text-transform:uppercase;color:var(--muted)}}
.insight{{background:linear-gradient(180deg,var(--insight-a),var(--insight-b));border:1px solid var(--line);
border-left:3px solid var(--product);border-radius:10px;padding:16px 18px;margin:6px 0 22px}}
.insight h3{{margin:0 0 8px;font-size:15px}}.insight p{{margin:6px 0;color:var(--text-soft)}}
.small{{font-size:11.5px}}.muted{{color:var(--muted)}}
.legend{{display:flex;gap:14px;flex-wrap:wrap;margin:10px 0 0;font-size:12px;color:var(--muted)}}
.legend b{{display:inline-block;width:9px;height:9px;border-radius:2px;margin-right:5px}}

/* ---- pipeline ---- */
.pipe-tabs{{display:flex;gap:6px;margin:10px 0 6px}}
.pipe-tabs button{{font:10px var(--mono);text-transform:uppercase;letter-spacing:.1em;
background:var(--panel2);color:var(--muted);border:1px solid var(--line);border-radius:7px;
padding:7px 12px;cursor:pointer}}
.pipe-tabs button.on{{color:var(--text);border-color:var(--btn-on-border);background:var(--btn-on-bg)}}
.tl-controls{{display:flex;justify-content:space-between;align-items:center;gap:12px;flex-wrap:wrap;margin:10px 0 6px}}
.tl-controls .pipe-tabs{{margin:0}}
.tl-fit{{display:flex;gap:6px;align-items:center}}
.tl-fit .tlf-lab{{font:9px var(--mono);letter-spacing:.12em;text-transform:uppercase;color:var(--muted)}}
.tl-fit button{{font:10px var(--mono);text-transform:uppercase;letter-spacing:.1em;
background:var(--panel2);color:var(--muted);border:1px solid var(--line);border-radius:7px;
padding:7px 12px;cursor:pointer}}
.tl-fit button.on{{color:var(--text);border-color:var(--btn-on-border);background:var(--btn-on-bg)}}
.prun{{font:9px var(--mono);text-transform:uppercase;letter-spacing:.06em;border:1px solid;
border-radius:4px;padding:1px 6px}}
.trun{{font:9px var(--mono);text-transform:uppercase;letter-spacing:.06em;border:1px solid;
border-radius:4px;padding:1px 5px;margin-left:2px}}
.stepfilter{{display:flex;gap:5px;flex-wrap:wrap;margin:0 0 14px}}
.stepfilter button{{font:9px var(--mono);text-transform:uppercase;letter-spacing:.08em;
background:var(--panel2);color:var(--muted);border:1px solid var(--line);border-radius:6px;
padding:3px 8px;cursor:pointer}}
.stepfilter button.on{{color:var(--text);border-color:var(--btn-on-border)}}
/* ---- swimlane timeline ---- */
.tl-legend{{display:flex;gap:13px;flex-wrap:wrap;align-items:center;margin:6px 0 8px;font-size:11px;color:var(--muted)}}
.tl-legend .tll-t{{font:9px var(--mono);letter-spacing:.12em;text-transform:uppercase}}
.tl-legend b{{display:inline-block;width:9px;height:9px;border-radius:2px;margin-right:5px}}
.tl{{display:none}}.tl.on{{display:block}}
.tlscroll{{overflow-x:auto;border:1px solid var(--line);border-radius:10px;background:var(--tl-scrollbg)}}
.tlsvg{{display:block;min-width:100%}}
/* theme-swappable SVG structural + accent fills (CSS overrides the baked
   presentation attributes; the baked values equal the Dark Original theme). */
.tlsvg .tl-lane-a{{fill:var(--tl-lane-a)}}
.tlsvg .tl-lane-b{{fill:var(--tl-lane-b)}}
.tlsvg .tl-tickline{{stroke:var(--tl-tick)}}
.tlsvg .tl-ticktext{{fill:var(--tl-ticktext)}}
.tlsvg .tl-bartext{{fill:var(--tl-bar-text)}}
.tlsvg .tl-breakband{{fill:var(--tl-break)}}
.tlsvg .tl-breakline{{stroke:var(--tl-break)}}
.tlsvg .tl-breaklab{{fill:var(--tl-break)}}
.tlsvg .role-driver{{fill:var(--role-driver)}}
.tlsvg .role-requirements{{fill:var(--role-requirements)}}
.tlsvg .role-plan{{fill:var(--role-plan)}}
.tlsvg .role-plan-critique{{fill:var(--role-plan-critique)}}
.tlsvg .role-plan-update{{fill:var(--role-plan-update)}}
.tlsvg .role-coding{{fill:var(--role-coding)}}
.tlsvg .role-validation{{fill:var(--role-validation)}}
.tlsvg .role-review{{fill:var(--role-review)}}
.tlsvg .cat-infra{{fill:var(--cat-infra)}}
.tlsvg .cat-product{{fill:var(--cat-product)}}
.tlsvg .cat-setup{{fill:var(--cat-setup)}}
.tlsvg .cat-other{{fill:var(--cat-other)}}
.tlsvg .lanehit{{cursor:help}}
/* theme chooser (top-right) */
.themepick{{position:fixed;top:14px;right:14px;z-index:200;display:flex;align-items:center;
background:var(--panel2);border:1px solid var(--line);border-radius:8px;padding:3px;box-shadow:0 6px 20px rgba(0,0,0,.35)}}
.themepick .tp-lab{{font:9px var(--mono);letter-spacing:.12em;text-transform:uppercase;color:var(--muted);padding:0 8px}}
.themepick button{{font:10px var(--mono);text-transform:uppercase;letter-spacing:.08em;background:transparent;
color:var(--muted);border:0;border-radius:6px;padding:6px 10px;cursor:pointer}}
.themepick button.on{{background:var(--btn-on-bg);color:var(--text)}}
/* model+harness line in the click inspector */
.tli-mh{{font:10px var(--mono);letter-spacing:.06em;color:var(--muted);margin:0 0 8px}}
.tli-mh b{{color:var(--text)}}
/* Geometry variants: gaps {{full|collapsed}} x off-run {{shown|hidden}}.
   Exactly one is display:block per (collapsed, hide-offrun) state combo.
   Off-run-hidden variants use a re-flowed x-mapping that reclaims the space
   the off-run commits occupied (space removed, not merely blanked). Views
   with no off-run commits alias fs=fh and cs=ch (same SVG, both classes). */
.tlvar{{display:none}}
.tl:not(.collapsed):not(.hide-offrun) .tlvar-fs{{display:block}}
.tl:not(.collapsed).hide-offrun .tlvar-fh{{display:block}}
.tl.collapsed:not(.hide-offrun) .tlvar-cs{{display:block}}
.tl.collapsed.hide-offrun .tlvar-ch{{display:block}}
/* Off-run diamonds are already omitted from the hidden-geometry variants;
   this is a belt-and-suspenders CSS hide for any that remain. */
.tl.hide-offrun .offrun{{display:none}}
/* Fit mode: scale the whole SVG to the panel width, no horizontal scroll */
.tl.fit .tlscroll{{overflow-x:hidden}}
.tl.fit .tlsvg{{width:100%;min-width:0;height:auto}}
.tlhint{{font-size:11.5px;color:var(--muted);margin:8px 0}}
.tltip{{position:fixed;z-index:100;display:none;pointer-events:none;background:var(--tip-bg);
border:1px solid var(--tip-border);border-radius:8px;padding:9px 10px;box-shadow:0 12px 40px rgba(0,0,0,.5);
max-width:340px;font-size:11.5px;line-height:1.45}}
.tltip strong{{color:var(--text);font:11px var(--mono)}}
.tltip .tl-time{{color:var(--text-dim2);font:10px var(--mono);margin:3px 0}}
.tltip .tl-m{{color:var(--text-soft2);margin:2px 0}}
.tltip .tl-cm{{color:var(--text-soft2);margin-top:4px}}
.tvb{{font:9px var(--mono);text-transform:uppercase;border:1px solid transparent;border-radius:4px;padding:1px 5px}}
.tcat{{font:9px var(--mono);text-transform:uppercase}}
.tlinspector{{border:1px solid var(--line);background:var(--panel);border-radius:10px;padding:14px 16px;min-height:70px;margin-bottom:6px}}
.tli-empty{{color:var(--muted);font-size:12.5px}}
.tli-head h4{{margin:0 0 4px;font-size:15px}}
.tli-sum{{color:var(--text-soft);font-size:12.5px;margin:4px 0 10px}}
.tli-kv{{display:grid;grid-template-columns:110px 1fr;gap:8px;padding:5px 0;border-bottom:1px solid var(--line2);font-size:12px}}
.tli-kv span:first-child{{color:var(--muted);font:10px var(--mono);text-transform:uppercase;letter-spacing:.06em}}
.tli-kv span:last-child{{color:var(--text-soft3)}}
.tli-h{{font:9px var(--mono);text-transform:uppercase;letter-spacing:.1em;color:var(--muted);margin:12px 0 5px}}
.tli-ck{{margin:0;padding-left:16px}}.tli-ck li{{font-size:12px;color:var(--text-check);margin:2px 0}}
.tli-c{{display:flex;gap:9px;align-items:baseline;font-size:12px;padding:2px 0}}
.tli-hash{{font:11px var(--mono);color:#8fa;flex:none;width:56px}}
.tli-msg{{color:var(--text-soft2);flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}}
.tli-sz{{font:10px var(--mono);flex:none}}
/* consumed/produced artifact disclosures */
.tli-sec{{font:9px var(--mono);text-transform:uppercase;letter-spacing:.1em;color:var(--muted);margin:14px 0 6px}}
.artdisc{{border:1px solid var(--line);background:var(--panel2);border-radius:7px;margin:0 0 5px;overflow:hidden}}
.artdisc summary{{cursor:pointer;list-style:none;display:flex;gap:9px;align-items:baseline;
padding:6px 9px;font-size:12px}}
.artdisc summary::-webkit-details-marker{{display:none}}
.artdisc summary::before{{content:'▸';color:var(--muted);font-size:10px}}
.artdisc[open] summary::before{{content:'▾'}}
.artdisc summary .art-name{{font:11px var(--mono);color:var(--text-soft3)}}
.artdisc summary .art-from{{font:10px var(--mono);color:var(--muted);margin-left:auto}}
.artbody{{margin:0;padding:9px 11px;font:11px var(--mono);white-space:pre-wrap;word-break:break-word;
max-height:320px;overflow:auto;color:var(--text-soft);background:var(--tl-scrollbg);border-top:1px solid var(--line)}}
/* tooltip checked-off bullet list */
.tiptip{{margin:3px 0 0;padding-left:16px;color:var(--text-check)}}
.tiptip li{{font-size:11px;margin:1px 0}}
.pipe-sub{{font-size:15px;color:var(--text-soft);margin:26px 0 10px;padding-top:16px;border-top:1px solid var(--line)}}
.vmetrics{{display:flex;flex-wrap:wrap;gap:12px;margin:7px 0 2px}}
.vm{{font:11px var(--mono);color:var(--text-metric)}}
.pipe{{display:none;position:relative;padding-left:26px}}
.pipe.on{{display:block}}
.pline{{position:absolute;left:6px;top:8px;bottom:8px;width:2px;background:linear-gradient(180deg,#1c2636,#263143,#1c2636)}}
.pvisit{{position:relative;margin:0 0 12px}}
.pdot{{position:absolute;left:-26px;top:14px;width:11px;height:11px;border-radius:50%}}
.pcard{{background:var(--panel);border:1px solid var(--line);border-radius:10px;padding:11px 13px}}
.pcard-head{{display:flex;align-items:center;gap:9px;flex-wrap:wrap}}
.pstage{{font:10px var(--mono);text-transform:uppercase;letter-spacing:.08em;border:1px solid;
border-radius:5px;padding:2px 7px}}
.pidx{{font:11px var(--mono);color:var(--muted)}}
.vbadge{{font:9px var(--mono);text-transform:uppercase;letter-spacing:.06em;border:1px solid;
border-radius:4px;padding:1px 6px}}
.vprog{{font:10px var(--mono);color:var(--text-soft2)}}
.ptime{{font:10px var(--mono);color:var(--text-dim);margin-left:auto}}
.vnext{{font:10px var(--mono);color:var(--muted)}}
.vsummary{{color:var(--text-soft);font-size:12.5px;margin:8px 0 2px}}
.vsub{{margin-top:9px;padding-top:8px;border-top:1px solid var(--line2)}}
.vsub-t{{font:9px var(--mono);text-transform:uppercase;letter-spacing:.1em;color:var(--muted);display:block;margin-bottom:5px}}
.pchecks{{margin:0;padding-left:16px}}
.pchecks li{{font-size:12px;color:var(--text-check);margin:2px 0}}
.pcommits{{display:flex;flex-direction:column;gap:3px}}
.pcommit{{display:flex;gap:8px;align-items:baseline;font-size:12px}}
.pc-hash{{font:11px var(--mono);flex:none;width:56px}}
.pc-msg{{color:var(--text-soft2);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}}

/* ---- summary run cards ---- */
.grid{{display:grid;grid-template-columns:1fr 1fr;gap:16px}}
@media(max-width:1000px){{.grid{{grid-template-columns:1fr}}.stats{{grid-template-columns:repeat(3,1fr)}}}}
.runcard{{background:var(--panel);border:1px solid var(--line);border-radius:12px;padding:16px}}
.rc-head{{display:flex;justify-content:space-between;gap:12px;align-items:flex-start}}
.rc-title{{font-size:16px;font-weight:600}}
.rc-sub{{color:var(--muted);font-size:12px;margin-top:2px}}
.rc-id{{font:10px var(--mono);color:var(--text-dim);margin-top:4px}}
.rc-prog{{text-align:right}}.rc-prog .big{{font-size:30px;font-weight:700;color:var(--product)}}
.rc-prog .big span{{font-size:15px;color:var(--muted);font-weight:400}}
.rc-dur{{margin-top:8px;font:11px var(--mono);color:var(--text-soft);white-space:nowrap}}
.rc-stats{{display:grid;grid-template-columns:repeat(6,1fr);gap:6px;margin:14px 0}}
.st{{background:var(--panel2);border:1px solid var(--line);border-radius:8px;padding:8px 4px;text-align:center}}
.st b{{display:block;font-size:17px}}.st span{{font:9px var(--mono);letter-spacing:.08em;text-transform:uppercase;color:var(--muted)}}
.rc-block{{margin-top:14px}}
.section-title{{font:10px var(--mono);letter-spacing:.16em;text-transform:uppercase;color:var(--muted);margin-bottom:7px}}
.catbar{{display:flex;height:14px;border-radius:7px;overflow:hidden;border:1px solid var(--line)}}
.catbar .cseg{{height:100%}}
.spark{{width:100%;height:90px;display:block;background:var(--panel2);border:1px solid var(--line);border-radius:8px}}
.pill{{font:9px var(--mono);text-transform:uppercase;letter-spacing:.08em;border-radius:5px;padding:2px 6px}}
.pill.live{{color:#3fb6a8;border:1px solid #3fb6a855;background:#3fb6a814}}
.pill.stopped{{color:#e0982e;border:1px solid #e0982e55;background:#e0982e14}}
footer{{margin-top:26px;color:var(--text-dim);font:10px var(--mono);letter-spacing:.1em}}
</style></head>
<body>
<div class="themepick" id="themePick">
  <span class="tp-lab">theme</span>
  <button data-theme="dark-original" class="on">Dark Original</button>
  <button data-theme="dark-diffusion">Dark Diffusion</button>
  <button data-theme="light-diffusion">Light Diffusion</button>
</div>
<div class="wrap">
  <div class="eyebrow">EasyLoop Reconstruction · generated {esc(gen_time)}</div>
  <h1>{subtitle_html}</h1>
  <p class="lede">Every run drives its own worktree against its own plan. Data read
  live from <code>~/.easyloop/runs</code> and each run's worktree git log, discovered
  per-run at build time.</p>

  <div class="stats">
    {stats_html}
  </div>

  <h2>The pipeline — every visit, in order</h2>
  <p class="lede small">The heart of an EasyLoop run: each agent visit in sequence, with its
  verdict, the commits it produced, and the plan boxes it knocked off. The <b>E2E</b> flow
  starts with requirements and validates each coding round; the <b>Easy</b> flow starts from
  a supplied spec and loops directly between coding and review. Each run shows only the
  lanes defined by its own copied skill.</p>
  <div class="tl-controls">
    <div class="pipe-tabs" id="pipeTabs">
      {tabs_html}
    </div>
    <div class="tl-fit" id="tlFit">
      <span class="tlf-lab">width:</span>
      <button class="on" data-mode="scroll">Scroll</button>
      <button data-mode="fit">Fit</button>
    </div>
    <div class="tl-fit" id="tlCollapse">
      <span class="tlf-lab">gaps:</span>
      <button class="on" data-mode="show">Full</button>
      <button data-mode="collapse">Collapse gaps</button>
    </div>
    <div class="tl-fit" id="tlOffrun">
      <span class="tlf-lab">off-run commits:</span>
      <button class="on" data-mode="show">Show</button>
      <button data-mode="hide">Hide</button>
    </div>
  </div>

  <div class="tl-legend">
    <span class="tll-t">roles:</span>
    <span><b style="background:var(--role-driver)"></b>driver</span>
    <span><b style="background:var(--role-requirements)"></b>requirements</span>
    <span><b style="background:var(--role-plan)"></b>plan</span>
    <span><b style="background:var(--role-plan-critique)"></b>critique</span>
    <span><b style="background:var(--role-plan-update)"></b>plan-update</span>
    <span><b style="background:var(--role-coding)"></b>coding</span>
    <span><b style="background:var(--role-validation)"></b>validation</span>
    <span><b style="background:var(--role-review)"></b>review</span>
    <span style="margin-left:auto">◆ commit (size = churn) · <b style="background:#5ad1ff"></b>driver steering msg · <span style="color:#ff9f43">⊕</span> compaction · ◹ hatched = fail/incomplete · ◯ review = boxes confirmed</span>
  </div>
  <div class="tl-legend" id="combLegend" style="display:none">
    <span class="tll-t">combined runs:</span>
    {comblegend_html}
    <span style="margin-left:auto">all runs on one wall-clock axis · full commit history</span>
  </div>
  {timelines_html}
  <div class="tlhint">Role swimlanes, time left → right — rebuilt from the old visualization's timeline.
  Hover a bar or commit diamond for detail; click to pin it in the inspector below.
  <b>Combined</b> merges every run on one wall-clock axis and plots the full commit history back to the first commit.
  Use <b>Scroll / Fit</b> (top right) to keep the fixed time-scale (horizontal scroll) or shrink the whole span to fit.</div>
  <div id="tlinspect" class="tlinspector"><div class="tli-empty">Select any visit bar or commit
  diamond above — its commits, commit size, and checkboxes suggested vs checked show here.</div></div>

  <h3 class="pipe-sub">Same run as a vertical pipeline</h3>
  <div class="stepfilter" id="stepFilter">
    <button class="on" data-step="all">all stages</button>
    <button data-step="requirements">requirements</button>
    <button data-step="plan">plan</button>
    <button data-step="plan-critique">critique</button>
    <button data-step="plan-update">plan-update</button>
    <button data-step="coding">coding</button>
    <button data-step="validation">validation</button>
    <button data-step="review">review</button>
  </div>
  {pipes_html}

  <h2>Run summaries</h2>
  <div class="legend">
    <span><b style="background:var(--infra)"></b>infra / plumbing</span>
    <span><b style="background:var(--product)"></b>product</span>
    <span><b style="background:var(--setup)"></b>setup / scaffold</span>
    <span><b style="background:var(--other)"></b>other</span>
  </div>
  <div class="grid" style="margin-top:12px">
    {cards_html}
  </div>

  <footer>SELF-CONTAINED · NO NETWORK · generated {esc(gen_time)} from ~/.easyloop + git log ·
  runs {esc(footer_ids)}</footer>
</div>
<div id="tltip" class="tltip"></div>
<script>
const VMETA={js_json(VMETA)};
const CMETA={js_json(CMETA)};
const MMETA={js_json(MMETA)};
const LANEMETA={js_json(LANEMETA)};
const BLOBS={js_json(BLOBS)};
const VCOL={{pass:'#3fb6a8',fail:'#ff5f6d',complete:'#3fb6a8',incomplete:'#e0982e'}};
const RUNCOL={js_json(RUNCOL)};
const RUNLAB={js_json(RUNLABEL)};
const esc=s=>String(s==null?'':s).replace(/[&<>]/g,c=>({{'&':'&amp;','<':'&lt;','>':'&gt;'}}[c]));
function runBadge(r){{if(!r)return'';const c=RUNCOL[r]||'#7f8da3';return `<span class="trun" style="color:${{c}};border-color:${{c}}55;background:${{c}}14">${{esc(RUNLAB[r]||r)}}</span>`;}}
// ---- theme chooser (Dark Original / Dark Diffusion / Light Diffusion) ----
const THEMES=['dark-original','dark-diffusion','light-diffusion'];
const themeBtns=[...document.querySelectorAll('#themePick button')];
function applyTheme(t){{
  if(!THEMES.includes(t)) t='dark-original';
  document.documentElement.setAttribute('data-theme',t);
  themeBtns.forEach(x=>x.classList.toggle('on',x.dataset.theme===t));
  try{{localStorage.setItem('gri_theme',t);}}catch(e){{}}
}}
themeBtns.forEach(b=>b.onclick=()=>applyTheme(b.dataset.theme));
(function initTheme(){{let t='dark-original';try{{const s=localStorage.getItem('gri_theme');if(s)t=s;}}catch(e){{}}applyTheme(t);}})();
// run toggle — controls BOTH the swimlane timeline and the vertical pipeline
const combLegend=document.getElementById('combLegend');
const tabs=[...document.querySelectorAll('#pipeTabs button')];
tabs.forEach(b=>b.onclick=()=>{{
  tabs.forEach(x=>x.classList.remove('on'));b.classList.add('on');
  document.querySelectorAll('.pipe').forEach(p=>
    p.classList.toggle('on', p.dataset.run===b.dataset.run));
  document.querySelectorAll('.tl').forEach(p=>
    p.classList.toggle('on', p.dataset.run===b.dataset.run));
  if(combLegend) combLegend.style.display = (b.dataset.run==='combined')?'flex':'none';
}});
// per-view Scroll / Fit toggle — applies to whichever timeline is shown
const fitBtns=[...document.querySelectorAll('#tlFit button')];
fitBtns.forEach(b=>b.onclick=()=>{{
  fitBtns.forEach(x=>x.classList.remove('on'));b.classList.add('on');
  const fit=b.dataset.mode==='fit';
  document.querySelectorAll('.tl').forEach(p=>p.classList.toggle('fit',fit));
}});
// Collapse-gaps toggle — removes empty spans >1h, applies to whichever view is shown
const colBtns=[...document.querySelectorAll('#tlCollapse button')];
colBtns.forEach(b=>b.onclick=()=>{{
  colBtns.forEach(x=>x.classList.remove('on'));b.classList.add('on');
  const collapse=b.dataset.mode==='collapse';
  document.querySelectorAll('.tl').forEach(p=>p.classList.toggle('collapsed',collapse));
}});
// Off-run commits toggle — show/hide commit diamonds that fired while no run
// was active (not covered by any active visit span); default Show.
const offBtns=[...document.querySelectorAll('#tlOffrun button')];
offBtns.forEach(b=>b.onclick=()=>{{
  offBtns.forEach(x=>x.classList.remove('on'));b.classList.add('on');
  const hide=b.dataset.mode==='hide';
  document.querySelectorAll('.tl').forEach(p=>p.classList.toggle('hide-offrun',hide));
}});
// ---- swimlane tooltip + inspector ----
const tip=document.getElementById('tltip');
const insp=document.getElementById('tlinspect');
function metric(v){{
  let m=[];
  if(v.nCommit) m.push(`◆ ${{v.nCommit}} commit${{v.nCommit!=1?'s':''}} · <b style="color:#3fb6a8">+${{v.ins}}</b> <b style="color:#ff5f6d">−${{v.del}}</b> · ${{v.files}} file${{v.files!=1?'s':''}}`);
  if(v.step==='coding') m.push(`☐ ${{v.suggested||0}} boxes suggested · <b style="color:#3fb6a8">☑ ${{v.review_confirmed||0}} confirmed by review</b>`);
  else if((v.step==='review'||v.step==='plan-update')&&v.confirmed) m.push(`☑ ${{v.confirmed}} newly confirmed done`);
  return m;
}}
// coding-round proposed boxes, colored by review outcome:
// green = accepted, red = rejected, yellow = not yet reviewed (pending).
function propStatusItems(arr,cap){{
  const COL={{accepted:'#3fb6a8',rejected:'#ff5f6d',pending:'#e0b93a'}};
  const MK={{accepted:'☑',rejected:'☒',pending:'☐'}};
  let items=arr.slice(0,cap).map(p=>`<li style="color:${{COL[p.s]||'#e0b93a'}}">${{MK[p.s]||'☐'}} ${{esc(p.t)}}</li>`).join('');
  let more=arr.length>cap?`<li style="color:var(--text-soft2)">+${{arr.length-cap}} more</li>`:'';
  return items+more;
}}
function visitTip(v){{
  let vb=v.verdict?`<span class="tvb" style="color:${{VCOL[v.vkind]||'#7f8da3'}}">${{esc(v.verdict)}}</span>`:'';
  let mm=metric(v).map(x=>`<div class="tl-m">${{x}}</div>`).join('');
  let done=(v.done!=null)?`<div class="tl-m">plan boxes: ${{v.done}}/${{v.total}}</div>`:'';
  let cks='';
  if((v.step==='review'||v.step==='plan-update')&&v.checks&&v.checks.length){{
    let items=v.checks.slice(0,12).map(c=>`<li>${{esc(c)}}</li>`).join('');
    let more=v.checks.length>12?`<li>+${{v.checks.length-12}} more</li>`:'';
    cks=`<div class="tl-m">Checked off:</div><ul class="tiptip">${{items}}${{more}}</ul>`;
  }} else if(v.step==='coding'&&v.proposed_status&&v.proposed_status.length){{
    cks=`<div class="tl-m">Boxes suggested <span style="color:#3fb6a8">accepted</span> · <span style="color:#e0b93a">pending</span> · <span style="color:#ff5f6d">rejected</span>:</div><ul class="tiptip">${{propStatusItems(v.proposed_status,14)}}</ul>`;
  }}
  let idl=(v.idle&&v.idle.length)?`<div class="tl-m">idle: <b style="color:#b6c1d4">${{esc(v.idle_total)}}</b> (${{v.idle.length}} gap${{v.idle.length>1?'s':''}} &gt; 10m)</div>`:'';
  return `<strong>${{esc(v.lab)}}</strong> ${{runBadge(v.run)}}${{vb}}<div class="tl-time">${{esc(v.t0)}} → ${{esc(v.t1)}} · ${{esc(v.dur)}}${{v.next?' · → '+esc(v.next):''}}</div>${{mm}}${{idl}}${{done}}${{cks}}`;
}}
function commitTip(c){{
  return `<strong>${{esc(c.h)}}</strong> ${{runBadge(c.run)}}<span class="tcat" style="color:${{catcol(c.cat)}}">${{esc(c.cat)}}</span><div class="tl-time">${{esc(c.t)}} · <b style="color:#3fb6a8">+${{c.ins}}</b> <b style="color:#ff5f6d">−${{c.del}}</b> · ${{c.files}}f</div><div class="tl-cm">${{esc(c.msg)}}</div>`;
}}
function catcol(c){{return({{infra:'#e0982e',product:'#3fb6a8',setup:'#7f8da3',other:'#8b6fc7'}})[c]||'#8b6fc7';}}
function clipJS(s,n){{s=String(s==null?'':s);return s.length<=n?s:s.slice(0,n-1)+'…';}}
function markerTip(m){{
  if(m.kind==='compaction') return `<strong>Compaction${{m.n?' #'+m.n:''}}</strong> ${{runBadge(m.run)}}<span class="tcat" style="color:#ff9f43">${{esc(m.src)}}${{m.engine?' · '+esc(m.engine):''}}</span><div class="tl-time">${{esc(m.t)}}${{m.sid?' · '+esc(m.sid):''}}</div><div class="tl-cm">context window compacted${{m.trigger?' ('+esc(m.trigger)+')':''}}</div>`;
  return `<strong>Steering message</strong> ${{runBadge(m.run)}}<span class="tcat" style="color:#5ad1ff">driver</span><div class="tl-time">${{esc(m.t)}}</div><div class="tl-cm">${{esc(clipJS(m.text,200))}}</div>`;
}}
function markerPanel(m){{
  if(m.kind==='compaction'){{const eng=m.engine==='claude'?'Claude Code <code>compact_boundary</code>':'Codex <code>type: compacted</code>';return `<div class="tli-head"><h4>Compaction${{m.n?' #'+m.n:''}} ${{runBadge(m.run)}}<span class="tcat" style="color:#ff9f43">${{esc(m.src)}}${{m.engine?' · '+esc(m.engine):''}}</span></h4></div><p class="tli-sum">Context-window compaction of the ${{esc(m.src)}} session (${{eng}} event)${{m.trigger?', trigger '+esc(m.trigger):''}}.</p>${{kv('time',esc(m.t))}}${{m.lane?kv('lane',esc(m.lane)):''}}${{m.engine?kv('engine',esc(m.engine)):''}}${{m.trigger?kv('trigger',esc(m.trigger)):''}}${{m.sid?kv('session',esc(m.sid)):''}}`;}}
  return `<div class="tli-head"><h4>Steering message ${{runBadge(m.run)}}<span class="tcat" style="color:#5ad1ff">driver</span></h4></div><p class="tli-sum">${{esc(m.text)}}</p>${{kv('time',esc(m.t))}}`;
}}
function dataFor(el){{
  const run=el.dataset.run;
  if(el.dataset.i!=null) return {{kind:'v',d:VMETA[run][+el.dataset.i]}};
  if(el.dataset.ci!=null) return {{kind:'c',d:CMETA[run][+el.dataset.ci]}};
  if(el.dataset.mi!=null) return {{kind:'m',d:MMETA[run][+el.dataset.mi]}};
  return null;
}}
document.querySelectorAll('.tlsvg .hit').forEach(el=>{{
  el.style.cursor='pointer';
  el.addEventListener('mousemove',e=>{{
    const g=dataFor(el); if(!g||!g.d){{tip.style.display='none';return;}}
    tip.innerHTML=g.kind==='v'?visitTip(g.d):g.kind==='c'?commitTip(g.d):markerTip(g.d);
    tip.style.display='block';
    let x=e.clientX+14,y=e.clientY+14;
    const w=tip.offsetWidth,h=tip.offsetHeight;
    if(x+w>innerWidth-8)x=e.clientX-w-14;
    if(y+h>innerHeight-8)y=e.clientY-h-14;
    tip.style.left=x+'px';tip.style.top=y+'px';
  }});
  el.addEventListener('mouseleave',()=>tip.style.display='none');
  el.addEventListener('click',()=>{{
    const g=dataFor(el); if(!g||!g.d)return;
    insp.innerHTML=g.kind==='v'?visitPanel(g.d):g.kind==='c'?commitPanel(g.d):markerPanel(g.d);
  }});
}});
// ---- row-heading (lane) hover + click: model + harness used for that role ----
function laneRows(lm){{
  if(lm.external) return lm.rows.map(r=>`<div class="tl-m">Source: <strong>${{esc(r.source||'—')}}</strong></div>`).join('');
  return lm.rows.map(r=>`<div class="tl-m">${{r.run?esc(r.run)+' · ':''}}Model: <strong>${{esc(r.model||'—')}}</strong> · Harness: <strong>${{esc(r.harness||'—')}}</strong></div>`).join('');
}}
function laneTip(lm){{return `<strong>${{esc(lm.role)}}</strong><div class="tl-time">${{lm.external?'external timeline row':'row heading — agent config'}}</div>${{lm.description?`<div class="tl-cm">${{esc(lm.description)}}</div>`:''}}${{laneRows(lm)}}`;}}
function lanePanel(lm){{
  if(lm.external) return `<div class="tli-head"><h4>${{esc(lm.role)}}</h4></div><p class="tli-sum">${{esc(lm.description||'Imported timeline intervals')}}</p>${{lm.rows.map(r=>kv('source',esc(r.source||'—'))).join('')}}`;
  let rows=lm.rows.map(r=>{{let pre=r.run?esc(r.run)+' ':'';return kv(pre+'model',esc(r.model||'—'))+kv(pre+'harness',esc(r.harness||'—'));}}).join('');
  return `<div class="tli-head"><h4>${{esc(lm.role)}}</h4></div><p class="tli-sum">Row heading — model &amp; coding harness for this role${{lm.rows.length>1?' (per run)':''}}.</p>${{rows}}`;
}}
document.querySelectorAll('.tlsvg .lanehit').forEach(el=>{{
  el.addEventListener('mousemove',e=>{{
    const m=LANEMETA[el.dataset.run]; const lm=m&&m[el.dataset.lane];
    if(!lm){{tip.style.display='none';return;}}
    tip.innerHTML=laneTip(lm); tip.style.display='block';
    let x=e.clientX+14,y=e.clientY+14; const w=tip.offsetWidth,h=tip.offsetHeight;
    if(x+w>innerWidth-8)x=e.clientX-w-14; if(y+h>innerHeight-8)y=e.clientY-h-14;
    tip.style.left=x+'px';tip.style.top=y+'px';
  }});
  el.addEventListener('mouseleave',()=>tip.style.display='none');
  el.addEventListener('click',()=>{{const m=LANEMETA[el.dataset.run];const lm=m&&m[el.dataset.lane];if(lm)insp.innerHTML=lanePanel(lm);}});
}});
function kv(k,val){{return `<div class="tli-kv"><span>${{k}}</span><span>${{val}}</span></div>`;}}
// Consumed / Produced artifact disclosures (collapsed <details> by default).
function artSec(title,arr){{
  if(!arr||!arr.length) return '';
  const items=arr.map(a=>`<details class="artdisc"><summary><span class="art-name">${{esc(a.name)}}</span><span class="art-from">${{esc(a.from)}}</span></summary><pre class="artbody">${{esc(BLOBS[a.blob]||'')}}</pre></details>`).join('');
  return `<div class="tli-sec">${{title}} (${{arr.length}})</div>${{items}}`;
}}
function visitPanel(v){{
  let vb=v.verdict?`<span class="tvb" style="color:${{VCOL[v.vkind]||'#7f8da3'}};border-color:${{(VCOL[v.vkind]||'#7f8da3')}}55">${{esc(v.verdict)}}</span>`:'';
  let rows=kv('window',`${{esc(v.t0)}} → ${{esc(v.t1)}}`)+kv('duration',esc(v.dur))+(v.next?kv('next',esc(v.next)):'');
  rows+=kv('commits',v.nCommit?`${{v.nCommit}} · <b style="color:#3fb6a8">+${{v.ins}}</b> <b style="color:#ff5f6d">−${{v.del}}</b> · ${{v.files}} files`:'none');
  if(v.step==='coding') rows+=kv('checkboxes',`${{v.suggested||0}} suggested · <b style="color:#3fb6a8">${{v.review_confirmed||0}} confirmed by following review</b>`);
  if((v.step==='review'||v.step==='plan-update')&&v.confirmed!=null) rows+=kv('checkboxes',`${{v.confirmed}} newly confirmed done`);
  if(v.done!=null) rows+=kv('plan total',`${{v.done}}/${{v.total}} boxes`);
  if(v.idle&&v.idle.length) rows+=kv('idle (dead time)',`<b style="color:#b6c1d4">${{esc(v.idle_total)}}</b> · ${{v.idle.length}} gap${{v.idle.length>1?'s':''}} &gt; 10 min`);
  if(v.external&&v.source) rows+=kv('source',esc(v.source));
  let hs=v.hashes&&v.hashes.length?`<div class="tli-h">commits</div>`+v.hashes.map(c=>`<div class="tli-c"><span class="tli-hash">${{esc(c.h)}}</span><span class="tli-msg">${{esc(c.msg)}}</span><span class="tli-sz"><b style="color:#3fb6a8">+${{c.ins}}</b> <b style="color:#ff5f6d">−${{c.del}}</b></span></div>`).join(''):'';
  let ck='';
  if(v.step==='coding'&&v.proposed_status&&v.proposed_status.length){{
    ck=`<div class="tli-h">boxes proposed — <span style="color:#3fb6a8">accepted</span> · <span style="color:#e0b93a">not yet reviewed</span> · <span style="color:#ff5f6d">rejected</span></div><ul class="tli-ck">`+propStatusItems(v.proposed_status,99)+`</ul>`;
  }} else if(v.checks&&v.checks.length){{
    ck=`<div class="tli-h">${{v.checks_kind==='proposed'?'boxes proposed':'boxes confirmed done'}}</div><ul class="tli-ck">`+v.checks.map(c=>`<li>${{esc(c)}}</li>`).join('')+`</ul>`;
  }}
  let idl=v.idle&&v.idle.length?`<div class="tli-h">idle / dead time (event gaps &gt; 10 min)</div><ul class="tli-ck">`+v.idle.map(s=>`<li><b style="color:#b6c1d4">${{esc(s.dur)}}</b> — ${{esc(s.t0)}} → ${{esc(s.t1)}}</li>`).join('')+`</ul>`:'';
  let art=artSec('Consumed',v.reads)+artSec('Produced',v.creates);
  let pt=v.prompt_blob?(BLOBS[v.prompt_blob]||''):'';
  let pr=pt?`<div class="tli-sec">User message sent to the agent</div><details class="artdisc"><summary><span class="art-name">prompt</span><span class="art-from">full</span></summary><pre class="artbody">${{esc(pt)}}</pre></details>`:'';
  let mh=(v.model||v.harness)?`<div class="tli-mh">${{esc(v.step)}} · <b>${{esc(v.model||'—')}}</b> · <b>${{esc(v.harness||'—')}}</b></div>`:'';
  return `<div class="tli-head"><h4>${{esc(v.lab)}} ${{runBadge(v.run)}}${{vb}}</h4></div>${{mh}}${{v.summary?`<p class="tli-sum">${{esc(v.summary)}}</p>`:''}}${{rows}}${{ck}}${{idl}}${{hs}}${{pr}}${{art}}`;
}}
function commitPanel(c){{
  return `<div class="tli-head"><h4>${{esc(c.h)}} ${{runBadge(c.run)}}<span class="tcat" style="color:${{catcol(c.cat)}}">${{esc(c.cat)}}</span></h4></div><p class="tli-sum">${{esc(c.msg)}}</p>${{c.run?kv('run',esc(RUNLAB[c.run]||c.run)):''}}${{kv('time',esc(c.t))}}${{kv('size',`<b style="color:#3fb6a8">+${{c.ins}}</b> <b style="color:#ff5f6d">−${{c.del}}</b> · ${{c.files}} files`)}}`;
}}
// stage filter
const sf=[...document.querySelectorAll('#stepFilter button')];
sf.forEach(b=>b.onclick=()=>{{
  sf.forEach(x=>x.classList.remove('on'));b.classList.add('on');
  const s=b.dataset.step;
  document.querySelectorAll('.pvisit').forEach(v=>
    v.style.display=(s==='all'||v.dataset.step===s)?'':'none');
}});
// ---- default view state on load: Combined · Fit · Collapse gaps · Hide off-run.
// Drive the existing button handlers so every class + .on state stays consistent.
(function initDefaults(){{
  const clickIt=sel=>{{const b=document.querySelector(sel); if(b) b.click();}};
  clickIt('#pipeTabs button[data-run="combined"]');
  clickIt('#tlFit button[data-mode="fit"]');
  clickIt('#tlCollapse button[data-mode="collapse"]');
  clickIt('#tlOffrun button[data-mode="hide"]');
}})();
</script>
</body></html>"""
  return HTMLDOC

# ================= per-run discovery ====================================
def run_dates(rec_window, run_id):
    """Candidate ("YYYY","MM","DD") date tuples to search for this run's
    sessions: the run-id's own date plus the wall-clock day(s) its visits span
    (a run can cross midnight)."""
    dates = set()
    # from the run id (YYYYMMDD_...)
    m = re.match(r'^(\d{4})(\d{2})(\d{2})', run_id)
    if m:
        dates.add((m.group(1), m.group(2), m.group(3)))
    for t in rec_window:
        if t:
            dt = datetime.fromtimestamp(t)
            dates.add((f"{dt.year:04d}", f"{dt.month:02d}", f"{dt.day:02d}"))
    # also include the day after the last activity (sessions may roll past midnight)
    if rec_window and rec_window[1]:
        dt = datetime.fromtimestamp(rec_window[1] + 86400)
        dates.add((f"{dt.year:04d}", f"{dt.month:02d}", f"{dt.day:02d}"))
    return sorted(dates)

def find_orchestrator(sids, dates, exclude_paths):
    """Auto-find the top-level driver: the session transcript (in any candidate
    home) that references the MOST of this run's sub-agent session ids. Scans
    line-by-line with bounded reads and early-stops a file once it matches every
    id. Returns {engine,id,path,matched} or None. Requires >=2 distinct matches
    so a sub-agent's own transcript (which only carries its own id) never wins."""
    sids = [s for s in sids if s]
    if not sids:
        return None
    sidset = set(sids)
    # candidate files: codex sessions in the run's date dirs (both homes) +
    # claude project files whose mtime falls on/around the run window.
    cands = []
    for y, mo, d in dates:
        for home in CODEX_HOMES:
            day = os.path.join(home, "sessions", y, mo, d)
            if os.path.isdir(day):
                for fn in os.listdir(day):
                    if fn.endswith(".jsonl"):
                        cands.append(os.path.join(day, fn))
    day_epochs = set()
    for y, mo, d in dates:
        try:
            day_epochs.add(datetime(int(y), int(mo), int(d)).timestamp())
        except Exception:
            pass
    for home in CLAUDE_HOMES:
        proj = os.path.join(home, "projects")
        if not os.path.isdir(proj):
            continue
        for root, _dirs, files in os.walk(proj):
            for fn in files:
                if not fn.endswith(".jsonl"):
                    continue
                p = os.path.join(root, fn)
                mt = mtime(p)
                if any(abs(mt - de) < 2 * 86400 for de in day_epochs):
                    cands.append(p)
    excl = set(exclude_paths or [])
    best = None
    for path in cands:
        if path in excl:
            continue
        matched = set()
        try:
            with open(path, encoding="utf-8", errors="replace") as f:
                for ln in f:
                    if len(ln) > 500000:
                        continue
                    for s in sidset - matched:
                        if s in ln:
                            matched.add(s)
                    if len(matched) == len(sidset):
                        break
        except OSError:
            continue
        if len(matched) >= 2 and (best is None or len(matched) > best["matched"]):
            m = _UUID_RE.search(os.path.basename(path))
            eng = "codex" if os.sep + "sessions" + os.sep in path else "claude"
            best = {"engine": eng, "id": m.group(1) if m else "",
                    "path": path, "matched": len(matched)}
            if best["matched"] == len(sidset):
                # a perfect match is almost certainly the driver; keep scanning
                # only to see if a later file matches the same count (rare) — but
                # early accept to stay efficient.
                pass
    return best

_CD_RE = re.compile(r'--(?:cd|add-dir)\s+(/[^\s"\'`,\\]+)')

def find_working_dir(orch_path):
    """Recover the working_dir from the orchestrator transcript: the CLI commands
    it issued carry `--cd <working_dir>` / `--add-dir <working_dir>` tokens. Pick
    the most common absolute path that is NOT under ~/.easyloop/runs (those are
    run_dir --add-dir grants). Returns a path or None."""
    if not orch_path or not os.path.exists(orch_path):
        return None
    from collections import Counter
    counts = Counter()
    try:
        with open(orch_path, encoding="utf-8", errors="replace") as f:
            for ln in f:
                if "--cd" not in ln and "--add-dir" not in ln:
                    continue
                if len(ln) > 500000:
                    continue
                for p in _CD_RE.findall(ln):
                    if p.startswith(RUNS):
                        continue
                    counts[p] += 1
    except OSError:
        return None
    if not counts:
        return None
    return counts.most_common(1)[0][0]

def discover_run(run_arg, index):
    """Given a run id or full path, return a run record with everything the build
    needs. `index` is 0-based order (for tag/color/label)."""
    if os.path.isdir(run_arg):
        run_dir = os.path.abspath(run_arg)
    else:
        run_dir = os.path.join(RUNS, run_arg)
    run_id = os.path.basename(run_dir.rstrip(os.sep))
    tag = "run%d" % (index + 1)
    # start time from the dir name (YYYYMMDD_HHMMSS)
    start_dt = None
    m = re.match(r'^(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})', run_id)
    if m:
        try:
            start_dt = datetime(*[int(x) for x in m.groups()])
        except Exception:
            start_dt = None
    # scan visits (baseline applied later, once we know run order)
    run = scan_run(run_dir)
    window = run_window(run)
    dates = run_dates(window, run_id)
    io_map = parse_step_io(run_dir)
    models_map = parse_step_models(run_dir)
    workflow = parse_workflow(run_dir)
    lanes = ["driver"] + [lane for lane in GORDER
                          if lane != "driver" and lane in io_map]
    # resolve each visit's session -> engine + transcript path
    sub_paths = []
    for v in run["visits"]:
        path, engine = resolve_session(v.get("sid"), dates)
        v["engine"] = engine
        v["sess_path"] = path
        if path:
            sub_paths.append(path)
    sids = sorted({v["sid"] for v in run["visits"] if v.get("sid")})
    orch = find_orchestrator(sids, dates, sub_paths)
    # driver/orchestrator lane model+harness (best-effort; never blocks build)
    models_map["driver"] = driver_model(orch)
    wt = find_working_dir(orch["path"]) if orch else None
    if not wt:
        # fallback: a recorded cwd in any sub-agent codex session
        for v in run["visits"]:
            if v.get("engine") == "codex" and v.get("sess_path"):
                cw = _codex_session_cwd(v["sess_path"])
                if cw and not cw.startswith(RUNS):
                    wt = cw; break
    # status heuristic: live if last activity is within ~90 min of now
    last = window[1]
    status = "live" if (last and (datetime.now().timestamp() - last) < 5400) else "stopped"
    start_hhmm = start_dt.strftime("%H:%M") if start_dt else ""
    start_day = start_dt.strftime("%b %-d") if start_dt else run_id[:8]
    label_short = "Run %d" % (index + 1)
    label = "%s · %s" % (label_short, run_id)
    ncoding = len(run["steps"].get("coding", []))
    sub = "%s flow · %s · started %s · %d coding visit%s" % (
        workflow["label"], start_day, start_hhmm or "?", ncoding,
        "" if ncoding == 1 else "s")
    return {
        "arg": run_arg, "path": run_dir, "id": run_id, "tag": tag,
        "index": index, "run": run, "window": window, "dates": dates,
        "io_map": io_map, "models": models_map, "workflow": workflow,
        "lanes": lanes, "orch": orch, "wt": wt, "status": status,
        "start_dt": start_dt, "label": label, "label_short": label_short,
        "sub": sub, "sids": sids, "sub_paths": sub_paths,
    }

def _codex_session_cwd(path):
    """Read the `cwd` recorded in a Codex rollout's session_meta head line."""
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            head = f.readline()[:2000]
        m = re.search(r'"cwd"\s*:\s*"([^"]+)"', head)
        return m.group(1) if m else None
    except OSError:
        return None

# ================= run listing / CLI ===================================
def list_runs():
    """Return [(run_id, start_hhmm, visit_count)] for every run dir, newest first."""
    out = []
    if not os.path.isdir(RUNS):
        return out
    for name in sorted(os.listdir(RUNS), reverse=True):
        rd = os.path.join(RUNS, name)
        if not os.path.isdir(rd):
            continue
        nvis = 0
        adir = os.path.join(rd, "agents")
        if os.path.isdir(adir):
            for role in os.listdir(adir):
                rp = os.path.join(adir, role)
                if os.path.isdir(rp):
                    nvis += sum(1 for x in os.listdir(rp)
                                if os.path.isdir(os.path.join(rp, x)))
        m = re.match(r'^(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})', name)
        when = ""
        if m:
            try:
                when = datetime(*[int(x) for x in m.groups()]).strftime("%Y-%m-%d %H:%M")
            except Exception:
                when = ""
        out.append((name, when, nvis))
    return out

def print_run_list():
    rows = list_runs()
    print("Available runs under %s:" % RUNS)
    for run_id, when, nvis in rows:
        print("  %-20s  %-16s  %3d visits" % (run_id, when or "?", nvis))
    return rows

# ================= build ================================================
def build(run_args, timeline_row_paths=None):
    records = [discover_run(a, i) for i, a in enumerate(run_args)]
    set_run_identity(records)
    external_rows = load_external_timeline_rows(timeline_row_paths or [])

    # baseline chaining: each run inherits the prior run's final checked boxes
    # (so inherited boxes aren't double-counted as newly confirmed). Re-scan in
    # order applying the running baseline.
    baseline = set()
    for rec in records:
        rec["run"] = scan_run(rec["path"], baseline=baseline)
        # re-resolve sessions onto the fresh visit objects
        for v in rec["run"]["visits"]:
            path, engine = resolve_session(v.get("sid"), rec["dates"])
            v["engine"] = engine; v["sess_path"] = path
            v["run"] = rec["tag"]
            mh = rec["models"].get(v["step"], {})
            v["model"] = mh.get("model", ""); v["harness"] = mh.get("harness", "")
        rec["window"] = run_window(rec["run"])
        baseline = rec["run"].get("final_checked") or baseline

    assign = make_assign_run(records)
    assign_all = make_assign_run_all(records, assign)

    # since-date for per-run day commits = earliest run's midnight
    starts = [r["window"][0] for r in records if r["window"][0]]
    since_dt = datetime.fromtimestamp(min(starts)) if starts else datetime.now()
    since = since_dt.strftime("%Y-%m-%d 00:00")

    # distinct worktrees among runs (for shared-worktree reproduction + support
    # of runs on different worktrees)
    wts = []
    for rec in records:
        if rec["wt"] and rec["wt"] not in wts:
            wts.append(rec["wt"])

    # day commits (union over distinct worktrees, deduped by hash), tagged/sized
    day_commits = {}
    for wt in wts:
        sizes = git_numstat(wt, since)
        for c in git_log(wt, since):
            if c["h"] in day_commits:
                continue
            c["cat"] = categorize(c["msg"]); c["t"] = fmt_time(c["iso"])
            ins, dele, files = sizes.get(c["h"], (0, 0, 0))
            c["ins"] = ins; c["del"] = dele; c["files"] = files
            c["run"] = assign(c["ct"])
            day_commits[c["h"]] = c
    day_commits = list(day_commits.values())

    # per-run commit lists + attribution
    for rec in records:
        ci = [c for c in day_commits if c["run"] == rec["tag"]]
        rec["commits"] = ci
        attribute_commits(rec["run"], ci)

    # markers: sub-agent compactions per run + driver events from each run's
    # discovered orchestrator (dedup driver paths shared across runs).
    markers_all = []
    seen_orch = set()
    for rec in records:
        markers_all += gather_subagent_compactions(rec)
    # how many runs share each orchestrator path (runs continued in the same
    # top-level session share one; a fresh pilot is unique to its run)
    orch_paths = [rec["orch"]["path"] for rec in records
                  if rec.get("orch") and rec["orch"].get("path")]
    for rec in records:
        orch = rec["orch"]
        if orch and orch["engine"] == "codex" and orch["path"] not in seen_orch:
            seen_orch.add(orch["path"])
            evs = gather_driver_events(orch["path"], assign_all)
            # A driver session unique to ONE run owns all its steering events,
            # even when the kickoff is stamped a few seconds before that run's
            # first visit mtime (otherwise the prior run's 30-min trailing slack
            # steals it and the new run shows no driver/steering at all).
            if orch_paths.count(orch["path"]) == 1:
                for e in evs:
                    e["run"] = rec["tag"]
            markers_all += evs
    for _m in markers_all:
        _m.setdefault("run", records[0]["tag"])

    # combined visits (all runs, one wall-clock axis)
    combined_visits = sorted(
        [v for rec in records for v in rec["run"]["visits"]],
        key=lambda x: (x.get("start") or x.get("mtime") or 0))

    run_tags = {rec["id"]: rec["tag"] for rec in records}
    combined_external, combined_external_lanes = external_visits(external_rows, run_tags)
    combined_timeline_visits = sorted(
        combined_visits + combined_external,
        key=lambda x: (x.get("start") or x.get("mtime") or 0))

    # FULL branch history (union over distinct worktrees), tagged with pre/runK
    all_commits = {}
    for wt in wts:
        sizes = git_numstat_full(wt)
        for c in git_log_full(wt):
            if c["h"] in all_commits:
                continue
            c["cat"] = categorize(c["msg"]); c["t"] = fmt_time(c["iso"])
            ins, dele, files = sizes.get(c["h"], (0, 0, 0))
            c["ins"] = ins; c["del"] = dele; c["files"] = files
            c["run"] = assign_all(c["ct"])
            all_commits[c["h"]] = c
    combined_commits = sorted(all_commits.values(), key=lambda x: x["ct"])

    # per-run off-run re-tag (Task A) + artifact resolution (Task B)
    for rec in records:
        pre = retag_offrun(rec["run"], rec["commits"])
        rec["pre_offrun"] = pre
        resolve_artifacts(rec["run"]["visits"], rec["io_map"])
        resolve_prompts(rec["run"]["visits"])
        resolve_idle(rec["run"]["visits"])

    # markers split per run tag + combined
    markers_by = {rec["tag"]: sorted([m for m in markers_all if m["run"] == rec["tag"]],
                                     key=lambda x: x["ct"]) for rec in records}
    markers_combined = sorted(markers_all, key=lambda x: x["ct"])

    # metas + cards + pipelines + timelines
    metas = [build_meta(rec) for rec in records]
    VMETA, CMETA, MMETA = {}, {}, {}
    timelines = []
    pipes = []
    cards = []
    for rec, meta in zip(records, metas):
        active = (rec["index"] == 0)
        pipes.append(pipeline(rec["run"], meta, active))
        ext_visits, ext_lanes = external_visits(external_rows, {rec["id"]: rec["tag"]})
        timeline_visits = sorted(
            rec["run"]["visits"] + ext_visits,
            key=lambda x: (x.get("start") or x.get("mtime") or 0))
        tl, vm, cm, mm = build_timeline(
            timeline_visits, meta["commits"], meta["id"], active,
            markers=markers_by[rec["tag"]], extra_lanes=ext_lanes,
            base_lanes=rec["lanes"])
        timelines.append(tl)
        VMETA[meta["id"]] = vm; CMETA[meta["id"]] = cm; MMETA[meta["id"]] = mm
        cards.append(run_card(meta))
    # combined tab (always present; for 1 run it is that run on the full axis)
    pipes.append(pipeline({"visits": combined_visits}, {"id": "combined"},
                          False, show_run=True))
    combined_lanes = [lane for lane in GORDER
                      if any(lane in rec["lanes"] for rec in records)]
    tlC, vmC, cmC, mmC = build_timeline(combined_timeline_visits, combined_commits,
                                        "combined", False, combined=True,
                                        markers=markers_combined,
                                        extra_lanes=combined_external_lanes,
                                        base_lanes=combined_lanes)
    timelines.append(tlC)
    VMETA["combined"] = vmC; CMETA["combined"] = cmC; MMETA["combined"] = mmC

    # ---- per-lane model+harness meta (Feature 1: row-heading hover / inspector).
    # Per-run: one row per lane (that run's own setup/SKILL.md). Combined: the
    # DISTINCT (model,harness) values across runs for each lane, tagged by run.
    LANEMETA = {}
    for rec in records:
        lm = {}
        for lane in GORDER:
            mh = rec["models"].get(lane)
            if not mh:
                continue
            lm[lane] = {"role": STEPLABEL.get(lane, lane),
                        "rows": [{"run": None, "model": mh.get("model", ""),
                                  "harness": mh.get("harness", "")}]}
        LANEMETA[rec["id"]] = lm
    lmc = {}
    for lane in GORDER:
        rows, seen = [], set()
        for rec in records:
            mh = rec["models"].get(lane)
            if not mh:
                continue
            key = (mh.get("model", ""), mh.get("harness", ""))
            if key in seen:
                continue
            seen.add(key)
            rows.append({"run": rec["label_short"], "model": mh.get("model", ""),
                         "harness": mh.get("harness", "")})
        if rows:
            lmc[lane] = {"role": STEPLABEL.get(lane, lane), "rows": rows}
    LANEMETA["combined"] = lmc

    # External row headings explain their provenance rather than pretending to
    # be agent/model configuration. Only rows occupied in a view are present.
    def add_external_lane_meta(target, row, run_id=None):
        events = row["events"] if run_id is None else [
            e for e in row["events"] if e["run_id"] == run_id]
        if not events:
            return
        target[row["lane"]] = {
            "role": row["label"], "external": True,
            "description": row.get("description") or "Imported timeline intervals",
            "rows": [{"run": None, "source": p} for p in row.get("files", [])],
        }
    for rec in records:
        for row in external_rows:
            add_external_lane_meta(LANEMETA[rec["id"]], row, rec["id"])
    for row in external_rows:
        add_external_lane_meta(LANEMETA["combined"], row)

    # ---- HTML fragments ----
    def tab_label(rec):
        return "%s · %s (%s)" % (rec["label_short"], rec["workflow"]["label"],
                            rec["start_dt"].strftime("%H:%M") if rec["start_dt"] else rec["id"])
    tabs = []
    for i, rec in enumerate(records):
        oncls = ' class="on"' if i == 0 else ''
        tabs.append('<button%s data-run="%s">%s</button>'
                    % (oncls, esc(rec["id"]), esc(tab_label(rec))))
    tabs.append('<button data-run="combined">Combined (%d run%s)</button>'
                % (len(records), "" if len(records) == 1 else "s"))
    tabs_html = "\n      ".join(tabs)

    comb = []
    for rec in records:
        c = RUNCOL[rec["tag"]]
        comb.append(f'<span><b style="background:{c}"></b>{esc(rec["label_short"])} '
                    f'visit / commit outline</span>')
    comb.append('<span><b style="background:#7f8da3"></b>pre-run commit</span>')
    comblegend_html = "\n    ".join(comb)

    timelines_html = "\n  ".join(timelines)
    pipes_html = "\n  ".join(pipes)
    cards_html = "\n    ".join(cards)

    # stats tiles: total runs, total day-commits, then per-run boxes-done (cap at
    # a sensible number of tiles so the 6-col grid stays tidy)
    tot_commits = sum(len(rec["commits"]) for rec in records)
    stat_tiles = [f'<div class="tile"><b>{len(records)}</b><span>runs</span></div>',
                  f'<div class="tile"><b>{tot_commits}</b><span>commits</span></div>']
    for rec, meta in list(zip(records, metas))[:4]:
        stat_tiles.append(
            f'<div class="tile"><b>{meta["done"]}<span style="font-size:15px">'
            f'/{meta["total"] or "?"}</span></b><span>{esc(rec["label_short"].lower())} '
            f'boxes done</span></div>')
    for rec, meta in list(zip(records, metas))[:4]:
        stat_tiles.append(
            f'<div class="tile" title="{esc(meta["span"])}"><b style="font-size:20px">'
            f'{esc(meta["dur"])}</b><span>{esc(rec["label_short"].lower())} '
            f'wall-clock</span></div>')
    stats_html = "\n    ".join(stat_tiles)

    # Per-run "what these runs show" commentary removed — this is a generic tool.
    insight_html = ""

    if len(records) == 1:
        subtitle_html = f'One run, <span>{esc(records[0]["id"])}.</span>'
    else:
        subtitle_html = f'{len(records)} runs, <span>one view.</span>'
    footer_ids = " · ".join(rec["id"] for rec in records)
    title = "EasyLoop Insights / " + ", ".join(rec["id"] for rec in records)
    gen_time = datetime.now().strftime("%Y-%m-%d %H:%M")

    doc = render_doc(records, metas, tabs_html, timelines_html, comblegend_html,
                     pipes_html, cards_html, stats_html, insight_html, title,
                     subtitle_html, footer_ids, VMETA, CMETA, MMETA, LANEMETA, gen_time)

    with open(OUT, "w", encoding="utf-8") as f:
        f.write(doc)

    # ---- diagnostics ----
    print("wrote", OUT, f"({len(doc)} bytes)")
    _sub = [m for m in markers_all if m.get("src") == "sub-agent"]
    print("sub-agent compactions — codex: %d claude: %d" % (
        sum(1 for m in _sub if m.get("engine") == "codex"),
        sum(1 for m in _sub if m.get("engine") == "claude")))
    for rec, meta in zip(records, metas):
        orch = rec["orch"]
        orch_s = ("%s %s (matched %d)" % (orch["engine"], orch["id"][:13], orch["matched"])
                  if orch else "none found")
        nres = sum(1 for v in rec["run"]["visits"] if v.get("sess_path"))
        print("%s %s: %d visits (%d sessions resolved) · commits %d (%d off-run) "
              "· done %d/%d · orchestrator %s · working_dir %s" % (
                  rec["tag"], rec["id"], len(rec["run"]["visits"]), nres,
                  len(rec["commits"]), rec["pre_offrun"], meta["done"],
                  meta["total"], orch_s, rec["wt"] or "unavailable"))
        for lane in GORDER:
            mh = rec["models"].get(lane)
            if mh:
                print("    %-14s model %-16s harness %s" % (
                    lane, mh.get("model") or "(unknown)", mh.get("harness") or "(unknown)"))
    print("artifact blobs:", len(BLOBS))
    if external_rows:
        print("external timeline rows:", len(external_rows), "from",
              len(timeline_row_paths or []), "file(s)")
    _prec = sum(1 for c in combined_commits if c.get("run") == "pre")
    print("combined visits:", len(combined_visits), "commits", len(combined_commits),
          f"(pre-run {_prec})")

def resolve_run_args(argv):
    """Turn CLI args into a list of run ids/paths. With no args: default to the
    most-recent 2 runs (after printing the list)."""
    if argv:
        return argv
    rows = print_run_list()
    default = [r[0] for r in rows[:2]]   # rows are newest-first
    default = list(reversed(default))    # build oldest-first for stable tags
    print("\nNo run specified — defaulting to the most recent %d run(s): %s"
          % (len(default), ", ".join(default)))
    return default

def parse_cli(argv):
    """Split positional run arguments from repeatable --timeline-rows files."""
    runs, row_paths = [], []
    i = 0
    while i < len(argv):
        arg = argv[i]
        if arg == "--timeline-rows":
            if i + 1 >= len(argv):
                raise ValueError("--timeline-rows requires a JSON file path")
            row_paths.append(argv[i + 1]); i += 2; continue
        if arg.startswith("--timeline-rows="):
            path = arg.split("=", 1)[1]
            if not path:
                raise ValueError("--timeline-rows requires a JSON file path")
            row_paths.append(path); i += 1; continue
        if arg.startswith("-"):
            raise ValueError(f"unknown option: {arg}")
        runs.append(arg); i += 1
    return runs, row_paths

def main(argv):
    if argv and argv[0] in ("--list", "-l", "list"):
        print_run_list()
        return 0
    if argv and argv[0] in ("--help", "-h"):
        print(__doc__)
        return 0
    try:
        positional, timeline_row_paths = parse_cli(argv)
        run_args = resolve_run_args(positional)
    except ValueError as exc:
        print("error:", exc, file=sys.stderr)
        return 2
    if not run_args:
        print("No runs found under", RUNS)
        return 1
    try:
        build(run_args, timeline_row_paths=timeline_row_paths)
    except ValueError as exc:
        print("error:", exc, file=sys.stderr)
        return 2
    return 0

if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
