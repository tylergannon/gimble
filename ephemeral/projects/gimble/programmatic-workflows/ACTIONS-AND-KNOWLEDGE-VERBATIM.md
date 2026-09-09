I think it would be totally okay to say that whenever we author a workflow, we'd use static analysis to pull out the graph nodes, and then basically use an agent session to put together the edges and give it some spatiality.  Or maybe we can use static analysis to put in edges wherever they can be strongly inferred, and then ask an agent to fill in additional conceptual edges?  The point is not to have a graph, the point is to help organic-brained humies to understand it. &#x20;



I think since we're moving to just using ordinary Go for the actual orchestration, the number of "shapes" maybe goes down.



- Command
- Agent Task
- Supervisor/Coach agent

Honestly am I missing anything?

So we would need to possibly come up with a different set of tool shapes.  I like the idea of keeping to rules of thumb about the numbers of how many tools there are.  I think some people like working with FIVE and SEVEN as possible magic numbers of items.  if there are seven different primitive action types, or maybe five, I think it'll help to keep cognitive load down when authoring them.



Like it makes sense possibly to differentiate a CODING task from a VALIDATION task from a INDEXING task.



That make sense?  Maybe you guys should review the whole semantic index topic briefly.

Currently it's sort of thought of as an afterthought or a separate concern, that there should be a semantic index where the agent can find information.  But I think that should be a very important consideration on any task that's difficult enough to warrant a workflow, you better fucking make sure that you have the broader context local and searchable so that coding agents aren't inventing anything nor looking for things that aren't already present.  So IDK maybe RESEARCH and INDEXING are companions in the same way that CODING and VALIDATION are companions.
