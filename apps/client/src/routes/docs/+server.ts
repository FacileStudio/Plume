import type { RequestHandler } from './$types';

const html = `<!doctype html>
<html>
<head>
  <title>Plume API</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <style>body { margin: 0; }</style>
</head>
<body>
  <script id="api-reference" data-url="/api/docs/openapi.json"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`;

export const prerender = false;

export const GET: RequestHandler = () =>
	new Response(html, { headers: { 'content-type': 'text/html; charset=utf-8' } });
