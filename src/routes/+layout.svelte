<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import { pb } from '$lib/pocketbase.js';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { afterNavigate } from '$app/navigation';

	let { children } = $props();

	let user = $state<any>(null);
	let tasks = $state<any[]>([]);
	let loading = $state(true);

	async function loadData() {
		try {
			// Skip authRefresh call if no token exists
			if (!pb.getAuthToken()) {
				user = null;
				tasks = [];
				return;
			}
			user = await pb.getMe();
			if (user) {
				tasks = await pb.listTasks();
			} else {
				// Clear tasks if user is null (truly unauthenticated)
				tasks = [];
			}
		} catch (e) {
			console.error('Failed to load data', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadData();
	});

	afterNavigate(() => {
		// Reload user data after navigation from login/signup
		loadData();
	});

	// Realtime subscription for tasks
	$effect(() => {
		if (!user) return;

		let unsub: (() => Promise<void>) | null = null;

		pb.subscribeToTasks((e: any) => {
			if (e.action === 'create') {
				tasks = [e.record, ...tasks];
			} else if (e.action === 'update') {
				if (e.record.archived === true) {
					tasks = tasks.filter((t) => t.id !== e.record.id);
				} else {
					tasks = tasks.map((t) => (t.id === e.record.id ? e.record : t));
				}
			} else if (e.action === 'delete') {
				tasks = tasks.filter((t) => t.id !== e.record.id);
			}
		}).then((u) => {
			unsub = u;
		});

		return () => {
			if (unsub) unsub();
		};
	});

	function getInitials(u: any): string {
		if (u.name) {
			return u.name.split(' ').map((n: string) => n[0]).join('').slice(0, 2).toUpperCase();
		}
		if (u.email) {
			return u.email.slice(0, 2).toUpperCase();
		}
		return 'U';
	}

	async function logout() {
		try {
			pb.signOut();
			window.location.href = '/login';
		} catch (e) {
			console.error('Logout failed', e);
		}
	}
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>fromai</title>
</svelte:head>

<div class="app">
	<aside class="sidebar">
		<div class="sidebar-header">
			<a href="/" class="logo">
				<span class="logo-icon">$</span>
				<span class="logo-text">fromai</span>
			</a>
		</div>

		{#if loading}
			<div class="loading">
				<div class="spinner"></div>
				<span>Loading...</span>
			</div>
		{:else if !user}
			<div class="auth-section">
				<p class="auth-title">Ready to code?</p>
				<p class="auth-sub">Sign in to view your assigned tasks.</p>
				<a href="/login" class="btn-primary">Sign In</a>
			</div>

			<div class="sidebar-footer">
				<a href="/how-it-works" class="settings-link">
					<span class="settings-icon">?</span>
					How It Works
				</a>
				<a href="/install" class="settings-link">
					<span class="settings-icon">⬇</span>
					Install CLI
				</a>
			</div>
		{:else}
			<div class="user-section">
				<div class="user-row">
					<div class="avatar">{getInitials(user)}</div>
					<div class="user-meta">
						<span class="user-name">{user.name || user.email?.split('@')[0] || 'User'}</span>
						<span class="user-email">{user.email}</span>
					</div>
				</div>
				<button class="logout-btn" onclick={logout} title="Log out">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
				</button>
			</div>

			<div class="nav-section">
				<div class="nav-header">
					<span class="nav-label">Your Tasks</span>
					<a href="/new" class="new-link">+ New</a>
				</div>
				{#if tasks.length === 0}
					<div class="empty-state">
						<p>No tasks yet.</p>
						<span>Tasks will appear here when assigned.</span>
					</div>
				{:else}
					<ul class="task-list">
						{#each tasks as task}
							<li class:active={$page.params.id === task.id}>
								<a href="/{task.id}">
									<div class="task-info">
										<span class="task-title">{task.title}</span>
										<span class="task-lang">{task.language}</span>
									</div>
									<span class="task-status" class:completed={task.status === 'completed'}></span>
								</a>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<div class="sidebar-footer">
				<a href="/how-it-works" class="settings-link">
					<span class="settings-icon">?</span>
					How It Works
				</a>
				<a href="/install" class="settings-link">
					<span class="settings-icon">⬇</span>
					Install CLI
				</a>
				<a href="/settings" class="settings-link">
					<span class="settings-icon">⚙</span>
					Settings
				</a>
			</div>
		{/if}
	</aside>
	<main>
		<div class="terminal-chrome">
			<div class="terminal-titlebar">
				<div class="terminal-dots">
					<div class="dot red"></div>
					<div class="dot yellow"></div>
					<div class="dot green"></div>
				</div>
				<span class="terminal-title">fromai — zsh</span>
			</div>
			<div class="terminal-body">
				{@render children()}
			</div>
		</div>
	</main>
</div>

<style>
	:global(body) {
		margin: 0;
		font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Fira Code', 'JetBrains Mono', monospace;
		background: #000;
		color: #e2e8f0;
	}

	.app {
		display: flex;
		height: 100vh;
		overflow: hidden;
	}

	/* ── Sidebar ── */
	.sidebar {
		width: 260px;
		background: #000;
		border-right: 1px solid #1a1a1a;
		display: flex;
		flex-direction: column;
		overflow-y: auto;
	}

	/* Header */
	.sidebar-header {
		padding: 16px 18px 12px;
	}

	.logo {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		text-decoration: none;
		font-weight: 700;
		font-size: 0.9rem;
		letter-spacing: -0.3px;
	}

	.logo-icon {
		color: #238636;
	}

	.logo-text {
		color: #e2e8f0;
	}

	/* Loading */
	.loading {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 20px;
		color: #8b949e;
		font-size: 0.8rem;
	}

	.spinner {
		width: 12px;
		height: 12px;
		border: 2px solid #1a1a1a;
		border-top-color: #3fb950;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	/* Auth section */
	.auth-section {
		padding: 20px;
		margin-top: 8px;
	}

	.auth-section::before {
		content: '$ ';
		color: #238636;
	}

	.auth-title {
		margin: 0 0 6px;
		font-size: 0.9rem;
		color: #e2e8f0;
		font-weight: 600;
	}

	.auth-sub {
		margin: 0 0 18px;
		font-size: 0.75rem;
		color: #8b949e;
		line-height: 1.5;
	}

	.btn-primary {
		display: inline-block;
		background: #238636;
		color: #000;
		text-decoration: none;
		padding: 6px 16px;
		font-size: 0.8rem;
		font-weight: 600;
		transition: background 0.15s;
	}

	.btn-primary:hover {
		background: #2ea043;
	}

	/* User section */
	.user-section {
		padding: 10px 14px;
		margin: 0 10px 10px;
		background: #0d1117;
		border: 1px solid #1a1a1a;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
	}

	.user-row {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
		flex: 1;
	}

	.avatar {
		width: 28px;
		height: 28px;
		background: #238636;
		color: #000;
		font-size: 0.65rem;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.user-meta {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.user-name {
		font-size: 0.75rem;
		font-weight: 600;
		color: #e2e8f0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.user-email {
		font-size: 0.65rem;
		color: #8b949e;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.logout-btn {
		background: transparent;
		border: none;
		color: #8b949e;
		padding: 2px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		transition: color 0.15s;
	}

	.logout-btn:hover {
		color: #da3633;
	}

	/* Nav section */
	.nav-section {
		flex: 1;
		overflow-y: auto;
		padding: 0 10px 16px;
	}

	.nav-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 6px 8px;
		margin-bottom: 4px;
	}

	.nav-label {
		font-size: 0.7rem;
		color: #6e7681;
	}

	.nav-label::before {
		content: '$ ';
		color: #238636;
	}

	.new-link {
		color: #3fb950;
		text-decoration: none;
		font-size: 0.75rem;
		font-weight: 600;
		padding: 2px 6px;
		transition: background 0.15s;
	}

	.new-link:hover {
		background: rgba(35, 134, 54, 0.12);
	}

	/* Empty state */
	.empty-state {
		padding: 24px 12px;
		color: #8b949e;
	}

	.empty-state::before {
		content: '$ ';
		color: #238636;
	}

	.empty-state p {
		margin: 0 0 4px;
		font-size: 0.8rem;
		color: #c9d1d9;
		display: inline;
	}

	.empty-state span {
		font-size: 0.7rem;
		color: #6e7681;
	}

	/* Task list */
	.task-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 1px;
	}

	.task-list li a {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		padding: 8px 10px;
		color: #c9d1d9;
		text-decoration: none;
		transition: background 0.12s;
		position: relative;
	}

	.task-list li a::before {
		content: '> ';
		color: #6e7681;
		font-size: 0.7rem;
		flex-shrink: 0;
	}

	.task-list li a:hover {
		background: rgba(48, 54, 61, 0.3);
	}

	.task-list li a:hover::before {
		color: #3fb950;
	}

	.task-list li.active a {
		background: rgba(35, 134, 54, 0.1);
		color: #3fb950;
	}

	.task-list li.active a::before {
		content: '> ';
		color: #238636;
	}

	.task-info {
		display: flex;
		flex-direction: column;
		gap: 1px;
		min-width: 0;
		flex: 1;
	}

	.task-title {
		font-size: 0.75rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.task-lang {
		font-size: 0.6rem;
		color: #6e7681;
		text-transform: lowercase;
	}

	.task-status {
		width: 6px;
		height: 6px;
		background: #d29922;
		flex-shrink: 0;
	}

	.task-status.completed {
		background: #238636;
	}

	/* ── Terminal Main Window ── */
	main {
		flex: 1;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		background: #000;
	}

	.terminal-chrome {
		flex: 1;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		border-left: 1px solid #1a1a1a;
	}

	.terminal-titlebar {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 14px;
		background: #0d1117;
		border-bottom: 1px solid #1a1a1a;
	}

	.terminal-dots {
		display: flex;
		gap: 6px;
		flex-shrink: 0;
	}

	.dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
	}

	.dot.red { background: #ff5f56; }
	.dot.yellow { background: #ffbd2e; }
	.dot.green { background: #27c93f; }

	.terminal-title {
		flex: 1;
		text-align: center;
		font-size: 0.7rem;
		color: #8b949e;
		padding-right: 40px;
	}

	.terminal-body {
		flex: 1;
		overflow: hidden;
	}

	/* Sidebar footer */
	.sidebar-footer {
		margin-top: auto;
		padding: 10px 6px 6px;
		border-top: 1px solid #1a1a1a;
	}

	.settings-link {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 10px;
		color: #8b949e;
		text-decoration: none;
		font-size: 0.7rem;
		transition: color 0.12s;
	}

	.settings-link:hover {
		color: #c9d1d9;
	}

	.settings-icon {
		font-size: 0.8rem;
	}
</style>
