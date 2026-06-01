<script lang="ts">
	import { pb } from '$lib/pocketbase.js';
	import { BADGE_CATALOG } from '$lib/gamification.js';
	import { onMount } from 'svelte';

	let tasks = $state<any[]>([]);
	let userStats = $state<any>(null);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		if (!pb.getAuthToken()) {
			window.location.href = '/login';
			return;
		}
		try {
			const [tasksData, statsData] = await Promise.all([
				pb.listTasks(),
				pb.getMyStats()
			]);
			tasks = tasksData;
			userStats = statsData;
		} catch (e) {
			error = 'Failed to load data';
		} finally {
			loading = false;
		}
	});

	const currentStreak = $derived(userStats?.current_streak ?? 0);
	const bestStreak = $derived(userStats?.best_streak ?? 0);
	const earnedBadgeCount = $derived((userStats?.badges ?? []).length);

	const heatmapWeeks = $derived(() => {
		const dateCounts: Record<string, number> = {};
		tasks.filter(t => t.status === 'completed' && t.completed_at).forEach(t => {
			const d = new Date(t.completed_at).toISOString().slice(0, 10);
			dateCounts[d] = (dateCounts[d] || 0) + 1;
		});

		const today = new Date();
		today.setHours(0, 0, 0, 0);

		const startDate = new Date(today);
		startDate.setDate(startDate.getDate() - 89);

		const dayOfWeek = startDate.getDay();
		const mondayOffset = dayOfWeek === 0 ? -6 : 1 - dayOfWeek;
		startDate.setDate(startDate.getDate() + mondayOffset);

		const weeks = [];
		const cursor = new Date(startDate);
		const end = new Date(today);

		while (cursor <= end) {
			const week = [];
			for (let i = 0; i < 7; i++) {
				const key = cursor.toISOString().slice(0, 10);
				const count = dateCounts[key] || 0;
				const level = count === 0 ? 0 : count === 1 ? 1 : count <= 3 ? 2 : 3;
				const label = count === 0 ? '' : `${count} completion${count !== 1 ? 's' : ''}`;
				week.push({ date: key, count, level, label });
				cursor.setDate(cursor.getDate() + 1);
			}
			weeks.push(week);
		}

		return weeks;
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
		<!-- Streak -->
		<section class="section">
			<h2>$ Streak</h2>
			<div class="streak-row">
				<span class="streak-fire">🔥</span>
				<span class="streak-value">{currentStreak} day{currentStreak !== 1 ? 's' : ''}</span>
			</div>
			<div class="stat-row">
				<span class="label">best streak</span>
				<span class="value">{bestStreak} day{bestStreak !== 1 ? 's' : ''}</span>
			</div>
		</section>

		<!-- Activity Heatmap -->
		<section class="section">
			<h2>$ Activity</h2>
			{#if tasks.filter(t => t.status === 'completed').length === 0}
				<p class="empty">No completed tasks yet</p>
			{:else}
				<div class="heatmap">
					<div class="heatmap-labels">
						<span>Mon</span>
						<span></span>
						<span>Wed</span>
						<span></span>
						<span>Fri</span>
						<span></span>
						<span>Sun</span>
					</div>
					<div class="heatmap-grid">
						{#each heatmapWeeks() as week}
							<div class="heatmap-week">
								{#each week as day}
									<div
										class="heatmap-cell level-{day.level}"
										title="{day.date}: {day.label || 'no activity'}"
									></div>
								{/each}
							</div>
						{/each}
					</div>
				</div>
				<div class="heatmap-legend">
					<span class="legend-label">Less</span>
					<span class="legend-cell level-0"></span>
					<span class="legend-cell level-1"></span>
					<span class="legend-cell level-2"></span>
					<span class="legend-cell level-3"></span>
					<span class="legend-label">More</span>
				</div>
			{/if}
		</section>

		<!-- Badges -->
		<section class="section">
			<h2>$ Badges</h2>
			<div class="stat-row">
				<span class="label">earned</span>
				<span class="value">
					<a href="/badges" class="badge-link">{earnedBadgeCount} / {BADGE_CATALOG.length}</a>
				</span>
			</div>
		</section>

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

	.streak-row {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 6px;
	}

	.streak-fire {
		font-size: 1.2rem;
	}

	.streak-value {
		font-size: 0.85rem;
		color: #e2e8f0;
		font-weight: 600;
	}

	.heatmap {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		padding-bottom: 4px;
	}

	.heatmap-labels {
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
		gap: 2px;
		padding-top: 0;
	}

	.heatmap-labels span {
		font-size: 0.55rem;
		color: #8b949e;
		line-height: 10px;
		height: 10px;
		display: flex;
		align-items: center;
	}

	.heatmap-grid {
		display: flex;
		gap: 2px;
	}

	.heatmap-week {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.heatmap-cell {
		width: 10px;
		height: 10px;
		border-radius: 1px;
	}

	.heatmap-cell.level-0 { background: #161b22; }
	.heatmap-cell.level-1 { background: #0e4429; }
	.heatmap-cell.level-2 { background: #006d32; }
	.heatmap-cell.level-3 { background: #26a641; }

	.heatmap-legend {
		display: flex;
		align-items: center;
		gap: 4px;
		margin-top: 8px;
		justify-content: flex-end;
	}

	.legend-label {
		font-size: 0.6rem;
		color: #8b949e;
	}

	.legend-cell {
		width: 10px;
		height: 10px;
		border-radius: 1px;
		display: inline-block;
	}

	.legend-cell.level-0 { background: #161b22; border: 1px solid #21262d; }
	.legend-cell.level-1 { background: #0e4429; }
	.legend-cell.level-2 { background: #006d32; }
	.legend-cell.level-3 { background: #26a641; }

	.badge-link {
		color: #58a6ff;
		text-decoration: none;
		font-size: 0.8rem;
	}

	.badge-link:hover {
		text-decoration: underline;
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