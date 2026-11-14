<script lang="ts">
	import { Cursor, Roll, Stage, Window, cursorAt, grow, type Act } from '@facile/mise-en-scene';
	import { Badge, Button, USER_COLORS } from '@facile/muse';
	import { tick } from 'svelte';
	import SignDemoPaper from './SignDemoPaper.svelte';
	import { t } from '$lib/i18n/index.svelte';

	type Status = 'draft' | 'pending' | 'completed';

	const tone = { draft: 'neutral', pending: 'info', completed: 'success' } as const;
	const signer = 'Camille Martin';

	let placed = $state(false);
	let status = $state<Status>('draft');
	let inked = $state(false);
	let cursor = $state(cursorAt(700, 420));
	let tool = $state<HTMLElement>();
	let send = $state<HTMLElement>();

	const signerState = $derived(
		inked ? t('status.signed') : status === 'pending' ? t('status.pending') : t('status.draft')
	);

	function reset() {
		placed = false;
		status = 'draft';
		inked = false;
		cursor = cursorAt(700, 420);
	}

	function poster() {
		placed = true;
		status = 'completed';
		inked = true;
	}

	async function script(act: Act) {
		await act.wait(700);
		cursor.visible = true;
		await act.move(cursor, act.point(tool, 0.5, 0.6));
		await act.click(cursor, () => (placed = true));
		await act.wait(700);
		await tick();
		await act.move(cursor, act.point(send, 0.5, 0.6));
		await act.click(cursor, () => (status = 'pending'));
		await act.wait(300);
		cursor.visible = false;
		await act.wait(1100);
		inked = true;
		await act.wait(1100);
		status = 'completed';
		await act.wait(2200);
		reset();
		await act.wait(600);
	}
</script>

{#snippet signerRow()}
	<div class="sd-signer">
		<span class="sd-dot" style="background: {USER_COLORS[0]}"></span>
		<span class="sd-name">{signer}</span>
		<Roll value={signerState} class="sd-state" />
	</div>
{/snippet}

{#snippet badge()}
	<Badge tone={tone[status]}>{t(`status.${status}`)}</Badge>
{/snippet}

<Stage
	width={960}
	height={580}
	compact={{ below: 640, width: 358, height: 560 }}
	{script}
	{reset}
	{poster}
	label={t('landing.demo.label')}
>
	{#snippet children({ compact })}
		<Window
			style={compact
				? 'top:12px;left:12px;width:334px;height:536px'
				: 'top:24px;left:60px;width:840px;height:532px'}
			trailing={badge}
		>
			<div class="sd-body" class:sd-compact={compact}>
				<div class="sd-pages"><SignDemoPaper {compact} {placed} {inked} {signer} /></div>
				<div class="sd-panel">
					<div class="sd-group">
						<p class="sd-heading">{t('editor.signer')}</p>
						{@render signerRow()}
					</div>
					<div class="sd-group">
						<p class="sd-heading">{t('editor.addField')}</p>
						<span bind:this={tool} class="sd-tool">
							<Button variant="secondary" size="sm" icon="solar:pen-new-round-linear" class="w-full justify-start">
								{t('field.type.signature')}
							</Button>
						</span>
					</div>
					{#if !compact}
						<div class="sd-group">
							<p class="sd-heading">{t('editor.placed', { count: placed ? 1 : 0 })}</p>
							{#if placed}
								<div class="sd-placed" in:grow={{ duration: 320 }}>
									<span class="sd-row">
										<span class="sd-dot sd-dot-small" style="background: {USER_COLORS[0]}"></span>
										<span class="sd-name">{t('field.type.signature')}</span>
										<span class="sd-page">{t('editor.pageShort', { page: 1 })}</span>
									</span>
								</div>
							{/if}
						</div>
					{/if}
					<span bind:this={send} class="sd-send">
						<Button icon="solar:plain-linear" class="w-full">{t('doc.send')}</Button>
					</span>
				</div>
			</div>
		</Window>
		<Cursor {cursor} />
	{/snippet}
</Stage>

<style>
	.sd-body {
		display: flex;
		flex: 1;
		min-height: 0;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
	}
	.sd-compact {
		flex-direction: column;
	}
	.sd-pages {
		display: flex;
		flex: 1;
		align-items: flex-start;
		justify-content: center;
		min-height: 0;
		padding: 28px 24px 0;
		overflow: hidden;
		background: var(--color-fc-surface);
	}
	.sd-compact .sd-pages {
		padding: 16px 16px 0;
	}
	.sd-panel {
		display: flex;
		flex: none;
		flex-direction: column;
		gap: 24px;
		width: 272px;
		padding: 20px 16px;
	}
	.sd-compact .sd-panel {
		gap: 12px;
		width: auto;
		padding: 14px 16px 16px;
	}
	.sd-group {
		display: grid;
		gap: 8px;
	}
	.sd-heading {
		margin: 0;
		font-size: 13px;
		font-weight: 700;
	}
	.sd-compact .sd-heading {
		display: none;
	}
	.sd-signer {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: 44px;
		padding: 0 12px;
		border-radius: var(--radius-fc-md);
		background: var(--color-fc-surface);
		font-size: 13px;
	}
	.sd-dot {
		width: 12px;
		height: 12px;
		flex: none;
		border-radius: 999px;
	}
	.sd-name {
		flex: 1;
		min-width: 0;
		white-space: nowrap;
	}
	.sd-signer :global(.sd-state) {
		color: var(--color-fc-fg-muted);
		font-size: 12px;
		white-space: nowrap;
	}
	.sd-placed {
		padding-bottom: 2px;
	}
	.sd-row {
		display: flex;
		align-items: center;
		gap: 8px;
		min-height: 40px;
		padding: 0 10px;
		border-radius: var(--radius-fc-md);
		background: var(--color-fc-surface);
		font-size: 13px;
	}
	.sd-dot-small {
		width: 10px;
		height: 10px;
	}
	.sd-page {
		color: var(--color-fc-fg-muted);
		font-size: 12px;
	}
	.sd-tool,
	.sd-send {
		display: block;
	}
	.sd-send {
		margin-top: auto;
	}
</style>
