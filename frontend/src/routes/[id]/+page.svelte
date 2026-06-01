<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { pb } from '$lib/pocketbase.js';
	import { registerSvelteLanguage } from '$lib/monaco-svelte.js';
	import ContextModal from './ContextModal.svelte';
	let { data } = $props<{ data: { id: string } }>();

	let task = $state<any>(null);
	let loading = $state(true);
	let error = $state('');
	let editorContainer = $state<HTMLDivElement | null>(null);
	let editor: any;
	let saving = $state(false);
	let saveStatus = $state('');
	let finishing = $state(false);
	let confirmArchive = $state(false);
	let archiving = $state(false);
	let deletingPermanently = $state(false);
	let monacoReady = $state(false);
	let showContextModal = $state(false);
	let hasChanges = $state(false);

	// Derived state for read-only mode (completed tasks)
	let isReadOnly = $derived(task?.status === 'completed');

	function getMonacoLanguage(lang: string): string {
		const map: Record<string, string> = {
			typescript: 'typescript',
			javascript: 'javascript',
			python: 'python',
			go: 'go',
			rust: 'rust',
			java: 'java',
			cpp: 'cpp',
			svelte: 'svelte'
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
		// Reset per-task UI state when navigating to a different task
		confirmArchive = false;
		saveStatus = '';
		error = '';
		saving = false;
		finishing = false;
		archiving = false;
		deletingPermanently = false;
		if (editor) {
			editor.dispose();
			editor = null;
		}
		loadTask(id);
	});

	// Realtime subscription for this task
	$effect(() => {
		const id = data.id;
		if (!id) return;

		let unsub: (() => Promise<void>) | null = null;

		pb.subscribeToTask(id, (e: any) => {
			// Delete event from another session — redirect away from stale page
			if (e.action === 'delete') {
				goto('/');
				return;
			}
			// Archived from another session — redirect to home
			if (e.action === 'update' && e.record?.archived === true) {
				goto('/');
				return;
			}
			// Normal update/create — refresh local state
			if (e.record?.id === id) {
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
				const monaco = (window as any).monaco;
				registerSvelteLanguage(monaco);
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
			editor.updateOptions({ readOnly: isReadOnly });
		} else {
			editor = monaco.editor.create(editorContainer, {
				value: code,
				language: lang,
				theme: 'vs-dark',
				automaticLayout: true,
				fontSize: 14,
				minimap: { enabled: false },
				scrollBeyondLastLine: false,
				readOnly: isReadOnly
			});

			// Add Ctrl/Cmd+S save binding for non-completed tasks
			if (!isReadOnly) {
				editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
					save();
				});

				// Track changes to enable/disable Finish button
				hasChanges = editor.getValue() !== task.starter_code;
				editor.onDidChangeModelContent(() => {
					hasChanges = editor.getValue() !== task.starter_code;
				});
			}
		}
	});

	async function save(): Promise<boolean> {
		if (!editor || !task) return false;
		saving = true;
		saveStatus = '';
		const code = editor.getValue();
		try {
			task = await pb.updateTaskCode(data.id, code);
			saveStatus = 'Saved';
			saving = false;
			setTimeout(() => (saveStatus = ''), 2000);
			return true;
		} catch (e: any) {
			saveStatus = e.message || 'Save failed';
			saving = false;
			setTimeout(() => (saveStatus = ''), 2000);
			return false;
		}
	}

	async function finish() {
		if (!editor || !task) return;
		finishing = true;
		const saved = await save();
		if (!saved) {
			finishing = false;
			return;
		}
		try {
			task = await pb.submitTask(data.id);
		} catch (e: any) {
			saveStatus = e.message || 'Finish failed';
		}
		finishing = false;
	}

	function promptArchive() {
		confirmArchive = true;
	}

	async function doArchive() {
		archiving = true;
		try {
			await pb.archiveTask(data.id);
			goto('/');
		} catch (e: any) {
			saveStatus = e.message || 'Archive failed';
			archiving = false;
		}
	}

	async function doDeletePermanently() {
		deletingPermanently = true;
		try {
			await pb.deleteTask(data.id);
			goto('/');
		} catch (e: any) {
			saveStatus = e.message || 'Delete failed';
			deletingPermanently = false;
		}
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
				<button class="view-context-btn" onclick={() => showContextModal = true}>
					<span class="prompt">$</span> view context
				</button>
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
				<button onclick={save} disabled={saving || isReadOnly}>
					{saving ? 'Saving...' : 'Save'}
				</button>
				<button class="finish" onclick={finish} disabled={finishing || task.status === 'completed' || !hasChanges}>
					{finishing ? 'Finishing...' : 'Finish'}
				</button>
				{#if saveStatus}
					<span class="save-status">{saveStatus}</span>
				{/if}
				{#if isReadOnly}
					<span class="readonly-badge">read-only</span>
				{/if}
				{#if !hasChanges && !isReadOnly}
					<span class="no-changes-hint">Make changes before finishing</span>
				{/if}
				<span class="spacer"></span>
				{#if confirmArchive}
					<button class="archive" onclick={doArchive} disabled={archiving || deletingPermanently}>
						{archiving ? 'Archiving...' : 'Archive'}
					</button>
					<button class="delete-permanent" onclick={doDeletePermanently} disabled={archiving || deletingPermanently}>
						{deletingPermanently ? 'Deleting...' : 'Delete permanently'}
					</button>
					<button class="cancel" onclick={() => confirmArchive = false} disabled={archiving || deletingPermanently}>Cancel</button>
				{:else}
					<button class="archive collapsed" onclick={promptArchive}>Archive</button>
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

<ContextModal open={showContextModal} onClose={() => showContextModal = false} content={task?.description ?? ''} />

<style>
	.task-view { display: flex; flex-direction: column; height: 100%; font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Fira Code', 'JetBrains Mono', monospace; }
	.loading, .error { padding: 20px; color: #8b949e; font-size: 0.8rem; }
	.error { color: #f85149; }
	.info-panel { padding: 14px 18px; border-bottom: 1px solid #1a1a1a; background: #0d1117; flex-shrink: 0; }
	.info-header { display: flex; justify-content: space-between; align-items: center; gap: 12px; margin-bottom: 10px; }
	.info-header h2 { margin: 0; font-size: 0.9rem; color: #e2e8f0; font-weight: 600; }
	.info-header h2::before { content: '$ cat '; color: #238636; font-weight: 400; }
	.status-badge { font-size: 0.65rem; padding: 2px 8px; background: #1a1a1a; color: #8b949e; text-transform: uppercase; font-weight: 600; }
	.status-badge.completed { background: rgba(35, 134, 54, 0.15); color: #3fb950; }
	.meta { display: flex; gap: 16px; font-size: 0.7rem; color: #8b949e; margin-bottom: 12px; }
	.meta code { background: #000; padding: 1px 5px; color: #58a6ff; font-family: inherit; }
	.grade { background: #000; border: 1px solid #1a1a1a; padding: 8px 12px; margin-bottom: 12px; font-size: 0.8rem; }
	.grade::before { content: '$ grade\A'; color: #238636; white-space: pre; }
	.grade p { margin: 6px 0 0; color: #8b949e; white-space: pre-wrap; line-height: 1.5; }
	.actions { display: flex; align-items: center; gap: 8px; }
	.actions button { background: #1f6feb; color: #fff; border: none; padding: 6px 14px; font-size: 0.8rem; font-weight: 600; cursor: pointer; font-family: inherit; }
	.actions button:hover:not(:disabled) { background: #388bfd; }
	.actions button.finish { background: #238636; color: #000; }

	.actions button.finish:hover:not(:disabled) { background: #2ea043; }
	.actions button:disabled { opacity: 0.4; cursor: not-allowed; }
	.actions .spacer { flex: 1; }
	.actions button.archive { background: #238636; color: #000; }
	.actions button.archive:hover:not(:disabled) { background: #2ea043; }
	.actions button.archive.collapsed { background: transparent; color: #8b949e; border: 1px solid #8b949e; }
	.actions button.archive.collapsed:hover:not(:disabled) { color: #e2e8f0; border-color: #e2e8f0; }
	.actions button.delete-permanent { background: transparent; color: #f85149; border: 1px solid #f85149; }
	.actions button.delete-permanent:hover:not(:disabled) { background: rgba(248, 81, 73, 0.1); }
	.actions button.cancel { background: transparent; color: #8b949e; border: none; padding: 6px 10px; }
	.actions button.cancel:hover:not(:disabled) { color: #e2e8f0; }

	.save-status { font-size: 0.75rem; color: #3fb950; }
	.readonly-badge { font-size: 0.65rem; padding: 2px 8px; background: rgba(210, 153, 34, 0.15); color: #d29922; font-weight: 600; }
	.no-changes-hint { font-size: 0.7rem; color: #d29922; }
	.editor-wrapper { flex: 1; overflow: hidden; position: relative; border-top: 1px solid #1a1a1a; }
	.monaco-container { position: absolute; inset: 0; }
	.view-context-btn { background: transparent; color: #58a6ff; border: 1px solid #1a1a1a; padding: 4px 12px; font-size: 0.75rem; font-family: inherit; cursor: pointer; text-align: left; margin-bottom: 12px; }
	.view-context-btn:hover { background: rgba(88, 166, 255, 0.1); border-color: #58a6ff; }
	.view-context-btn .prompt { color: #238636; margin-right: 6px; }
</style>
