<script lang="ts">
	import { onMount } from 'svelte';

	let { url, class: className = '' }: { url: string; class?: string } = $props();
	let container: HTMLDivElement;
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		const pdfjsLib = await import('pdfjs-dist');
		pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
			'pdfjs-dist/build/pdf.worker.min.mjs',
			import.meta.url
		).toString();

		try {
			const pdf = await pdfjsLib.getDocument(url).promise;
			loading = false;

			for (let i = 1; i <= pdf.numPages; i++) {
				const page = await pdf.getPage(i);
				const viewport = page.getViewport({ scale: 1.5 });
				const canvas = document.createElement('canvas');
				canvas.width = viewport.width;
				canvas.height = viewport.height;
				canvas.style.width = '100%';
				canvas.style.height = 'auto';
				canvas.classList.add('rounded-lg', 'border', 'mb-4');
				container.appendChild(canvas);

				await page.render({
					canvasContext: canvas.getContext('2d')!,
					canvas,
					viewport
				}).promise;
			}
		} catch {
			error = 'Failed to load PDF preview';
			loading = false;
		}
	});
</script>

<div bind:this={container} class={className}>
	{#if loading}
		<div class="flex items-center justify-center py-12 text-muted-foreground">
			<span class="text-sm">Loading document preview…</span>
		</div>
	{/if}
	{#if error}
		<div class="flex items-center justify-center py-8 text-sm text-muted-foreground">
			{error}
		</div>
	{/if}
</div>
