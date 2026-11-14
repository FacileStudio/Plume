<script lang="ts">
	import { Bar, Stage, Tile, Window, type Act } from '@facile/mise-en-scene';
	import { Badge, Icon, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	const trail = [
		{ time: '09:12', w: '70%' },
		{ time: '09:40', w: '56%' },
		{ time: '10:05', w: '64%' }
	];

	let shown = $state(1);
	let verified = $state(false);

	function reset() {
		shown = 1;
		verified = false;
	}

	function poster() {
		shown = trail.length;
		verified = true;
	}

	async function script(act: Act) {
		await act.wait(700);
		for (let i = 2; i <= trail.length; i++) {
			shown = i;
			await act.wait(700);
		}
		await act.wait(400);
		verified = true;
		await act.wait(2400);
		reset();
		await act.wait(500);
	}
</script>

<Tile title={t('landing.card.proof.title')} body={t('landing.card.proof.text')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<Window style="top:12px;left:20px;width:280px;height:216px">
				<div class="pt-body">
					<p class="pt-head">
						<Icon icon={icons.shield} size={16} />
						<span>{t('landing.demo.certificate')}</span>
						<span class="pt-pdf">PDF</span>
					</p>
					<p class="pt-label">{t('landing.demo.audit')}</p>
					<ul class="pt-trail">
						{#each trail as row, index (row.time)}
							<li class="pt-row" class:pt-on={index < shown}>
								<span class="pt-time">{row.time}</span>
								<Bar w={row.w} />
							</li>
						{/each}
					</ul>
					<span class="pt-badge" class:pt-on={verified}>
						<Badge tone="success">{t('landing.demo.verified')}</Badge>
					</span>
				</div>
			</Window>
		{/snippet}
	</Stage>
</Tile>

<style>
	.pt-body {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 12px 14px;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.pt-head {
		display: flex;
		align-items: center;
		gap: 6px;
		margin: 0;
		font-weight: 700;
	}
	.pt-pdf {
		margin-left: auto;
		color: var(--color-fc-fg-muted);
		font-size: 11px;
		font-weight: 400;
	}
	.pt-label {
		margin: 2px 0 0;
		color: var(--color-fc-fg-muted);
		font-size: 11px;
	}
	.pt-trail {
		display: grid;
		gap: 6px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.pt-row {
		display: flex;
		align-items: center;
		gap: 10px;
		height: 24px;
		padding: 0 8px;
		border-radius: var(--radius-fc-sm);
		background: var(--color-fc-surface);
		opacity: 0;
		transform: translateX(-6px);
		transition:
			opacity 0.3s ease,
			transform 0.3s ease;
	}
	.pt-time {
		flex: none;
		color: var(--color-fc-fg-muted);
		font-family: var(--font-fc-mono);
		font-size: 11px;
	}
	.pt-badge {
		align-self: flex-end;
		opacity: 0;
		transform: scale(0.9);
		transition:
			opacity 0.3s ease,
			transform 0.3s ease;
	}
	.pt-on {
		opacity: 1;
		transform: none;
	}
</style>
