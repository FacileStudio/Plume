<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		FeatureSplit,
		ForDevelopers,
		LANDING_HEADING,
		LandingCta,
		LandingFooter,
		LandingHeader,
		PricingPlans,
		SignInWithFacile,
		WorksWith,
		icons,
		type DeveloperRow,
		type FeatureItem,
		type WorksWithApp
	} from '@facile/muse';
	import LandingCards from '$lib/components/landing/LandingCards.svelte';
	import SignDemo from '$lib/components/landing/SignDemo.svelte';
	import { currentUser } from '$lib';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	const signInHref = '/api/auth/oidc';
	const brandIcon = 'solar:pen-new-square-bold-duotone';
	const repo = 'https://github.com/FacileStudio/Plume';
	const width = 'mx-auto w-full max-w-5xl px-4 sm:px-6';

	let ready = $state(false);

	const features: FeatureItem[] = $derived([
		{ icon: 'solar:layers-linear', title: t('landing.feature.fields.title'), text: t('landing.feature.fields.text') },
		{ icon: icons.userCircle, title: t('landing.feature.clients.title'), text: t('landing.feature.clients.text') },
		{ icon: icons.usersGroup, title: t('landing.feature.spaces.title'), text: t('landing.feature.spaces.text') }
	]);

	const apps: WorksWithApp[] = $derived([
		{ name: 'Ardoise', icon: 'solar:bill-list-bold-duotone', line: t('landing.with.ardoise'), href: 'https://ardoise.facile.studio' },
		{ name: 'Nuage', icon: 'solar:cloud-bold-duotone', line: t('landing.with.nuage'), href: 'https://nuage.facile.studio' },
		{ name: 'Bulle', icon: 'solar:smile-circle-bold-duotone', line: t('landing.with.bulle'), href: 'https://bulle.facile.studio' }
	]);

	const devRows: DeveloperRow[] = $derived([
		{ kind: 'api', text: t('landing.dev.api'), href: `${repo}/blob/main/docs/api.md` },
		{ icon: icons.bolt, title: 'Webhooks', text: t('landing.dev.webhooks') },
		{ kind: 'mcp', text: t('landing.dev.mcp'), href: `${repo}/blob/main/docs/mcp.md` },
		{ kind: 'source', text: t('landing.dev.source'), href: repo }
	]);

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
				<SignInWithFacile href={signInHref} size="sm" />
			{/snippet}
		</LandingHeader>

		<main id="main">
			<section class="{width} pt-20 pb-12 text-center sm:pt-24">
				<h1 class="font-fc-title text-fc-3xl font-bold tracking-tight">{t('landing.heroTitle')}</h1>
				<p class="mx-auto mt-5 max-w-xl text-fc-md text-fc-fg-muted">{t('landing.heroText')}</p>
				<LandingCta {signInHref} align="center" class="mt-8" />
			</section>

			<section class="{width} pb-8">
				<SignDemo />
			</section>

			<LandingCards class="{width} py-16" />

			<WorksWith class="{width} py-12" {apps} headingClass={LANDING_HEADING} />

			<FeatureSplit id="features" class="{width} py-16" title={t('landing.features.title')} {features} />

			<PricingPlans
				id="pricing"
				class="{width} py-16"
				title={t('landing.pricing')}
				headingClass={LANDING_HEADING}
				highlight={['ai.access']}
				highlightOnly
				notes={{ '*': t('landing.plan.included') }}
				recommended="pro"
				suiteHref="https://suite.facile.studio/#tarifs"
			/>

			<section class="{width} py-20 text-center">
				<h2 class={LANDING_HEADING}>{t('landing.ctaTitle')}</h2>
				<p class="mt-3 text-fc-fg-muted">{t('landing.ctaText')}</p>
				<LandingCta {signInHref} align="center" class="mt-8" />
			</section>

			<div class="{width} pb-20">
				<ForDevelopers rows={devRows} description={t('landing.dev.line')} />
			</div>
		</main>

		<LandingFooter
			appName="Plume"
			icon={brandIcon}
			tagline={t('login.tagline')}
			legal={{ onLocaleChange: rememberLocale }}
		/>
	</div>
{/if}
