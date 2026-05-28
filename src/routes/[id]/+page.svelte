<script lang="ts">
	import { onMount } from 'svelte';
	import { pb } from '$lib/pocketbase.js';
	let { data } = $props<{ data: { id: string } }>();

	let task = $state<any>(null);
	let loading = $state(true);
	let error = $state('');
	let editorContainer = $state<HTMLDivElement | null>(null);
	let editor: any;
	let saving = $state(false);
	let saveStatus = $state('');
	let finishing = $state(false);
	let monacoReady = $state(false);

	function getMonacoLanguage(lang: string): string {
		const map: Record<string, string> = {
			typescript: 'typescript',
			javascript: 'javascript',
			python: 'python',
			go: 'go',
			rust: 'rust',
			java: 'java',
			cpp: 'cpp'
		};
		return map[lang] ?? 'plaintext';
	}

	async function loadTask(id: string) {
		loading = true;
		error = '';
		try {
			task = await pb.getTask(id);
		} catch (e: any) {
			error = e.message || 'Failed to load task';
			task = null;
		} finally {
			loading = false;
		}
	}

	// Reload task whenever the route ID changes
	$effect(() => {
		const id = data.id;
		loadTask(id);
	});

	// Realtime subscription for this task
	$effect(() => {
		const id = data.id;
		if (!id) return;

		let unsub: (() => Promise<void>) | null = null;

		pb.subscribeToTask(id, (e: any) => {
			if (e.record.id === id) {
				task = e.record;
			}
		}).then((u) => {
			unsub = u;
		});

		return () => {
			if (unsub) unsub();
		};
	});

	// Load Monaco script once on mount
	onMount(() => {
		const script = document.createElement('script');
		script.src = 'https://cdn.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs/loader.js';
		script.onload = () => {
			(window as any).require.config({
				paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs' }
			});
			(window as any).require(['vs/editor/editor.main'], () => {
				monacoReady = true;
			});
		};
		document.head.appendChild(script);

		return () => {
			if (editor) editor.dispose();
			script.remove();
		};
	});

	// Create or swap editor model when Monaco is ready and task changes
	$effect(() => {
		if (!monacoReady || !editorContainer || !task) return;

		const monaco = (window as any).monaco;
		const code = task.code ?? '';
		const lang = getMonacoLanguage(task.language ?? 'typescript');

		if (editor) {
			const oldModel = editor.getModel();
			if (oldModel) oldModel.dispose();
			const newModel = monaco.editor.createModel(code, lang);
			editor.setModel(newModel);
		} else {
			editor = monaco.editor.create(editorContainer, {
				value: code,
				language: lang,
				theme: 'vs-dark',
				automaticLayout: true,
				fontSize: 14,
				minimap: { enabled: false },
				scrollBeyondLastLine: false
			});
		}
	});

	async function save() {
		if (!editor || !task) return;
		saving = true;
		saveStatus = '';
		const code = editor.getValue();
		try {
			task = await pb.updateTaskCode(data.id, code);
			saveStatus = 'Saved';
		} catch (e: any) {
			saveStatus = e.message || 'Save failed';
		}
		saving = false;
		setTimeout(() => (saveStatus = ''), 2000);
	}

	async function finish() {
		if (!editor || !task) return;
		finishing = true;
		await save();
		try {
			task = await pb.submitTask(data.id);
		} catch (e: any) {
			saveStatus = e.message || 'Finish failed';
		}
		finishing = false;
	}
</script>

<div class="task-view">
	{#if loading}
		<div class="loading">Loading task...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else if task}
		<div class="info-panel">
			<div class="info-header">
				<h2>{task.title}</h2>
				<span class="status-badge" class:completed={task.status === 'completed'}>{task.status}</span>
			</div>
			{#if task.description}
				<p class="description">{task.description}</p>
			{/if}
			<div class="meta">
				<span>Language: <code>{task.language}</code></span>
				<span>Updated: {new Date(task.updated_at).toLocaleString()}</span>
			</div>
			{#if task.grade}
				<div class="grade">
					<strong>Grade:</strong> {task.grade}
					{#if task.feedback}
						<p>{task.feedback}</p>
					{/if}
				</div>
			{/if}
			<div class="actions">
				<button onclick={save} disabled={saving}>
					{saving ? 'Saving...' : 'Save'}
				</button>
				<button class="finish" onclick={finish} disabled={finishing || task.status === 'completed'}>
					{finishing ? 'Finishing...' : 'Finish'}
				</button>
				{#if saveStatus}
					<span class="save-status">{saveStatus}</span>
				{/if}
			</div>
		</div>
		<div class="editor-wrapper">
			<div bind:this={editorContainer} class="monaco-container"></div>
		</div>
	{:else}
		<div class="error">Task not found</div>
	{/if}
</div>

<style>
	.task-view {
		display: flex;
		flex-direction: column;
		height: 100%;
		font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Fira Code', 'JetBrains Mono', monospace;
	}

	.loading, .error {
		padding: 20px;
		color: #8b949e;
		font-size: 0.8rem;
	}

	.error {
		color: #f85149;
	}

	.info-panel {
		padding: 14px 18px;
		border-bottom: 1px solid #1a1a1a;
		background: #0d1117;
		flex-shrink: 0;
	}

	.info-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		margin-bottom: 10px;
	}

	.info-header h2 {
		margin: 0;
		font-size: 0.9rem;
		color: #e2e8f0;
		font-weight: 600;
	}

	.info-header h2::before {
		content: '$ cat ';
		color: #238636;
		font-weight: 400;
	}

	.status-badge {
		font-size: 0.65rem;
		padding: 2px 8px;
		background: #1a1a1a;
		color: #8b949e;
		text-transform: uppercase;
		font-weight: 600;
	}

	.status-badge.completed {
		background: rgba(35, 134, 54, 0.15);
		color: #3fb950;
	}

	.description {
		margin: 0 0 12px;
		color: #8b949e;
		font-size: 0.8rem;
		line-height: 1.5;
		white-space: pre-wrap;
	}

	.meta {
		display: flex;
		gap: 16px;
		font-size: 0.7rem;
		color: #8b949e;
		margin-bottom: 12px;
	}

	.meta code {
		background: #000;
		padding: 1px 5px;
		color: #58a6ff;
		font-family: inherit;
	}

	.grade {
		background: #000;
		border: 1px solid #1a1a1a;
		padding: 8px 12px;
		margin-bottom: 12px;
		font-size: 0.8rem;
	}

	.grade::before {
		content: '$ grade\A';
		color: #238636;
		white-space: pre;
	}

	.grade p {
		margin: 6px 0 0;
		color: #8b949e;
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.actions button {
		background: #1f6feb;
		color: #fff;
		border: none;
		padding: 6px 14px;
		font-size: 0.8rem;
		font-weight: 600;
		cursor: pointer;
		font-family: inherit;
	}

	.actions button:hover:not(:disabled) {
		background: #388bfd;
	}

	.actions button.finish {
		background: #238636;
		color: #000;
	}

	.actions button.finish:hover:not(:disabled) {
		background: #2ea043;
	}

	.actions button:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.save-status {
		font-size: 0.75rem;
		color: #3fb950;
	}

	.editor-wrapper {
		flex: 1;
		overflow: hidden;
		position: relative;
		border-top: 1px solid #1a1a1a;
	}

	.monaco-container {
		position: absolute;
		inset: 0;
	}
</style>
