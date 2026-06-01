import PocketBase from 'pocketbase';

const POCKETBASE_URL = import.meta.env.VITE_POCKETBASE_URL || 'http://127.0.0.1:8090';

const client = new PocketBase(POCKETBASE_URL);

// Migrate old raw token to SDK auth store
if (typeof window !== 'undefined') {
	const oldToken = localStorage.getItem('pb_token');
	if (oldToken) {
		client.authStore.save(oldToken, null);
		localStorage.removeItem('pb_token');
	}
}

function getBaseURL(): string {
	return POCKETBASE_URL;
}

export class PocketBaseClient {
	private pb: PocketBase;

	constructor(pbInstance: PocketBase) {
		this.pb = pbInstance;
	}

	async signIn(email: string, password: string) {
		const data = await this.pb.collection('users').authWithPassword(email, password);
		return data;
	}

	async signUp(email: string, password: string, name?: string) {
		return this.pb.collection('users').create({
			email,
			password,
			passwordConfirm: password,
			name,
		});
	}

	async signOut() {
		this.pb.authStore.clear();
	}

	async getMe() {
		try {
			const data = await this.pb.collection('users').authRefresh();
			return data.record;
		} catch {
			return null;
		}
	}

	private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
		const res = await fetch(`${getBaseURL()}${path}`, {
			...options,
			headers: {
				'Authorization': this.getAuthToken(),
				'Content-Type': 'application/json',
				...(options.headers || {}),
			},
		});
		if (!res.ok) {
			const err = await res.text();
			throw new Error(`Request failed: ${res.status} ${err}`);
		}
		return res.json();
	}

	async createTask(data: {
		title: string;
		description: string;
		starter_code: string;
		language: string;
	}): Promise<any> {
		return this.request('/api/tasks', {
			method: 'POST',
			body: JSON.stringify(data),
		});
	}

	async getTask(id: string): Promise<any> {
		return this.request(`/api/tasks/${id}`);
	}

	async listTasks() {
		return this.request<any[]>('/api/tasks');
	}

	async updateTaskCode(id: string, code: string): Promise<any> {
		return this.request(`/api/tasks/${id}`, {
			method: 'PATCH',
			body: JSON.stringify({ code }),
		});
	}

	async submitTask(id: string): Promise<any> {
		return this.request(`/api/tasks/${id}/submit`, {
			method: 'POST',
		});
	}

	async gradeTask(id: string, grade: string, feedback?: string): Promise<any> {
		return this.request(`/api/tasks/${id}/grade`, {
			method: 'POST',
			body: JSON.stringify({ grade, feedback }),
		});
	}

	async archiveTask(id: string): Promise<any> {
		return this.request(`/api/tasks/${id}/archive`, {
			method: 'POST',
		});
	}

	async deleteTask(id: string): Promise<any> {
		return this.request(`/api/tasks/${id}/delete`, {
			method: 'POST',
		});
	}

	async subscribeToTasks(callback: (e: any) => void) {
		return this.pb.collection('tasks').subscribe('*', callback);
	}

	async unsubscribeFromTasks() {
		return this.pb.collection('tasks').unsubscribe();
	}

	async subscribeToTask(id: string, callback: (e: any) => void) {
		return this.pb.collection('tasks').subscribe(id, callback);
	}

	async unsubscribeFromTask(id: string) {
		return this.pb.collection('tasks').unsubscribe(id);
	}

	getAuthToken(): string {
		return this.pb.authStore.token;
	}

	async getAPIKey(): Promise<string | null> {
		try {
			const res = await fetch(`${getBaseURL()}/api/me/api-key`, {
				headers: {
					'Authorization': this.getAuthToken(),
					'Content-Type': 'application/json',
				},
			});
			if (res.ok) {
				const data = await res.json();
				return data.api_key;
			}
			return null;
		} catch {
			return null;
		}
	}

	async regenerateAPIKey(): Promise<string | null> {
		try {
			const res = await fetch(`${getBaseURL()}/api/me/api-key`, {
				method: 'POST',
				headers: {
					'Authorization': this.getAuthToken(),
					'Content-Type': 'application/json',
				},
			});
			if (res.ok) {
				const data = await res.json();
				return data.api_key;
			}
			return null;
		} catch {
			return null;
		}
	}

	async getMyStats() {
		return this.request('/api/me/stats');
	}

	async getTodayChallenge() {
		try {
			return await this.request('/api/challenges/today');
		} catch {
			return null;
		}
	}

	async listChallenges() {
		return this.request('/api/challenges');
	}

	async startChallenge(id: string) {
		return this.request(`/api/challenges/${id}/start`, {
			method: 'POST',
		});
	}

	subscribeToUserStats(statsId: string, callback: (data: any) => void) {
		return this.pb.collection('user_stats').subscribe(statsId, (e) => {
			callback(e.record);
		});
	}
}

export const pb = new PocketBaseClient(client);
