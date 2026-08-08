<script lang="ts">
	import '../app.css';
	import { Toaster } from 'svelte-sonner';
	import { browser } from '$app/environment';
	import { theme } from '$lib/theme.svelte';

	let { children } = $props();

	if (browser) theme.restore();
</script>

<Toaster richColors />
{@render children()}
