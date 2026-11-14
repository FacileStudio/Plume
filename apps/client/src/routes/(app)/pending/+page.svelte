<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Alert, Button, EmptyState, Spinner, icons, toast } from '@facile/muse';
	import { ApiError, api, type PendingAction } from '$lib';
	import PendingCard from '$lib/components/PendingCard.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let actions = $state<PendingAction[]>([]);
	let loading = $state(true);
	let busy = $state('');
	let notOwnSession = $state(false);

	const focused = $derived(page.url.searchParams.get('action') ?? '');
	const ordered = $derived([
		...actions.filter((action) => action.id === focused),
		...actions.filter((action) => action.id !== focused)
	]);

	onMount(load);

	/**
	 * Reports whether err is porte's refusal of a session that is not the person's own
	 * browser session: a cookie from before the upgrade, a CLI or API token session, or
	 * the mobile or desktop app. Signing in again in the browser fixes it.
	 */
	function refusedSession(err: unknown): boolean {
		notOwnSession = err instanceof ApiError && err.status === 403;
		return notOwnSession;
	}

	/** Loads the open actions. */
	async function load() {
		loading = true;
		try {
			actions = (await api.pending.list()).actions;
			notOwnSession = false;
		} catch (err) {
			if (!refusedSession(err)) toast.danger(t('pending.loadFailed'));
		}
		loading = false;
	}

	const doneKeys = {
		send_document: 'pending.done.send',
		remind_signer: 'pending.done.remind',
		delete_document: 'pending.done.delete',
		delete_client: 'pending.done.deleteClient',
		delete_field: 'pending.done.deleteField'
	} as const;

	/** The toast once an action of this kind ran. */
	function doneMessage(action: PendingAction): string {
		return t(doneKeys[action.kind] ?? 'pending.done.delete');
	}

	/** Runs one action once, then drops it from the list. */
	async function confirm(action: PendingAction) {
		busy = action.id;
		try {
			await api.pending.confirm(action.id);
			toast.success(doneMessage(action));
			actions = actions.filter((item) => item.id !== action.id);
		} catch (err) {
			if (!refusedSession(err)) {
				toast.danger(err instanceof Error ? err.message : t('pending.failed'));
				await load();
			}
		}
		busy = '';
	}

	/** Closes one action without running it. */
	async function dismiss(action: PendingAction) {
		busy = action.id;
		try {
			await api.pending.dismiss(action.id);
			toast.success(t('pending.dismissed'));
			actions = actions.filter((item) => item.id !== action.id);
		} catch (err) {
			if (!refusedSession(err)) toast.danger(err instanceof Error ? err.message : t('pending.failed'));
		}
		busy = '';
	}
</script>

<svelte:head>
	<title>{t('pending.title')}</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-2xl flex-col gap-6">
	<div class="flex flex-col gap-2">
		<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('pending.heading')}</h1>
		<p class="text-fc-sm text-fc-fg-muted">{t('pending.intro')}</p>
	</div>

	{#if notOwnSession}
		<Alert tone="warning" class="flex flex-col items-start gap-3">
			<p>{t('pending.signInAgain')}</p>
			<Button href="/api/auth/oidc" variant="primary">{t('pending.signIn')}</Button>
		</Alert>
	{/if}

	{#if loading}
		<div class="flex h-48 items-center justify-center"><Spinner /></div>
	{:else if actions.length === 0}
		<EmptyState icon={icons.check} title={t('pending.none')} description={t('pending.noneBody')} />
	{:else}
		<section class="flex flex-col gap-4">
			{#each ordered as action (action.id)}
				<PendingCard {action} busy={busy === action.id} onconfirm={confirm} ondismiss={dismiss} />
			{/each}
		</section>
	{/if}
</div>
