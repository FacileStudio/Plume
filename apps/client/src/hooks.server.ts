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

		const init: RequestInit & { duplex?: 'half' } = {
			method: event.request.method,
			headers,
			redirect: 'manual'
		};

		if (event.request.method !== 'GET' && event.request.method !== 'HEAD') {
			init.body = event.request.body;
			init.duplex = 'half';
		}

		const res = await fetch(url, init);

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

		responseHeaders.delete('content-encoding');
		responseHeaders.delete('content-length');
		responseHeaders.delete('transfer-encoding');

		return new Response(res.body, {
			status: res.status,
			statusText: res.statusText,
			headers: responseHeaders
		});
	}

	return resolve(event);
};
