import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';

import Overlay from './alert-dialog-overlay.svelte';
import Content from './alert-dialog-content.svelte';
import Title from './alert-dialog-title.svelte';
import Description from './alert-dialog-description.svelte';
import Action from './alert-dialog-action.svelte';
import Cancel from './alert-dialog-cancel.svelte';
import Footer from './alert-dialog-footer.svelte';
import Header from './alert-dialog-header.svelte';

const Root = AlertDialogPrimitive.Root;
const Trigger = AlertDialogPrimitive.Trigger;
const Portal = AlertDialogPrimitive.Portal;

export {
	Root,
	Trigger,
	Portal,
	Overlay,
	Content,
	Title,
	Description,
	Action,
	Cancel,
	Footer,
	Header,
};
