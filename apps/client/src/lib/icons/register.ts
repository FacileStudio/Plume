import type { IconData } from '@facile/muse';

interface IconifyCollection {
	prefix: string;
	width?: number;
	height?: number;
	icons: Record<string, { body: string; width?: number; height?: number }>;
	aliases?: Record<string, { parent: string }>;
}

/**
 * Turns a bundled Iconify collection into the map muse's `registerIcons` takes, aliases included,
 * so muse's inline `Icon` draws every glyph Plume ships instead of an empty box.
 */
export function museIconData(collection: IconifyCollection): Record<string, IconData> {
	const glyph = (name: string): IconData | undefined => {
		const icon = collection.icons[name];
		if (!icon) return undefined;
		return { body: icon.body, width: icon.width ?? collection.width ?? 24, height: icon.height ?? collection.height ?? 24 };
	};
	const entries = [
		...Object.keys(collection.icons).map((name) => [name, glyph(name)] as const),
		...Object.entries(collection.aliases ?? {}).map(([name, alias]) => [name, glyph(alias.parent)] as const)
	];
	return Object.fromEntries(
		entries.flatMap(([name, data]) => (data ? [[`${collection.prefix}:${name}`, data]] : []))
	);
}
