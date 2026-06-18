import type { Handle } from '@sveltejs/kit';

const API_URL = process.env.API_URL ?? 'http://localhost:4000';

export const handle: Handle = async ({ event, resolve }) => {
	if (event.url.pathname.startsWith('/api/')) {
		const url = `${API_URL}${event.url.pathname}${event.url.search}`;

		const headers = new Headers();
		for (const [key, value] of event.request.headers.entries()) {
			if (key === 'host') continue;
			headers.set(key, value);
		}

		headers.set('X-Real-IP', event.getClientAddress());

		let body: ArrayBuffer | undefined;
		if (event.request.method !== 'GET' && event.request.method !== 'HEAD') {
			body = await event.request.arrayBuffer();
		}

		const res = await fetch(url, {
			method: event.request.method,
			headers,
			body,
			redirect: 'manual'
		});

		const responseHeaders = new Headers();
		for (const [key, value] of res.headers.entries()) {
			responseHeaders.append(key, value);
		}
		const setCookies = res.headers.getSetCookie?.();
		if (setCookies) {
			responseHeaders.delete('set-cookie');
			for (const cookie of setCookies) {
				responseHeaders.append('set-cookie', cookie);
			}
		}

		return new Response(res.body, {
			status: res.status,
			statusText: res.statusText,
			headers: responseHeaders
		});
	}

	return resolve(event);
};
