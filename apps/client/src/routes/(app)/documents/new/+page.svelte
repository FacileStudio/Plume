<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import {
		Button,
		Dropzone,
		Field,
		IconButton,
		Input,
		SettingsSection,
		UploadProgress,
		icons,
		toast
	} from '@facile/muse';
	import { spaceStore } from '$lib/stores/space.svelte';
	import { t } from '$lib/i18n/index.svelte';

	const MAX_FILE_SIZE = 10 * 1024 * 1024;

	let name = $state('');
	let nameError = $state('');
	let files = $state<File[]>([]);
	let uploadStatus = $state<'pending' | 'uploading' | 'error'>('pending');
	let uploadError = $state('');
	let signers = $state<{ name: string; email: string }[]>([{ name: '', email: '' }]);
	let submitting = $state(false);

	const file = $derived(files[0] ?? null);

	const uploadItems = $derived(
		file
			? [
					{
						id: 'pdf',
						name: file.name,
						size: file.size,
						progress: uploadStatus === 'uploading' ? 0 : 100,
						status: uploadStatus,
						error: uploadError || undefined
					}
				]
			: []
	);

	function addSigner() {
		signers = [...signers, { name: '', email: '' }];
	}

	function removeSigner(index: number) {
		signers = signers.filter((_, i) => i !== index);
	}

	function clearFile() {
		if (submitting) return;
		files = [];
		uploadStatus = 'pending';
		uploadError = '';
	}

	function rejectFiles(rejections: { file: File; reason: 'type' | 'size' | 'count' }[]) {
		const first = rejections[0];
		if (!first) return;
		if (first.reason === 'type') {
			toast.danger(t('docNew.onlyPdf'));
			return;
		}
		if (first.reason === 'size') {
			toast.danger(
				t('docNew.tooLarge', {
					name: first.file.name,
					size: (first.file.size / 1024 / 1024).toFixed(1)
				})
			);
			return;
		}
		toast.danger(t('docNew.onlyOne'));
	}

	function acceptFiles() {
		uploadStatus = 'pending';
		uploadError = '';
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		nameError = '';

		if (!name.trim()) {
			nameError = t('docNew.nameRequired');
			return;
		}
		if (!file) {
			toast.danger(t('docNew.uploadAPdf'));
			return;
		}

		submitting = true;
		uploadStatus = 'uploading';
		uploadError = '';
		try {
			const doc = await api.documents.create(name, file, spaceStore.activeId);

			const validSigners = signers.filter((s) => s.name.trim() && s.email.trim());
			for (const signer of validSigners) {
				await api.signers.add(doc.id, signer.name, signer.email);
			}

			goto(`/documents/${doc.id}`);
		} catch (e) {
			const message = e instanceof Error ? e.message : t('docNew.createFailed');
			uploadStatus = 'error';
			uploadError = message;
			toast.danger(message);
			submitting = false;
		}
	}
</script>

<svelte:head><title>{t('docNew.title')}</title></svelte:head>

<div class="flex max-w-2xl flex-col gap-10">
	<div class="flex flex-col gap-3">
		<Button
			href="/documents"
			variant="ghost"
			size="sm"
			icon={icons.chevronLeft}
			class="self-start px-2"
		>
			{t('common.backToDocuments')}
		</Button>
		<div class="flex flex-col gap-1">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('docNew.heading')}</h1>
			<p class="text-fc-sm text-fc-fg-muted">{t('docNew.subtitle')}</p>
		</div>
	</div>

	<form onsubmit={submit} class="flex flex-col gap-10">
		<SettingsSection title={t('docNew.section.document')} description={t('docNew.section.documentText')}>
			<Field label={t('docNew.name')} error={nameError}>
				<Input bind:value={name} placeholder={t('docNew.namePlaceholder')} required />
			</Field>

			<div class="flex flex-col gap-2">
				<span class="text-fc-sm text-fc-fg">{t('docNew.pdfFile')}</span>
				<Dropzone
					bind:files
					accept=".pdf,application/pdf"
					maxSize={MAX_FILE_SIZE}
					disabled={submitting}
					label={t('docNew.dropLabel')}
					browseLabel={t('docNew.browse')}
					dropAnnouncement={t('docNew.dropAnnouncement')}
					hint={t('docNew.hint')}
					onFiles={acceptFiles}
					onReject={rejectFiles}
				/>
				{#if uploadItems.length > 0}
					<UploadProgress items={uploadItems} showTotal={false} onCancel={clearFile} />
				{/if}
			</div>
		</SettingsSection>

		<SettingsSection title={t('docNew.section.signers')} description={t('docNew.section.signersText')}>
			{#snippet actions()}
				<Button variant="outline" size="sm" icon={icons.plus} onclick={addSigner}>
					{t('docNew.addSigner')}
				</Button>
			{/snippet}

			{#each signers as signer, i (i)}
				<div class="flex items-end gap-3 pt-4 first:pt-0">
					<div class="grid min-w-0 flex-1 gap-3 sm:grid-cols-2">
						<Field label={t('common.name')}>
							<Input bind:value={signer.name} placeholder={t('docNew.signerNamePlaceholder')} />
						</Field>
						<Field label={t('common.email')}>
							<Input
								bind:value={signer.email}
								type="email"
								placeholder={t('docNew.signerEmailPlaceholder')}
							/>
						</Field>
					</div>
					{#if signers.length > 1}
						<IconButton
							variant="danger"
							aria-label={t('docNew.removeSigner', { number: i + 1 })}
							onclick={() => removeSigner(i)}
						>
							<iconify-icon icon={icons.remove} width="18" height="18" class="block"
							></iconify-icon>
						</IconButton>
					{/if}
				</div>
			{/each}
		</SettingsSection>

		<Button type="submit" size="lg" icon={icons.plus} disabled={submitting} class="w-full">
			{submitting ? t('common.creating') : t('docNew.submit')}
		</Button>
	</form>
</div>
