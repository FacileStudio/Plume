import type { Row } from '../row';

/** Strings of the dashboard, the document list, the new document form and the document page. */
export const documents = {
	'dashboard.title': ['Dashboard — Plume', 'Tableau de bord — Plume', 'Panel — Plume', 'Übersicht — Plume'],
	'dashboard.heading': ['Dashboard', 'Tableau de bord', 'Panel', 'Übersicht'],
	'dashboard.subtitle': [
		'Everything waiting on a signature, at a glance.',
		'Tout ce qui attend une signature, en un coup d’œil.',
		'Todo lo que espera una firma, de un vistazo.',
		'Alles, was auf eine Unterschrift wartet, auf einen Blick.'
	],
	'dashboard.verify': ['Verify a document', 'Vérifier un document', 'Verificar un documento', 'Dokument prüfen'],
	'dashboard.newDocument': ['New document', 'Nouveau document', 'Nuevo documento', 'Neues Dokument'],
	'dashboard.stat.total': ['Total documents', 'Documents au total', 'Documentos en total', 'Dokumente insgesamt'],
	'dashboard.stat.pending': [
		'Pending signatures',
		'Signatures en attente',
		'Firmas pendientes',
		'Ausstehende Unterschriften'
	],
	'dashboard.stat.completed': ['Completed', 'Terminés', 'Completados', 'Abgeschlossen'],
	'dashboard.chart.created': ['Documents created', 'Documents créés', 'Documentos creados', 'Erstellte Dokumente'],
	'dashboard.chart.series': ['Documents', 'Documents', 'Documentos', 'Dokumente'],
	'dashboard.chart.byStatus': ['By status', 'Par statut', 'Por estado', 'Nach Status'],
	'dashboard.chart.center': ['documents', 'documents', 'documentos', 'Dokumente'],
	'dashboard.recent': ['Recent documents', 'Documents récents', 'Documentos recientes', 'Neueste Dokumente'],
	'dashboard.recentHint': [
		'The ten most recent in this space.',
		'Les dix plus récents de cet espace.',
		'Los diez más recientes de este espacio.',
		'Die zehn neuesten in diesem Bereich.'
	],
	'dashboard.viewAll': [
		'View all documents',
		'Voir tous les documents',
		'Ver todos los documentos',
		'Alle Dokumente anzeigen'
	],
	'documents.title': ['Documents — Plume', 'Documents — Plume', 'Documentos — Plume', 'Dokumente — Plume'],
	'documents.heading': ['Documents', 'Documents', 'Documentos', 'Dokumente'],
	'documents.subtitle': [
		'Everything you have sent out for signature, and everything still in draft.',
		'Tout ce que vous avez envoyé à signer, et tout ce qui est encore en brouillon.',
		'Todo lo que has enviado a firmar y todo lo que sigue en borrador.',
		'Alles, was du zur Unterschrift versendet hast, und alles, was noch ein Entwurf ist.'
	],
	'documents.empty.title': ['No documents yet', 'Aucun document pour l’instant', 'Aún no hay documentos', 'Noch keine Dokumente'],
	'documents.empty.dashboardText': [
		'Upload a PDF, drop a few fields on it, and send it out for signature.',
		'Importez un PDF, placez quelques champs et envoyez-le à signer.',
		'Sube un PDF, coloca algunos campos y envíalo a firmar.',
		'Lade ein PDF hoch, platziere ein paar Felder und sende es zur Unterschrift.'
	],
	'documents.empty.text': [
		'Upload a PDF, place the fields, and send it out for signature.',
		'Importez un PDF, placez les champs et envoyez-le à signer.',
		'Sube un PDF, coloca los campos y envíalo a firmar.',
		'Lade ein PDF hoch, platziere die Felder und sende es zur Unterschrift.'
	],
	'documents.col.document': ['Document', 'Document', 'Documento', 'Dokument'],
	'documents.col.signers': ['Signers', 'Signataires', 'Firmantes', 'Unterzeichner'],
	'documents.signerCount': [
		{ one: '{count} signer', other: '{count} signers' },
		{ one: '{count} signataire', many: '{count} signataires', other: '{count} signataires' },
		{ one: '{count} firmante', many: '{count} firmantes', other: '{count} firmantes' },
		{ one: '{count} Unterzeichner', other: '{count} Unterzeichner' }
	],
	'documents.deleteAria': ['Delete {name}', 'Supprimer {name}', 'Eliminar {name}', '{name} löschen'],
	'documents.deleted': ['Document deleted', 'Document supprimé', 'Documento eliminado', 'Dokument gelöscht'],
	'documents.deleteFailed': [
		'Failed to delete the document',
		'Impossible de supprimer le document',
		'No se pudo eliminar el documento',
		'Das Dokument konnte nicht gelöscht werden'
	],
	'documents.loadFailed': [
		'Failed to load documents',
		'Impossible de charger les documents',
		'No se pudieron cargar los documentos',
		'Die Dokumente konnten nicht geladen werden'
	],
	'documents.confirm.title': ['Delete {name}?', 'Supprimer {name} ?', '¿Eliminar {name}?', '{name} löschen?'],
	'documents.confirm.thisDocument': ['this document', 'ce document', 'este documento', 'dieses Dokument'],
	'documents.confirm.text': [
		'The document, its signers and every field placed on it are removed permanently. This cannot be undone.',
		'Le document, ses signataires et tous les champs placés dessus sont supprimés définitivement. Cette action est irréversible.',
		'El documento, sus firmantes y todos los campos colocados en él se eliminan de forma permanente. No se puede deshacer.',
		'Das Dokument, seine Unterzeichner und alle darauf platzierten Felder werden endgültig entfernt. Das lässt sich nicht rückgängig machen.'
	],
	'documents.confirm.label': ['Delete document', 'Supprimer le document', 'Eliminar documento', 'Dokument löschen'],
	'docNew.title': ['New Document — Plume', 'Nouveau document — Plume', 'Nuevo documento — Plume', 'Neues Dokument — Plume'],
	'docNew.heading': ['New document', 'Nouveau document', 'Nuevo documento', 'Neues Dokument'],
	'docNew.subtitle': [
		'Upload the PDF and list who has to sign it. Fields are placed on the next screen.',
		'Importez le PDF et indiquez qui doit le signer. Les champs se placent à l’écran suivant.',
		'Sube el PDF e indica quién debe firmarlo. Los campos se colocan en la pantalla siguiente.',
		'Lade das PDF hoch und gib an, wer unterschreiben muss. Die Felder platzierst du im nächsten Schritt.'
	],
	'docNew.section.document': ['Document', 'Document', 'Documento', 'Dokument'],
	'docNew.section.documentText': [
		'The PDF your signers will receive.',
		'Le PDF que recevront vos signataires.',
		'El PDF que recibirán tus firmantes.',
		'Das PDF, das deine Unterzeichner erhalten.'
	],
	'docNew.name': ['Document name', 'Nom du document', 'Nombre del documento', 'Dokumentname'],
	'docNew.namePlaceholder': [
		'Contract, NDA, Agreement…',
		'Contrat, NDA, accord…',
		'Contrato, NDA, acuerdo…',
		'Vertrag, NDA, Vereinbarung…'
	],
	'docNew.nameRequired': [
		'Document name is required',
		'Le nom du document est obligatoire',
		'El nombre del documento es obligatorio',
		'Der Dokumentname ist erforderlich'
	],
	'docNew.pdfFile': ['PDF file', 'Fichier PDF', 'Archivo PDF', 'PDF-Datei'],
	'docNew.dropLabel': ['Drop a PDF here', 'Déposez un PDF ici', 'Suelta un PDF aquí', 'PDF hier ablegen'],
	'docNew.browse': ['Browse', 'Parcourir', 'Examinar', 'Durchsuchen'],
	'docNew.dropAnnouncement': [
		'Release to add the PDF',
		'Relâchez pour ajouter le PDF',
		'Suelta para añadir el PDF',
		'Loslassen, um das PDF hinzuzufügen'
	],
	'docNew.hint': [
		'PDF only · 10 MB maximum',
		'PDF uniquement · 10 Mo maximum',
		'Solo PDF · 10 MB como máximo',
		'Nur PDF · maximal 10 MB'
	],
	'docNew.onlyPdf': [
		'Only PDF files are accepted',
		'Seuls les fichiers PDF sont acceptés',
		'Solo se aceptan archivos PDF',
		'Es werden nur PDF-Dateien akzeptiert'
	],
	'docNew.tooLarge': [
		'{name} is too large ({size} MB). Maximum size is 10 MB.',
		'{name} est trop volumineux ({size} Mo). La taille maximale est de 10 Mo.',
		'{name} es demasiado grande ({size} MB). El tamaño máximo es de 10 MB.',
		'{name} ist zu groß ({size} MB). Die maximale Größe beträgt 10 MB.'
	],
	'docNew.onlyOne': [
		'Only one PDF can be attached to a document',
		'Un document ne peut contenir qu’un seul PDF',
		'Solo se puede adjuntar un PDF a un documento',
		'An ein Dokument kann nur ein PDF angehängt werden'
	],
	'docNew.uploadAPdf': [
		'Please upload a PDF file',
		'Veuillez importer un fichier PDF',
		'Sube un archivo PDF',
		'Bitte lade eine PDF-Datei hoch'
	],
	'docNew.createFailed': [
		'Failed to create the document',
		'Impossible de créer le document',
		'No se pudo crear el documento',
		'Das Dokument konnte nicht erstellt werden'
	],
	'docNew.section.signers': ['Signers', 'Signataires', 'Firmantes', 'Unterzeichner'],
	'docNew.section.signersText': [
		'Everyone who has to sign. You can add more later, while the document is still a draft.',
		'Toutes les personnes qui doivent signer. Vous pourrez en ajouter plus tard, tant que le document est un brouillon.',
		'Todas las personas que deben firmar. Puedes añadir más después, mientras el documento siga siendo un borrador.',
		'Alle, die unterschreiben müssen. Du kannst später weitere hinzufügen, solange das Dokument ein Entwurf ist.'
	],
	'docNew.addSigner': ['Add signer', 'Ajouter un signataire', 'Añadir firmante', 'Unterzeichner hinzufügen'],
	'docNew.signerNamePlaceholder': ['Jane Doe', 'Camille Martin', 'Ana García', 'Erika Mustermann'],
	'docNew.signerEmailPlaceholder': ['jane@example.com', 'camille@exemple.fr', 'ana@ejemplo.es', 'erika@beispiel.de'],
	'docNew.removeSigner': [
		'Remove signer {number}',
		'Retirer le signataire {number}',
		'Quitar al firmante {number}',
		'Unterzeichner {number} entfernen'
	],
	'docNew.submit': ['Create document', 'Créer le document', 'Crear documento', 'Dokument erstellen'],
	'doc.titleFallback': ['Plume', 'Plume', 'Plume', 'Plume'],
	'doc.fieldsPrepared': [
		{ one: '{count} field prepared', other: '{count} fields prepared' },
		{ one: '{count} champ préparé', many: '{count} champs préparés', other: '{count} champs préparés' },
		{ one: '{count} campo preparado', many: '{count} campos preparados', other: '{count} campos preparados' },
		{ one: '{count} Feld vorbereitet', other: '{count} Felder vorbereitet' }
	],
	'doc.noFieldsPrepared': [
		'No fields prepared yet',
		'Aucun champ préparé pour l’instant',
		'Aún no hay campos preparados',
		'Noch keine Felder vorbereitet'
	],
	'doc.createdOn': ['Created {date}', 'Créé le {date}', 'Creado el {date}', 'Erstellt am {date}'],
	'doc.client': ['Client', 'Client', 'Cliente', 'Kunde'],
	'doc.linkedClient': ['Linked client', 'Client associé', 'Cliente asociado', 'Verknüpfter Kunde'],
	'doc.noClient': ['No client', 'Aucun client', 'Sin cliente', 'Kein Kunde'],
	'doc.downloadCertificate': [
		'Download certificate',
		'Télécharger le certificat',
		'Descargar el certificado',
		'Zertifikat herunterladen'
	],
	'doc.downloadAudit': ['Download audit', 'Télécharger l’audit', 'Descargar la auditoría', 'Prüfprotokoll herunterladen'],
	'doc.downloadFailed': ['Download failed', 'Échec du téléchargement', 'Error en la descarga', 'Download fehlgeschlagen'],
	'doc.editFields': [
		'Edit fields ({count})',
		'Modifier les champs ({count})',
		'Editar campos ({count})',
		'Felder bearbeiten ({count})'
	],
	'doc.prepareFields': ['Prepare fields', 'Préparer les champs', 'Preparar campos', 'Felder vorbereiten'],
	'doc.send': ['Send for signing', 'Envoyer à signer', 'Enviar a firmar', 'Zur Unterschrift senden'],
	'doc.sent': ['Document sent', 'Document envoyé', 'Documento enviado', 'Dokument gesendet'],
	'doc.sendFailed': [
		'Failed to send the document',
		'Impossible d’envoyer le document',
		'No se pudo enviar el documento',
		'Das Dokument konnte nicht gesendet werden'
	],
	'doc.loadFailed': [
		'Failed to load the document',
		'Impossible de charger le document',
		'No se pudo cargar el documento',
		'Das Dokument konnte nicht geladen werden'
	],
	'doc.fieldsReloadFailed': [
		'Failed to reload the fields',
		'Impossible de recharger les champs',
		'No se pudieron recargar los campos',
		'Die Felder konnten nicht neu geladen werden'
	],
	'doc.order.title': ['Signing order', 'Ordre de signature', 'Orden de firma', 'Signaturreihenfolge'],
	'doc.order.draft': [
		'Choose how signers are invited to sign this document.',
		'Choisissez comment les signataires sont invités à signer ce document.',
		'Elige cómo se invita a los firmantes a firmar este documento.',
		'Lege fest, wie die Unterzeichner zum Unterschreiben dieses Dokuments eingeladen werden.'
	],
	'doc.order.sequentialOn': [
		'Signers are invited one at a time in order.',
		'Les signataires sont invités un par un, dans l’ordre.',
		'Los firmantes son invitados de uno en uno, en orden.',
		'Die Unterzeichner werden nacheinander in der festgelegten Reihenfolge eingeladen.'
	],
	'doc.order.sequentialOff': [
		'All signers were invited simultaneously.',
		'Tous les signataires ont été invités en même temps.',
		'Todos los firmantes fueron invitados a la vez.',
		'Alle Unterzeichner wurden gleichzeitig eingeladen.'
	],
	'doc.order.label': ['Sequential signing', 'Signature séquentielle', 'Firma secuencial', 'Sequenzielle Unterschrift'],
	'doc.order.labelText': [
		'Signers are invited one at a time in order. Signer N+1 is invited only after signer N completes.',
		'Les signataires sont invités un par un, dans l’ordre. Le signataire N+1 n’est invité qu’après la signature du signataire N.',
		'Los firmantes son invitados de uno en uno, en orden. El firmante N+1 solo recibe la invitación cuando el firmante N ha terminado.',
		'Die Unterzeichner werden nacheinander eingeladen. Unterzeichner N+1 wird erst eingeladen, wenn Unterzeichner N fertig ist.'
	],
	'doc.order.failed': [
		'Failed to change the signing order',
		'Impossible de modifier l’ordre de signature',
		'No se pudo cambiar el orden de firma',
		'Die Signaturreihenfolge konnte nicht geändert werden'
	],
	'doc.clientFailed': [
		'Failed to update client',
		'Impossible de mettre à jour le client',
		'No se pudo actualizar el cliente',
		'Der Kunde konnte nicht aktualisiert werden'
	],
	'doc.signers.title': ['Signers', 'Signataires', 'Firmantes', 'Unterzeichner'],
	'doc.signers.count': [
		{ one: '{count} signer on this document.', other: '{count} signers on this document.' },
		{
			one: '{count} signataire sur ce document.',
			many: '{count} signataires sur ce document.',
			other: '{count} signataires sur ce document.'
		},
		{
			one: '{count} firmante en este documento.',
			many: '{count} firmantes en este documento.',
			other: '{count} firmantes en este documento.'
		},
		{
			one: '{count} Unterzeichner für dieses Dokument.',
			other: '{count} Unterzeichner für dieses Dokument.'
		}
	],
	'doc.signers.addDisabled': [
		'Signers can only be added before signing starts',
		'Les signataires ne peuvent être ajoutés qu’avant le début de la signature',
		'Solo se pueden añadir firmantes antes de que empiece la firma',
		'Unterzeichner lassen sich nur vor dem Start der Unterschrift hinzufügen'
	],
	'doc.signers.emptyTitle': ['No signers yet', 'Aucun signataire pour l’instant', 'Aún no hay firmantes', 'Noch keine Unterzeichner'],
	'doc.signers.emptyText': [
		'Add at least one signer before this document can be sent.',
		'Ajoutez au moins un signataire avant d’envoyer ce document.',
		'Añade al menos un firmante antes de poder enviar este documento.',
		'Füge mindestens einen Unterzeichner hinzu, bevor dieses Dokument gesendet werden kann.'
	],
	'doc.signers.fieldsAssigned': [
		{ one: '{count} field assigned', other: '{count} fields assigned' },
		{ one: '{count} champ attribué', many: '{count} champs attribués', other: '{count} champs attribués' },
		{ one: '{count} campo asignado', many: '{count} campos asignados', other: '{count} campos asignados' },
		{ one: '{count} Feld zugewiesen', other: '{count} Felder zugewiesen' }
	],
	'doc.signers.noFieldsAssigned': [
		'No fields assigned',
		'Aucun champ attribué',
		'Ningún campo asignado',
		'Keine Felder zugewiesen'
	],
	'doc.signers.copyLink': ['Copy link', 'Copier le lien', 'Copiar enlace', 'Link kopieren'],
	'doc.signers.copied': ['Copied!', 'Copié !', '¡Copiado!', 'Kopiert!'],
	'doc.signers.copyFailed': ['Copy failed', 'Échec de la copie', 'Error al copiar', 'Kopieren fehlgeschlagen'],
	'doc.signers.resend': ['Resend email', 'Renvoyer l’email', 'Reenviar correo', 'E-Mail erneut senden'],
	'doc.signers.reminded': ['Reminder sent', 'Rappel envoyé', 'Recordatorio enviado', 'Erinnerung gesendet'],
	'doc.signers.remindFailed': [
		'Failed to send reminder',
		'Impossible d’envoyer le rappel',
		'No se pudo enviar el recordatorio',
		'Die Erinnerung konnte nicht gesendet werden'
	],
	'doc.signers.sentAgo': ['Sent {when}', 'Envoyé {when}', 'Enviado {when}', 'Gesendet {when}'],
	'doc.signers.openedDocument': [
		'Opened the document {when}',
		'A ouvert le document {when}',
		'Abrió el documento {when}',
		'Hat das Dokument {when} geöffnet'
	],
	'doc.signers.openedEmail': [
		'Opened the email {when}',
		'A ouvert l’email {when}',
		'Abrió el correo {when}',
		'Hat die E-Mail {when} geöffnet'
	],
	'doc.signers.waiting': [
		'Waiting for previous signer',
		'En attente du signataire précédent',
		'Esperando al firmante anterior',
		'Wartet auf den vorherigen Unterzeichner'
	],
	'doc.signers.signedOn': ['Signed {date}', 'Signé le {date}', 'Firmado el {date}', 'Unterschrieben am {date}'],
	'doc.stage.signed': ['signed', 'signé', 'firmado', 'unterschrieben'],
	'doc.stage.declined': ['declined', 'refusé', 'rechazado', 'abgelehnt'],
	'doc.stage.notSent': ['not sent', 'non envoyé', 'no enviado', 'nicht gesendet'],
	'doc.stage.waiting': ['waiting', 'en attente', 'en espera', 'wartet'],
	'doc.stage.openedDocument': [
		'opened document',
		'document ouvert',
		'documento abierto',
		'Dokument geöffnet'
	],
	'doc.stage.openedEmail': ['opened email', 'email ouvert', 'correo abierto', 'E-Mail geöffnet'],
	'doc.stage.sent': ['sent', 'envoyé', 'enviado', 'gesendet'],
	'doc.notFound.title': ['Document not found', 'Document introuvable', 'Documento no encontrado', 'Dokument nicht gefunden'],
	'doc.notFound.text': [
		'It may have been deleted, or it belongs to another space.',
		'Il a peut-être été supprimé, ou il appartient à un autre espace.',
		'Puede que se haya eliminado o que pertenezca a otro espacio.',
		'Es wurde möglicherweise gelöscht oder gehört zu einem anderen Bereich.'
	],
	'doc.addSigner.title': ['Add signer', 'Ajouter un signataire', 'Añadir firmante', 'Unterzeichner hinzufügen'],
	'doc.addSigner.fill': [
		'Fill from client',
		'Remplir depuis un client',
		'Rellenar desde un cliente',
		'Aus Kunde übernehmen'
	],
	'doc.addSigner.fillHelp': [
		'Copies the client’s name and email into the fields below.',
		'Copie le nom et l’email du client dans les champs ci-dessous.',
		'Copia el nombre y el correo del cliente en los campos de abajo.',
		'Übernimmt Namen und E-Mail des Kunden in die Felder unten.'
	],
	'doc.addSigner.pick': ['Pick a client…', 'Choisir un client…', 'Elegir un cliente…', 'Kunden auswählen…'],
	'doc.addSigner.required': [
		'Name and email are required',
		'Le nom et l’email sont obligatoires',
		'El nombre y el correo son obligatorios',
		'Name und E-Mail sind erforderlich'
	],
	'doc.addSigner.failed': [
		'Failed to add the signer',
		'Impossible d’ajouter le signataire',
		'No se pudo añadir el firmante',
		'Der Unterzeichner konnte nicht hinzugefügt werden'
	],
	'doc.addSigner.added': ['Signer added', 'Signataire ajouté', 'Firmante añadido', 'Unterzeichner hinzugefügt']
} satisfies Record<string, Row>;
