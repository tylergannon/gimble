Here's the key claim that we're making, here.  PLEase write down everything here, verbatim, as well as your own cleaned up version for the documentaiton. &#x20;



The DAG-style workflow definition is fine as long as workflows are very simple, but the attempt to model workflow in general is hard.


- The art of agentic workflow is rapidly changing and with it the tactics and specific primitives need to be moving at the same rate.&#x20;
- **Postulate**: A workflow shorthand will evolve (perhaps asymptotically?) towards becoming a pseudo-programming language.
- The reaction we're choosing is to abandon the effort to *model* workflow, except *as a program*, in a programming language.
- At some point the context will expand large enough that it shouldn't be entirely sent to the agent with every request.  At that time it should spill over onto the disk and just be a place where the agent can search for more information.
- The previous finding / observation is taken in light with the already prevalent notion of a semantic index.  We've already discussed semantic index quite a bit and we should even have skills or workflows around building them.  I think we actually should basically consider the program context to directly be a mutating / evolving addition to the first / root node of the semantic index.

I think letting ourselves think of workflows as just Go programs alleviates us of many trappings of
the DAG-style workflow definition and also gives us a really helpful metaphor for what to build, so
that programming it will *feel very natural*.

In other words, if we lean further towards the metaphor of a *program*, the "context" that we're discussing right now is actually like the stack memory, or the "in-context" information.  The frame data should basically have the same priority and possibly the same way of defining and possibly the same way of being presented along with any given agent request.  What I mean is, we always want to make sure that the agent's context includes the current context information as well as the highest level index in whatever semantic index this agent is expected to draw from, along with instructions about when and how to use it.  It will truly and merely be index files and other files, all of which are plain text and/or images etc, stuff that's legible and searchable by the agents.  That make sense?  So like the context is the runtime state, if you will, or the part that we want agents to be reasoning about, but we also can give it a place to hunt for other info it expects.  *JUST LIKE A REGULAR ASS PROGRAM*
