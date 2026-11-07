<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Space } from '$lib';
	import {
		Badge,
		Button,
		Card,
		ConfirmModal,
		EmptyState,
		Spinner,
		Table,
		icons,
		toast
	} from '@facile/muse';
	import { spaceStore } from '$lib/stores/space.svelte';
	import { formatShortDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	let spaces = $state<Space[]>([]);
	let loading = $state(true);
	let confirmLeave = $state(false);
	let pendingLeave = $state<Space | null>(null);
	let registreUrl = $state('');

	const roleTone = { owner: 'owner', admin: 'admin', member: 'neutral' } as const;

	function askLeave(space: Space) {
		pendingLeave = space;
		confirmLeave = true;
	}

	async function runLeave() {
		const target = pendingLeave;
		if (!target) return;
		try {
			await api.spaces.leave(target.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('spaces.leaveFailed'));
			throw e;
		}
		spaces = spaces.filter((s) => s.id !== target.id);
		spaceStore.spaces = spaces;
		if (spaceStore.activeId === target.id) {
			spaceStore.activeId = null;
		}
		pendingLeave = null;
		toast.success(t('spaces.left'));
	}

	onMount(async () => {
		try {
			const [list, cfg] = await Promise.all([
				api.spaces.list(),
				fetch('/api/auth/config').then((r) => r.json())
			]);
			spaces = list;
			spaceStore.spaces = spaces;
			registreUrl = cfg.registre_spaces_url ?? '';
		} catch {}
		loading = false;
	});
</script>

<svelte:head><title>{t('spaces.title')}</title></svelte:head>

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-2">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('spaces.heading')}</h1>
			<p class="text-fc-sm text-fc-fg-muted">
				{t('spaces.subtitle')}
				{#if registreUrl}{t('spaces.subtitleRegistre')}{/if}
			</p>
		</div>
		{#if registreUrl}
			<Button href={registreUrl} icon={icons.space}>{t('common.manageInRegistre')}</Button>
		{:else}
			<Button href="/spaces/new" icon={icons.plus}>{t('spaces.new')}</Button>
		{/if}
	</div>

	{#if loading}
		<div class="flex min-h-[40dvh] items-center justify-center">
			<Spinner size="lg" />
		</div>
	{:else if spaces.length === 0}
		<EmptyState
			icon={icons.space}
			title={t('spaces.empty.title')}
			description={registreUrl ? t('spaces.empty.registre') : t('spaces.empty.local')}
		>
			{#if registreUrl}
				<Button href={registreUrl} variant="outline" icon={icons.space}>
					{t('common.manageInRegistre')}
				</Button>
			{:else}
				<Button href="/spaces/new" variant="outline" icon={icons.plus}>{t('spaces.new')}</Button>
			{/if}
		</EmptyState>
	{:else}
		<div class="hidden md:block">
			<Table>
				<thead>
					<tr>
						<th scope="col">{t('spaces.col.space')}</th>
						<th scope="col">{t('spaces.col.role')}</th>
						<th scope="col">{t('common.created')}</th>
						<th aria-label={t('common.actions')}></th>
					</tr>
				</thead>
				<tbody>
					{#each spaces as space (space.id)}
						<tr>
							<td>
								<a
									href="/spaces/{space.id}"
									class="flex min-w-0 flex-col rounded-fc-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
								>
									<span class="truncate font-medium text-fc-fg">{space.name}</span>
									{#if space.description}
										<span class="truncate text-fc-xs text-fc-fg-muted">{space.description}</span>
									{/if}
								</a>
							</td>
							<td><Badge tone={roleTone[space.role]}>{t(`role.${space.role}`)}</Badge></td>
							<td class="whitespace-nowrap text-fc-fg-muted">{formatShortDate(space.created_at)}</td>
							<td class="text-right">
								{#if !space.registre_url}
									<Button
										variant="ghost-danger"
										size="sm"
										icon={icons.logout}
										aria-label={t('spaces.leaveAria', { name: space.name })}
										onclick={() => askLeave(space)}
									>
										{t('spaces.leave')}
									</Button>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		</div>

		<div class="flex flex-col gap-2 md:hidden">
			{#each spaces as space (space.id)}
				<Card class="flex flex-col gap-3">
					<div class="flex min-w-0 items-start justify-between gap-3">
						<a
							href="/spaces/{space.id}"
							class="flex min-w-0 flex-col rounded-fc-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
						>
							<span class="truncate text-fc-sm font-medium text-fc-fg">{space.name}</span>
							<span class="text-fc-xs text-fc-fg-muted">{formatShortDate(space.created_at)}</span>
						</a>
						<Badge tone={roleTone[space.role]}>{t(`role.${space.role}`)}</Badge>
					</div>
					{#if !space.registre_url}
						<div class="flex items-center justify-end">
							<Button
								variant="ghost-danger"
								size="lg"
								icon={icons.logout}
								aria-label={t('spaces.leaveAria', { name: space.name })}
								onclick={() => askLeave(space)}
							>
								{t('spaces.leave')}
							</Button>
						</div>
					{/if}
				</Card>
			{/each}
		</div>
	{/if}
</div>

<ConfirmModal
	bind:open={confirmLeave}
	tone="danger"
	title={t('spaces.confirm.title', { name: pendingLeave?.name ?? t('spaces.confirm.thisSpace') })}
	description={t('spaces.confirm.text')}
	confirmLabel={t('spaces.confirm.label')}
	cancelLabel={t('spaces.confirm.cancel')}
	onConfirm={runLeave}
	onCancel={() => (pendingLeave = null)}
/>
