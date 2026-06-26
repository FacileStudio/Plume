<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib';
	import type { Field, Signer, CreateFieldRequest } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import Icon from '@iconify/svelte';

	let { documentId, signers, onclose }: {
		documentId: number;
		signers: Signer[];
		onclose: () => void;
	} = $props();

	const SIGNER_COLORS = ['#3b82f6', '#22c55e', '#f97316', '#a855f7', '#ec4899'];
	const FIELD_DEFAULTS: Record<string, { width: number; height: number }> = {
		signature: { width: 200, height: 50 },
		text: { width: 150, height: 30 },
		date: { width: 120, height: 30 },
		checkbox: { width: 30, height: 30 }
	};

	let fields = $state<Field[]>([]);
	let selectedSignerId = $state<number>(0);

	$effect(() => {
		if (signers.length > 0 && selectedSignerId === 0) {
			selectedSignerId = signers[0].id;
		}
	});
	let selectedFieldId = $state<number | null>(null);
	let loading = $state(true);
	let pdfError = $state('');
	let pagesContainer = $state<HTMLDivElement>();
	let currentPage = $state(1);

	let pages = $state<{ num: number; width: number; height: number }[]>([]);
	let pageCanvases = $state<Map<number, HTMLCanvasElement>>(new Map());

	let dragState = $state<{
		fieldId: number;
		type: 'move' | 'resize';
		startX: number;
		startY: number;
		origX: number;
		origY: number;
		origW: number;
		origH: number;
		pageNum: number;
		pointerId: number;
	} | null>(null);

	let observer: IntersectionObserver | null = null;

	function signerColor(signerId: number): string {
		const idx = signers.findIndex((s) => s.id === signerId);
		return SIGNER_COLORS[idx % SIGNER_COLORS.length];
	}

	function signerName(signerId: number): string {
		return signers.find((s) => s.id === signerId)?.name ?? 'Unknown';
	}

	function fieldsForPage(pageNum: number): Field[] {
		return fields.filter((f) => f.page === pageNum);
	}

	async function loadFields() {
		fields = await api.fields.list(documentId);
	}

	async function addField(fieldType: string) {
		if (!selectedSignerId) return;
		const defaults = FIELD_DEFAULTS[fieldType];
		const pageInfo = pages.find((p) => p.num === currentPage) ?? pages[0];
		if (!pageInfo) return;

		const req: CreateFieldRequest = {
			signer_id: selectedSignerId,
			field_type: fieldType,
			page: pageInfo.num,
			x: (50 / pageInfo.width) * 100,
			y: (50 / pageInfo.height) * 100,
			width: (defaults.width / pageInfo.width) * 100,
			height: (defaults.height / pageInfo.height) * 100,
			required: true,
			label: ''
		};

		const created = await api.fields.create(documentId, req);
		fields = [...fields, created];
		selectedFieldId = created.id;
	}

	async function deleteField(fieldId: number) {
		await api.fields.delete(documentId, fieldId);
		fields = fields.filter((f) => f.id !== fieldId);
		if (selectedFieldId === fieldId) selectedFieldId = null;
	}

	async function persistField(field: Field) {
		await api.fields.update(documentId, field.id, {
			field_type: field.field_type,
			page: field.page,
			x: field.x,
			y: field.y,
			width: field.width,
			height: field.height,
			required: field.required,
			label: field.label || ''
		});
	}

	async function renameField(field: Field, label: string) {
		field.label = label;
		await persistField(field);
	}

	function handlePointerDown(e: PointerEvent, field: Field, type: 'move' | 'resize') {
		e.preventDefault();
		e.stopPropagation();
		selectedFieldId = field.id;

		const pageEl = (e.currentTarget as HTMLElement).closest('[data-page]') as HTMLElement;
		const pageNum = Number(pageEl.dataset.page);

		dragState = {
			fieldId: field.id,
			type,
			startX: e.clientX,
			startY: e.clientY,
			origX: field.x,
			origY: field.y,
			origW: field.width,
			origH: field.height,
			pageNum,
			pointerId: e.pointerId
		};
	}

	function findPageUnderPointer(clientX: number, clientY: number): { pageNum: number; rect: DOMRect } | null {
		if (!pagesContainer) return null;
		const pageEls = pagesContainer.querySelectorAll<HTMLElement>('[data-page]');
		for (const el of pageEls) {
			const rect = el.getBoundingClientRect();
			if (clientX >= rect.left && clientX <= rect.right && clientY >= rect.top && clientY <= rect.bottom) {
				return { pageNum: Number(el.dataset.page), rect };
			}
		}
		return null;
	}

	function handlePointerMove(e: PointerEvent) {
		if (!dragState) return;
		e.preventDefault();

		const field = fields.find((f) => f.id === dragState!.fieldId);
		if (!field) return;

		if (dragState.type === 'move') {
			const hit = findPageUnderPointer(e.clientX, e.clientY);
			if (hit) {
				if (hit.pageNum !== dragState.pageNum) {
					dragState.startX = e.clientX;
					dragState.startY = e.clientY;
					dragState.origX = field.x;
					dragState.origY = field.y;
					dragState.pageNum = hit.pageNum;
					field.page = hit.pageNum;
				}

				const dxPct = ((e.clientX - dragState.startX) / hit.rect.width) * 100;
				const dyPct = ((e.clientY - dragState.startY) / hit.rect.height) * 100;
				field.x = Math.max(0, Math.min(100 - field.width, dragState.origX + dxPct));
				field.y = Math.max(0, Math.min(100 - field.height, dragState.origY + dyPct));
			}
		} else {
			const pageEl = pagesContainer?.querySelector(`[data-page="${dragState.pageNum}"]`) as HTMLElement;
			if (!pageEl) return;
			const rect = pageEl.getBoundingClientRect();
			const dxPct = ((e.clientX - dragState.startX) / rect.width) * 100;
			const dyPct = ((e.clientY - dragState.startY) / rect.height) * 100;
			field.width = Math.max(2, Math.min(100 - field.x, dragState.origW + dxPct));
			field.height = Math.max(2, Math.min(100 - field.y, dragState.origH + dyPct));
		}
	}

	async function handlePointerUp() {
		if (!dragState) return;
		const field = fields.find((f) => f.id === dragState!.fieldId);
		dragState = null;
		if (field) await persistField(field);
	}

	function appendCanvas(node: HTMLElement, canvas: HTMLCanvasElement) {
		node.appendChild(canvas);
		return {
			destroy() {
				canvas.remove();
			}
		};
	}

	function handleOverlayClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			selectedFieldId = null;
		}
	}

	function setupPageObserver() {
		if (!pagesContainer) return;
		observer = new IntersectionObserver(
			(entries) => {
				let best: { pageNum: number; ratio: number } | null = null;
				for (const entry of entries) {
					const pageNum = Number((entry.target as HTMLElement).dataset.page);
					if (!best || entry.intersectionRatio > best.ratio) {
						best = { pageNum, ratio: entry.intersectionRatio };
					}
				}
				if (best && best.ratio > 0) {
					currentPage = best.pageNum;
				}
			},
			{ root: pagesContainer, threshold: [0, 0.25, 0.5, 0.75, 1] }
		);
		const pageEls = pagesContainer.querySelectorAll('[data-page]');
		for (const el of pageEls) observer.observe(el);
	}

	onMount(async () => {
		const pdfjsLib = await import('pdfjs-dist');
		pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
			'pdfjs-dist/build/pdf.worker.min.mjs',
			import.meta.url
		).toString();

		try {
			const token = localStorage.getItem('token');
			const response = await fetch(`/api/documents/${documentId}/file`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (!response.ok) throw new Error('Failed to load PDF');

			const arrayBuffer = await response.arrayBuffer();
			const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;

			for (let i = 1; i <= pdf.numPages; i++) {
				const page = await pdf.getPage(i);
				const viewport = page.getViewport({ scale: 1.5 });
				pages.push({ num: i, width: viewport.width, height: viewport.height });

				const canvas = document.createElement('canvas');
				canvas.width = viewport.width;
				canvas.height = viewport.height;
				canvas.style.width = '100%';
				canvas.style.height = 'auto';
				canvas.style.display = 'block';
				pageCanvases.set(i, canvas);

				await page.render({
					canvasContext: canvas.getContext('2d')!,
					canvas,
					viewport
				}).promise;
			}

			await loadFields();
		} catch {
			pdfError = 'Failed to load document';
		}
		loading = false;

		requestAnimationFrame(() => setupPageObserver());
	});

	onDestroy(() => {
		observer?.disconnect();
	});
