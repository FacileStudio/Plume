import type { Handle } from '@sveltejs/kit';

const API_URL = process.env.API_URL ?? 'http://localhost:4000';

export const handle: Handle = async ({ event, resolve }) => {
	if (event.url.pathname.startsWith('/api/e/')) {
		const path = event.url.pathname.slice(4);
		const url = `${API_URL}${path}${event.url.search}`;

		if (event.request.method === 'OPTIONS') {
			return new Response(null, {
				status: 204,
				headers: {
					'Access-Control-Allow-Origin': '*',
					'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
					'Access-Control-Allow-Headers': 'Content-Type',
					'Access-Control-Max-Age': '86400'
				}
			});
		}

		const proxyHeaders: Record<string, string> = { 'Content-Type': 'text/plain' };
		const cfCountry = event.request.headers.get('cf-ipcountry');
		if (cfCountry) proxyHeaders['CF-IPCountry'] = cfCountry;
		const userAgent = event.request.headers.get('user-agent');
		if (userAgent) proxyHeaders['User-Agent'] = userAgent;

		const res = await fetch(url, {
			method: event.request.method,
			headers: proxyHeaders,
			body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
				? await event.request.text()
				: undefined
		});

		const body = await res.text();
		return new Response(body, {
			status: res.status,
			headers: {
				'Access-Control-Allow-Origin': '*',
				'Content-Type': res.headers.get('Content-Type') ?? 'application/json',
				'Cache-Control': 'no-cache, no-store'
			}
		});
	}

	if (event.url.pathname.startsWith('/api/')) {
		const path = event.url.pathname.slice(4);
		const url = `${API_URL}${path}${event.url.search}`;

		const headers = new Headers();
		for (const [key, value] of event.request.headers.entries()) {
			if (key === 'host') continue;
			headers.set(key, value);
		}

		const cookieHeader = event.request.headers.get('cookie');
		if (cookieHeader) {
			headers.set('cookie', cookieHeader);
		}

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
