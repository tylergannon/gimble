1. `chapters.md` names the frame fixture at the wrong path:

   - `chapters.md`: “from `prove/fixtures/frame-preamble.txt`”
   - `chapters/04-library/SPRINT-02.md`: “frame preamble (`fixtures/frame-preamble.txt`)”

   These resolve to different locations; only the latter exists. Change `chapters.md`. Owning node: `decompose`.

2. P7’s verifier contradicts the universal proof rule:

   - `decisions.md` decision 41: “Every promise needs a judgment over recorded evidence by a model that is not the coder's”
   - `declaration.md` P7 verifier: “Timeline `steer` verdict with a delivered disposition; diff of the stage output before and after.”

   The declared verifier is only timeline inspection and a diff, with no independent model judgment. Change `declaration.md`. Owning node: `brief`.

ROUTE: fail


