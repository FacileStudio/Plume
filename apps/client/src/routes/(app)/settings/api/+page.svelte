<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Alert,
		Button,
		ConfirmModal,
		Drawer,
		Input,
		SecretField,
		SettingsRow,
		SettingsSection,
		StatusDot,
		Table,
		icons,
		toast
	} from '@facile/muse';
	import { api, type ApiToken } from '$lib/backend';

	let tokens = $state<ApiToken[]>([]);
	let busy = $state(false);
	let createOpen = $state(false);
	let revoking = $state<ApiToken | null>(null);
	let revokeOpen = $state(false);
	let label = $state('');
	let issued = $state('');

	onMount(load);

	async function load() {
		try {
			tokens = await api.auth.apiTokens();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to load tokens');
		}
	}

	function openCreate() {
		issued = '';
		label = '';
		createOpen = true;
	}

	async function generate() {
		busy = true;
		try {
			const created = await api.auth.createApiToken(label.trim());
			issued = created.token;
			toast.success('Token generated');
			await load();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to generate the token');
		} finally {
			busy = false;
		}
	}

	async function revoke() {
		if (!revoking) return;
		const target = revoking;
		busy = true;
		try {
			await api.auth.revokeApiToken(target.id);
			toast.success('Token revoked');
			await load();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to revoke the token');
		} finally {
			busy = false;
		}
	}

	function formatDate(value: string) {
		return new Date(value).toLocaleDateString();
	}
</script>

<svelte:head><title>API — Plume</title></svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Reference"
		description="Every route this instance serves, generated from the handlers themselves rather than kept by hand."
	>
		{#snippet actions()}
			<Button href="/docs" target="_blank" rel="noopener" icon={icons.code}>Open the reference</Button>
		{/snippet}

		<SettingsRow
			label="Base path"
			description="Every endpoint is scoped under it, on this same origin."
		>
			<span class="font-fc-mono text-fc-sm text-fc-fg">/api</span>
		</SettingsRow>

		<SettingsRow
			label="OpenAPI document"
			description="The machine-readable schema behind the reference — feed it to a client generator."
		>
			<Button variant="outline" href="/docs/openapi.json" target="_blank" rel="noopener" icon={icons.download}>
				openapi.json
			</Button>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection
		title="Authentication"
		description="How a request proves who it is."
	>
		<SettingsRow
			label="Scheme"
			description="Send the token as Authorization: Bearer <token>. Browser logins and API tokens are accepted the same way."
		>
			<span class="font-fc-mono text-fc-sm text-fc-fg">Bearer</span>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection
		title="API tokens"
		description="Named machine credentials for other apps and scripts, such as Ardoise. A token acts as your account, has no expiry and survives a sign-out."
	>
		{#snippet actions()}
			<Button icon={icons.plus} onclick={openCreate}>Generate a token</Button>
		{/snippet}

		{#if tokens.length}
			<Table>
				<thead>
					<tr>
						<th>Name</th>
						<th>Created</th>
						<th>Last used</th>
						<th class="text-right">Action</th>
					</tr>
				</thead>
				<tbody>
					{#each tokens as token (token.id)}
						<tr>
							<td class="font-fc-mono">{token.label}</td>
							<td>{formatDate(token.created_at)}</td>
							<td>{formatDate(token.last_used_at)}</td>
							<td class="text-right">
								<Button
									variant="ghost-danger"
									size="sm"
									icon={icons.revoke}
									disabled={busy}
									onclick={() => {
										revoking = token;
										revokeOpen = true;
									}}
								>
									Revoke
								</Button>
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		{:else}
			<SettingsRow
				label="No token"
				description="A token is shown once, when you create it. Plume stores only a hash and cannot show it again."
			>
				<StatusDot tone="neutral" label="None generated" />
			</SettingsRow>
		{/if}
	</SettingsSection>
</div>

<Drawer
	bind:open={createOpen}
	title="Generate an API token"
	description="Name it after the app or machine that will use it. The name is the only way to know later what you are revoking."
	onClose={() => (issued = '')}
>
	{#if issued}
		<div class="flex flex-col gap-4">
			<Alert tone="warning" title="Copy it now">
				This token is never shown again. Plume keeps only a hash, so once this drawer closes you
				will have to generate another.
			</Alert>
			<SecretField value={issued} label="Token" visible autoHideMs={0} />
		</div>
	{:else}
		<div class="flex flex-col gap-2">
			<label for="token-label" class="text-fc-sm font-medium text-fc-fg">Token name</label>
			<Input id="token-label" bind:value={label} placeholder="ardoise, ci, laptop" />
			<p class="text-fc-xs text-fc-fg-muted">
				A token carries everything your account can do and has no scope or expiry. Revoke it as
				soon as the app that uses it is gone.
			</p>
		</div>
	{/if}

	{#snippet footer()}
		{#if issued}
			<Button variant="outline" onclick={() => (createOpen = false)}>Close</Button>
		{:else}
			<Button variant="outline" onclick={() => (createOpen = false)}>Cancel</Button>
			<Button icon={icons.key} disabled={busy || !label.trim()} onclick={generate}>
				{busy ? 'Generating…' : 'Generate'}
			</Button>
		{/if}
	{/snippet}
</Drawer>

<ConfirmModal
	bind:open={revokeOpen}
	title="Revoke this token?"
	description="Every app or script using {revoking?.label ?? 'it'} stops working at once, and a revoked token cannot be restored. You will have to generate a new one and paste it into each place that used the old one."
	confirmLabel="Revoke"
	cancelLabel="Cancel"
	tone="danger"
	icon={icons.revoke}
	onConfirm={revoke}
/>
