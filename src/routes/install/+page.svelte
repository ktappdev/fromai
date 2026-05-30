<script lang="ts">
	let copied = $state<string | null>(null);

	async function copyCommand(cmd: string, id: string) {
		try {
			await navigator.clipboard.writeText(cmd);
			copied = id;
			setTimeout(() => (copied = null), 2000);
		} catch (e) {
			console.error('Failed to copy', e);
		}
	}
</script>

<div class="install">
	<h1>Install fai CLI</h1>
	<p class="intro">Install the fromai CLI to manage tasks from your terminal.</p>

	<section class="method">
		<h2>Homebrew</h2>
		<p class="desc">Install via Homebrew tap (recommended for macOS/Linux).</p>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>brew install ktappdev/tap/fai</code>
			<button onclick={() => copyCommand('brew install ktappdev/tap/fai', 'brew-install')} class="copy-btn">
				{copied === 'brew-install' ? '✓' : '📋'}
			</button>
		</div>
	</section>

	<section class="method">
		<h2>Direct Download</h2>
		<p class="desc">Download a pre-built binary for your platform.</p>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>curl -L -o fai &lt;RELEASE_URL&gt;</code>
			<button onclick={() => copyCommand('curl -L -o fai <RELEASE_URL>', 'download')} class="copy-btn">
				{copied === 'download' ? '✓' : '📋'}
			</button>
		</div>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>chmod +x fai</code>
			<button onclick={() => copyCommand('chmod +x fai', 'chmod')} class="copy-btn">
				{copied === 'chmod' ? '✓' : '📋'}
			</button>
		</div>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>sudo mv fai /usr/local/bin/</code>
			<button onclick={() => copyCommand('sudo mv fai /usr/local/bin/', 'move')} class="copy-btn">
				{copied === 'move' ? '✓' : '📋'}
			</button>
		</div>
	</section>

	<section class="method">
		<h2>Go Install</h2>
		<p class="desc">Install directly from source using Go toolchain.</p>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>go install github.com/kentaylor/fromai/cli/cmd/fai@latest</code>
			<button onclick={() => copyCommand('go install github.com/kentaylor/fromai/cli/cmd/fai@latest', 'go-install')} class="copy-btn">
				{copied === 'go-install' ? '✓' : '📋'}
			</button>
		</div>
	</section>

	<section class="method">
		<h2>Build from Source</h2>
		<p class="desc">Clone and build manually.</p>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>git clone https://github.com/kentaylor/fromai.git</code>
			<button onclick={() => copyCommand('git clone https://github.com/kentaylor/fromai.git', 'clone')} class="copy-btn">
				{copied === 'clone' ? '✓' : '📋'}
			</button>
		</div>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>cd fromai/cli</code>
			<button onclick={() => copyCommand('cd fromai/cli', 'cd-cli')} class="copy-btn">
				{copied === 'cd-cli' ? '✓' : '📋'}
			</button>
		</div>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>go build -o fai ./cmd/fai</code>
			<button onclick={() => copyCommand('go build -o fai ./cmd/fai', 'build')} class="copy-btn">
				{copied === 'build' ? '✓' : '📋'}
			</button>
		</div>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>sudo mv fai /usr/local/bin/</code>
			<button onclick={() => copyCommand('sudo mv fai /usr/local/bin/', 'move-build')} class="copy-btn">
				{copied === 'move-build' ? '✓' : '📋'}
			</button>
		</div>
	</section>

	<section class="method">
		<h2>Verify Installation</h2>
		<p class="desc">Confirm fai is installed and working.</p>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>fai --help</code>
			<button onclick={() => copyCommand('fai --help', 'help')} class="copy-btn">
				{copied === 'help' ? '✓' : '📋'}
			</button>
		</div>
	</section>

	<section class="method">
		<h2>Configure</h2>
		<p class="desc">Set up authentication with your API key.</p>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>fai init --key &lt;your-api-key&gt; --base-url &lt;backend-url&gt;</code>
			<button onclick={() => copyCommand('fai init --key <your-api-key> --base-url <backend-url>', 'init')} class="copy-btn">
				{copied === 'init' ? '✓' : '📋'}
			</button>
		</div>
		<p class="note">Get your API key from the <a href="/settings">Settings</a> page.</p>
		<div class="code-block">
			<span class="prompt">$</span>
			<code>fai whoami</code>
			<button onclick={() => copyCommand('fai whoami', 'whoami')} class="copy-btn">
				{copied === 'whoami' ? '✓' : '📋'}
			</button>
		</div>
		<p class="note">Verify authentication is working.</p>
	</section>
</div>

<style>
	.install {
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

	.intro {
		color: #8b949e;
		font-size: 0.75rem;
		margin: 0 0 28px;
	}

	.method {
		margin-bottom: 32px;
	}

	.method:last-child {
		margin-bottom: 0;
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

	.desc {
		color: #8b949e;
		font-size: 0.7rem;
		margin: 0 0 12px;
	}

	.code-block {
		display: flex;
		align-items: center;
		gap: 10px;
		background: #0d1117;
		border: 1px solid #1a1a1a;
		padding: 8px 12px;
		margin-bottom: 6px;
	}

	.code-block:last-child {
		margin-bottom: 0;
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

	.copy-btn {
		background: transparent;
		border: none;
		color: #8b949e;
		padding: 2px 6px;
		font-size: 0.8rem;
		cursor: pointer;
		flex-shrink: 0;
		transition: color 0.12s;
	}

	.copy-btn:hover {
		color: #c9d1d9;
	}

	.note {
		color: #6e7681;
		font-size: 0.7rem;
		margin: 10px 0 0 0;
	}

	.note a {
		color: #3fb950;
		text-decoration: none;
	}

	.note a:hover {
		text-decoration: underline;
	}
</style>