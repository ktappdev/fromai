<script lang="ts">
	import { onMount } from 'svelte';
	import { pb } from '$lib/pocketbase.js';
	import { getBadgeDef } from '$lib/gamification.js';

	let user = $state<any>(null);
	let loading = $state(true);

	let headline = $state('');
	const fullHeadline = 'keep the human in the loop... kinda';
	let typeIndex = $state(0);

	let leakedCode = $state<string[]>([]);
	const snippets = [
		'function solve(arr) {',
		'  return arr.sort((a,b) => a - b);',
		'}',
		'const result = solve(input);',
		'if (grade > 90) {',
		'  console.log("excellent");',
		'}',
	];

	let stats = $state<any>(null);
	let todayChallenge = $state<any>(null);

	onMount(() => {
		const hasToken = !!pb.getAuthToken();
		if (!hasToken) {
			user = null;
			loading = false;
		} else {
			pb.getMe()
				.then((u) => { user = u; })
				.catch(() => { user = null; })
				.finally(() => {
					if (!user) user = { email: '' };
					loading = false;
				});

			// Fetch gamification data
			pb.getMyStats().then((s) => { stats = s; }).catch(() => {});
			pb.getTodayChallenge().then((c) => { todayChallenge = c; }).catch(() => {});
		}

		const typeInterval = setInterval(() => {
			if (typeIndex < fullHeadline.length) {
				headline = fullHeadline.slice(0, typeIndex + 1);
				typeIndex += 1;
			} else {
				clearInterval(typeInterval);
			}
		}, 60);

		const leakInterval = setInterval(() => {
			const snippet = snippets[Math.floor(Math.random() * snippets.length)];
			const left = Math.random() * 80;
			const top = Math.random() * 80;
			const id = Date.now().toString();
			leakedCode = [...leakedCode, `${id}|${snippet}|${left}|${top}`];
			setTimeout(() => {
				leakedCode = leakedCode.filter((c) => !c.startsWith(id));
			}, 4000);
		}, 800);

		return () => {
			clearInterval(typeInterval);
			clearInterval(leakInterval);
		};
	});
</script>

