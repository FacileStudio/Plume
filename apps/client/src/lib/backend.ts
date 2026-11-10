import { ApiError, createApi, redirectToLogin, type CallOptions } from '@facile/client';

export { ApiError };

const LEGACY_TOKEN_KEYS = ['plume.token', 'token'];

const apiFetch = createApi({ onUnauthorized: redirectToLogin });

/**
 * forgetLegacyToken deletes the bearer token older builds kept in localStorage.
 * The browser signs in with porte's HttpOnly cookie only, so the token is
 * never read again and should not outlive this version.
 */
export function forgetLegacyToken() {
	try {
		for (const key of LEGACY_TOKEN_KEYS) localStorage.removeItem(key);
	} catch {
		return;
	}
}

/**
 * currentUser asks the API who the caller is, and is the only honest answer to
 * "am I signed in". It probes quietly, so a public page that asks is not sent
 * to the login page.
 *
 * Returns null when the server says the caller is not authenticated, and
 * rethrows anything else so a transient failure is not mistaken for a logout.
 */
export async function currentUser(): Promise<UserProfile | null> {
	try {
		return await request<UserProfile>('GET', '/auth/me', undefined, { quiet: true });
	} catch (err) {
		if (err instanceof ApiError && (err.status === 401 || err.status === 403)) return null;
		throw err;
	}
}

/** logout ends the cookie session, which only the server can revoke. */
export async function logout() {
	try {
		await request<void>('POST', '/auth/logout', undefined, { quiet: true });
	} catch {}
	forgetLegacyToken();
}

function request<T>(method: string, path: string, body?: unknown, call?: CallOptions): Promise<T> {
	return apiFetch<T>(path, { method, body: body ? JSON.stringify(body) : undefined }, call);
}

/**
 * upload sends a multipart body. A refusal without the API's envelope, such as
 * the proxy's 413 for a file over the limit, keeps the readable message the
 * upload forms have always shown.
 */
async function upload<T>(method: string, path: string, formData: FormData, call?: CallOptions): Promise<T> {
	try {
		return await apiFetch<T>(path, { method, body: formData }, call);
	} catch (err) {
		if (!(err instanceof ApiError) || err.code) throw err;
		if (err.status === 413) throw new Error('File is too large. Maximum upload size is 50 MB.');
		throw new Error('Upload failed — please try again');
	}
}

export interface ApiToken {
	id: number;
	label: string;
	created_at: string;
	last_used_at: string;
}

export interface ApiTokenCreated {
	id: number;
	label: string;
	token: string;
}

