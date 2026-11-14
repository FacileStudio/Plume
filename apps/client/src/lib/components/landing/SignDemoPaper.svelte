<script lang="ts">
	import { Bar, grow } from '@facile/mise-en-scene';
	import { USER_COLORS } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let { compact, placed, inked, signer }: {
		compact: boolean;
		placed: boolean;
		inked: boolean;
		signer: string;
	} = $props();
</script>

<div class="sd-paper" class:sd-paper-compact={compact}>
	<p class="sd-doc">{t('landing.demo.document')}</p>
	<div class="sd-lines">
		<Bar w="92%" /><Bar w="86%" /><Bar w="90%" /><Bar w="54%" />
	</div>
	<div class="sd-lines">
		<Bar w="88%" /><Bar w="94%" /><Bar w="70%" />
	</div>
	{#if !compact}
		<div class="sd-lines">
			<Bar w="90%" /><Bar w="82%" /><Bar w="40%" />
		</div>
	{/if}
	<div class="sd-slot">
		{#if placed}
			<div class="sd-field" style="--sd-signer: {USER_COLORS[0]}" in:grow={{ duration: 320 }}>
				<svg class="sd-ink" class:sd-inked={inked} viewBox="0 0 180 44" aria-hidden="true">
					<path
						pathLength="1"
						d="M8 30c10-18 18-22 20-14s-6 18 0 16 12-22 18-20-4 20 4 18 10-12 16-10 0 12 8 10 14-14 22-12 6 10 14 8 18-8 30-6 20 2 32-2"
					/>
				</svg>
				<span class="sd-field-label" class:sd-gone={inked}>{t('field.type.signature')} · {signer}</span>
			</div>
		{/if}
		<Bar w="38%" />
	</div>
</div>

<style>
	.sd-paper {
		display: flex;
		flex-direction: column;
		gap: 22px;
		width: 400px;
		height: 520px;
		padding: 36px 40px;
		border-radius: var(--radius-fc-md);
		background: var(--color-fc-page);
		box-shadow: 0 12px 32px -12px rgb(0 0 0 / 0.18);
	}
	.sd-paper-compact {
		gap: 16px;
		width: 100%;
		height: 360px;
		padding: 24px;
	}
	.sd-doc {
		margin: 0;
		font-family: var(--font-fc-title);
		font-size: 17px;
		font-weight: 700;
		white-space: nowrap;
	}
	.sd-lines {
		display: grid;
		gap: 9px;
	}
	.sd-slot {
		display: grid;
		gap: 10px;
		margin-top: auto;
		margin-bottom: 40px;
	}
	.sd-paper-compact .sd-slot {
		margin-bottom: 24px;
	}
	.sd-field {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 200px;
		height: 50px;
		border-radius: var(--radius-fc-xs);
		background: color-mix(in oklab, var(--sd-signer) 24%, transparent);
		font-size: 12px;
		font-weight: 700;
	}
	.sd-field-label {
		transition: opacity 0.3s ease;
	}
	.sd-gone {
		opacity: 0;
	}
	.sd-ink {
		position: absolute;
		inset: 3px 10px;
		width: calc(100% - 20px);
		height: calc(100% - 6px);
		fill: none;
		stroke: var(--color-fc-fg);
		stroke-width: 2.2;
		stroke-linecap: round;
		stroke-linejoin: round;
		stroke-dasharray: 1;
		stroke-dashoffset: 1;
	}
	.sd-inked {
		stroke-dashoffset: 0;
		transition: stroke-dashoffset 1s ease-in-out;
	}
	@media (prefers-reduced-motion: reduce) {
		.sd-inked,
		.sd-field-label {
			transition: none;
		}
	}
</style>
