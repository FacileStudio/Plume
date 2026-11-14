import mdi from '../src/lib/icons/mdi.json';
import solar from '../src/lib/icons/solar.json';
import { museIconData } from '../src/lib/icons/register';

const problems: string[] = [];
for (const collection of [solar, mdi]) {
	const data = museIconData(collection);
	const names = [...Object.keys(collection.icons), ...Object.keys(collection.aliases ?? {})];
	for (const name of names) {
		const glyph = data[`${collection.prefix}:${name}`];
		if (!glyph?.body || !glyph.width || !glyph.height) problems.push(`${collection.prefix}:${name}`);
	}
}

if (problems.length > 0) {
	console.error(`icons muse cannot draw: ${problems.join(', ')}`);
	process.exit(1);
}
console.log('every bundled icon and alias resolves for muse');
