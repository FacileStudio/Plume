import type { Row } from '../row';

/** Strings of the client pages. */
export const clients = {
	'clients.title': ['Clients — Plume', 'Clients — Plume', 'Clientes — Plume', 'Kunden — Plume'],
	'clients.heading': ['Clients', 'Clients', 'Clientes', 'Kunden'],
	'clients.subtitle': [
		'The people and companies you send documents to.',
		'Les personnes et les sociétés à qui vous envoyez des documents.',
		'Las personas y empresas a las que envías documentos.',
		'Die Personen und Firmen, denen du Dokumente sendest.'
	],
	'clients.new': ['New client', 'Nouveau client', 'Nuevo cliente', 'Neuer Kunde'],
	'clients.empty.title': ['No clients yet', 'Aucun client pour l’instant', 'Aún no hay clientes', 'Noch keine Kunden'],
	'clients.empty.text': [
		'Add your first one and it will show up here, along with every document linked to it.',
		'Ajoutez le premier : il apparaîtra ici avec tous les documents qui lui sont liés.',
		'Añade el primero y aparecerá aquí junto con todos los documentos vinculados.',
		'Füge den ersten hinzu, dann erscheint er hier mit allen verknüpften Dokumenten.'
	],
	'clients.col.client': ['Client', 'Client', 'Cliente', 'Kunde'],
	'clients.noEmail': ['No email', 'Aucun email', 'Sin correo', 'Keine E-Mail'],
	'clients.noContact': [
		'No contact details',
		'Aucune coordonnée',
		'Sin datos de contacto',
		'Keine Kontaktdaten'
	],
	'clients.deleted': ['Client deleted', 'Client supprimé', 'Cliente eliminado', 'Kunde gelöscht'],
	'clients.deleteFailed': [
		'Failed to delete client',
		'Impossible de supprimer le client',
		'No se pudo eliminar el cliente',
		'Der Kunde konnte nicht gelöscht werden'
	],
	'clients.confirm.title': ['Delete {name}?', 'Supprimer {name} ?', '¿Eliminar {name}?', '{name} löschen?'],
	'clients.confirm.thisClient': ['this client', 'ce client', 'este cliente', 'diesen Kunden'],
	'clients.confirm.text': [
		'This cannot be undone. Documents linked to this client are kept, but unlinked.',
		'Cette action est irréversible. Les documents liés à ce client sont conservés, mais dissociés.',
		'No se puede deshacer. Los documentos vinculados a este cliente se conservan, pero sin vínculo.',
		'Das lässt sich nicht rückgängig machen. Mit diesem Kunden verknüpfte Dokumente bleiben erhalten, werden aber getrennt.'
	],
	'clients.confirm.label': ['Delete client', 'Supprimer le client', 'Eliminar cliente', 'Kunden löschen'],
	'clients.confirm.cancel': ['Keep client', 'Conserver le client', 'Conservar cliente', 'Kunden behalten'],
	'clientNew.title': ['New Client — Plume', 'Nouveau client — Plume', 'Nuevo cliente — Plume', 'Neuer Kunde — Plume'],
	'clientNew.back': ['Back to clients', 'Retour aux clients', 'Volver a los clientes', 'Zurück zu den Kunden'],
	'clientNew.subtitle': [
		'Only the name is required — everything else can be filled in later.',
		'Seul le nom est obligatoire, le reste peut être complété plus tard.',
		'Solo el nombre es obligatorio; el resto se puede completar después.',
		'Nur der Name ist erforderlich, alles andere kannst du später ergänzen.'
	],
	'clientNew.nameRequired': [
		'Client name is required',
		'Le nom du client est obligatoire',
		'El nombre del cliente es obligatorio',
		'Der Kundenname ist erforderlich'
	],
	'clientNew.created': ['Client created', 'Client créé', 'Cliente creado', 'Kunde erstellt'],
	'clientNew.failed': [
		'Failed to create client',
		'Impossible de créer le client',
		'No se pudo crear el cliente',
		'Der Kunde konnte nicht erstellt werden'
	],
	'clientNew.namePlaceholder': ['Jane Doe', 'Camille Martin', 'Ana García', 'Erika Mustermann'],
	'clientNew.emailPlaceholder': ['jane@example.com', 'camille@exemple.fr', 'ana@ejemplo.es', 'erika@beispiel.de'],
	'clientNew.companyPlaceholder': ['Acme Inc.', 'Acme SARL', 'Acme S.L.', 'Acme GmbH'],
	'clientNew.phonePlaceholder': ['+1 555 000 0000', '+33 1 23 45 67 89', '+34 600 000 000', '+49 30 1234567'],
	'clientNew.notesHelp': [
		'Optional — anything worth remembering about this client.',
		'Facultatif : tout ce qui vaut la peine d’être noté sur ce client.',
		'Opcional: cualquier cosa que convenga recordar sobre este cliente.',
		'Optional: alles, was man über diesen Kunden festhalten sollte.'
	],
	'clientNew.notesPlaceholder': [
		'Prefers PDFs, invoices monthly…',
		'Préfère les PDF, facture chaque mois…',
		'Prefiere PDF, factura cada mes…',
		'Bevorzugt PDFs, Rechnung monatlich…'
	],
	'clientNew.submit': ['Create client', 'Créer le client', 'Crear cliente', 'Kunden erstellen'],
	'client.titleFallback': ['Client', 'Client', 'Cliente', 'Kunde'],
	'client.back': ['Clients', 'Clients', 'Clientes', 'Kunden'],
	'client.nameRequired': [
		'Client name is required',
		'Le nom du client est obligatoire',
		'El nombre del cliente es obligatorio',
		'Der Kundenname ist erforderlich'
	],
	'client.updated': ['Client updated', 'Client mis à jour', 'Cliente actualizado', 'Kunde aktualisiert'],
	'client.updateFailed': [
		'Failed to update client',
		'Impossible de mettre à jour le client',
		'No se pudo actualizar el cliente',
		'Der Kunde konnte nicht aktualisiert werden'
	],
	'client.details': ['Details', 'Détails', 'Detalles', 'Details'],
	'client.detailsText': [
		'Contact information and notes.',
		'Coordonnées et notes.',
		'Datos de contacto y notas.',
		'Kontaktdaten und Notizen.'
	],
	'client.notesPlaceholder': [
		'Anything worth remembering about this client…',
		'Tout ce qui vaut la peine d’être noté sur ce client…',
		'Cualquier cosa que convenga recordar sobre este cliente…',
		'Alles, was man über diesen Kunden festhalten sollte…'
	],
	'client.documents': ['Documents', 'Documents', 'Documentos', 'Dokumente'],
	'client.documentCount': [
		{ one: '{count} document linked to this client.', other: '{count} documents linked to this client.' },
		{
			one: '{count} document lié à ce client.',
			many: '{count} documents liés à ce client.',
			other: '{count} documents liés à ce client.'
		},
		{
			one: '{count} documento vinculado a este cliente.',
			many: '{count} documentos vinculados a este cliente.',
			other: '{count} documentos vinculados a este cliente.'
		},
		{
			one: '{count} Dokument mit diesem Kunden verknüpft.',
			other: '{count} Dokumente mit diesem Kunden verknüpft.'
		}
	],
	'client.noDocuments': [
		'No documents linked yet',
		'Aucun document lié pour l’instant',
		'Aún no hay documentos vinculados',
		'Noch keine Dokumente verknüpft'
	],
	'client.noDocumentsText': [
		'Pick this client when you create a document and it will appear here.',
		'Choisissez ce client en créant un document et il apparaîtra ici.',
		'Elige este cliente al crear un documento y aparecerá aquí.',
		'Wähle diesen Kunden beim Erstellen eines Dokuments aus, dann erscheint es hier.'
	]
} satisfies Record<string, Row>;
