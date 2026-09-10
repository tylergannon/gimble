# Direction corrections — verbatim

Captured from Tyler's corrections to the first synthesis on 2026-09-09.

## Pseudocode as the API-design ideal

This misses the point of the "pseudocode" claim.  The point of the pseudocode is that if the workflow definition can be read as a page or two that resembles pseudocode, that's an ideal for readability that should guide the API development.

## Telemetry should prioritize bad runs

mmmmmmmm I guess so probably.  But I think it's more important to focus on being able to study why it wandered, failed, cost too much, or needed unwanted steering.

## Recovery is desired, but no compass mechanism exists

I don't know that "the compass idea" really has any traction.  Unless I forgot parts of it?  My point is that "I would like to be able to get a wayward workflow to be able to steer towards the goal when it's lost the point.  I'm not entirely sure that we have any ideas yet, other than just good steering from supervisors.

## The substantive reasons for Go

The main arguments are pretty simple really.  Like my postulate about workflow definition languages over time needing to become a programming language.  Like Go being basically defined for async shit.  Like the fact that it just turns out to be really really hard to write workflows in the graph language.  Not easy, like intended.

## Keep the README current until the refactor exists

not a problem.  we haven't realized this yet.  we can rewrite the readme when the refactor happens.
