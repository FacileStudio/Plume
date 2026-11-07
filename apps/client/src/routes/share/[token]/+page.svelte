<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { SigningPayload, Field as SigningField, CompletedField, SigningStatus, SigningRosterEntry } from '$lib';
	import {
		Alert,
		Badge,
		Button,
		Card,
		Checkbox,
		ConfirmModal,
		Divider,
		EmptyState,
		Field,
		Input,
		LocaleSwitch,
		Spinner,
		icons,
		toast
	} from '@facile/muse';
	import SignaturePad from '$lib/components/signature-pad.svelte';
	import { formatDateTime } from '$lib/i18n/format';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	let payload = $state<SigningPayload | null>(null);
	let loading = $state(true);
	let notFound = $state(false);
	let submitting = $state(false);
	let signed = $state(false);
	let declined = $state(false);
	let error = $state('');
	let declineError = $state('');
	let confirmDecline = $state(false);
	let fieldValues = $state<Record<string, string>>({});
	let signingStatus = $state<SigningStatus | null>(null);
	let validated = $state(false);

	const signParticipants = $derived(
		(signingStatus?.signers ?? []).filter((s) => s.role !== 'viewer')
	);
	const signedCount = $derived(signParticipants.filter((s) => s.status === 'signed').length);
	const totalSigners = $derived(signParticipants.length);
	const someoneDeclined = $derived(
		(signingStatus?.signers ?? []).some((s) => s.status === 'declined')
	);
	const isComplete = $derived(signingStatus?.document.status === 'completed');
	const pendingNames = $derived(
		signParticipants.filter((s) => s.status === 'pending').map((s) => s.name)
	);

	const token = $derived((page.params as Record<string, string>).token);
	const fileUrl = $derived(api.signing.fileUrl(token));

	async function loadStatus(t: string) {
		try {
			signingStatus = await api.signing.status(t);
		} catch {}
	}

	function formatDate(value: string | null): string {
		if (!value) return '';
		try {
			return formatDateTime(value);
		} catch {
			return value;
		}
	}

	function roleName(role: string): string {
		switch (role) {
			case 'signer':
			case 'viewer':
			case 'approver':
				return t(`role.${role}`);
			default:
				return role;
		}
	}

	function rosterMeta(entry: SigningRosterEntry): { icon: string; tone: string; label: string } {
		if (entry.status === 'signed') {
			return {
				icon: icons.check,
				tone: 'text-fc-success',
				label: entry.signed_at
					? t('sign.roster.signedOn', { date: formatDate(entry.signed_at) })
					: t('sign.roster.signed')
			};
		}
		if (entry.status === 'declined') {
			return { icon: icons.error, tone: 'text-fc-danger', label: t('sign.roster.declined') };
		}
		if (entry.role === 'viewer') {
			return { icon: icons.eye, tone: 'text-fc-fg-muted', label: t('sign.roster.viewer') };
		}
		return { icon: icons.clock, tone: 'text-fc-fg-muted', label: t('sign.roster.waiting') };
	}

	let pdfContainer = $state<HTMLDivElement>(undefined!);
	let pdfPages = $state<{ num: number; width: number; height: number }[]>([]);
	let pdfCanvases = $state<Map<number, HTMLCanvasElement>>(new Map());
	let pdfLoading = $state(true);
	let pdfFailed = $state(false);
	let activeFieldId = $state<number | null>(null);

	function fieldsForPage(pageNum: number): SigningField[] {
		return (payload?.fields ?? []).filter((f) => f.page === pageNum);
	}

	function completedFieldsForPage(pageNum: number): CompletedField[] {
		return (payload?.completed_fields ?? []).filter((f) => f.page === pageNum);
	}

	function scrollToField(fieldId: number) {
		activeFieldId = fieldId;
		const el = pdfContainer?.querySelector(`[data-field-id="${fieldId}"]`);
		if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' });
	}

	function appendCanvas(node: HTMLElement, canvas: HTMLCanvasElement) {
		node.appendChild(canvas);
		return { destroy() { canvas.remove(); } };
	}

	function initFields(fields: SigningField[]) {
		const values: Record<string, string> = {};
		for (const f of fields) {
			values[String(f.id)] = f.value ?? '';
		}
		fieldValues = values;
	}

	function fieldLabel(f: SigningField): string {
		switch (f.field_type) {
			case 'signature':
			case 'text':
			case 'date':
			case 'checkbox':
				return t(`field.type.${f.field_type}`);
			default:
				return t('field.type.other');
		}
	}

	function labelFor(f: SigningField): string {
		return `${f.label || fieldLabel(f)}${f.required ? ' *' : ''}`;
	}

	// Mirrors the server's completion guard: a checkbox has to be ticked, everything else
	// only has to be non-blank. The server still re-checks — this is the affordance, not the
	// enforcement — so the two rules staying in step is a courtesy, not a security boundary.
	function isFilled(field: SigningField, value: string): boolean {
		if (field.field_type === 'checkbox') return value === 'true';
		return value.trim() !== '';
	}

	const missingFields = $derived(
		validated
			? (payload?.fields ?? []).filter(
					(f) => f.required && !isFilled(f, fieldValues[String(f.id)] ?? '')
				)
			: []
	);
	const missingIds = $derived(missingFields.map((f) => f.id));

	function focusField(fieldId: number) {
		const host = document.querySelector<HTMLElement>(`[data-field-control="${fieldId}"]`);
		if (!host) return;
		const control = host.matches('input, [tabindex]')
			? host
			: host.querySelector<HTMLElement>('input, [tabindex]');
		control?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		control?.focus();
	}

	async function signDocument() {
		validated = true;
		const missing = (payload?.fields ?? []).filter(
			(f) => f.required && !isFilled(f, fieldValues[String(f.id)] ?? '')
		);
		if (missing.length > 0) {
			error = '';
			toast.danger(
				missing.length === 1 ? t('sign.missing.one') : t('sign.missing.many', { count: missing.length })
			);
			focusField(missing[0].id);
			return;
		}

		submitting = true;
		error = '';
		try {
			await api.signing.sign(token, fieldValues);
			signed = true;
			await loadStatus(token);
		} catch (e: any) {
			error = e.message;
		}
		submitting = false;
	}

	async function declineDocument() {
		submitting = true;
		declineError = '';
		try {
			await api.signing.decline(token);
			declined = true;
			await loadStatus(token);
		} catch (e: any) {
			declineError = e?.message ?? t('sign.declineFailed');
			throw e;
		} finally {
			submitting = false;
		}
	}

	onMount(async () => {
		try {
			payload = await api.signing.get(token);
			if (payload.signer.status === 'signed') {
				signed = true;
				await loadStatus(token);
			} else if (payload.signer.status === 'declined') {
				declined = true;
				await loadStatus(token);
			} else {
				initFields(payload.fields);

				const pdfjsLib = await import('pdfjs-dist');
				pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
					'pdfjs-dist/build/pdf.worker.min.mjs',
					import.meta.url
				).toString();

				try {
					const pdf = await pdfjsLib.getDocument(fileUrl).promise;
					for (let i = 1; i <= pdf.numPages; i++) {
						const pg = await pdf.getPage(i);
						const viewport = pg.getViewport({ scale: 1.5 });
						pdfPages.push({ num: i, width: viewport.width, height: viewport.height });
						const canvas = document.createElement('canvas');
						canvas.width = viewport.width;
						canvas.height = viewport.height;
						canvas.style.width = '100%';
						canvas.style.height = 'auto';
						canvas.style.display = 'block';
						pdfCanvases.set(i, canvas);
						await pg.render({ canvasContext: canvas.getContext('2d')!, canvas, viewport }).promise;
					}
				} catch {
					pdfFailed = pdfPages.length === 0;
				}
				pdfLoading = false;
			}
		} catch {
			try {
				signingStatus = await api.signing.status(token);
				if (signingStatus.signer.status === 'declined') {
					declined = true;
				} else {
					signed = true;
				}
			} catch {
				notFound = true;
			}
		}
		loading = false;
	});
