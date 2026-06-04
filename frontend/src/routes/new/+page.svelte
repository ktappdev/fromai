<script lang="ts">
	import { goto } from '$app/navigation';
	import { pb } from '$lib/pocketbase.js';
	import { onMount } from 'svelte';

	let title = $state('');
	let description = $state('');
	let starterCode = $state('');
	let language = $state('typescript');
	let submitting = $state(false);
	let error = $state('');

	onMount(async () => {
		// Only redirect if no token exists (truly unauthenticated)
		if (!pb.getAuthToken()) {
			goto('/login');
			return;
		}
		// Try to get user, but don't redirect on transient failures
		await pb.getMe();
	});

	async function createTask(e: Event) {
		e.preventDefault();
		if (!title.trim()) {
			error = 'Title is required';
			return;
		}
		submitting = true;
		error = '';

		try {
			const result = await pb.createTask({
				title: title.trim(),
				description: description.trim(),
				starter_code: starterCode,
				language
			});
			goto(`/${result.id}`);
		} catch (e: any) {
			error = e.message || 'Failed to create task';
		}
		submitting = false;
	}
</script>

<div class="welcome">
	<h2>Create New Task</h2>
	{#if error}
		<div class="error">{error}</div>
	{/if}
	<form onsubmit={createTask}>
		<div class="field">
			<label for="title">Title *</label>
			<input id="title" type="text" bind:value={title} placeholder="e.g. Implement a binary search" />
		</div>
		<div class="field">
			<label for="description">Description</label>
			<textarea id="description" bind:value={description} rows="4" placeholder="Explain what needs to be done..."></textarea>
		</div>
		<div class="field">
			<label for="language">Language</label>
			<select id="language" bind:value={language}>
				<option value="typescript">TypeScript</option>
				<option value="javascript">JavaScript</option>
				<option value="python">Python</option>
				<option value="go">Go</option>
				<option value="rust">Rust</option>
				<option value="java">Java</option>
				<option value="cpp">C++</option>
				<option value="plaintext">Plain text / pseudo-code</option>
			</select>
		</div>
		<div class="field">
			<label for="starter">Starter Code</label>
			<textarea id="starter" bind:value={starterCode} rows="8" placeholder="Paste starter code here..."></textarea>
		</div>
		<button type="submit" disabled={submitting}>
			{submitting ? 'Creating...' : 'Create Task'}
		</button>
	</form>
</div>

<style>
	.welcome {
		max-width: 640px;
		margin: 32px auto;
		padding: 0 24px;
		font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Fira Code', 'JetBrains Mono', monospace;
	}

	.welcome h2 {
		margin-top: 0;
		color: #e2e8f0;
		font-size: 0.9rem;
		font-weight: 600;
	}

	.welcome h2::before {
		content: '$ ';
		color: #238636;
	}

	.error {
		background: rgba(218, 54, 51, 0.1);
		color: #f85149;
		padding: 8px 12px;
		border-left: 2px solid #da3633;
		margin-bottom: 16px;
		font-size: 0.8rem;
	}

	.field {
		margin-bottom: 16px;
	}

	.field label {
		display: block;
		margin-bottom: 4px;
		font-size: 0.75rem;
		color: #8b949e;
	}

	.field label::before {
		content: '> ';
		color: #238636;
	}

	.field input,
	.field textarea,
	.field select {
		width: 100%;
		padding: 8px 10px;
		border: 1px solid #1a1a1a;
		background: #0d1117;
		color: #e2e8f0;
		font-family: inherit;
		font-size: 0.85rem;
		box-sizing: border-box;
		outline: none;
		transition: border-color 0.15s;
	}

	.field input:focus,
	.field textarea:focus,
	.field select:focus {
		border-color: #238636;
	}

	.field textarea {
		resize: vertical;
	}

	button[type='submit'] {
		background: #238636;
		color: #000;
		border: none;
		padding: 8px 16px;
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		font-family: inherit;
	}

	button[type='submit']:hover:not(:disabled) {
		background: #2ea043;
	}

	button[type='submit']:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	@media (max-width: 767px) {
		.welcome {
			margin: 16px;
			padding: 0;
		}

		.field input,
		.field textarea,
		.field select {
			font-size: 16px;
			padding: 12px 10px;
		}

		button[type='submit'] {
			font-size: 16px;
			padding: 12px 16px;
			min-height: 44px;
			width: 100%;
		}
	}
</style>
