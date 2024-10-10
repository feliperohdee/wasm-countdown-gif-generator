import './wasm_exec';
import wasm from './main.wasm';

export default {
    async fetch(req, env, ctx) {
		const cache = caches.default;
		const url = new URL(req.url);

        if (url.pathname === '/favicon.ico') {
            return new Response(null, {
                status: 204
            });
        }

		const cached = await cache.match(req);

		if (cached) {
			return cached;
		}

        const go = new globalThis.Go();
        const instance = await WebAssembly.instantiate(wasm, go.importObject);

		go.run(instance);
;
        const base64 = globalThis.build({
			background: url.searchParams.get('background') || url.searchParams.get('bg') || '000',
			color: url.searchParams.get('color') || 'fff',
			date: new Date(url.searchParams.get('date') || '2025-01-01').toISOString(),
			kind: url.searchParams.get('kind') || 'rounded'
		});

        const binaryString = atob(base64);
        const bytes = new Uint8Array(binaryString.length);

        for (let i = 0; i < binaryString.length; i++) {
            bytes[i] = binaryString.charCodeAt(i);
        }

        const res = new Response(bytes, {
            headers: {
				'cache-control': 'public, max-age=3600',
                'content-type': 'image/gif'
            }
        });

		ctx.waitUntil(cache.put(req, res.clone()));

		return res;
    }
};