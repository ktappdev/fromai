<script lang="ts">
	import { goto } from '$app/navigation';
	import { pb } from '$lib/pocketbase.js';

	let email = $state('');
	let password = $state('');
	let flow = $state<'signIn' | 'signUp'>('signIn');
	let loading = $state(false);
	let error = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!email.trim() || !password.trim()) {
			error = 'Email and password are required';
			return;
		}
		loading = true;
		error = '';

		try {
			if (flow === 'signIn') {
				await pb.signIn(email.trim(), password);
			} else {
				await pb.signUp(email.trim(), password);
				await pb.signIn(email.trim(), password);
			}
			goto('/');
		} catch (e: any) {
			error = e.message || 'Authentication failed';
		}
		loading = false;
	}
</script>

<div class="auth-page">
	<div class="auth-card">
		<h2>{flow === 'signIn' ? 'Sign In' : 'Sign Up'}</h2>
		{#if error}
			<div class="error">{error}</div>
		{/if}
		<form onsubmit={handleSubmit}>
			<div class="field">
				<label for="email">Email</label>
				<input id="email" type="email" bind:value={email} placeholder="you@example.com" />
			</div>
			<div class="field">
				<label for="password">Password</label>
				<input id="password" type="password" bind:value={password} placeholder="••••••••" />
			</div>
			<button type="submit" disabled={loading}>
				{loading ? 'Working...' : flow === 'signIn' ? 'Sign In' : 'Sign Up'}
			</button>
		</form>
		<p class="switch">
			{flow === 'signIn' ? "Don't have an account?" : 'Already have an account?'}
			<button class="link" onclick={() => flow = flow === 'signIn' ? 'signUp' : 'signIn'}>
				{flow === 'signIn' ? 'Sign Up' : 'Sign In'}
			</button>
		</p>
	</div>
</div>

<style>
	.auth-page {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		background: #000;
		font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Fira Code', 'JetBrains Mono', monospace;
	}

	.auth-card {
		background: #000;
		border: 1px solid #1a1a1a;
		padding: 32px;
		width: 100%;
		max-width: 420px;
		position: relative;
	}

	.auth-card::before {
		content: '$ login';
		position: absolute;
		top: -10px;
		left: 12px;
		background: #000;
		padding: 0 6px;
		font-size: 0.7rem;
		color: #238636;
	}

	.auth-card h2 {
		margin: 0 0 20px;
		color: #e2e8f0;
		font-size: 0.95rem;
		font-weight: 600;
	}

	.error {
		background: rgba(218, 54, 51, 0.1);
		color: #f85149;
		padding: 8px 12px;
		border-left: 2px solid #da3633;
		margin-bottom: 16px;
		font-size: 0.8rem;
	}

	.field {
		margin-bottom: 16px;
	}

	.field label {
		display: block;
		margin-bottom: 4px;
		font-size: 0.75rem;
		color: #8b949e;
	}

	.field label::before {
		content: '> ';
		color: #238636;
	}

	.field input {
		width: 100%;
		padding: 8px 10px;
		border: 1px solid #1a1a1a;
		background: #0d1117;
		color: #e2e8f0;
		font-family: inherit;
		font-size: 0.85rem;
		box-sizing: border-box;
		outline: none;
		transition: border-color 0.15s;
	}

	.field input:focus {
		border-color: #238636;
	}

	button[type='submit'] {
		width: 100%;
		background: #238636;
		color: #000;
		border: none;
		padding: 8px;
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		margin-top: 4px;
		font-family: inherit;
	}

	button[type='submit']:hover:not(:disabled) {
		background: #2ea043;
	}

	button[type='submit']:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.switch {
		margin-top: 16px;
		text-align: center;
		color: #8b949e;
		font-size: 0.8rem;
	}

	.switch::before {
		content: '$ ';
		color: #238636;
	}

	.link {
		background: none;
		border: none;
		color: #3fb950;
		cursor: pointer;
		padding: 0;
		font-size: 0.8rem;
		text-decoration: none;
		font-family: inherit;
	}

	.link:hover {
		text-decoration: underline;
	}
</style>
