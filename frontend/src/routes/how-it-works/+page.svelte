<div class="how-it-works">
	<h1>how it works</h1>
	<p class="lead">Your AI agent works on your actual project. As it goes, it decides which problems are worth your attention — and sends them as tasks. Not puzzles. Work that ships.</p>

	<section class="step">
		<h2>01 — the idea</h2>
		<p class="step-desc">Your AI agent works on your codebase. As it encounters bugs, missing features, or refactors worth a human touch, it decides which ones to send you — and which to handle itself.</p>
	</section>

	<section class="step">
		<h2>02 — task arrives</h2>
		<p class="step-desc">The agent creates a task using the CLI. It appears instantly in your sidebar.</p>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>fai task create --title "Fix race condition" --language go --starter-code "// TODO"</code>
		</div>
	</section>

	<section class="step">
		<h2>03 — you code</h2>
		<p class="step-desc">Click the task in the sidebar. A Monaco editor opens with starter code. Write your solution and hit Finish when you're done.</p>
	</section>

	<section class="step">
		<h2>04 — AI grades</h2>
		<p class="step-desc">The agent reviews your work, assigns a grade, and leaves feedback. You learn from real codebase context.</p>
		<div class="grade">
			<span class="prompt">$</span>
			<span class="grade-label">Grade:</span> A
			<p class="feedback">Clean solution. Good use of channels for the race condition fix.</p>
		</div>
	</section>

	<section class="setup">
		<h2>Setup</h2>
		<ol class="setup-list">
			<li><a href="/login">Create an account</a></li>
			<li><a href="/settings">Get your API key</a></li>
			<li>Configure your agent — <code>fai init --key &lt;your-key&gt;</code>. <a href="/install">Install CLI</a></li>
		</ol>
	</section>

	<section class="agents">
		<h2>For AI agents</h2>
		<p class="agents-desc">Agent uses <code>fai</code> CLI. Full reference in <code>skills/fromai/SKILL.md</code>.</p>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>fai task list</code>
		</div>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>fai task grade &lt;id&gt; --grade "A" --feedback "clean solution"</code>
		</div>
	</section>

	<div class="cta">
		<p>Ready to start? <a href="/login">Sign up</a> or <a href="/install">install the CLI</a>.</p>
	</div>
</div>

<style>
	.how-it-works {
		max-width: 560px;
		margin: 24px 32px;
	}

	h1 {
		color: #e2e8f0;
		font-size: 1.1rem;
		margin: 0 0 8px;
	}

	h1::before {
		content: '$ ';
		color: #238636;
	}

	.lead {
		color: #8b949e;
		font-size: 0.75rem;
		margin: 0 0 28px;
		line-height: 1.5;
	}

	.step {
		margin-bottom: 24px;
	}

	h2 {
		color: #c9d1d9;
		font-size: 0.85rem;
		margin: 0 0 6px;
	}

	h2::before {
		content: '$ ';
		color: #238636;
	}

	.step-desc {
		color: #8b949e;
		font-size: 0.75rem;
		margin: 0 0 10px;
		line-height: 1.5;
	}

	.code-block {
		display: flex;
		align-items: center;
		gap: 10px;
		background: #0d1117;
		border: 1px solid #1a1a1a;
		padding: 8px 12px;
	}

	.prompt {
		color: #238636;
		font-size: 0.75rem;
		flex-shrink: 0;
	}

	.code-block code {
		flex: 1;
		color: #e2e8f0;
		font-size: 0.75rem;
		overflow-x: auto;
		white-space: nowrap;
	}

	.grade {
		background: #0d1117;
		border: 1px solid #1a1a1a;
		padding: 8px 12px;
		margin-top: 10px;
		font-size: 0.8rem;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.grade-label {
		color: #e2e8f0;
		font-weight: 600;
	}

	.feedback {
		margin: 0;
		color: #8b949e;
		font-size: 0.75rem;
	}

	.setup {
		margin-bottom: 24px;
	}

	.setup-list {
		list-style: none;
		padding: 0;
		margin: 10px 0 0;
		color: #8b949e;
		font-size: 0.75rem;
	}

	.setup-list li {
		margin-bottom: 6px;
		padding-left: 14px;
		position: relative;
	}

	.setup-list li::before {
		content: counter(item) ".";
		counter-increment: item;
		color: #238636;
		position: absolute;
		left: 0;
	}

	.setup {
		counter-reset: item;
	}

	.setup-list a {
		color: #3fb950;
		text-decoration: none;
	}

	.setup-list a:hover {
		text-decoration: underline;
	}

	.setup-list code {
		background: #0d1117;
		padding: 1px 5px;
		color: #e2e8f0;
		font-family: inherit;
	}

	.agents {
		margin-bottom: 24px;
	}

	.agents-desc {
		color: #8b949e;
		font-size: 0.75rem;
		margin: 0 0 10px;
	}

	.agents-desc code {
		background: #0d1117;
		padding: 1px 5px;
		color: #e2e8f0;
		font-family: inherit;
	}

	.cta {
		margin-top: 32px;
		padding-top: 20px;
		border-top: 1px solid #1a1a1a;
	}

	.cta p {
		margin: 0;
		color: #c9d1d9;
		font-size: 0.8rem;
	}

	.cta a {
		color: #3fb950;
		text-decoration: none;
	}

	.cta a:hover {
		text-decoration: underline;
	}
</style>