<script lang="ts">
	let { open, onClose, content }: {
		open: boolean;
		onClose: () => void;
		content: string;
	} = $props();

	const labels = ['File:', 'Purpose:', 'Problem:', 'Property:', 'Constraints:', 'Success:', 'Available tools:'];

	function formatContent(text: string): string {
		let formatted = text;
		for (const label of labels) {
			formatted = formatted.replace(label, `<span class="label">${label}</span><br><br>`);
		}
		return formatted;
	}
</script>

{#if open}
	<div class="modal-overlay" role="dialog" aria-modal="true" tabindex="-1" onclick={onClose} onkeydown={(e) => e.key === 'Escape' && onClose()}>
		<div class="modal" onmousedown={(e) => e.stopPropagation()} role="document">
			<div class="modal-header">
				<span class="modal-title">task context</span>
				<button class="modal-close" onclick={onClose} aria-label="Close modal">×</button>
			</div>
			<div class="modal-content">
				<pre>{@html formatContent(content)}</pre>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.8);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal {
		background: #0d1117;
		border: 1px solid #1a1a1a;
		max-width: 700px;
		width: 90%;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 16px;
		border-bottom: 1px solid #1a1a1a;
		background: #000;
	}

	.modal-title {
		color: #e2e8f0;
		font-size: 0.8rem;
		font-weight: 600;
	}

	.modal-title::before {
		content: '$ cat ';
		color: #238636;
		font-weight: 400;
	}

	.modal-close {
		background: transparent;
		color: #8b949e;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		padding: 0;
		line-height: 1;
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: inherit;
	}

	.modal-close:hover {
		color: #e2e8f0;
	}

	.modal-content {
		padding: 16px;
		overflow-y: auto;
	}

	.modal-content pre {
		margin: 0;
		color: #8b949e;
		font-size: 0.8rem;
		line-height: 1.6;
		white-space: pre-wrap;
		word-wrap: break-word;
		font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Fira Code', 'JetBrains Mono', monospace;
	}

	.label {
		color: #238636;
		font-weight: 600;
	}
</style>