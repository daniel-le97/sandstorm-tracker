import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

// https://vitejs.dev/config/
export default defineConfig( {
	base: '/ui/',
	build: {
		outDir: '../assets/ui',
		emptyOutDir: true,
		rollupOptions: {
			output: {
				// Ensure all assets are accessible
				manualChunks: undefined,
			},
		},
	},
	plugins: [
		preact( {
			prerender: {
				enabled: true,
				renderTarget: '#app',
				additionalPrerenderRoutes: [ '/404', '/login', '/servers', '/matches', '/players', '/weapons' ],
				previewMiddlewareEnabled: true,
				previewMiddlewareFallback: '/404',
			},
		} ),
	],
} );
