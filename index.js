import './wasm_exec';
import wasm from './main.wasm';

const toNumber = (value, fallback) => {
	if (value === '' || value === null) {
		return fallback;
	}

	const number = Number(value);

	return isNaN(number) ? fallback : number;
};

export default {
    async fetch(req, env, ctx) {
        try {
            const cache = caches.default;
            const url = new URL(req.url);

            if (url.pathname === '/favicon.ico') {
                return new Response(null, {
                    status: 204
                });
            }

            const cached = await cache.match(req);

            if (cached && env.ENV === 'production') {
                return cached;
            }

            const go = new globalThis.Go();
            const instance = await WebAssembly.instantiate(wasm, go.importObject);

            go.run(instance);
            // const base64 = globalThis.build({
            //     background: url.searchParams.get('background') || url.searchParams.get('bg') || '000',
            //     color: url.searchParams.get('color') || 'fff',
            //     date: new Date(url.searchParams.get('date') || '2025-01-01').toISOString(),
            //     gmt: toNumber(url.searchParams.get('gmt'), 0),
            //     frames: toNumber(url.searchParams.get('frames'), 10),
            //     lang: url.searchParams.get('lang') || 'en',
            //     kind: url.searchParams.get('kind') || 'rounded'
            // });

			// const base64 = globalThis.buildLedBanner({
			// 	forward: url.searchParams.get('forward') === 'true',
			// 	height: toNumber(url.searchParams.get('height'), 200),
			// 	delay: toNumber(url.searchParams.get('delay'), 100),
			// 	spaceSize: toNumber(url.searchParams.get('spaceSize'), 5),
			// 	speed: toNumber(url.searchParams.get('speed'), 5),
			// 	text: url.searchParams.get('text') || 'BLACK FRIDAY',
			// 	width: toNumber(url.searchParams.get('width'), 800),
            //     background: url.searchParams.get('background') || url.searchParams.get('bg') || 'fff',
            //     color: url.searchParams.get('color') || '000',
            //     frames: toNumber(url.searchParams.get('frames'), 15),
            // });
			
			// const base64 = globalThis.buildFlashingText({
			// 	forward: url.searchParams.get('forward') === 'true',
			// 	height: toNumber(url.searchParams.get('height'), 200),
			// 	delay: toNumber(url.searchParams.get('delay'), 200),
			// 	spaceSize: toNumber(url.searchParams.get('spaceSize'), 5),
			// 	speed: toNumber(url.searchParams.get('speed'), 5),
			// 	text: url.searchParams.get('text') || 'BLACK FRIDAY',
			// 	width: toNumber(url.searchParams.get('width'), 800),
            //     background: url.searchParams.get('background') || url.searchParams.get('bg') || 'fff',
            //     color: url.searchParams.get('color') || '000',
            //     frames: toNumber(url.searchParams.get('frames'), 15),
            // });

			const base64 = globalThis.buildTypingText({});

            const binaryString = atob(base64);
            const bytes = new Uint8Array(binaryString.length);

            for (let i = 0; i < binaryString.length; i++) {
                bytes[i] = binaryString.charCodeAt(i);
            }

            const res = new Response(bytes, {
                headers: {
                    'cache-control': 'public, max-age=3600',
                    'content-type': 'image/gif',
                    'content-length': bytes.length,
                    'content-disposition': 'inline'
                }
            });

            ctx.waitUntil(cache.put(req, res.clone()));

            return res;
        } catch (err) {
			return Response.json({
				error: err.message || err.toString()
			}, {
				status: 500
			});
		}
    }
};