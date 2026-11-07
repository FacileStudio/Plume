import { locale, t } from './index.svelte';

/** Formats a date or instant with `Intl.DateTimeFormat` in the current locale. */
export function formatDate(value: Date | string | number, options: Intl.DateTimeFormatOptions = {}): string {
	return new Intl.DateTimeFormat(locale(), options).format(new Date(value));
}

/** A short calendar date such as "Nov 7, 2026", in the current locale. */
export function formatShortDate(value: Date | string | number): string {
	return formatDate(value, { month: 'short', day: 'numeric', year: 'numeric' });
}

/** A date with its clock time, in the current locale. */
export function formatDateTime(value: Date | string | number): string {
	return formatDate(value, { dateStyle: 'medium', timeStyle: 'short' });
}

/** The short month name of a date, in the current locale. */
export function formatMonth(value: Date | string | number): string {
	return formatDate(value, { month: 'short' });
}

/** Formats a number in the current locale. */
export function formatNumber(value: number): string {
	return new Intl.NumberFormat(locale()).format(value);
}

/** How long ago an instant was, such as "5 min. ago", in the current locale. */
export function formatRelative(value: Date | string | number): string {
	const diff = Date.now() - new Date(value).getTime();
	if (diff < 60_000) return t('time.justNow');
	const format = new Intl.RelativeTimeFormat(locale(), { numeric: 'always', style: 'short' });
	const minutes = Math.floor(diff / 60_000);
	if (minutes < 60) return format.format(-minutes, 'minute');
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return format.format(-hours, 'hour');
	return format.format(-Math.floor(hours / 24), 'day');
}
