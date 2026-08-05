import type { Handle } from '@sveltejs/kit';
import { handle } from '../src/hooks.server.ts';

type ProxyInit = RequestInit & { duplex?: 'half' };

function assert(condition: boolean, message: string): void {
	if (!condition) throw new Error(message);
}

async function proxy(
	request: Request,
	upstreamHeaders: Record<string, string>
): Promise<{ response: Response; init: ProxyInit }> {
	let init: ProxyInit = {};
	globalThis.fetch = (async (_url: string, requestInit: ProxyInit) => {
		init = requestInit;
		return new Response('{"ok":true}', { status: 200, headers: upstreamHeaders });
	}) as unknown as typeof fetch;

	const response = await handle({
		event: {
			url: new URL(request.url),
			request,
			getClientAddress: () => '127.0.0.1'
		},
		resolve: async () => new Response('not proxied')
	} as unknown as Parameters<Handle>[0]);

	return { response, init };
}

const proxied = await proxy(new Request('http://localhost:3000/api/documents'), {
	'content-type': 'application/json',
	'content-encoding': 'gzip',
	'content-length': '999',
	'transfer-encoding': 'chunked',
	'content-range': 'bytes 0-10/11'
});

assert(
	proxied.response.headers.get('content-encoding') === null,
	'content-encoding must be stripped: undici already decoded the body'
);
assert(proxied.response.headers.get('content-length') === null, 'content-length must be stripped');
assert(
	proxied.response.headers.get('transfer-encoding') === null,
	'transfer-encoding must be stripped'
);
assert(
	proxied.response.headers.get('content-range') === 'bytes 0-10/11',
	'content-range must be preserved for range requests'
);
assert(
	proxied.response.headers.get('content-type') === 'application/json',
	'content-type must be preserved'
);

const upload = await proxy(
	new Request('http://localhost:3000/api/documents', { method: 'POST', body: 'pdf-bytes' }),
	{ 'content-type': 'application/json' }
);

assert(upload.init.duplex === 'half', 'streamed request bodies require duplex: half');
assert(
	upload.init.body instanceof ReadableStream,
	'request body must be streamed, not buffered into memory'
);

console.log('proxy header + streaming checks passed');
