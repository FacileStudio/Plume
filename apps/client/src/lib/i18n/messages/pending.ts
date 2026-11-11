import type { Row } from '../row';

/** The page where the person confirms what an AI client asked Plume to do. */
export const pending = {
	'pending.title': ['Pending actions · Plume', 'Actions en attente · Plume', 'Acciones pendientes · Plume', 'Ausstehende Aktionen · Plume'],
	'pending.heading': ['Pending actions', 'Actions en attente', 'Acciones pendientes', 'Ausstehende Aktionen'],
	'pending.intro': [
		'An AI client asked Plume to do these. Nothing happens until you confirm.',
		'Un client IA a demandé ces actions à Plume. Rien ne se passe tant que vous ne confirmez pas.',
		'Un cliente de IA pidió estas acciones a Plume. No ocurre nada hasta que confirmes.',
		'Ein KI-Client hat Plume um diese Aktionen gebeten. Nichts passiert, bevor Sie bestätigen.'
	],
	'pending.none': ['Nothing is waiting', 'Rien en attente', 'Nada pendiente', 'Nichts ausstehend'],
	'pending.noneBody': [
		'Actions an AI client asks for appear here until you confirm or dismiss them.',
		'Les actions demandées par un client IA apparaissent ici jusqu’à ce que vous les confirmiez ou les ignoriez.',
		'Las acciones que pide un cliente de IA aparecen aquí hasta que las confirmes o descartes.',
		'Aktionen, um die ein KI-Client bittet, erscheinen hier, bis Sie sie bestätigen oder verwerfen.'
	],
	'pending.kind.send': ['Send for signature', 'Envoyer pour signature', 'Enviar para firma', 'Zur Unterschrift senden'],
	'pending.kind.remind': ['Remind a signer', 'Relancer un signataire', 'Recordar a un firmante', 'Unterzeichner erinnern'],
	'pending.kind.delete': ['Delete document', 'Supprimer le document', 'Eliminar documento', 'Dokument löschen'],
	'pending.askedBy': ['Asked by {client}', 'Demandé par {client}', 'Pedido por {client}', 'Angefragt von {client}'],
	'pending.sendNote': [
		'Plume emails each signer a private signing link.',
		'Plume envoie à chaque signataire un lien de signature privé.',
		'Plume envía a cada firmante un enlace de firma privado.',
		'Plume schickt jedem Unterzeichner einen privaten Signaturlink.'
	],
	'pending.remindNote': [
		'Plume emails {name} their signing link again.',
		'Plume renvoie à {name} son lien de signature.',
		'Plume vuelve a enviar a {name} su enlace de firma.',
		'Plume schickt {name} den Signaturlink erneut.'
	],
	'pending.deleteNote': [
		'The document, its signers and its file are deleted for good.',
		'Le document, ses signataires et son fichier sont supprimés définitivement.',
		'El documento, sus firmantes y su archivo se eliminan para siempre.',
		'Das Dokument, seine Unterzeichner und seine Datei werden endgültig gelöscht.'
	],
	'pending.confirm.send': ['Send', 'Envoyer', 'Enviar', 'Senden'],
	'pending.confirm.remind': ['Send the reminder', 'Envoyer la relance', 'Enviar el recordatorio', 'Erinnerung senden'],
	'pending.confirm.delete': ['Delete the document', 'Supprimer le document', 'Eliminar el documento', 'Dokument löschen'],
	'pending.dismiss': ['Dismiss', 'Ignorer', 'Descartar', 'Verwerfen'],
	'pending.changed': [
		'This changed after the request. Ask the AI client again.',
		'Ceci a changé après la demande. Redemandez au client IA.',
		'Esto cambió después de la solicitud. Vuelve a pedírselo al cliente de IA.',
		'Das wurde nach der Anfrage geändert. Fragen Sie den KI-Client erneut.'
	],
	'pending.missing': [
		'This document no longer exists.',
		'Ce document n’existe plus.',
		'Este documento ya no existe.',
		'Dieses Dokument existiert nicht mehr.'
	],
	'pending.done.send': ['Document sent', 'Document envoyé', 'Documento enviado', 'Dokument gesendet'],
	'pending.done.remind': ['Reminder sent', 'Relance envoyée', 'Recordatorio enviado', 'Erinnerung gesendet'],
	'pending.done.delete': ['Document deleted', 'Document supprimé', 'Documento eliminado', 'Dokument gelöscht'],
	'pending.dismissed': ['Request dismissed', 'Demande ignorée', 'Solicitud descartada', 'Anfrage verworfen'],
	'pending.loadFailed': [
		'Could not load the pending actions.',
		'Impossible de charger les actions en attente.',
		'No se pudieron cargar las acciones pendientes.',
		'Die ausstehenden Aktionen konnten nicht geladen werden.'
	],
	'pending.failed': [
		'That did not work. Try again.',
		'Cela n’a pas fonctionné. Réessayez.',
		'No funcionó. Inténtalo de nuevo.',
		'Das hat nicht geklappt. Versuchen Sie es erneut.'
	]
} as const satisfies Record<string, Row>;