</script>

<svelte:head>
	<title>
		{payload ? t('sign.title', { name: payload.document.name }) : t('sign.titleFallback')}
	</title>
</svelte:head>

{#snippet progressAndRoster()}
	{#if signingStatus && signingStatus.signers.length > 0}
		<Card class="flex w-full flex-col gap-4 text-left">
			{#if totalSigners > 0}
				<div class="flex flex-col gap-2">
					<div class="flex items-center justify-between gap-3 text-fc-sm">
						<span class="font-medium text-fc-fg">{t('sign.progress')}</span>
						<span class="text-fc-fg-muted">
							{t('sign.progressCount', { signed: signedCount, total: totalSigners })}
						</span>
					</div>
					<div
						class="h-2 w-full overflow-hidden rounded-fc-pill bg-fc-surface"
						role="progressbar"
						aria-valuemin={0}
						aria-valuemax={totalSigners}
						aria-valuenow={signedCount}
						aria-label={t('sign.progress')}
					>
						<div
							class="h-full rounded-fc-pill bg-fc-success transition-all duration-500 motion-reduce:transition-none"
							style="width: {totalSigners ? (signedCount / totalSigners) * 100 : 0}%"
						></div>
					</div>
				</div>
				<Divider />
			{/if}

			<ul class="flex flex-col gap-3">
				{#each signingStatus.signers as entry (entry.name + '-' + entry.order_num)}
					{@const meta = rosterMeta(entry)}
					<li class="flex items-start gap-3">
						<span class="mt-0.5 shrink-0 {meta.tone}">
							<iconify-icon icon={meta.icon} width="18" height="18" class="block"></iconify-icon>
						</span>
						<div class="min-w-0 flex-1">
							<div class="flex flex-wrap items-center gap-1.5">
								<span class="truncate text-fc-sm font-medium text-fc-fg">{entry.name}</span>
								{#if entry.is_you}
									<Badge tone="accent">{t('sign.you')}</Badge>
								{/if}
								{#if entry.role !== 'signer'}
									<Badge tone="neutral">{roleName(entry.role)}</Badge>
								{/if}
							</div>
							<p class="text-fc-xs text-fc-fg-muted">{meta.label}</p>
						</div>
					</li>
				{/each}
			</ul>

			{#if isComplete}
				<Alert tone="success">{t('sign.finalized')}</Alert>
			{:else if someoneDeclined}
				<Alert tone="danger">{t('sign.stopped')}</Alert>
			{:else if pendingNames.length > 0}
				<Alert tone="neutral">{t('sign.waitingOn', { names: pendingNames.join(', ') })}</Alert>
			{/if}
		</Card>
	{/if}
{/snippet}

<div class="flex min-h-[100dvh] flex-col bg-fc-page text-fc-fg">
	<header class="flex items-center justify-between gap-2 px-4 py-4 sm:px-6">
		<div class="flex items-center gap-2">
			<iconify-icon icon="solar:pen-new-square-bold-duotone" width="24" height="24" class="block"
			></iconify-icon>
			<span class="text-fc-lg font-semibold tracking-tight">Plume</span>
		</div>
		<LocaleSwitch onChange={rememberLocale} />
	</header>

	<main
		class="flex flex-1 justify-center p-4 sm:p-6 {loading || notFound || signed || declined
			? 'items-center'
			: 'items-start'}"
	>
		{#if loading}
			<Spinner size="lg" label={t('common.loadingDocument')} />
		{:else if notFound}
			<div class="w-full max-w-md">
				<EmptyState
					icon={icons.eyeClosed}
					title={t('sign.notFound.title')}
					description={t('sign.notFound.text')}
				/>
			</div>
		{:else if signed}
			<div class="flex w-full max-w-md flex-col gap-4">
				<Alert
					tone="success"
					title={isComplete ? t('sign.done.completed') : t('sign.done.signed')}
				>
					{#if isComplete}
						{t('sign.done.allSigned')}
					{:else if totalSigners > 1}
						{t('sign.done.waiting')}
					{:else}
						{t('sign.done.close')}
					{/if}
				</Alert>

				{@render progressAndRoster()}

				<Button
					variant="secondary"
					size="lg"
					href={fileUrl}
					download
					icon={icons.download}
					class="w-full"
				>
					{t('common.download')}
				</Button>
			</div>
		{:else if declined}
			<div class="flex w-full max-w-md flex-col gap-4">
				<Alert tone="danger" title={t('sign.declined.title')}>{t('sign.declined.text')}</Alert>

				{@render progressAndRoster()}
			</div>
		{:else if payload}
			<div class="flex w-full max-w-6xl flex-col gap-4 lg:flex-row lg:gap-8">
				<div
					bind:this={pdfContainer}
					class="max-h-[60dvh] min-w-0 flex-1 overflow-y-auto rounded-fc-md bg-fc-component p-2 sm:p-4 lg:max-h-[calc(100dvh-10rem)]"
				>
					{#if pdfLoading}
						<div class="flex items-center justify-center gap-3 py-12 text-fc-sm text-fc-fg-muted">
							<Spinner size="sm" />
							{t('sign.loadingPreview')}
						</div>
					{:else if pdfFailed}
						<div class="py-8">
							<Alert tone="warning" title={t('sign.previewUnavailable')}>
								{t('sign.previewUnavailableText')}
							</Alert>
						</div>
					{:else}
						{#each pdfPages as pg (pg.num)}
							<div class="relative mx-auto mb-4" style="max-width: {pg.width}px;" data-page={pg.num}>
								{#if pdfCanvases.get(pg.num)}
									{@const canvas = pdfCanvases.get(pg.num)!}
									<div use:appendCanvas={canvas}></div>
								{/if}
								<div class="pointer-events-none absolute inset-0">
									{#each fieldsForPage(pg.num) as field (field.id)}
										{@const isActive = activeFieldId === field.id}
										{@const val = fieldValues[String(field.id)] || ''}
										<div
											data-field-id={field.id}
											class="absolute flex flex-col items-center justify-center overflow-hidden rounded-fc-xs border-2 text-fc-xs text-fc-info transition-all duration-300 motion-reduce:transition-none
												{isActive ? 'border-solid border-fc-info bg-fc-info/20 ring-3 ring-fc-info/15' : ''}
												{!isActive && val ? 'border-dashed border-fc-info bg-fc-info/10' : ''}
												{!isActive && !val ? 'border-dashed border-fc-info/40 bg-fc-info/10' : ''}"
											style="left: {field.x}%; top: {field.y}%; width: {field.width}%; height: {field.height}%;"
										>
											{#if val && field.field_type === 'signature'}
												{#if val.startsWith('data:image/')}
													<img src={val} alt={t('sign.signatureAlt')} class="h-full w-full object-contain p-0.5" />
												{:else}
													<span class="truncate px-1 font-serif italic">{val}</span>
												{/if}
											{:else if val && field.field_type === 'checkbox'}
												<span class="text-fc-sm font-bold">{val === 'true' ? '✓' : ''}</span>
											{:else if val}
												<span class="truncate px-1 font-medium">{val}</span>
											{:else}
												<span class="truncate px-1 font-medium opacity-70">
													{field.label || fieldLabel(field)}
												</span>
											{/if}
										</div>
									{/each}
									{#each completedFieldsForPage(pg.num) as cf (cf.id)}
										<div
											class="absolute flex flex-col items-center justify-center overflow-hidden rounded-fc-xs border border-fc-success/40 bg-fc-success/10 text-fc-xs text-fc-success transition-all duration-300 motion-reduce:transition-none"
											style="left: {cf.x}%; top: {cf.y}%; width: {cf.width}%; height: {cf.height}%;"
										>
											<span class="truncate px-1 font-medium opacity-80">{cf.signer_name}</span>
											{#if cf.field_type === 'signature'}
												{#if cf.value.startsWith('data:image/')}
													<img
														src={cf.value}
														alt={t('sign.signatureBy', { name: cf.signer_name })}
														class="h-full w-full object-contain p-0.5 opacity-60"
													/>
												{:else}
													<span class="truncate px-1 font-serif italic opacity-70">{cf.value}</span>
												{/if}
											{:else if cf.value && cf.field_type !== 'checkbox'}
												<span class="truncate px-1 opacity-70">{cf.value}</span>
											{:else if cf.field_type === 'checkbox' && cf.value === 'true'}
												<span class="opacity-70">✓</span>
											{/if}
										</div>
									{/each}
								</div>
							</div>
						{/each}
					{/if}
				</div>

				<div class="flex w-full shrink-0 flex-col gap-4 lg:w-80">
					<Card class="flex flex-col gap-4">
						<div class="flex flex-col gap-1">
							<h1 class="text-fc-lg font-semibold text-fc-fg">{payload.document.name}</h1>
							<p class="text-fc-sm text-fc-fg-muted">
								{t('sign.signingAs', { name: payload.signer.name })}
							</p>
						</div>

						<Button
							variant="secondary"
							size="lg"
							href={fileUrl}
							download
							icon={icons.download}
							class="w-full"
						>
							{t('common.download')}
						</Button>

						{#if payload.fields.length > 0}
							<Divider />

							<div class="flex flex-col gap-4">
								{#each payload.fields as field (field.id)}
									{@const isMissing = missingIds.includes(field.id)}
									{#if field.field_type === 'signature'}
										<div class="flex flex-col gap-1.5">
											<p class="text-fc-sm text-fc-fg">{labelFor(field)}</p>
											<div
												role="group"
												tabindex="-1"
												data-field-control={field.id}
												aria-label={labelFor(field)}
												onfocusin={() => scrollToField(field.id)}
											>
												<SignaturePad bind:value={fieldValues[String(field.id)]} />
											</div>
											{#if isMissing}
												<span class="text-fc-xs text-fc-danger">{t('sign.signatureRequired')}</span>
											{/if}
										</div>
									{:else if field.field_type === 'date'}
										<Field
											label={labelFor(field)}
											error={isMissing ? t('sign.fieldRequired') : undefined}
											data-field-control={field.id}
										>
											<div class="flex items-center gap-2">
												<Input
													type="date"
													bind:value={fieldValues[String(field.id)]}
													onfocus={() => scrollToField(field.id)}
													class="flex-1"
												/>
												<Button
													variant="secondary"
													size="lg"
													icon={icons.calendar}
													aria-label={t('sign.useToday')}
													onclick={() => {
														fieldValues[String(field.id)] = new Date().toISOString().split('T')[0];
														scrollToField(field.id);
													}}
												>
													{t('sign.today')}
												</Button>
											</div>
										</Field>
									{:else if field.field_type === 'checkbox'}
										<div class="flex flex-col gap-1.5">
											<Checkbox
												label={labelFor(field)}
												checked={fieldValues[String(field.id)] === 'true'}
												onchange={(e) => {
													fieldValues[String(field.id)] = (e.currentTarget as HTMLInputElement).checked
														? 'true'
														: 'false';
												}}
												onfocus={() => scrollToField(field.id)}
												data-field-control={field.id}
												aria-invalid={isMissing}
												class="min-h-11"
											/>
											{#if isMissing}
												<span class="text-fc-xs text-fc-danger">{t('sign.checkboxRequired')}</span>
											{/if}
										</div>
									{:else}
										<Field
											label={labelFor(field)}
											error={isMissing ? t('sign.fieldRequired') : undefined}
											data-field-control={field.id}
										>
											<Input
												bind:value={fieldValues[String(field.id)]}
												placeholder={t('sign.enterText')}
												onfocus={() => scrollToField(field.id)}
											/>
										</Field>
									{/if}
								{/each}
							</div>
						{/if}

						{#if missingFields.length > 0}
							<Alert tone="danger" title={t('sign.missing.title')}>
								{t('sign.missing.list', {
									names: missingFields.map((f) => f.label || fieldLabel(f)).join(', ')
								})}
							</Alert>
						{/if}

						{#if error}
							<Alert tone="danger" title={t('sign.error.title')}>{error}</Alert>
						{/if}

						<Divider />

						<div class="flex flex-col gap-2 sm:flex-row">
							<Button
								size="lg"
								icon={icons.edit}
								onclick={signDocument}
								disabled={submitting}
								class="flex-1"
							>
								{#if submitting}
									<Spinner size="sm" />
								{/if}
								{t('sign.submit')}
							</Button>
							<Button
								variant="secondary"
								size="lg"
								onclick={() => (confirmDecline = true)}
								disabled={submitting}
							>
								{t('sign.decline')}
							</Button>
						</div>
					</Card>
				</div>
			</div>
		{/if}
	</main>

	<footer
		class="flex items-center justify-center gap-1.5 px-4 py-4 text-fc-xs text-fc-fg-muted sm:px-6"
	>
		<iconify-icon icon="solar:pen-new-square-bold-duotone" width="16" height="16" class="block"
		></iconify-icon>
		{t('common.footer.poweredBy')}
		<a href="/" class="font-medium text-fc-fg hover:opacity-80">Plume</a>
	</footer>
</div>

<ConfirmModal
	bind:open={confirmDecline}
	tone="danger"
	title={t('sign.confirm.title')}
	description={t('sign.confirm.text')}
	confirmLabel={t('sign.confirm.label')}
	cancelLabel={t('sign.confirm.cancel')}
	onConfirm={declineDocument}
	onCancel={() => (declineError = '')}
>
	{#if declineError}
		<Alert tone="danger">{declineError}</Alert>
	{/if}
</ConfirmModal>
