<script lang="ts">
	import { Bar, Roll, Stage, Tile, type Act } from '@facile/mise-en-scene';
	import { Icon } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let signed = $state(false);
	let filed = $state(false);

	function reset() {
		signed = false;
		filed = false;
	}

	function poster() {
		signed = true;
		filed = true;
	}

	async function script(act: Act) {
		await act.wait(1200);
		signed = true;
		await act.wait(1100);
		filed = true;
		await act.wait(2600);
		reset();
		await act.wait(600);
	}
</script>

<Tile title={t('landing.card.ardoise.title')} body={t('landing.card.ardoise.text')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<div class="at-card" style="top:14px;left:20px;width:280px">
				<p class="at-app"><Icon icon="solar:bill-list-bold-duotone" size={16} />Ardoise</p>
				<div class="at-row">
					<span class="at-name">{t('landing.demo.document')}</span>
					<span class="at-state" class:at-signed={signed}>
						<Roll value={signed ? t('status.signed') : t('landing.demo.sentFromArdoise')} />
					</span>
				</div>
			</div>
			<div class="at-card" style="top:108px;left:20px;width:280px">
				<p class="at-app"><Icon icon="solar:cloud-bold-duotone" size={16} />Nuage<span class="at-path">/ Plume</span></p>
				<div class="at-row at-small">
					<Icon icon="solar:file-text-linear" size={16} />
					<Bar w="44%" />
					<span class="at-ext">PDF</span>
				</div>
				<div class="at-row at-small at-file" class:at-on={filed}>
					<Icon icon="solar:file-text-linear" size={16} />
					<Bar w="60%" />
					<span class="at-ext">PDF</span>
				</div>
			</div>
		{/snippet}
	</Stage>
</Tile>

<style>
	.at-card {
		position: absolute;
		display: grid;
		gap: 10px;
		padding: 12px 14px;
		border-radius: var(--radius-fc-lg);
		background: var(--color-fc-page);
		box-shadow: 0 10px 28px -12px rgb(0 0 0 / 0.25);
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.at-app {
		display: flex;
		align-items: center;
		gap: 6px;
		margin: 0;
		font-weight: 700;
	}
	.at-path {
		color: var(--color-fc-fg-muted);
		font-weight: 400;
	}
	.at-row {
		display: flex;
		align-items: center;
		gap: 8px;
		height: 34px;
		padding: 0 10px;
		border-radius: var(--radius-fc-md);
		background: var(--color-fc-surface);
	}
	.at-small {
		height: 30px;
	}
	.at-name {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.at-state {
		flex: none;
		color: var(--color-fc-info);
		font-size: 11px;
		white-space: nowrap;
		transition: color 0.3s ease;
	}
	.at-signed {
		color: var(--color-fc-success);
	}
	.at-file {
		opacity: 0;
		transform: translateY(6px);
		transition:
			opacity 0.35s ease,
			transform 0.35s ease;
	}
	.at-on {
		opacity: 1;
		transform: none;
	}
	.at-ext {
		margin-left: auto;
		color: var(--color-fc-fg-muted);
		font-size: 11px;
	}
</style>
