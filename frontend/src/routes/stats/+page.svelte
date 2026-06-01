<script lang="ts">
	import { pb } from '$lib/pocketbase.js';
	import { onMount } from 'svelte';

	let tasks = $state<any[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		if (!pb.getAuthToken()) {
			window.location.href = '/login';
			return;
		}
		try {
			tasks = await pb.listTasks();
		} catch (e) {
			error = 'Failed to load tasks';
		} finally {
			loading = false;
		}
	});

	const pending = $derived(tasks.filter(t => t.status !== 'completed').length);
	const completed = $derived(tasks.filter(t => t.status === 'completed').length);
	const total = $derived(tasks.length);
	const completionRate = $derived(total > 0 ? Math.round((completed / total) * 100) : 0);

	const gradeValue = (g: string): number => {
		const map: Record<string, number> = { A: 4, B: 3, C: 2, D: 1, F: 0 };
		return map[g] ?? 0;
	};

	const completedTasks = $derived(tasks.filter(t => t.status === 'completed' && t.grade));
	const avgGrade = $derived(
		completedTasks.length > 0
			? completedTasks.reduce((sum, t) => sum + gradeValue(t.grade), 0) / completedTasks.length
			: 0
	);

	const gradeDistribution = $derived(() => {
		const grades = ['A', 'B', 'C', 'D', 'F'];
		const counts = grades.map(g => ({
			grade: g,
			count: completedTasks.filter(t => t.grade === g).length
		}));
		const max = Math.max(...counts.map(c => c.count), 1);
		return counts.map(c => ({
			...c,
			bar: '█'.repeat(Math.round((c.count / max) * 20)).padEnd(20, '░')
		}));
	});

	const languageBreakdown = $derived(() => {
		const counts: Record<string, number> = {};
		tasks.forEach(t => {
			const lang = t.language || 'unknown';
			counts[lang] = (counts[lang] || 0) + 1;
		});
		return Object.entries(counts)
			.map(([lang, count]) => ({ lang, count }))
			.sort((a, b) => b.count - a.count);
	});

	const recentTasks = $derived(() => {
		return [...completedTasks]
			.sort((a, b) => new Date(b.updated).getTime() - new Date(a.updated).getTime())
			.slice(0, 5);
	});

	const gradeColor = (g: string): string => {
		const colors: Record<string, string> = {
			A: '#3fb950',
			B: '#58a6ff',
			C: '#d29922',
			D: '#f0883e',
			F: '#f85149'
		};
		return colors[g] ?? '#8b949e';
	};
</script>

<div class="stats">
	<h1>$ Stats</h1>

	{#if loading}
		<p class="loading">Loading...</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else}
		<!-- Overview -->
		<section class="section">
			<h2>$ Overview</h2>
			{#if total === 0}
				<p class="empty">No tasks yet</p>
			{:else}
				<div class="stat-row">
					<span class="label">pending</span>
					<span class="value">{pending}</span>
				</div>
				<div class="stat-row">
					<span class="label">completed</span>
					<span class="value">{completed}</span>
				</div>
				<div class="stat-row">
					<span class="label">total</span>
					<span class="value">{total}</span>
				</div>
				<div class="stat-row">
					<span class="label">completion</span>
					<span class="value">{completionRate}%</span>
				</div>
			{/if}
		</section>

		<!-- Grades -->
		<section class="section">
			<h2>$ Grades</h2>
			{#if completedTasks.length === 0}
				<p class="empty">No graded tasks yet</p>
			{:else}
				<div class="stat-row">
					<span class="label">average</span>
					<span class="value">{avgGrade.toFixed(2)}</span>
				</div>
				<div class="grade-bars">
					{#each gradeDistribution() as { grade, count, bar }}
						<div class="grade-row">
							<span class="grade-label" style="color: {gradeColor(grade)}">{grade}</span>
							<span class="bar-text">{bar}</span>
							<span class="grade-count">{count}</span>
						</div>
					{/each}
				</div>
			{/if}
		</section>

		<!-- Languages -->
		<section class="section">
			<h2>$ Languages</h2>
			{#if total === 0}
				<p class="empty">No tasks yet</p>
			{:else}
				{#each languageBreakdown() as { lang, count }}
					<div class="stat-row">
						<span class="label">{lang}</span>
						<span class="value">{count}</span>
					</div>
				{/each}
			{/if}
		</section>

		<!-- Recent -->
		<section class="section">
			<h2>$ Recent</h2>
			{#if recentTasks.length === 0}
				<p class="empty">No completed tasks yet</p>
			{:else}
				{#each recentTasks() as task}
					<div class="task-row">
						<span class="task-title">{task.title}</span>
						<span class="task-grade" style="color: {gradeColor(task.grade)}">{task.grade}</span>
					</div>
				{/each}
			{/if}
		</section>
	{/if}
</div>

<style>
	.stats {
		max-width: 560px;
		margin: 24px 32px;
	}

	h1 {
		color: #238636;
		font-size: 1.1rem;
		margin: 0 0 24px;
	}

	h2 {
		color: #238636;
		font-size: 0.85rem;
		margin: 0 0 12px;
	}

	.section {
		padding-bottom: 16px;
		margin-bottom: 16px;
		border-bottom: 1px solid #1a1a1a;
	}

	.section:last-child {
		border-bottom: none;
		margin-bottom: 0;
	}

	.loading, .empty, .error {
		font-size: 0.8rem;
		color: #8b949e;
		margin: 0;
	}

	.error {
		color: #f85149;
	}

	.stat-row {
		display: flex;
		justify-content: space-between;
		margin-bottom: 6px;
	}

	.label {
		font-size: 0.75rem;
		color: #8b949e;
	}

	.value {
		font-size: 0.8rem;
		color: #e2e8f0;
	}

	.grade-bars {
		margin-top: 8px;
	}

	.grade-row {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 4px;
	}

	.grade-label {
		font-size: 0.75rem;
		font-weight: 600;
		width: 12px;
	}

	.bar-text {
		font-size: 0.75rem;
		color: #58a6ff;
		flex: 1;
	}

	.grade-count {
		font-size: 0.75rem;
		color: #e2e8f0;
		width: 20px;
		text-align: right;
	}

	.task-row {
		display: flex;
		justify-content: space-between;
		margin-bottom: 6px;
	}

	.task-title {
		font-size: 0.75rem;
		color: #e2e8f0;
		flex: 1;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.task-grade {
		font-size: 0.75rem;
		font-weight: 600;
		flex-shrink: 0;
		margin-left: 10px;
	}
</style>