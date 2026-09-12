<script lang="ts">
	import { guide } from './guide.remote';
</script>

<svelte:head>
	<title>Gimble — agent workflows in Go</title>
	<meta
		name="description"
		content="Gimble is a Go library for running agent workflows as ordinary Go code."
	/>
</svelte:head>

<article>
	<svelte:boundary>
		{@const name = await guide()}
		<p class="eyebrow">{name}</p>
	</svelte:boundary>
	<h1 data-testid="title">Agent workflows in Go</h1>
	<p class="lead">
		Gimble is a Go library for running agent work from ordinary Go code. You write the
		workflow. Gimble runs the agent sessions, keeps their work inside named scopes, and
		records what happened.
	</p>

	<section aria-labelledby="what-it-does">
		<h2 id="what-it-does">What it does</h2>
		<ul>
			<li>Starts a named run and writes a durable JSONL record for it.</li>
			<li>Creates agent sessions through a harness adapter such as Codex or Claude Code.</li>
			<li>Runs turns, lets another goroutine steer or interrupt a running turn, and records the result.</li>
			<li>Gives normal Go concurrency a visible shape with named scopes and groups.</li>
		</ul>
	</section>

	<section aria-labelledby="what-you-write">
		<h2 id="what-you-write">What you write</h2>
		<p>
			The workflow is regular Go: functions, loops, conditionals, errors, and the tools you
			already use. Gimble does not hide a critique round, retry policy, or delivery process
			behind a special workflow function. Put that logic directly in the program that needs it.
		</p>
		<pre><code>ctx = gimble.Project(ctx, ".")

err := gimble.Run(ctx, "repair", func(ctx context.Context) error &#123;
	worker := gimble.NewSession(ctx, "worker", adapter, model, workdir)
	answer, err := worker.Generate[gimble.Text](ctx, "Fix the failing test.")
	if err != nil &#123;
		return err
	&#125;
	fmt.Println(answer)
	return nil
&#125;)</code></pre>
		<p>
		<code>adapter</code> is the harness-specific implementation. The workflow stays the same
			whether it talks to Codex, Claude Code, or an adapter you provide.
		</p>
	</section>

	<section aria-labelledby="start">
		<h2 id="start">Start here</h2>
		<ol>
			<li>Install the library in a Go project.</li>
			<li>Create a project context and a <code>Run</code>.</li>
			<li>Create a named session, then call <code>Generate</code> with the prompt you want.</li>
			<li>Use <code>Scope</code> and <code>Group</code> when the work has a bounded step or concurrent branches.</li>
		</ol>
		<pre><code>go get github.com/tylergannon/gimble
go doc -all github.com/tylergannon/gimble</code></pre>
	</section>

	<section aria-labelledby="limits">
		<h2 id="limits">What it does not do</h2>
		<p>
			Gimble does not choose your process for you. It is not a hosted agent product, a new
			programming language, or a library of named management tactics. It is a small runtime for
			the parts that are hard to get right repeatedly: agent sessions, cancellation, bounded
			concurrency, and a record of the run.
		</p>
	</section>
</article>

<style>
	article { max-width: 48rem; }
	.eyebrow { margin: 0; font-weight: 700; letter-spacing: 0.04em; text-transform: uppercase; color: #2b5fd9; }
	h1 { margin: 0.25rem 0 0; font-size: clamp(2rem, 6vw, 3rem); line-height: 1.1; }
	.lead { font-size: 1.2rem; }
	section { margin-top: 2.5rem; }
	pre { overflow-x: auto; padding: 1rem; border-radius: 0.4rem; background: #171b22; color: #f5f7fa; }
</style>
