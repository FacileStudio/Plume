import type { Handle } from '@sveltejs/kit';

const API_URL = process.env.API_URL ?? 'http://localhost:4000';

export const handle: Handle = async ({ event, resolve }) => {
	if (event.url.pathname.startsWith('/api/')) {
		const path = event.url.pathname.slice(4);
		const url = `${API_URL}${path}${event.url.search}`;

		const headers = new Headers();
		for (const [key, value] of event.request.headers.entries()) {
			if (key === 'host') continue;
			headers.set(key, value);
		}

		headers.set('X-Real-IP', event.getClientAddress());

		const res = await fetch(url, {
			method: event.request.method,
			headers,
			body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
				? event.request.body
				: undefined,
			redirect: 'manual',
			// @ts-expect-error duplex needed for streaming request body
			duplex: 'half'
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
