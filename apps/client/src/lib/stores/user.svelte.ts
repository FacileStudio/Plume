import type { UserProfile } from '$lib/backend';

let current = $state<UserProfile | null>(null);

export const userStore = {
	get value() { return current; },
	set value(u: UserProfile | null) { current = u; }
};
