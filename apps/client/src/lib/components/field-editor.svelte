<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api, getToken } from '$lib';
	import type { Field, Signer, CreateFieldRequest } from '$lib';
	import {
		Alert,
		Badge,
		Button,
		ConfirmModal,
		IconButton,
		Input,
		Spinner,
		USER_COLORS,
		cn,
		icons,
		toast
	} from '@facile/muse';

	let { documentId, signers, onclose }: {
		documentId: number;
		signers: Signer[];
		onclose: () => void;
	} = $props();

	const FIELD_DEFAULTS: Record<string, { width: number; height: number }> = {
		signature: { width: 200, height: 50 },
		text: { width: 150, height: 30 },
		date: { width: 120, height: 30 },
		checkbox: { width: 30, height: 30 }
	};

	const FIELD_TYPES: { type: string; label: string; icon: string }[] = [
		{ type: 'signature', label: 'Signature', icon: 'solar:pen-new-round-linear' },
		{ type: 'text', label: 'Text', icon: 'solar:text-linear' },
		{ type: 'date', label: 'Date', icon: icons.calendar },
		{ type: 'checkbox', label: 'Checkbox', icon: 'solar:check-square-linear' }
	];

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

	let confirmDelete = $state(false);
	let pendingDeleteId = $state<number | null>(null);

	let pages = $state<{ num: number; width: number; height: number }[]>([]);
	let pageCanvases = $state<Map<number, HTMLCanvasElement>>(new Map());

	let dragState = $state<{
		fieldId: number;
		type: 'move' | 'resize';
		startX: number;
		startY: number;
		origW: number;
		origH: number;
		// Where inside the field box the pointer grabbed, as a fraction (0-1)
		// of the box's own width/height. Kept page-size independent so the
		// grab point stays under the cursor across pages of differing sizes.
		grabFracX: number;
		grabFracY: number;
		pageNum: number;
		pointerId: number;
	} | null>(null);

	let observer: IntersectionObserver | null = null;

	function signerColor(signerId: number): string {
		const idx = signers.findIndex((s) => s.id === signerId);
		return USER_COLORS[idx % USER_COLORS.length];
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

	// Center a new field on the part of the page the user is actually looking at:
	// the vertical band of the current page that is visible inside the scroll
	// container, rather than a fixed spot near the top of the page.
	function centeredPlacement(pageNum: number, widthPct: number, heightPct: number): { x: number; y: number } {
		const pageEl = pagesContainer?.querySelector<HTMLElement>(`[data-page="${pageNum}"]`);
		let cyPct = 50;
		if (pageEl) {
			const pageRect = pageEl.getBoundingClientRect();
			const containerRect = pagesContainer?.getBoundingClientRect();
			let visTop = pageRect.top;
			let visBottom = pageRect.bottom;
			if (containerRect) {
				visTop = Math.max(pageRect.top, containerRect.top);
				visBottom = Math.min(pageRect.bottom, containerRect.bottom);
			}
			if (visBottom <= visTop) {
				visTop = pageRect.top;
				visBottom = pageRect.bottom;
			}
			if (pageRect.height > 0) {
				cyPct = (((visTop + visBottom) / 2 - pageRect.top) / pageRect.height) * 100;
			}
		}
		const x = Math.max(0, Math.min(100 - widthPct, 50 - widthPct / 2));
		const y = Math.max(0, Math.min(100 - heightPct, cyPct - heightPct / 2));
		return { x, y };
	}

	async function addField(fieldType: string) {
		if (!selectedSignerId) return;
		const defaults = FIELD_DEFAULTS[fieldType];
		const pageInfo = pages.find((p) => p.num === currentPage) ?? pages[0];
		if (!pageInfo) return;

		const widthPct = (defaults.width / pageInfo.width) * 100;
		const heightPct = (defaults.height / pageInfo.height) * 100;
		const { x, y } = centeredPlacement(pageInfo.num, widthPct, heightPct);

		const req: CreateFieldRequest = {
			signer_id: selectedSignerId,
			field_type: fieldType,
			page: pageInfo.num,
			x,
			y,
			width: widthPct,
			height: heightPct,
			required: true,
			label: ''
		};

		try {
			const created = await api.fields.create(documentId, req);
			fields = [...fields, created];
			selectedFieldId = created.id;
		} catch {
			toast.danger('Could not add the field');
		}
	}

	function askDelete(fieldId: number) {
		pendingDeleteId = fieldId;
		confirmDelete = true;
	}

	async function deleteField() {
		const fieldId = pendingDeleteId;
		if (fieldId === null) return;
		try {
			await api.fields.delete(documentId, fieldId);
			fields = fields.filter((f) => f.id !== fieldId);
			if (selectedFieldId === fieldId) selectedFieldId = null;
			pendingDeleteId = null;
		} catch {
			toast.danger('Could not delete the field');
			throw new Error('delete failed');
		}
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

		// The field box carries role="button"; from either the box itself (move)
		// or the resize handle inside it (resize) this resolves to the box, so we
		// can measure where within the box the pointer grabbed.
		const fieldEl = (e.currentTarget as HTMLElement).closest('[role="button"]') as HTMLElement | null;
		const fieldRect = fieldEl?.getBoundingClientRect();
		const grabFracX = fieldRect && fieldRect.width > 0 ? (e.clientX - fieldRect.left) / fieldRect.width : 0.5;
		const grabFracY = fieldRect && fieldRect.height > 0 ? (e.clientY - fieldRect.top) / fieldRect.height : 0.5;

		dragState = {
			fieldId: field.id,
			type,
			startX: e.clientX,
			startY: e.clientY,
			origW: field.width,
			origH: field.height,
			grabFracX,
			grabFracY,
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
					dragState.pageNum = hit.pageNum;
					field.page = hit.pageNum;
				}

				// Position the box absolutely from the pointer, keeping the grab
				// point under the cursor. Field size in pixels is derived from the
				// page currently under the pointer, so crossing to a differently
				// sized page never makes the box jump away from the cursor.
				const fieldWpx = (field.width / 100) * hit.rect.width;
				const fieldHpx = (field.height / 100) * hit.rect.height;
				const leftPx = e.clientX - dragState.grabFracX * fieldWpx;
				const topPx = e.clientY - dragState.grabFracY * fieldHpx;
				const xPct = ((leftPx - hit.rect.left) / hit.rect.width) * 100;
				const yPct = ((topPx - hit.rect.top) / hit.rect.height) * 100;
				field.x = Math.max(0, Math.min(100 - field.width, xPct));
				field.y = Math.max(0, Math.min(100 - field.height, yPct));
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
			const token = getToken();
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

<div class="fixed inset-0 z-50 flex h-dvh flex-col bg-fc-page">
	<div class="flex items-center justify-between gap-3 border-b border-fc-border px-4 py-3">
		<div class="flex min-w-0 items-center gap-3">
			<h2 class="truncate text-fc-lg font-semibold text-fc-fg">Prepare fields</h2>
			{#if pages.length > 1}
				<Badge tone="neutral">Page {currentPage} / {pages.length}</Badge>
			{/if}
		</div>
		<Button icon={icons.check} onclick={onclose}>Save &amp; close</Button>
	</div>

	{#if loading}
		<div class="flex flex-1 items-center justify-center">
			<Spinner size="lg" />
		</div>
	{:else if pdfError}
		<div class="flex flex-1 items-start justify-center p-6">
			<Alert tone="danger" class="w-full max-w-fc-sm">{pdfError}</Alert>
		</div>
	{:else}
		<div class="flex min-h-0 flex-1 flex-col-reverse md:flex-row">
			<div
				bind:this={pagesContainer}
				class="pdf-pages min-h-0 flex-1 overflow-y-auto bg-fc-surface p-4 sm:p-6"
			>
				{#each pages as pg}
					<div
						class="relative mx-auto mb-6 rounded-fc-md shadow-lg"
						style="max-width: {pg.width}px;"
						data-page={pg.num}
					>
						{#if pageCanvases.get(pg.num)}
							{@const canvas = pageCanvases.get(pg.num)!}
							<div class="pdf-canvas-host overflow-hidden rounded-fc-md" use:appendCanvas={canvas}></div>
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
									class="absolute flex cursor-grab select-none items-center justify-center rounded-fc-xs text-fc-xs font-medium"
									style="
										left: {field.x}%;
										top: {field.y}%;
										width: {field.width}%;
										height: {field.height}%;
										background: color-mix(in oklab, {color} {isSelected ? 40 : 24}%, transparent);
										border: 2px {isSelected ? 'solid' : 'dashed'} {color};
										color: var(--page-ink);
									"
									onpointerdown={(e) => handlePointerDown(e, field, 'move')}
								>
									<span class="pointer-events-none truncate px-1">
										{field.label || field.field_type} &middot; {signerName(field.signer_id)}
									</span>

									{#if isSelected}
										<button
											type="button"
											aria-label="Delete field"
											class="absolute -right-2 -top-2 flex size-5 items-center justify-center rounded-fc-pill bg-fc-danger text-fc-danger-fg transition-transform hover:scale-110"
											onpointerdown={(e: PointerEvent) => {
												e.preventDefault();
												e.stopPropagation();
												askDelete(field.id);
											}}
										>
											<iconify-icon icon={icons.close} width="14" height="14" class="pointer-events-none block"></iconify-icon>
										</button>

										<!-- svelte-ignore a11y_no_static_element_interactions -->
										<div
											class="absolute bottom-0 right-0 flex size-4 cursor-nwse-resize items-end justify-end"
											onpointerdown={(e) => handlePointerDown(e, field, 'resize')}
										>
											<svg width="10" height="10" viewBox="0 0 10 10" class="pointer-events-none">
												<path d="M10 0 L10 10 L0 10 Z" fill={color} opacity="0.9" />
											</svg>
										</div>
									{/if}
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>

			<div
				class="flex w-full shrink-0 flex-col gap-6 overflow-y-auto border-b border-fc-border p-4 md:w-72 md:border-b-0 md:border-l"
			>
				<div class="flex flex-col gap-2">
					<p class="text-fc-sm font-medium text-fc-fg">Signer</p>
					<div class="flex flex-col gap-1">
						{#each signers as signer, i}
							{@const color = USER_COLORS[i % USER_COLORS.length]}
							{@const isActive = selectedSignerId === signer.id}
							<button
								type="button"
								class={cn(
									'flex min-h-11 w-full items-center gap-2.5 rounded-fc-md px-3 py-2 text-left text-fc-sm transition-colors',
									isActive
										? 'bg-fc-accent font-medium text-fc-accent-fg'
										: 'text-fc-fg hover:bg-fc-surface'
								)}
								onclick={() => (selectedSignerId = signer.id)}
							>
								<span class="size-3 shrink-0 rounded-fc-pill" style="background: {color};"></span>
								<span class="min-w-0 flex-1 truncate">{signer.name}</span>
							</button>
						{/each}
					</div>
				</div>

				<div class="flex flex-col gap-2">
					<p class="text-fc-sm font-medium text-fc-fg">Add field</p>
					{#each FIELD_TYPES as ft}
						<Button
							variant="outline"
							icon={ft.icon}
							class="justify-start"
							onclick={() => addField(ft.type)}
						>
							{ft.label}
						</Button>
					{/each}
				</div>

				{#if fields.length > 0}
					<div class="flex flex-col gap-2">
						<p class="text-fc-sm font-medium text-fc-fg">Placed fields ({fields.length})</p>
						<div class="flex flex-col gap-1">
							{#each fields as field}
								{@const color = signerColor(field.signer_id)}
								{@const isActive = selectedFieldId === field.id}
								<div
									class={cn(
										'flex min-h-11 items-center gap-2 rounded-fc-md px-2 py-1 transition-colors',
										isActive ? 'bg-fc-surface' : 'hover:bg-fc-surface'
									)}
								>
									<span class="size-2.5 shrink-0 rounded-fc-pill" style="background: {color};"></span>
									{#if isActive}
										<Input
											value={field.label ?? ''}
											placeholder={field.field_type}
											aria-label="Field label"
											class="min-w-0 flex-1"
											onchange={(e) => renameField(field, (e.currentTarget as HTMLInputElement).value)}
										/>
									{:else}
										<button
											type="button"
											class="min-w-0 flex-1 truncate text-left text-fc-sm text-fc-fg"
											onclick={() => (selectedFieldId = field.id)}
										>
											{field.label || field.field_type}
										</button>
									{/if}
									<span class="shrink-0 text-fc-xs text-fc-fg-muted">p{field.page}</span>
									<IconButton
										variant="danger"
										aria-label="Delete field"
										onclick={() => askDelete(field.id)}
									>
										<iconify-icon icon={icons.remove} width="16" height="16" class="block"></iconify-icon>
									</IconButton>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<ConfirmModal
	bind:open={confirmDelete}
	tone="danger"
	title="Delete this field?"
	description="The field is removed from the document. Signers will no longer be asked to fill it."
	confirmLabel="Delete"
	onConfirm={deleteField}
	onCancel={() => (pendingDeleteId = null)}
/>

<style>
	/*
	 * The pages column renders PDF paper, which is white in both themes, so the field
	 * boxes drawn on top of it cannot take their label colour from `fc-fg` — that flips
	 * to near-white in dark mode and disappears against the page. Pinned ink, matching
	 * the signature pad's paper surface.
	 */
	.pdf-pages {
		--page-ink: oklch(0.145 0 0);
	}
</style>