<div class="landing">
	{#if !loading && user}
		<div class="dashboard">
			<h1>Welcome back</h1>
			<p class="sub">Your AI works on your codebase and finds real issues.<br />Those become your tasks — bugs to fix, features to build, refactors to finish.<br />Nothing abstract. Just work that ships.</p>

			<!-- Streak Banner -->
			{#if stats}
				<div class="streak-banner">
					{#if stats.current_streak > 0}
						<span class="streak-icon">🔥</span>
						<span class="streak-text">{stats.current_streak}-day streak</span>
						{#if stats.best_streak > stats.current_streak}
							<span class="best-streak">(best: {stats.best_streak})</span>
						{/if}
					{:else}
						<span class="streak-text">Start completing tasks to build your streak!</span>
					{/if}
				</div>
			{/if}

			<!-- Today's Challenge Card -->
			<div class="challenge-card">
				<h2>Today's Challenge</h2>
				{#if todayChallenge}
					<div class="challenge-content">
						<div class="challenge-header">
							<span class="challenge-title">{todayChallenge.title}</span>
							<span class="difficulty-badge {todayChallenge.difficulty || 'medium'}">{todayChallenge.difficulty || 'medium'}</span>
						</div>
						<div class="challenge-meta">
							{#if todayChallenge.language}
								<span class="meta-tag">{todayChallenge.language}</span>
							{/if}
							{#if todayChallenge.category}
								<span class="meta-tag">{todayChallenge.category}</span>
							{/if}
						</div>
						<div class="challenge-actions">
							{#if todayChallenge.completed}
								<span class="completed-badge">Completed ✓</span>
								{#if todayChallenge.completion?.grade}
									<span class="grade-badge">Grade: {todayChallenge.completion.grade}</span>
								{/if}
							{:else if todayChallenge.completion?.task}
								<a href="/{todayChallenge.completion.task}" class="btn primary">Continue</a>
							{:else}
								<button class="btn primary" onclick={async () => {
									const result: any = await pb.startChallenge(todayChallenge.id);
									if (result?.task?.id) {
										window.location.href = `/${result.task.id}`;
									}
								}}>Start Challenge</button>
							{/if}
						</div>
					</div>
				{:else}
					<div class="no-challenge">
						<span class="no-challenge-text">No challenge today — create a task instead</span>
						<a href="/new" class="btn ghost">Create Task</a>
					</div>
				{/if}
			</div>

			<!-- Badge Preview -->
			{#if stats && stats.badges && stats.badges.length > 0}
				<div class="badge-preview">
					<h2>Badges</h2>
					<div class="badge-grid">
						{#each stats.badges.slice(0, 6) as badgeId}
							{@const badge = getBadgeDef(badgeId)}
							{#if badge}
								<div class="badge-item" title={badge.description}>
									<span class="badge-icon">{badge.icon}</span>
									<span class="badge-name">{badge.name}</span>
								</div>
							{/if}
						{/each}
					</div>
					{#if stats.badges.length > 6}
						<a href="/badges" class="more-link">+{stats.badges.length - 6} more</a>
					{/if}
				</div>
			{:else}
				<div class="badge-preview">
					<h2>Badges</h2>
					<div class="no-badges">
						<span class="no-badges-text">Complete tasks to earn badges</span>
						<a href="/badges" class="more-link">View all badges</a>
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<div class="hero">
			{#each leakedCode as leak (leak)}
				{@const [, text, l, t] = leak.split('|')}
				<div class="leak" style="left: {l}%; top: {t}%">{text}</div>
			{/each}
			<div class="content">
				<div class="brand">
					<span class="prompt">$</span>
					<span class="title">fromai</span>
					<span class="cursor">_</span>
				</div>
				<p class="tagline">{headline}<span class="caret"></span></p>
				<p class="subtext">Not puzzles. Real issues your AI found in your codebase.<br />Bugs, features, refactors — tasks that ship.</p>
				<div class="cta">
					<a href="/login" class="btn primary">Sign In</a>
					<a href="/login" class="btn ghost">Get Started</a>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.landing {
		height: 100%;
		overflow: hidden;
		position: relative;
	}

	.hero {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		position: relative;
		overflow: hidden;
		background: #000;
	}

	.hero::before {
		content: '';
		position: absolute;
		inset: 0;
		background-image:
			linear-gradient(rgba(48, 54, 61, 0.1) 1px, transparent 1px),
			linear-gradient(90deg, rgba(48, 54, 61, 0.1) 1px, transparent 1px);
		background-size: 40px 40px;
		pointer-events: none;
	}

	.leak {
		position: absolute;
		font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
		font-size: 0.75rem;
		color: rgba(139, 148, 158, 0.18);
		white-space: pre;
		pointer-events: none;
		animation: fadeInOut 4s ease-in-out forwards;
		user-select: none;
	}

	@keyframes fadeInOut {
		0% { opacity: 0; transform: translateY(4px); }
		20% { opacity: 1; }
		80% { opacity: 1; }
		100% { opacity: 0; transform: translateY(-4px); }
	}

	.content {
		position: relative;
		z-index: 1;
		text-align: center;
		padding: 0 24px;
	}

	.brand {
		font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
		font-size: 2.6rem;
		font-weight: 700;
		letter-spacing: -1px;
		margin-bottom: 12px;
	}

	.prompt {
		color: #238636;
		margin-right: 8px;
	}

	.title {
		color: #e2e8f0;
	}

	.cursor {
		color: #58a6ff;
		animation: blink 1s step-end infinite;
	}

	@keyframes blink {
		50% { opacity: 0; }
	}

	.tagline {
		font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
		font-size: 1.05rem;
		color: #8b949e;
		margin: 0 0 16px;
		min-height: 1.5em;
	}

	.subtext {
		font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
		font-size: 0.8rem;
		color: #6e7681;
		margin: 0 0 32px;
		line-height: 1.6;
	}

	.caret {
		display: inline-block;
		width: 2px;
		height: 1em;
		background: #58a6ff;
		vertical-align: text-bottom;
		margin-left: 2px;
		animation: blink 1s step-end infinite;
	}

	.cta {
		display: flex;
		gap: 12px;
		justify-content: center;
		flex-wrap: wrap;
	}

	.btn {
		display: inline-block;
		padding: 8px 20px;
		font-size: 0.85rem;
		text-decoration: none;
		font-weight: 500;
		transition: background 0.15s;
	}

	.btn.primary {
		background: #238636;
		color: #000;
	}

	.btn.primary:hover {
		background: #2ea043;
	}

	.btn.ghost {
		background: transparent;
		color: #c9d1d9;
		border: 1px solid #1a1a1a;
	}

	.btn.ghost:hover {
		border-color: #30363d;
		background: rgba(139, 148, 158, 0.06);
	}

	.dashboard {
		max-width: 640px;
		margin: 40px auto;
		padding: 0 24px;
		text-align: left;
	}

	.dashboard h1 {
		margin: 0 0 8px;
		color: #e2e8f0;
		font-size: 1rem;
		font-weight: 600;
	}

	.dashboard h1::before {
		content: '$ ';
		color: #238636;
	}

	.dashboard .sub {
		color: #8b949e;
		font-size: 0.8rem;
		margin: 0;
	}

	.streak-banner {
		background: rgba(35, 134, 54, 0.1);
		border: 1px solid rgba(35, 134, 54, 0.3);
		padding: 12px 16px;
		border-radius: 6px;
		margin: 24px 0;
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.streak-icon {
		font-size: 1.2rem;
	}

	.streak-text {
		color: #e2e8f0;
		font-size: 0.9rem;
		font-weight: 500;
	}

	.best-streak {
		color: #8b949e;
		font-size: 0.8rem;
	}

	.challenge-card {
		background: #0d1117;
		border: 1px solid #30363d;
		border-radius: 8px;
		padding: 20px;
		margin: 24px 0;
	}

	.challenge-card h2 {
		margin: 0 0 16px;
		color: #e2e8f0;
		font-size: 0.95rem;
		font-weight: 600;
	}

	.challenge-card h2::before {
		content: '$ ';
		color: #238636;
	}

	.challenge-content {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.challenge-header {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
	}

	.challenge-title {
		color: #e2e8f0;
		font-size: 1rem;
		font-weight: 500;
	}

	.difficulty-badge {
		padding: 2px 8px;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
		text-transform: uppercase;
	}

	.difficulty-badge.easy {
		background: rgba(35, 134, 54, 0.2);
		color: #3fb950;
	}

	.difficulty-badge.medium {
		background: rgba(187, 128, 9, 0.2);
		color: #d29922;
	}

	.difficulty-badge.hard {
		background: rgba(218, 54, 51, 0.2);
		color: #f85149;
	}

	.challenge-meta {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}

	.meta-tag {
		background: #161b22;
		border: 1px solid #30363d;
		padding: 4px 8px;
		border-radius: 4px;
		color: #8b949e;
		font-size: 0.75rem;
	}

	.challenge-actions {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-top: 4px;
	}

	.completed-badge {
		color: #3fb950;
		font-size: 0.9rem;
		font-weight: 500;
	}

	.grade-badge {
		background: rgba(35, 134, 54, 0.2);
		color: #3fb950;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 0.8rem;
		font-weight: 500;
	}

	.no-challenge {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 16px;
		text-align: center;
	}

	.no-challenge-text {
		color: #8b949e;
		font-size: 0.85rem;
	}

	.badge-preview {
		background: #0d1117;
		border: 1px solid #30363d;
		border-radius: 8px;
		padding: 20px;
		margin: 24px 0;
	}

	.badge-preview h2 {
		margin: 0 0 16px;
		color: #e2e8f0;
		font-size: 0.95rem;
		font-weight: 600;
	}

	.badge-preview h2::before {
		content: '$ ';
		color: #238636;
	}

	.badge-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
		gap: 12px;
	}

	.badge-item {
		background: #161b22;
		border: 1px solid #30363d;
		padding: 12px;
		border-radius: 6px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		text-align: center;
		transition: border-color 0.15s;
	}

	.badge-item:hover {
		border-color: #238636;
	}

	.badge-icon {
		font-size: 1.5rem;
	}

	.badge-name {
		color: #e2e8f0;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.more-link {
		color: #58a6ff;
		font-size: 0.8rem;
		text-decoration: none;
		margin-top: 12px;
		display: inline-block;
	}

	.more-link:hover {
		text-decoration: underline;
	}

	.no-badges {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 16px;
		text-align: center;
	}

	.no-badges-text {
		color: #8b949e;
		font-size: 0.85rem;
	}

	@media (max-width: 640px) {
		.dashboard {
			padding: 0 16px;
			margin: 24px auto;
		}

		.badge-grid {
			grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
		}
	}
</style>
