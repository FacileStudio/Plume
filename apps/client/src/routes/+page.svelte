<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Icon, LandingFooter, LandingHeader, PricingPlans, SignInWithFacile } from '@facile/muse';
	import SignDemo from '$lib/components/landing/SignDemo.svelte';
	import { currentUser } from '$lib';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	const signInHref = '/api/auth/oidc';
	const brandIcon = 'solar:pen-new-square-bold-duotone';

	let ready = $state(false);

	const features = [
		{ icon: 'solar:layers-linear', id: 'fields' },
		{ icon: 'solar:users-group-rounded-linear', id: 'order' },
		{ icon: 'solar:shield-check-linear', id: 'proof' },
		{ icon: 'solar:server-linear', id: 'hosted' }
	] as const;

	onMount(async () => {
		if (await currentUser().catch(() => null)) {
			goto('/dashboard');
			return;
		}
		ready = true;
	});
</script>

<svelte:head>
	<title>{t('landing.title')}</title>
	<meta name="description" content={t('landing.description')} />
</svelte:head>

{#if ready}
	<div class="min-h-dvh overflow-x-clip bg-fc-page text-fc-fg">
		<LandingHeader appName="Plume" icon={brandIcon} skipHref="#main">
			{#snippet actions()}
				<SignInWithFacile href={signInHref} />
			{/snippet}
		</LandingHeader>

		<main id="main">
			<section class="mx-auto max-w-5xl px-4 pt-20 pb-12 text-center sm:px-6 sm:pt-24">
				<h1 class="font-fc-title text-fc-3xl font-bold tracking-tight">{t('landing.heroTitle')}</h1>
				<p class="mx-auto mt-5 max-w-xl text-fc-md text-fc-fg-muted">{t('landing.heroText')}</p>
				<div class="mt-8 flex justify-center">
					<SignInWithFacile href={signInHref} />
				</div>
			</section>

			<section class="mx-auto max-w-5xl px-4 pb-8 sm:px-6">
				<SignDemo />
			</section>

			<section class="mx-auto max-w-5xl px-4 py-12 sm:px-6">
				<ul class="grid gap-4 sm:grid-cols-2">
					{#each features as feature (feature.id)}
						<li class="rounded-fc-lg bg-fc-surface p-6">
							<span class="mb-3 flex size-10 items-center justify-center rounded-fc-md bg-fc-component">
								<Icon icon={feature.icon} size={20} />
							</span>
							<h2 class="text-fc-lg font-bold">{t(`landing.feature.${feature.id}.title`)}</h2>
							<p class="mt-1.5 text-fc-sm text-fc-fg-muted">{t(`landing.feature.${feature.id}.text`)}</p>
						</li>
					{/each}
				</ul>
			</section>

			<PricingPlans title={t('landing.pricing')} class="mx-auto w-full max-w-5xl px-4 py-16 sm:px-6" />

			<section class="mx-auto max-w-5xl px-4 py-20 text-center sm:px-6">
				<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight">{t('landing.ctaTitle')}</h2>
				<p class="mt-3 text-fc-fg-muted">{t('landing.ctaText')}</p>
				<div class="mt-8 flex justify-center">
					<SignInWithFacile href={signInHref} />
				</div>
			</section>
		</main>

		<LandingFooter
			appName="Plume"
			icon={brandIcon}
			tagline={t('login.tagline')}
			legal={{ onLocaleChange: rememberLocale }}
		/>
	</div>
{/if}
