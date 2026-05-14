<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';

	let {
		value = $bindable(''),
		width = 400,
		height = 160
	}: {
		value: string;
		width?: number;
		height?: number;
	} = $props();

	let canvas = $state<HTMLCanvasElement>(undefined!);
	let ctx = $state<CanvasRenderingContext2D>(undefined!);
	let drawing = $state(false);
	let hasStrokes = $state(false);
	let points: { x: number; y: number; time: number }[] = [];
	let strokes: { x: number; y: number; time: number }[][] = [];
	let ratio = 1;

	function getPos(e: MouseEvent | TouchEvent): { x: number; y: number } {
		const rect = canvas.getBoundingClientRect();
		const scaleX = canvas.width / ratio / rect.width;
		const scaleY = canvas.height / ratio / rect.height;

		if ('touches' in e) {
			const touch = e.touches[0];
			return {
				x: (touch.clientX - rect.left) * scaleX,
				y: (touch.clientY - rect.top) * scaleY
			};
		}
		return {
			x: (e.clientX - rect.left) * scaleX,
			y: (e.clientY - rect.top) * scaleY
		};
	}

	function getStrokeWidth(velocity: number): number {
		const minWidth = 1.2;
		const maxWidth = 3.5;
		const v = Math.min(velocity, 6);
		return maxWidth - ((maxWidth - minWidth) * v) / 6;
	}

	function drawQuadratic(
		p0: { x: number; y: number },
		p1: { x: number; y: number },
		p2: { x: number; y: number },
		w1: number,
		w2: number
	) {
		const steps = 8;
		for (let i = 0; i < steps; i++) {
			const t1 = i / steps;
			const t2 = (i + 1) / steps;

			const x1 = (1 - t1) * (1 - t1) * p0.x + 2 * (1 - t1) * t1 * p1.x + t1 * t1 * p2.x;
			const y1 = (1 - t1) * (1 - t1) * p0.y + 2 * (1 - t1) * t1 * p1.y + t1 * t1 * p2.y;
			const x2 = (1 - t2) * (1 - t2) * p0.x + 2 * (1 - t2) * t2 * p1.x + t2 * t2 * p2.x;
			const y2 = (1 - t2) * (1 - t2) * p0.y + 2 * (1 - t2) * t2 * p1.y + t2 * t2 * p2.y;
			const w = w1 + (w2 - w1) * t1;

			ctx.beginPath();
			ctx.moveTo(x1 * ratio, y1 * ratio);
			ctx.lineTo(x2 * ratio, y2 * ratio);
			ctx.lineWidth = w * ratio;
			ctx.stroke();
		}
	}

	function startStroke(e: MouseEvent | TouchEvent) {
		e.preventDefault();
		drawing = true;
		const pos = getPos(e);
		points = [{ ...pos, time: Date.now() }];
	}

	function moveStroke(e: MouseEvent | TouchEvent) {
		if (!drawing) return;
		e.preventDefault();
		const pos = getPos(e);
		const now = Date.now();
		points.push({ ...pos, time: now });

		const len = points.length;
		if (len < 3) {
			ctx.beginPath();
			ctx.moveTo(points[0].x * ratio, points[0].y * ratio);
			ctx.lineTo(pos.x * ratio, pos.y * ratio);
			ctx.lineWidth = 2.5 * ratio;
			ctx.stroke();
			return;
		}

		const p0 = points[len - 3];
		const p1 = points[len - 2];
		const p2 = points[len - 1];

		const mid1 = { x: (p0.x + p1.x) / 2, y: (p0.y + p1.y) / 2 };
		const mid2 = { x: (p1.x + p2.x) / 2, y: (p1.y + p2.y) / 2 };

		const dt = (p2.time - p0.time) || 1;
		const dx = p2.x - p0.x;
		const dy = p2.y - p0.y;
		const velocity = Math.sqrt(dx * dx + dy * dy) / dt * 10;

		const w = getStrokeWidth(velocity);
		drawQuadratic(mid1, p1, mid2, w, w);
	}

	function endStroke() {
		if (!drawing) return;
		drawing = false;

		if (points.length > 0) {
			strokes.push([...points]);
			hasStrokes = strokes.length > 0;
			exportImage();
		}
		points = [];
	}

	function exportImage() {
		const tempCanvas = document.createElement('canvas');
		tempCanvas.width = canvas.width;
		tempCanvas.height = canvas.height;
		const tempCtx = tempCanvas.getContext('2d')!;
		tempCtx.fillStyle = '#ffffff';
		tempCtx.fillRect(0, 0, tempCanvas.width, tempCanvas.height);
		tempCtx.drawImage(canvas, 0, 0);
		value = tempCanvas.toDataURL('image/png');
	}

	function redrawAll() {
		ctx.clearRect(0, 0, canvas.width, canvas.height);
		ctx.strokeStyle = '#000000';
		ctx.lineCap = 'round';
		ctx.lineJoin = 'round';

		for (const stroke of strokes) {
			if (stroke.length < 2) continue;
			for (let i = 2; i < stroke.length; i++) {
				const p0 = stroke[i - 2];
				const p1 = stroke[i - 1];
				const p2 = stroke[i];
				const mid1 = { x: (p0.x + p1.x) / 2, y: (p0.y + p1.y) / 2 };
				const mid2 = { x: (p1.x + p2.x) / 2, y: (p1.y + p2.y) / 2 };
				const dt = (p2.time - p0.time) || 1;
				const dx = p2.x - p0.x;
				const dy = p2.y - p0.y;
				const velocity = Math.sqrt(dx * dx + dy * dy) / dt * 10;
				const w = getStrokeWidth(velocity);
				drawQuadratic(mid1, p1, mid2, w, w);
			}
			if (stroke.length === 2) {
				ctx.beginPath();
				ctx.moveTo(stroke[0].x * ratio, stroke[0].y * ratio);
				ctx.lineTo(stroke[1].x * ratio, stroke[1].y * ratio);
				ctx.lineWidth = 2.5 * ratio;
				ctx.stroke();
			}
		}
	}

	function clear() {
		strokes = [];
		points = [];
		hasStrokes = false;
		ctx.clearRect(0, 0, canvas.width, canvas.height);
		value = '';
	}

	function undo() {
		if (strokes.length === 0) return;
		strokes.pop();
		hasStrokes = strokes.length > 0;
		redrawAll();
		if (hasStrokes) {
			exportImage();
		} else {
			value = '';
		}
	}

	onMount(() => {
		ratio = window.devicePixelRatio || 1;
		canvas.width = width * ratio;
		canvas.height = height * ratio;
		ctx = canvas.getContext('2d')!;
		ctx.strokeStyle = '#000000';
		ctx.lineCap = 'round';
		ctx.lineJoin = 'round';
	});
