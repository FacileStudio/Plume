<script lang="ts">
	import { Badge, Button, Card, icons } from '@facile/muse';
	import type { PendingAction } from '$lib';
	import { formatDateTime } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	let {
		action,
		busy,
		onconfirm,
		ondismiss
	}: {
		action: PendingAction;
		busy: boolean;
		onconfirm: (action: PendingAction) => void;
		ondismiss: (action: PendingAction) => void;
	} = $props();

	const kind = $derived(
		action.kind === 'send_document' ? 'send' : action.kind === 'remind_signer' ? 'remind' : 'delete'
	);
	const tone = $derived(kind === 'delete' ? 'danger' : kind === 'send' ? 'accent' : 'info');
	const doc = $derived(action.document);
	const statusKey = $derived(
		({ draft: 'status.draft', pending: 'status.pending', completed: 'status.completed', declined: 'status.declined' } as const)[
			(doc?.status ?? 'draft') as 'draft' | 'pending' | 'completed' | 'declined'
		] ?? 'status.draft'
	);

	/** Names the client that asked: the host of a client URL, else its id. */
	function clientName(clientId: string): string {
		if (clientId === 'bulle') return 'Bulle';
		try {
			return new URL(clientId).host;
		} catch {
			return clientId || '?';
		}
	}

	/** The line under the target that says what confirming does. */
	function note(): string {
		if (kind === 'send') return t('pending.sendNote');
		if (kind === 'remind') return t('pending.remindNote', { name: action.signer?.name ?? '' });
		return t('pending.deleteNote');
	}
</script>

<Card class="flex flex-col gap-4">
	<div class="flex flex-wrap items-center justify-between gap-2">
		<Badge {tone} class="shrink-0 whitespace-nowrap">{t(`pending.kind.${kind}`)}</Badge>
		<span class="text-fc-xs text-fc-fg-muted">
			{t('pending.askedBy', { client: clientName(action.client_id) })} · {formatDateTime(action.created_at)}
		</span>
	</div>

	{#if doc}
		<a href={`/documents/${doc.id}`} class="flex items-start gap-3 rounded-fc-md bg-fc-surface p-4">
			<iconify-icon icon="solar:document-linear" width="24" height="24" class="shrink-0 text-fc-fg-muted"></iconify-icon>
			<div class="flex min-w-0 flex-1 flex-col gap-1">
				<p class="break-words text-fc-md font-medium text-fc-fg">{doc.name}</p>
				{#if action.signer}
					<p class="break-words text-fc-sm text-fc-fg-muted">{action.signer.name} · {action.signer.email}</p>
				{:else}
					{#each doc.signers as signer (signer)}
						<p class="break-words text-fc-sm text-fc-fg-muted">{signer}</p>
					{/each}
				{/if}
			</div>
			<Badge class="shrink-0 whitespace-nowrap">{t(statusKey)}</Badge>
		</a>
		<p class="text-fc-sm text-fc-fg-muted">{note()}</p>
	{:else}
		<p class="break-words text-fc-sm text-fc-fg-muted">{action.summary}</p>
	{/if}

	{#if action.state === 'changed'}
		<p class="text-fc-sm text-fc-warning">{t('pending.changed')}</p>
	{:else if action.state === 'missing'}
		<p class="text-fc-sm text-fc-warning">{t('pending.missing')}</p>
	{/if}

	<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
		<Button variant="ghost" class="w-full sm:w-auto [@media(pointer:coarse)]:h-11" disabled={busy} onclick={() => ondismiss(action)}>
			{t('pending.dismiss')}
		</Button>
		{#if action.state === 'ready'}
			<Button
				variant={kind === 'delete' ? 'danger' : 'primary'}
				class="w-full sm:w-auto [@media(pointer:coarse)]:h-11"
				disabled={busy}
				icon={kind === 'delete' ? icons.remove : icons.mail}
				onclick={() => onconfirm(action)}
			>
				{t(`pending.confirm.${kind}`)}
			</Button>
		{/if}
	</div>
</Card>
