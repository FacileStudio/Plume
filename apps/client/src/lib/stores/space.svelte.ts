import type { Space } from '$lib/backend';

const STORAGE_KEY = 'plume_active_space_id';

function loadPersistedId(): number | null {
	if (typeof window === 'undefined') return null;
	const raw = localStorage.getItem(STORAGE_KEY);
	if (!raw) return null;
	const parsed = Number(raw);
	return Number.isFinite(parsed) ? parsed : null;
}

let spaces = $state<Space[]>([]);
let activeId = $state<number | null>(loadPersistedId());

export const spaceStore = {
	get spaces() { return spaces; },
	set spaces(v: Space[]) { spaces = v; },

	get activeId() { return activeId; },
	set activeId(id: number | null) {
		activeId = id;
		if (typeof window !== 'undefined') {
			if (id !== null) {
				localStorage.setItem(STORAGE_KEY, String(id));
			} else {
				localStorage.removeItem(STORAGE_KEY);
			}
		}
	},

	get active(): Space | null {
		if (activeId === null) return null;
		return spaces.find((s) => s.id === activeId) ?? null;
	},

	clear() {
		spaces = [];
		activeId = null;
		if (typeof window !== 'undefined') {
			localStorage.removeItem(STORAGE_KEY);
		}
	}
};
