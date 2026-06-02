<script lang="ts">
	import { pb } from '$lib/pocketbase.js';
	import { onMount } from 'svelte';

	let apiKey = $state('');
	let loading = $state(true);
	let copied = $state(false);
	let regenerating = $state(false);

	let telegramStatus = $state<{ connected: boolean; chat_id: string } | null>(null);
	let telegramCode = $state('');
	let telegramLoading = $state(false);
	let telegramError = $state('');
	let telegramSuccess = $state('');
	let skillCopied = $state(false);

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

	async function loadTelegramStatus() {
		try {
			const status = await pb.getTelegramStatus();
			telegramStatus = status;
		} catch (e) {
			console.error('Failed to load Telegram status', e);
		}
	}

	async function verifyTelegram() {
		if (!telegramCode.trim()) return;
		telegramLoading = true;
		telegramError = '';
		telegramSuccess = '';
		try {
			const result = await pb.verifyTelegramCode(telegramCode.trim());
			telegramStatus = result;
			telegramCode = '';
			telegramSuccess = 'Telegram connected successfully!';
			setTimeout(() => (telegramSuccess = ''), 3000);
		} catch (e: any) {
			telegramError = e.message || 'Verification failed';
		} finally {
			telegramLoading = false;
		}
	}

	async function unsubscribeTelegram() {
		telegramLoading = true;
		telegramError = '';
		telegramSuccess = '';
		try {
			await pb.unsubscribeTelegram();
			telegramStatus = { connected: false, chat_id: '' };
			telegramSuccess = 'Unsubscribed from Telegram notifications';
			setTimeout(() => (telegramSuccess = ''), 3000);
		} catch (e) {
			telegramError = 'Failed to unsubscribe';
		} finally {
			telegramLoading = false;
		}
	}

	async function downloadSkill() {
		try {
			const res = await fetch('https://raw.githubusercontent.com/ktappdev/fromai/main/skills/fromai/SKILL.md');
			const blob = await res.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'SKILL.md';
			a.click();
			URL.revokeObjectURL(url);
		} catch (e) {
			console.error('Download failed', e);
		}
	}

	onMount(async () => {
		console.log('[Settings] onMount - auth check starting');
		const token = pb.getAuthToken();
		console.log('[Settings] token present:', !!token);

		// Skip auth check if no token exists (truly unauthenticated)
		if (!token) {
			console.log('[Settings] no token, skipping settings load');
			loading = false;
			return;
		}

		// Always try to load settings data if token exists (token may still be valid for API calls even if authRefresh fails)
		console.log('[Settings] loading settings data (API key and Telegram status)');
		loadKey();
		loadTelegramStatus();

		// Verify session is valid for logging/debugging purposes
		try {
			console.log('[Settings] calling getMe()');
			const user = await pb.getMe();
			console.log('[Settings] getMe result:', user ? 'user found' : 'user is null');
		} catch (e) {
			console.error('[Settings] getMe failed (settings data may still load):', e);
		}
	});
</script>

