<script lang="ts">
	import { Bar } from '@facile/mise-en-scene';
	import { Button, USER_COLORS } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let {
		shown,
		inked,
		compact,
		button = $bindable()
	}: {
		shown: boolean;
		inked: boolean;
		compact: boolean;
		button?: HTMLElement;
	} = $props();
</script>

<div
	class="sp-phone"
	class:sp-shown={shown}
	style={compact ? 'top:96px;left:69px' : 'top:118px;left:664px'}
	aria-hidden="true"
>
	<div class="sp-notch"></div>
	<p class="sp-doc">{t('landing.demo.document')}</p>
	<div class="sp-lines">
		<Bar w="94%" /><Bar w="88%" /><Bar w="92%" /><Bar w="60%" />
	</div>
	<div class="sp-lines">
		<Bar w="90%" /><Bar w="72%" />
	</div>
	<div class="sp-field" style="--sp-signer: {USER_COLORS[0]}">
		<svg class="sp-ink" class:sp-inked={inked} viewBox="0 0 180 44">
			<path
				pathLength="1"
				d="M8 30c10-18 18-22 20-14s-6 18 0 16 12-22 18-20-4 20 4 18 10-12 16-10 0 12 8 10 14-14 22-12 6 10 14 8 18-8 30-6 20 2 32-2"
			/>
		</svg>
	</div>
	<span bind:this={button} class="sp-button">
		<Button size="sm" class="w-full">{t('sign.submit')}</Button>
	</span>
</div>

<style>
	.sp-phone {
		position: absolute;
		display: flex;
		flex-direction: column;
		gap: 14px;
		width: 220px;
		height: 420px;
		padding: 14px 16px 18px;
		border-radius: 30px;
		background: var(--color-fc-page);
		box-shadow:
			0 0 0 6px var(--color-fc-component),
			0 24px 48px -16px rgb(0 0 0 / 0.35);
		opacity: 0;
		transform: translateY(24px) scale(0.96);
		transition:
			opacity 360ms ease,
			transform 420ms cubic-bezier(0.2, 0.8, 0.2, 1);
		font-family: var(--font-fc-body);
		color: var(--color-fc-fg);
	}
	.sp-shown {
		opacity: 1;
		transform: none;
	}
	.sp-notch {
		align-self: center;
		width: 64px;
		height: 6px;
		border-radius: 999px;
		background: var(--color-fc-component);
	}
	.sp-doc {
		margin: 0;
		font-size: 13px;
		font-weight: 700;
	}
	.sp-lines {
		display: grid;
		gap: 8px;
	}
	.sp-field {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 72px;
		margin-top: auto;
		border-radius: var(--radius-fc-md);
		background: color-mix(in srgb, var(--sp-signer) 14%, transparent);
	}
	.sp-ink {
		width: 150px;
		height: 40px;
	}
	.sp-ink path {
		fill: none;
		stroke: var(--color-fc-fg);
		stroke-width: 2.4;
		stroke-linecap: round;
		stroke-dasharray: 1;
		stroke-dashoffset: 1;
	}
	.sp-inked path {
		stroke-dashoffset: 0;
		transition: stroke-dashoffset 900ms ease-out;
	}
	.sp-button {
		display: block;
	}
	@media (prefers-reduced-motion: reduce) {
		.sp-phone,
		.sp-inked path {
			transition: none;
		}
	}
</style>
