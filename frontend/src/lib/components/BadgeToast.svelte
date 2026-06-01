<script lang="ts">
	import type { BadgeDef } from '$lib/gamification.js';

	let { badge, visible, onDismiss }: { badge: BadgeDef; visible: boolean; onDismiss?: () => void } = $props();

	let show = $state(false);

	$effect(() => {
		show = visible;
		if (show) {
			const timer = setTimeout(() => {
				show = false;
				onDismiss?.();
			}, 4000);
			return () => clearTimeout(timer);
		}
	});
</script>

{#if show}
	<div class="toast-container">
		<div class="toast">
			<div class="toast-icon">{badge.icon}</div>
			<div class="toast-content">
				<div class="toast-title">Badge Earned!</div>
				<div class="toast-badge-name">{badge.name}</div>
				<div class="toast-description">{badge.description}</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.toast-container {
		position: fixed;
		bottom: 24px;
		right: 24px;
		z-index: 9999;
		animation: slideIn 0.3s ease-out;
	}

	@keyframes slideIn {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.toast {
		background: #0d1117;
		border: 1px solid #238636;
		border-radius: 4px;
		padding: 14px 16px;
		display: flex;
		align-items: flex-start;
		gap: 12px;
		min-width: 280px;
		max-width: 340px;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
		animation: fadeIn 0.3s ease-out;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	.toast-icon {
		font-size: 1.5rem;
		flex-shrink: 0;
	}

	.toast-content {
		flex: 1;
		min-width: 0;
	}

	.toast-title {
		font-size: 0.7rem;
		color: #238636;
		font-weight: 600;
		margin-bottom: 2px;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.toast-title::before {
		content: '$ ';
	}

	.toast-badge-name {
		font-size: 0.85rem;
		color: #e2e8f0;
		font-weight: 600;
		margin-bottom: 2px;
	}

	.toast-description {
		font-size: 0.7rem;
		color: #8b949e;
		line-height: 1.4;
	}
</style>