</script>

<div class="space-y-2">
	<div
		class="relative rounded-lg border-2 border-dashed border-border bg-white transition-colors"
		class:border-foreground={drawing}
		style="width: 100%; aspect-ratio: {width}/{height};"
	>
		<canvas
			bind:this={canvas}
			style="width: 100%; height: 100%; display: block; touch-action: none; cursor: crosshair;"
			onmousedown={startStroke}
			onmousemove={moveStroke}
			onmouseup={endStroke}
			onmouseleave={endStroke}
			ontouchstart={startStroke}
			ontouchmove={moveStroke}
			ontouchend={endStroke}
		></canvas>

		{#if !hasStrokes && !drawing}
			<div class="pointer-events-none absolute inset-0 flex items-center justify-center">
				<span class="text-sm text-muted-foreground/50">Draw your signature here</span>
			</div>
		{/if}

		<div class="absolute bottom-0 left-0 right-0 mx-4 border-t border-muted-foreground/20"></div>
	</div>

	<div class="flex items-center gap-2">
		<button
			type="button"
			onclick={undo}
			disabled={!hasStrokes}
			class="inline-flex items-center gap-1 rounded-md border px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
		>
			<Icon icon="solar:undo-left-linear" class="h-3.5 w-3.5" />
			Undo
		</button>
		<button
			type="button"
			onclick={clear}
			disabled={!hasStrokes}
			class="inline-flex items-center gap-1 rounded-md border px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
		>
			<Icon icon="solar:eraser-linear" class="h-3.5 w-3.5" />
			Clear
		</button>
	</div>
</div>