</script>

<svelte:window onpointermove={handlePointerMove} onpointerup={handlePointerUp} />

<div class="fixed inset-0 z-50 flex flex-col bg-background">
	<div class="flex items-center justify-between px-4 py-3 border-b bg-background">
		<div class="flex items-center gap-3">
			<h2 class="text-lg font-semibold">Prepare fields</h2>
			{#if pages.length > 1}
				<span class="text-sm text-muted-foreground">Page {currentPage} / {pages.length}</span>
			{/if}
		</div>
		<Button onclick={onclose}>
			<Icon icon="solar:check-circle-linear" class="h-4 w-4" />
			Save & close
		</Button>
	</div>

	{#if loading}
		<div class="flex flex-1 items-center justify-center">
			<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
		</div>
	{:else if pdfError}
		<div class="flex flex-1 items-center justify-center text-sm text-muted-foreground">
			{pdfError}
		</div>
	{:else}
		<div class="flex flex-1 overflow-hidden">
			<div bind:this={pagesContainer} class="flex-1 overflow-y-auto p-6 bg-muted/30">
				{#each pages as pg}
					<div
						class="relative mx-auto mb-6 shadow-lg"
						style="max-width: {pg.width}px;"
						data-page={pg.num}
					>
						{#if pageCanvases.get(pg.num)}
							{@const canvas = pageCanvases.get(pg.num)!}
							<div class="pdf-canvas-host" use:appendCanvas={canvas}></div>
						{/if}

						<div
							class="absolute inset-0"
							role="presentation"
							onclick={handleOverlayClick}
						>
							{#each fieldsForPage(pg.num) as field}
								{@const color = signerColor(field.signer_id)}
								{@const isSelected = selectedFieldId === field.id}
								<div
									role="button"
									tabindex="0"
									class="absolute flex items-center justify-center text-xs font-medium select-none cursor-grab"
									style="
										left: {field.x}%;
										top: {field.y}%;
										width: {field.width}%;
										height: {field.height}%;
										background: {isSelected ? `${color}30` : `${color}20`};
										border: 2px {isSelected ? 'solid' : 'dashed'} {color};
										color: {color};
									"
									onpointerdown={(e) => handlePointerDown(e, field, 'move')}
								>
									<span class="truncate px-1 pointer-events-none">
										{field.label || field.field_type} &middot; {signerName(field.signer_id)}
									</span>

									{#if isSelected}
										<button
											class="absolute -top-2 -right-2 h-5 w-5 rounded-full bg-destructive text-destructive-foreground flex items-center justify-center text-xs hover:scale-110 transition-transform"
											onpointerdown={(e: PointerEvent) => { e.preventDefault(); e.stopPropagation(); deleteField(field.id); }}
										>
											<Icon icon="solar:close-circle-bold" class="h-3.5 w-3.5 pointer-events-none" />
										</button>

										<!-- svelte-ignore a11y_no_static_element_interactions -->
										<div
											class="absolute bottom-0 right-0 h-4 w-4 cursor-nwse-resize flex items-end justify-end"
											onpointerdown={(e) => handlePointerDown(e, field, 'resize')}
										>
											<svg width="10" height="10" viewBox="0 0 10 10" class="pointer-events-none">
												<path d="M10 0 L10 10 L0 10 Z" fill={color} opacity="0.7" />
											</svg>
										</div>
									{/if}
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>

			<div class="w-72 border-l bg-background p-4 overflow-y-auto flex flex-col gap-4">
				<div>
					<p class="text-sm font-medium mb-1.5">Signer</p>
					<div class="flex flex-col gap-1">
						{#each signers as signer, i}
							{@const color = SIGNER_COLORS[i % SIGNER_COLORS.length]}
							<button
								class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm text-left transition-colors hover:bg-muted"
								class:bg-muted={selectedSignerId === signer.id}
								class:border-foreground={selectedSignerId === signer.id}
								onclick={() => (selectedSignerId = signer.id)}
							>
								<span class="h-3 w-3 rounded-full shrink-0" style="background: {color};"></span>
								<span class="truncate">{signer.name}</span>
							</button>
						{/each}
					</div>
				</div>

				<div class="flex flex-col gap-2">
					<p class="text-sm font-medium">Add field</p>
					<Button variant="outline" class="justify-start" onclick={() => addField('signature')}>
						<Icon icon="solar:pen-new-round-linear" class="h-4 w-4" />
						Signature
					</Button>
					<Button variant="outline" class="justify-start" onclick={() => addField('text')}>
						<Icon icon="solar:text-linear" class="h-4 w-4" />
						Text
					</Button>
					<Button variant="outline" class="justify-start" onclick={() => addField('date')}>
						<Icon icon="solar:calendar-linear" class="h-4 w-4" />
						Date
					</Button>
					<Button variant="outline" class="justify-start" onclick={() => addField('checkbox')}>
						<Icon icon="solar:check-square-linear" class="h-4 w-4" />
						Checkbox
					</Button>
				</div>

				{#if fields.length > 0}
					<div>
						<p class="text-sm font-medium mb-2">Placed fields ({fields.length})</p>
						<div class="flex flex-col gap-1.5">
							{#each fields as field}
								{@const color = signerColor(field.signer_id)}
								{@const isActive = selectedFieldId === field.id}
								<div
									class="flex items-center gap-1.5 rounded-md border px-3 py-2 transition-colors cursor-pointer hover:bg-muted"
									class:border-foreground={isActive}
									class:bg-muted={isActive}
									onclick={() => (selectedFieldId = field.id)}
								>
									<span class="h-2.5 w-2.5 rounded-full shrink-0" style="background: {color};"></span>
									{#if isActive}
										<input
											type="text"
											value={field.label || ''}
											placeholder={field.field_type}
											class="flex-1 min-w-0 bg-transparent text-sm border-none outline-none placeholder:text-muted-foreground"
											onclick={(e) => e.stopPropagation()}
											onchange={(e) => renameField(field, (e.currentTarget as HTMLInputElement).value)}
										/>
									{:else}
										<span class="flex-1 min-w-0 truncate text-sm">{field.label || field.field_type}</span>
									{/if}
									<span class="text-[10px] text-muted-foreground shrink-0">p{field.page}</span>
									<button
										class="rounded-md p-1 text-muted-foreground transition-colors hover:text-red-500 shrink-0"
										onclick={(e) => { e.stopPropagation(); deleteField(field.id); }}
									>
										<Icon icon="solar:trash-bin-trash-linear" class="h-3.5 w-3.5" />
									</button>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