export const api = {
	auth: {
		me: () => request<UserProfile>('GET', '/auth/me'),
		updateProfile: (data: { name: string; email: string; reminder_interval_days?: number }) =>
			request<UserProfile>('PUT', '/auth/me', data),
		syncProfile: () => request<{ status: string }>('POST', '/auth/sync-profile'),
		apiTokens: () => request<ApiToken[]>('GET', '/auth/api-tokens'),
		createApiToken: (label: string) =>
			request<ApiTokenCreated>('POST', '/auth/api-token', { label }),
		revokeApiToken: (id: number) => request<{ deleted: boolean }>('DELETE', `/auth/api-token/${id}`)
	},
	documents: {
		list: (spaceId?: number | null) =>
			request<Document[]>('GET', spaceId ? `/documents?space_id=${spaceId}` : '/documents'),
		get: (id: number) => request<Document>('GET', `/documents/${id}`),
		create: (name: string, file: File, spaceId?: number | null) => {
			const formData = new FormData();
			formData.append('name', name);
			formData.append('file', file);
			if (spaceId) formData.append('space_id', String(spaceId));
			return upload<Document>('POST', '/documents', formData);
		},
		delete: (id: number) => request<void>('DELETE', `/documents/${id}`),
		update: (id: number, data: { name?: string; file_name?: string; sequential?: boolean; client_id?: number | null }) =>
			request<Document>('PUT', `/documents/${id}`, data),
		send: (id: number) => request<Document>('POST', `/documents/${id}/send`),
		stats: (spaceId?: number | null) =>
			request<DocumentStats>('GET', spaceId ? `/documents/stats?space_id=${spaceId}` : '/documents/stats'),
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
		delete: (id: number) => request<void>('DELETE', `/webhooks/${id}`),
		test: (id: number) =>
			request<{ status: string }>('POST', `/webhooks/${id}/test`)
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
		status: (token: string) => request<SigningStatus>('GET', `/sign/${token}/status`),
		sign: (token: string, fields: Record<string, string>) =>
			request<{ status: string }>('POST', `/sign/${token}`, {
				fields: Object.entries(fields).map(([id, value]) => ({ field_id: Number(id), value }))
			}),
		decline: (token: string, reason?: string) =>
			request<{ status: string }>('POST', `/sign/${token}/decline`, { reason })
	},
	spaces: {
		list: () => request<Space[]>('GET', '/spaces'),
		get: (id: number) => request<Space>('GET', `/spaces/${id}`),
		create: (data: { name: string; description?: string }) =>
			request<Space>('POST', '/spaces', data),
		update: (id: number, data: { name: string; description?: string }) =>
			request<Space>('PUT', `/spaces/${id}`, data),
		delete: (id: number) => request<void>('DELETE', `/spaces/${id}`),
		leave: (id: number) => request<{ status: string }>('POST', `/spaces/${id}/leave`),
		members: {
			list: (spaceId: number) => request<SpaceMember[]>('GET', `/spaces/${spaceId}/members`),
			add: (spaceId: number, data: { email: string; role?: string }) =>
				request<SpaceMember>('POST', `/spaces/${spaceId}/members`, data),
			updateRole: (spaceId: number, memberId: number, role: string) =>
				request<SpaceMember>('PUT', `/spaces/${spaceId}/members/${memberId}`, { role }),
			remove: (spaceId: number, memberId: number) =>
				request<void>('DELETE', `/spaces/${spaceId}/members/${memberId}`)
		}
	},
	clients: {
		list: () => request<Client[]>('GET', '/clients'),
		get: (id: number) => request<Client>('GET', `/clients/${id}`),
		create: (data: { name: string; email?: string; company?: string; phone?: string; notes?: string }) =>
			request<Client>('POST', '/clients', data),
		update: (id: number, data: { name: string; email?: string; company?: string; phone?: string; notes?: string }) =>
			request<Client>('PUT', `/clients/${id}`, data),
		delete: (id: number) => request<void>('DELETE', `/clients/${id}`),
		documents: (id: number) => request<Document[]>('GET', `/documents?client_id=${id}`)
	},
	verify: {
		check: (file: File) => {
			const formData = new FormData();
			formData.append('file', file);
			return upload<VerifyResponse>('POST', '/verify', formData, { quiet: true });
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
	avatar_url: string;
	avatar_source?: string;
	locale: string;
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
	client_id?: number | null;
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
	avatar_url?: string;
	role: string;
	status: 'pending' | 'signed' | 'declined';
	token: string;
	order_num: number;
	signed_at: string | null;
	viewed_at: string | null;
	email_opened_at: string | null;
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

export interface SigningRosterEntry {
	name: string;
	role: string;
	status: string;
	order_num: number;
	signed_at: string | null;
	is_you: boolean;
}

export interface SigningStatus {
	document: {
		id: number;
		name: string;
		file_name: string;
		status: string;
	};
	signer: SigningRosterEntry;
	signers: SigningRosterEntry[];
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

export interface Space {
	id: number;
	name: string;
	description: string;
	owner_id: number;
	role: 'owner' | 'admin' | 'member';
	created_at: string;
	updated_at: string;
	registre_url: string;
}

export interface Client {
	id: number;
	name: string;
	email: string;
	company: string;
	phone: string;
	notes: string;
	owner_id: number;
	created_at: string;
	updated_at: string;
}

export interface SpaceMember {
	id: number;
	user_id: number;
	email: string;
	name: string;
	avatar_url: string;
	role: 'owner' | 'admin' | 'member';
	joined_at: string;
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
