<script lang="ts">
	import { Roll, Stage, Tile, Window, type Act } from '@facile/mise-en-scene';
	import { USER_COLORS } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	const signers = ['Camille Martin', 'Léa Dubois', 'Hugo Bernard'];

	let step = $state(0);
	let reminded = $state(false);

	function stateOf(index: number) {
		if (index < step) return t('status.signed');
		if (index === step) return reminded ? t('doc.signers.reminded') : t('status.pending');
		return t('landing.demo.queued');
	}

	function reset() {
		step = 0;
		reminded = false;
	}

	function poster() {
		step = signers.length;
		reminded = false;
	}

	async function script(act: Act) {
		await act.wait(900);
		step = 1;
		await act.wait(1300);
		reminded = true;
		await act.wait(1500);
		reminded = false;
		step = 2;
		await act.wait(1300);
		step = 3;
		await act.wait(2200);
		reset();
		await act.wait(500);
	}
</script>

<Tile title={t('landing.card.order.title')} body={t('landing.card.order.text')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<Window style="top:28px;left:20px;width:280px;height:184px">
				<ol class="ot-list">
					{#each signers as name, index (name)}
						<li class="ot-row" class:ot-done={index < step} class:ot-reminded={reminded && index === step}>
							<span class="ot-step" style="background: {USER_COLORS[index]}">{index + 1}</span>
							<span class="ot-name">{name}</span>
							<Roll value={stateOf(index)} class="ot-state" />
						</li>
					{/each}
				</ol>
			</Window>
		{/snippet}
	</Stage>
</Tile>

<style>
	.ot-list {
		display: grid;
		align-content: start;
		gap: 8px;
		margin: 0;
		padding: 12px;
		list-style: none;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.ot-row {
		display: flex;
		align-items: center;
		gap: 8px;
		height: 34px;
		padding: 0 10px;
		border-radius: var(--radius-fc-md);
		background: var(--color-fc-surface);
	}
	.ot-step {
		display: grid;
		place-items: center;
		width: 20px;
		height: 20px;
		flex: none;
		border-radius: 999px;
		color: var(--color-fc-accent-fg);
		font-size: 11px;
		font-weight: 700;
	}
	.ot-name {
		flex: 1;
		min-width: 0;
		white-space: nowrap;
	}
	.ot-row :global(.ot-state) {
		color: var(--color-fc-fg-muted);
		font-size: 11px;
		white-space: nowrap;
	}
	.ot-done :global(.ot-state) {
		color: var(--color-fc-success);
	}
	.ot-reminded :global(.ot-state) {
		color: var(--color-fc-info);
	}
</style>