<div class="settings">
	<h2>Settings</h2>

	<div class="section">
		<h3>CLI API Key</h3>
		<p class="desc">
			Use this key with the fromai CLI to authenticate from the terminal.
			Run <code>fai init --key &lt;key&gt;</code> to set it up.
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

	<div class="section">
		<h3>Telegram Notifications</h3>
		<p class="desc">
			Receive task notifications via Telegram.
			Open <code>@official_faibot</code> on Telegram and click Start to get your verification code.
		</p>

		{#if telegramStatus === null}
			<p class="muted">Loading...</p>
		{:else if !telegramStatus.connected}
			<div class="telegram-form">
				<div class="input-row">
					<span class="prompt">&gt; verification code</span>
					<input
						type="text"
						placeholder="6-digit code"
						bind:value={telegramCode}
						maxlength="6"
						disabled={telegramLoading}
					/>
					<button onclick={verifyTelegram} disabled={telegramLoading || !telegramCode.trim()}>
						{telegramLoading ? 'Verifying...' : 'Verify'}
					</button>
				</div>
				{#if telegramError}
					<p class="error">{telegramError}</p>
				{/if}
				{#if telegramSuccess}
					<p class="success">{telegramSuccess}</p>
				{/if}
			</div>
		{:else}
			<div class="telegram-connected">
				<p class="status">Notifications active ✓</p>
				<p class="chat-id">Chat ID: {telegramStatus.chat_id}</p>
				<button onclick={unsubscribeTelegram} disabled={telegramLoading} class="danger">
					{telegramLoading ? 'Unsubscribing...' : 'Unsubscribe'}
				</button>
				{#if telegramError}
					<p class="error">{telegramError}</p>
				{/if}
				{#if telegramSuccess}
					<p class="success">{telegramSuccess}</p>
				{/if}
			</div>
		{/if}
	</div>

	<div class="section">
		<h3>AI Agent Skill</h3>
		<p class="desc">
			Download the fromai skill file for your AI agent. The agent uses this to create well-scoped tasks with appropriate difficulty. Place it where your agent looks for skills:
			<code>Claude Code</code> → <code>.claude/skills/</code> · <code>Cursor</code> → <code>.cursor/rules/</code> · <code>Copilot</code> → <code>.github/copilot-instructions.md</code>
		</p>
		<div class="skill-download">
			<button onclick={downloadSkill} class="download-btn">⬇ Download SKILL.md</button>
			<span class="or-divider">or</span>
			<code class="skill-cmd">curl -O https://raw.githubusercontent.com/ktappdev/fromai/main/skills/fromai/SKILL.md</code>
			<button onclick={() => { navigator.clipboard.writeText('curl -O https://raw.githubusercontent.com/ktappdev/fromai/main/skills/fromai/SKILL.md'); skillCopied = true; setTimeout(() => skillCopied = false, 2000); }} class="copy-btn">
				{skillCopied ? '✓ Copied' : 'Copy'}
			</button>
		</div>
	</div>
</div>

<style>
	.settings {
		max-width: 520px;
		margin: 24px 32px;
	}

	/* Mobile Responsive */
	@media (max-width: 767px) {
		.settings {
			margin: 16px;
		}

		.key-row {
			flex-direction: column;
			align-items: stretch;
		}

		.key-display {
			width: 100%;
		}

		.key-actions {
			justify-content: stretch;
		}

		.key-actions button {
			flex: 1;
		}

		.skill-download {
			flex-direction: column;
			align-items: stretch;
		}

		.skill-cmd {
			width: 100%;
		}

		.input-row {
			flex-direction: column;
			align-items: stretch;
		}

		.input-row input {
			width: 100%;
			box-sizing: border-box;
		}
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

	.telegram-form,
	.telegram-connected {
		margin-top: 8px;
	}

	.input-row {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}

	.prompt {
		color: #238636;
		font-size: 0.7rem;
		flex-shrink: 0;
	}

	.input-row input {
		background: #000;
		color: #c9d1d9;
		border: 1px solid #30363d;
		padding: 5px 10px;
		font-size: 0.7rem;
		font-family: inherit;
		width: 80px;
		text-transform: uppercase;
		letter-spacing: 1px;
	}

	.input-row input:focus {
		outline: none;
		border-color: #238636;
	}

	.input-row input:disabled {
		opacity: 0.5;
	}

	.input-row button {
		background: #238636;
		color: white;
		border: 1px solid #238636;
		padding: 5px 12px;
		font-size: 0.7rem;
		cursor: pointer;
		font-family: inherit;
		transition: background 0.12s;
	}

	.input-row button:hover:not(:disabled) {
		background: #2ea043;
	}

	.input-row button:disabled {
		opacity: 0.5;
		cursor: default;
	}

	.telegram-connected .status {
		color: #238636;
		font-size: 0.75rem;
		margin: 0 0 4px;
	}

	.telegram-connected .chat-id {
		color: #8b949e;
		font-size: 0.7rem;
		margin: 0 0 10px;
	}

	.telegram-connected button {
		background: #1a1a1a;
		color: #c9d1d9;
		border: 1px solid #30363d;
		padding: 5px 12px;
		font-size: 0.7rem;
		cursor: pointer;
		font-family: inherit;
		transition: background 0.12s;
	}

	.telegram-connected button:hover:not(:disabled) {
		background: #30363d;
	}

	.telegram-connected button.danger:hover:not(:disabled) {
		background: #da3633;
		border-color: #da3633;
		color: white;
	}

	.telegram-connected button:disabled {
		opacity: 0.5;
		cursor: default;
	}

	.error {
		color: #da3633;
		font-size: 0.7rem;
		margin: 6px 0 0;
	}

	.success {
		color: #238636;
		font-size: 0.7rem;
		margin: 6px 0 0;
	}

	.skill-download {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}

	.download-btn {
		background: #238636;
		color: #000;
		text-decoration: none;
		padding: 6px 14px;
		font-size: 0.7rem;
		font-weight: 600;
		font-family: inherit;
		cursor: pointer;
		transition: background 0.15s;
		flex-shrink: 0;
		white-space: nowrap;
	}

	.download-btn:hover {
		background: #2ea043;
	}

	.or-divider {
		color: #484f58;
		font-size: 0.65rem;
	}

	.skill-cmd {
		flex: 1;
		min-width: 0;
		background: #000;
		color: #c9d1d9;
		padding: 8px 10px;
		font-size: 0.65rem;
		border: 1px solid #1a1a1a;
		word-break: break-all;
		line-height: 1.5;
	}

	.copy-btn {
		background: #1a1a1a;
		color: #c9d1d9;
		border: 1px solid #30363d;
		padding: 6px 12px;
		font-size: 0.7rem;
		cursor: pointer;
		font-family: inherit;
		transition: background 0.12s;
		flex-shrink: 0;
		white-space: nowrap;
	}

	.copy-btn:hover {
		background: #30363d;
	}
</style>