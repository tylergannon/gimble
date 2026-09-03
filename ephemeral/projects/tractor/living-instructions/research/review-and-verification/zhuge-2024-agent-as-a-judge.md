# Zhuge et al. — Agent-as-a-Judge: Evaluate Agents with Agents

## Purpose
An agentic judge with tools (graph, locate, read, retrieve, ask) that inspects an agent's produced workspace against hierarchical requirements; the nearest published relative of an "independent verifier with tools".

## Pinned
- arXiv:2410.10934v2 (2024-10-16). Code: https://github.com/metauto-ai/agent-as-a-judge. License: as posted.
- https://arxiv.org/abs/2410.10934

## Key concepts
- DevAI: 55 AI-dev tasks, 365 hierarchical requirements with dependencies (Sec. 3).
- Eight modules proposed; ablation kept Graph, Locate, Read, Retrieve, Ask and dropped Search, Memory, Planning as unstable — prior judgments cascaded errors (Sec. 4.1, Fig. 6).
- Alignment with human majority, OpenHands, black-box: Agent-as-a-Judge 90.44% vs LLM-as-a-Judge 60.38%; gray-box 90.16% vs 70.76% (Table 3). Gray-box = access to manually collected trajectories (Table 3 note).
- Cost: 118 min / $30.58 vs 86.5 h / $1,297.50 for humans (Sec. 4.4).
- Individual human evaluators disagree 10-30% pairwise; single-human error up to 23.77%, majority vote 6.01% (Sec. 3.2, Figs. 4-5).

## Bounded comparison
Unlike ours because the judge reads and locates artifacts but does not run the software (Sec. 4.1: the Read module parses 33 file formats); it is a reading verifier with tools, not an operating one.

## Gotchas
- The 90% number is alignment with a *human majority* on requirement satisfaction, not with ground-truth program behavior; the paper never executes the deliverables.
- Judge planning/memory hurt: chained sub-judgments propagate early mistakes (Sec. 4.1). A verifier that carries state across chapters inherits this risk.
- Single judge model; no cross-provider setup; no adversarial author. Nothing here tests collusion or gaming.
- Requirements were hand-annotated per task; the judge's precision depends on that decomposition existing.

## Recipe
- To structure a verifier's tool set, start at Sec. 4.1 (module ablation) and keep the five that survived.
- To calibrate what "verifier agrees with humans" means, start at Sec. 3.2 / Fig. 5 (human disagreement).
