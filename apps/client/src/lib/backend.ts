function getToken(): string | null {
	if (typeof window === 'undefined') return null;
	return localStorage.getItem('token');
}

export function setToken(token: string) {
	localStorage.setItem('token', token);
}

export function clearToken() {
	localStorage.removeItem('token');
}

export function isAuthenticated(): boolean {
	return getToken() !== null;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const headers: Record<string, string> = {
		'Content-Type': 'application/json'
	};

	const token = getToken();
	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const res = await fetch(`/api${path}`, {
		method,
		headers,
		body: body ? JSON.stringify(body) : undefined
	});

	if (res.status === 204) return undefined as T;

	const data = await res.json();
	if (!res.ok) {
		throw new Error(data?.error?.message ?? 'Request failed');
	}
	return data as T;
}

async function upload<T>(method: string, path: string, formData: FormData, opts?: { auth?: boolean }): Promise<T> {
	const headers: Record<string, string> = {};

	if (opts?.auth !== false) {
		const token = getToken();
		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}
	}

	const res = await fetch(`/api${path}`, {
		method,
		headers,
		body: formData
	});

	if (res.status === 204) return undefined as T;

	const data = await res.json();
	if (!res.ok) {
		throw new Error(data?.error?.message ?? 'Request failed');
	}
	return data as T;
}

export const api = {
	auth: {
		register: (email: string, password: string) =>
			request<{ user_id: string; token: string }>('POST', '/auth/register', { email, password }),
		login: (email: string, password: string) =>
			request<{ user_id: string; token: string }>('POST', '/auth/login', { email, password }),
		me: () => request<UserProfile>('GET', '/auth/me'),
		updateProfile: (data: { name: string; email: string; reminder_interval_days?: number }) =>
			request<UserProfile>('PUT', '/auth/me', data),
		changePassword: (currentPassword: string, newPassword: string) =>
			request<{ status: string }>('PUT', '/auth/password', {
				current_password: currentPassword,
				new_password: newPassword
			})
	},
	documents: {
		list: () => request<Document[]>('GET', '/documents'),
		get: (id: number) => request<Document>('GET', `/documents/${id}`),
		create: (name: string, file: File) => {
			const formData = new FormData();
			formData.append('name', name);
			formData.append('file', file);
			return upload<Document>('POST', '/documents', formData);
		},
		delete: (id: number) => request<void>('DELETE', `/documents/${id}`),
		update: (id: number, data: { name?: string; file_name?: string; sequential?: boolean }) =>
			request<Document>('PUT', `/documents/${id}`, data),
		send: (id: number) => request<Document>('POST', `/documents/${id}/send`),
		stats: () => request<DocumentStats>('GET', '/documents/stats'),
		certificateUrl: (id: number) => `/api/documents/${id}/certificate`,
		fileUrl: (id: number) => `/api/documents/${id}/file`,
		auditTrailUrl: (id: number) => `/api/documents/${id}/audit-trail`
	},
	signers: {
		list: (documentId: number) => request<Signer[]>('GET', `/documents/${documentId}/signers`),
		add: (documentId: number, name: string, email: string) =>
			request<Signer>('POST', `/documents/${documentId}/signers`, { name, email }),
		remove: (signerId: number) =>
			request<void>('DELETE', `/signers/${signerId}`),
		remind: (signerId: number) =>
			request<{ status: string; reminded_at: string }>('POST', `/signers/${signerId}/remind`)
	},
	fields: {
		list: (documentId: number) => request<Field[]>('GET', `/documents/${documentId}/fields`),
		create: (documentId: number, data: CreateFieldRequest) =>
			request<Field>('POST', `/documents/${documentId}/fields`, data),
		update: (documentId: number, fieldId: number, data: UpdateFieldRequest) =>
			request<Field>('PUT', `/documents/${documentId}/fields/${fieldId}`, data),
		delete: (documentId: number, fieldId: number) =>
			request<void>('DELETE', `/documents/${documentId}/fields/${fieldId}`)
	},
	webhooks: {
		list: () => request<Webhook[]>('GET', '/webhooks'),
		create: (data: { url: string; secret: string }) =>
			request<Webhook>('POST', '/webhooks', data),
		update: (id: number, data: { url: string; secret: string; enabled: boolean }) =>
			request<Webhook>('PUT', `/webhooks/${id}`, data),
		delete: (id: number) => request<void>('DELETE', `/webhooks/${id}`)
	},
	smtp: {
		get: () => request<SmtpConfig>('GET', '/smtp'),
		save: (data: { host: string; port: number; username: string; password: string; from_email: string; from_name: string }) =>
			request<SmtpConfig>('PUT', '/smtp', data),
		delete: () => request<void>('DELETE', '/smtp'),
		test: (to: string) => request<{ status: string }>('POST', '/smtp/test', { to })
	},
	signing: {
		fileUrl: (token: string) => `/api/sign/${token}/file`,
		get: (token: string) => request<SigningPayload>('GET', `/sign/${token}`),
		sign: (token: string, fields: Record<string, string>) =>
			request<{ status: string }>('POST', `/sign/${token}`, {
				fields: Object.entries(fields).map(([id, value]) => ({ field_id: Number(id), value }))
			}),
		decline: (token: string, reason?: string) =>
			request<{ status: string }>('POST', `/sign/${token}/decline`, { reason })
	},
	verify: {
		check: (file: File) => {
			const formData = new FormData();
			formData.append('file', file);
			return upload<VerifyResponse>('POST', '/verify', formData, { auth: false });
		},
		byHash: (hash: string) =>
			fetch(`/api/verify/${hash}`)
				.then(async (res) => {
					const data = await res.json();
					if (!res.ok) throw new Error(data?.error?.message ?? 'Request failed');
					return data as VerifyResponse;
				})
	}
};

