<script lang="ts">
	import { pb } from '$lib/pocketbase.js';
	import { onMount } from 'svelte';

	let apiKey = $state('');
	let loading = $state(true);
	let copied = $state(false);
	let regenerating = $state(false);

	async function loadKey() {
		try {
			const key = await pb.getAPIKey();
			if (key) {
				apiKey = key;
			}
		} catch (e) {
			console.error('Failed to load API key', e);
		} finally {
			loading = false;
		}
	}

	async function regenerate() {
		regenerating = true;
		try {
			const key = await pb.regenerateAPIKey();
			if (key) {
				apiKey = key;
			}
		} catch (e) {
			console.error('Failed to regenerate API key', e);
		} finally {
			regenerating = false;
		}
	}

	async function copyKey() {
		await navigator.clipboard.writeText(apiKey);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	onMount(() => {
		loadKey();
	});
</script>

<div class="settings">
	<h2>Settings</h2>

	<div class="section">
		<h3>CLI API Key</h3>
		<p class="desc">
			Use this key with the Coding Gym CLI to authenticate from the terminal.
			Run <code>cg init --key &lt;key&gt;</code> to set it up.
		</p>

		{#if loading}
			<p class="muted">Loading...</p>
		{:else}
			<div class="key-row">
				<code class="key-display">{apiKey ? apiKey.slice(0, 16) + '...' + apiKey.slice(-8) : 'No key generated'}</code>
				<div class="key-actions">
					<button onclick={copyKey} disabled={!apiKey}>
						{copied ? 'Copied!' : 'Copy'}
					</button>
					<button onclick={regenerate} disabled={regenerating} class="danger">
						{regenerating ? 'Generating...' : 'Regenerate'}
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.settings {
		max-width: 520px;
		margin: 24px 32px;
	}

	h2 {
		color: #e2e8f0;
		font-size: 0.95rem;
		margin: 0 0 20px;
	}

	h2::before {
		content: '$ ';
		color: #238636;
	}

	.section {
		background: #0d1117;
		border: 1px solid #1a1a1a;
		padding: 18px 20px;
	}

	h3 {
		color: #c9d1d9;
		font-size: 0.8rem;
		margin: 0 0 8px;
	}

	.desc {
		color: #8b949e;
		font-size: 0.7rem;
		margin: 0 0 14px;
		line-height: 1.5;
	}

	.desc code {
		background: #1a1a1a;
		padding: 1px 5px;
		font-size: 0.7rem;
	}

	.muted {
		color: #6e7681;
		font-size: 0.75rem;
	}

	.key-row {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
	}

	.key-display {
		flex: 1;
		min-width: 0;
		background: #000;
		color: #3fb950;
		padding: 6px 10px;
		font-size: 0.7rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		border: 1px solid #1a1a1a;
	}

	.key-actions {
		display: flex;
		gap: 6px;
		flex-shrink: 0;
	}

	.key-actions button {
		background: #1a1a1a;
		color: #c9d1d9;
		border: 1px solid #30363d;
		padding: 5px 12px;
		font-size: 0.7rem;
		cursor: pointer;
		font-family: inherit;
		transition: background 0.12s;
	}

	.key-actions button:hover:not(:disabled) {
		background: #30363d;
	}

	.key-actions button:disabled {
		opacity: 0.5;
		cursor: default;
	}

	.key-actions button.danger:hover:not(:disabled) {
		background: #da3633;
		border-color: #da3633;
		color: white;
	}
</style>