<script lang="ts">
	import { onMount } from 'svelte';
	import { pb } from '$lib/pocketbase.js';

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

	onMount(() => {
		pb.getMe()
			.then((u) => { user = u; })
			.catch(() => { user = null; })
			.finally(() => { loading = false; });

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
			<p class="sub">Pick a task from the sidebar to start coding.</p>
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
					<span class="title">coding gym</span>
					<span class="cursor">_</span>
				</div>
				<p class="tagline">{headline}<span class="caret"></span></p>
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
		margin: 0 0 32px;
		min-height: 1.5em;
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
</style>
