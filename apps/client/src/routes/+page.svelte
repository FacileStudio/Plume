<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { currentUser } from '$lib';

	const ghostLinkClass =
		'inline-flex h-9 items-center justify-center rounded-md px-4 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground';
	const primaryLinkClass =
		'inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90';
	const primaryCtaClass =
		'inline-flex h-11 items-center justify-center rounded-md bg-primary px-6 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90';
	const outlineCtaClass =
		'inline-flex h-11 items-center justify-center rounded-md border border-border bg-background px-6 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground';
	const featureCardClass = 'rounded-lg border border-border p-6';
	const featureIconClass =
		'mb-3 flex size-10 items-center justify-center rounded-md border border-border';

	let redirecting = $state(true);
	let ssoOnly = $state(false);

	onMount(async () => {
		/* This page is where OIDC_SUCCESS_URL lands, so it is the first thing an
		   SSO user sees after the identity provider sends them back. It used to
		   wait for a ?code= that only the CLI exchange flow ever produced — porte
		   v0.2.4+ issues a session cookie on the browser callback and redirects
		   with an empty URL — and to trust localStorage, which an SSO login never
		   writes. Asking the API is the only way to notice the person standing
		   here is signed in. */
		if (await currentUser().catch(() => null)) {
			goto('/dashboard');
			return;
		}

		try {
			const cfg = await fetch('/api/auth/config').then((r) => r.json());
			ssoOnly = cfg.sso_only ?? false;
		} catch {}

		redirecting = false;
	});
</script>

<svelte:head>
	<title>Plume — Document Signing</title>
	<meta
		name="description"
		content="Self-hosted document signing. Send, sign, and seal — no third party required."
	/>
</svelte:head>

{#if !redirecting}
	<div class="min-h-screen bg-background text-foreground">
		<header class="border-b border-border">
			<div class="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
				<div class="flex h-14 items-center gap-3">
					<iconify-icon
						icon="solar:pen-new-square-bold-duotone"
						width="28"
						height="28"
						class="block size-7"
					></iconify-icon>
					<span class="text-2xl font-bold font-heading tracking-tight">Plume</span>
				</div>
				<div class="flex items-center gap-2">
					<a href="/login" class={ghostLinkClass}>Log in</a>
					<a href={ssoOnly ? '/login' : '/login?tab=register'} class={primaryLinkClass}>
						{ssoOnly ? 'Sign in with Facile' : 'Get started'}
					</a>
				</div>
			</div>
		</header>

		<main>
			<section class="mx-auto max-w-5xl px-6 py-24 text-center">
				<h1 class="text-4xl font-bold font-heading tracking-tight leading-tight">
					Send it. Sign it.<br />Seal it.
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-lg text-muted-foreground leading-relaxed">
					Plume is a self-hosted document signing platform. Upload a PDF, place fields, send for
					signature — done.
				</p>
				<div class="mt-10 flex justify-center gap-3">
					<a href={ssoOnly ? '/login' : '/login?tab=register'} class={primaryCtaClass}>
						{ssoOnly ? 'Sign in with Facile' : 'Get started'}
						<iconify-icon
							icon="solar:arrow-right-linear"
							width="16"
							height="16"
							class="ml-2 block size-4"
						></iconify-icon>
					</a>
					<a href="/login" class={outlineCtaClass}>Log in</a>
				</div>
			</section>

			<div class="mx-auto max-w-5xl"><div class="h-px bg-border"></div></div>

			<section class="mx-auto max-w-5xl px-6 py-20">
				<div class="grid gap-6 md:grid-cols-3">
					<div class={featureCardClass}>
						<div class={featureIconClass}>
							<iconify-icon
								icon="solar:document-add-linear"
								width="20"
								height="20"
								class="block size-5"
							></iconify-icon>
						</div>
						<h3 class="text-base font-semibold">Upload &amp; place fields</h3>
						<p class="mt-1.5 text-sm text-muted-foreground">
							Upload any PDF, drag signature and text fields where you need them. Send in seconds.
						</p>
					</div>

					<div class={featureCardClass}>
						<div class={featureIconClass}>
							<iconify-icon
								icon="solar:shield-check-linear"
								width="20"
								height="20"
								class="block size-5"
							></iconify-icon>
						</div>
						<h3 class="text-base font-semibold">Legally binding</h3>
						<p class="mt-1.5 text-sm text-muted-foreground">
							PKI-based digital signatures with full audit trail. Tamper-proof, timestamped,
							verifiable.
						</p>
					</div>

					<div class={featureCardClass}>
						<div class={featureIconClass}>
							<iconify-icon
								icon="solar:lock-linear"
								width="20"
								height="20"
								class="block size-5"
							></iconify-icon>
						</div>
						<h3 class="text-base font-semibold">Self-hosted</h3>
						<p class="mt-1.5 text-sm text-muted-foreground">
							Your documents never leave your server. No cloud dependency, no data harvesting.
						</p>
					</div>
				</div>
			</section>

			<div class="mx-auto max-w-5xl"><div class="h-px bg-border"></div></div>

			<section class="mx-auto max-w-5xl px-6 py-20 text-center">
				<h2 class="text-3xl font-bold font-heading tracking-tight">
					{ssoOnly ? 'Ready to sign in?' : 'Ready to start?'}
				</h2>
				<p class="mt-4 text-muted-foreground">
					{ssoOnly
						? 'Use your Facile SSO to access Plume.'
						: 'Free to use. Self-hosted. No credit card required.'}
				</p>
				<a href={ssoOnly ? '/login' : '/login?tab=register'} class="mt-8 {primaryCtaClass}">
					{ssoOnly ? 'Sign in with Facile' : 'Create an account'}
				</a>
			</section>
		</main>

		<footer class="border-t border-border text-center">
			<div class="mx-auto max-w-5xl px-6 py-6 text-sm text-muted-foreground">
				© {new Date().getFullYear()} Plume by <a
					href="https://facile.studio"
					class="font-semibold underline hover:cursor-pointer">Facile.</a
				>
			</div>
		</footer>
	</div>
{/if}
