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
			error = 'Failed to load badges';
		} finally {
			loading = false;
		}
	});

	const earnedIds = $derived<Set<string>>(new Set(userStats?.badges ?? []));

	const badgeProgress = $derived(() => {
		const badgeCatalog = userStats?.badge_catalog ?? {};

		return BADGE_CATALOG.map(badge => {
			const earned = earnedIds.has(badge.id);
			let progress: { current: number; target: number } | null = null;

			if (!earned && badgeCatalog[badge.id]) {
				const catalogEntry = badgeCatalog[badge.id];
				progress = {
					current: catalogEntry.progress ?? 0,
					target: catalogEntry.target ?? 1
				};
			}

			return { ...badge, earned, progress };
		});
	});

	const earnedCount = $derived(earnedIds.size);
</script>

<div class="badges-page">
	<h1>$ Badges</h1>
	<p class="subtitle">{earnedCount} of {BADGE_CATALOG.length} earned</p>

	{#if loading}
		<p class="loading">Loading...</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else}
		<div class="badge-grid">
			{#each badgeProgress() as badge}
				<div class="badge-card" class:earned={badge.earned}>
					<div class="badge-header">
						<span class="badge-icon">{badge.icon}</span>
						<h3 class="badge-name">{badge.name}</h3>
					</div>
					<p class="badge-desc">{badge.description}</p>
					<div class="badge-footer">
						{#if badge.earned}
							<span class="earned-label">✓ Earned!</span>
						{:else if badge.progress}
							<div class="progress-bar-wrapper">
								<div class="progress-bar">
									<div
										class="progress-fill"
										style="width: {Math.min(100, Math.round((badge.progress.current / badge.progress.target) * 100))}%"
									></div>
								</div>
								<span class="progress-text">{badge.progress.current}/{badge.progress.target}</span>
							</div>
						{:else}
							<span class="locked-label">🔒 Locked</span>
						{/if}
					</div>
				</div>
			{/each}
		</div>

		<div class="back-link">
			<a href="/">← back to dashboard</a>
		</div>
	{/if}
</div>

<style>
	.badges-page {
		max-width: 720px;
		margin: 24px 32px;
	}

	h1 {
		color: #238636;
		font-size: 1.1rem;
		margin: 0;
	}

	.subtitle {
		font-size: 0.75rem;
		color: #8b949e;
		margin: 4px 0 24px;
	}

	.loading, .error {
		font-size: 0.8rem;
		margin: 0;
	}

	.loading {
		color: #8b949e;
	}

	.error {
		color: #f85149;
	}

	.badge-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 12px;
	}

	.badge-card {
		background: #161b22;
		border: 1px solid #30363d;
		border-radius: 6px;
		padding: 14px;
		display: flex;
		flex-direction: column;
	}

	.badge-card.earned {
		background: #0d1117;
		border-color: #238636;
	}

	.badge-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 6px;
	}

	.badge-icon {
		font-size: 1.3rem;
		line-height: 1;
	}

	.badge-card:not(.earned) .badge-icon {
		opacity: 0.35;
		filter: grayscale(0.8);
	}

	.badge-name {
		color: #e2e8f0;
		font-size: 0.8rem;
		font-weight: 600;
		margin: 0;
	}

	.badge-card:not(.earned) .badge-name {
		color: #8b949e;
	}

	.badge-desc {
		font-size: 0.7rem;
		color: #8b949e;
		margin: 0 0 10px;
		line-height: 1.4;
	}

	.badge-footer {
		margin-top: auto;
	}

	.earned-label {
		font-size: 0.7rem;
		color: #238636;
		font-weight: 600;
	}

	.locked-label {
		font-size: 0.7rem;
		color: #8b949e;
	}

	.progress-bar-wrapper {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.progress-bar {
		flex: 1;
		height: 6px;
		background: #21262d;
		border-radius: 3px;
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: #238636;
		border-radius: 3px;
		transition: width 0.3s;
	}

	.progress-text {
		font-size: 0.65rem;
		color: #8b949e;
		flex-shrink: 0;
		min-width: 36px;
		text-align: right;
	}

	.back-link {
		margin-top: 24px;
	}

	.back-link a {
		font-size: 0.75rem;
		color: #58a6ff;
		text-decoration: none;
	}

	.back-link a:hover {
		text-decoration: underline;
	}
</style>