export interface SmtpConfig {
	host: string;
	port: number;
	username: string;
	from_email: string;
	from_name: string;
	updated_at: string;
}

export interface UserProfile {
	id: string;
	email: string;
	name: string;
	reminder_interval_days: number;
	created_at: string;
}

export interface Document {
	id: number;
	name: string;
	status: 'draft' | 'pending' | 'completed' | 'declined';
	file_name: string;
	owner_id: number;
	sequential: boolean;
	signer_count?: number;
	created_at: string;
	updated_at: string;
}

export interface DocumentStats {
	total: number;
	pending: number;
	completed: number;
}

export interface Signer {
	id: number;
	document_id: number;
	name: string;
	email: string;
	role: string;
	status: 'pending' | 'signed' | 'declined';
	token: string;
	order_num: number;
	signed_at: string | null;
	last_reminded_at: string | null;
	ip_address?: string;
	user_agent?: string;
}

export interface Field {
	id: number;
	document_id: number;
	signer_id: number;
	field_type: 'signature' | 'text' | 'date' | 'checkbox';
	page: number;
	x: number;
	y: number;
	width: number;
	height: number;
	required: boolean;
	label: string;
	value: string | null;
}

export interface CreateFieldRequest {
	signer_id: number;
	field_type: string;
	page: number;
	x: number;
	y: number;
	width: number;
	height: number;
	required: boolean;
	label: string;
}

export interface UpdateFieldRequest {
	field_type: string;
	page: number;
	x: number;
	y: number;
	width: number;
	height: number;
	required: boolean;
	label: string;
}

export interface Webhook {
	id: number;
	url: string;
	enabled: boolean;
	last_sent_at: string | null;
	created_at: string;
	updated_at: string;
}

export interface CompletedField {
	id: number;
	signer_name: string;
	field_type: string;
	label: string;
	page: number;
	x: number;
	y: number;
	width: number;
	height: number;
	value: string;
}

export interface SigningPayload {
	document: {
		name: string;
		status: string;
	};
	signer: {
		name: string;
		email: string;
		status: string;
	};
	fields: Field[];
	completed_fields: CompletedField[];
}

export interface VerifyDocument {
	name: string;
	file_name: string;
	status: 'draft' | 'pending' | 'completed' | 'declined';
	created_at: string;
	completed_at?: string;
}

export interface VerifySigner {
	name: string;
	email: string;
	status: 'pending' | 'signed' | 'declined';
	signed_at?: string | null;
}

export interface VerifyResponse {
	match: boolean;
	hash: string;
	variant?: 'original' | 'signed';
	document?: VerifyDocument;
	signers?: VerifySigner[];
}
