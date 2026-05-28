import PocketBase from 'pocketbase';

const POCKETBASE_URL = 'http://127.0.0.1:8090';

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

	async createTask(data: {
		title: string;
		description: string;
		starter_code: string;
		language: string;
	}) {
		return this.pb.collection('tasks').create(data);
	}

	async getTask(id: string) {
		return this.pb.collection('tasks').getOne(id);
	}

	async listTasks() {
		return this.pb.collection('tasks').getFullList({
			sort: '-created_at',
		});
	}

	async updateTaskCode(id: string, code: string) {
		return this.pb.collection('tasks').update(id, { code });
	}

	async submitTask(id: string) {
		return this.pb.collection('tasks').update(id, { status: 'completed' });
	}

	async gradeTask(id: string, grade: string, feedback?: string) {
		return this.pb.collection('tasks').update(id, { grade, feedback });
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
}

export const pb = new PocketBaseClient(client);
