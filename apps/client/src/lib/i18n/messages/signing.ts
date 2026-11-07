import type { Row } from '../row';

/** Strings of the field editor, the signature pad, the public signing page and the verify page. */
export const signing = {
	'field.type.signature': ['Signature', 'Signature', 'Firma', 'Unterschrift'],
	'field.type.text': ['Text', 'Texte', 'Texto', 'Text'],
	'field.type.date': ['Date', 'Date', 'Fecha', 'Datum'],
	'field.type.checkbox': ['Checkbox', 'Case à cocher', 'Casilla', 'Kontrollkästchen'],
	'field.type.other': ['Field', 'Champ', 'Campo', 'Feld'],
	'editor.title': ['Prepare fields', 'Préparer les champs', 'Preparar campos', 'Felder vorbereiten'],
	'editor.pageOf': ['Page {page} / {total}', 'Page {page} / {total}', 'Página {page} / {total}', 'Seite {page} / {total}'],
	'editor.saveClose': ['Save & close', 'Enregistrer et fermer', 'Guardar y cerrar', 'Speichern und schließen'],
	'editor.loadFailed': [
		'Failed to load document',
		'Impossible de charger le document',
		'No se pudo cargar el documento',
		'Das Dokument konnte nicht geladen werden'
	],
	'editor.addFailed': [
		'Could not add the field',
		'Impossible d’ajouter le champ',
		'No se pudo añadir el campo',
		'Das Feld konnte nicht hinzugefügt werden'
	],
	'editor.deleteFailed': [
		'Could not delete the field',
		'Impossible de supprimer le champ',
		'No se pudo eliminar el campo',
		'Das Feld konnte nicht gelöscht werden'
	],
	'editor.unknownSigner': ['Unknown', 'Inconnu', 'Desconocido', 'Unbekannt'],
	'editor.deleteField': ['Delete field', 'Supprimer le champ', 'Eliminar campo', 'Feld löschen'],
	'editor.signer': ['Signer', 'Signataire', 'Firmante', 'Unterzeichner'],
	'editor.addField': ['Add field', 'Ajouter un champ', 'Añadir campo', 'Feld hinzufügen'],
	'editor.placed': [
		'Placed fields ({count})',
		'Champs placés ({count})',
		'Campos colocados ({count})',
		'Platzierte Felder ({count})'
	],
	'editor.fieldLabel': ['Field label', 'Libellé du champ', 'Etiqueta del campo', 'Feldbezeichnung'],
	'editor.pageShort': ['p{page}', 'p{page}', 'p{page}', 'S.{page}'],
	'editor.confirm.title': ['Delete this field?', 'Supprimer ce champ ?', '¿Eliminar este campo?', 'Dieses Feld löschen?'],
	'editor.confirm.text': [
		'The field is removed from the document. Signers will no longer be asked to fill it.',
		'Le champ est retiré du document. Les signataires ne devront plus le remplir.',
		'El campo se elimina del documento. Ya no se pedirá a los firmantes que lo rellenen.',
		'Das Feld wird aus dem Dokument entfernt. Die Unterzeichner müssen es nicht mehr ausfüllen.'
	],
	'pad.hint': [
		'Draw your signature here',
		'Dessinez votre signature ici',
		'Dibuja tu firma aquí',
		'Zeichne deine Unterschrift hier'
	],
	'pad.undo': ['Undo', 'Annuler', 'Deshacer', 'Rückgängig'],
	'pad.clear': ['Clear', 'Effacer', 'Borrar', 'Löschen'],
	'sign.title': ['Sign — {name} — Plume', 'Signer — {name} — Plume', 'Firmar — {name} — Plume', 'Unterschreiben — {name} — Plume'],
	'sign.titleFallback': ['Sign Document — Plume', 'Signer un document — Plume', 'Firmar documento — Plume', 'Dokument unterschreiben — Plume'],
	'sign.notFound.title': ['Link not found', 'Lien introuvable', 'Enlace no encontrado', 'Link nicht gefunden'],
	'sign.notFound.text': [
		'This signing link may be invalid, already used, or expired. Ask the sender for a fresh one.',
		'Ce lien de signature est peut-être invalide, déjà utilisé ou expiré. Demandez-en un nouveau à l’expéditeur.',
		'Puede que este enlace de firma no sea válido, ya se haya usado o haya caducado. Pide uno nuevo al remitente.',
		'Dieser Signaturlink ist möglicherweise ungültig, bereits verwendet oder abgelaufen. Bitte den Absender um einen neuen.'
	],
	'sign.done.completed': ['Document completed', 'Document terminé', 'Documento completado', 'Dokument abgeschlossen'],
	'sign.done.signed': [
		'Document signed successfully',
		'Document signé avec succès',
		'Documento firmado correctamente',
		'Dokument erfolgreich unterschrieben'
	],
	'sign.done.allSigned': [
		'All parties have signed. The document is complete.',
		'Toutes les parties ont signé. Le document est terminé.',
		'Todas las partes han firmado. El documento está completo.',
		'Alle Parteien haben unterschrieben. Das Dokument ist abgeschlossen.'
	],
	'sign.done.waiting': [
		'Your signature is recorded. Waiting for the remaining signers to complete it.',
		'Votre signature est enregistrée. En attente des autres signataires.',
		'Tu firma está registrada. Esperando a que terminen los demás firmantes.',
		'Deine Unterschrift ist erfasst. Es fehlen noch die übrigen Unterzeichner.'
	],
	'sign.done.close': [
		'Your signature is recorded. You can close this page.',
		'Votre signature est enregistrée. Vous pouvez fermer cette page.',
		'Tu firma está registrada. Puedes cerrar esta página.',
		'Deine Unterschrift ist erfasst. Du kannst diese Seite schließen.'
	],
	'sign.declined.title': ['Document declined', 'Document refusé', 'Documento rechazado', 'Dokument abgelehnt'],
	'sign.declined.text': [
		'You have declined to sign this document.',
		'Vous avez refusé de signer ce document.',
		'Has rechazado firmar este documento.',
		'Du hast die Unterschrift unter diesem Dokument abgelehnt.'
	],
	'sign.progress': ['Signing progress', 'Avancement des signatures', 'Progreso de la firma', 'Signaturfortschritt'],
	'sign.progressCount': [
		'{signed} of {total} signed',
		'{signed} sur {total} signé(s)',
		'{signed} de {total} firmados',
		'{signed} von {total} unterschrieben'
	],
	'sign.you': ['You', 'Vous', 'Tú', 'Du'],
	'sign.roster.signedOn': ['Signed · {date}', 'Signé · {date}', 'Firmado · {date}', 'Unterschrieben · {date}'],
	'sign.roster.signed': ['Signed', 'Signé', 'Firmado', 'Unterschrieben'],
	'sign.roster.declined': ['Declined', 'Refusé', 'Rechazado', 'Abgelehnt'],
	'sign.roster.viewer': ['Viewer', 'Lecteur', 'Lector', 'Betrachter'],
	'sign.roster.waiting': [
		'Waiting for signature',
		'En attente de signature',
		'Esperando la firma',
		'Wartet auf Unterschrift'
	],
	'sign.finalized': [
		'Everyone has signed. This document is finalized.',
		'Tout le monde a signé. Ce document est finalisé.',
		'Todos han firmado. Este documento está finalizado.',
		'Alle haben unterschrieben. Dieses Dokument ist abgeschlossen.'
	],
	'sign.stopped': [
		'The signing process was stopped because a signer declined.',
		'Le processus de signature a été arrêté car un signataire a refusé.',
		'El proceso de firma se detuvo porque un firmante lo rechazó.',
		'Der Signaturprozess wurde beendet, weil ein Unterzeichner abgelehnt hat.'
	],
	'sign.waitingOn': [
		'Waiting on {names} to finalize the document.',
		'En attente de {names} pour finaliser le document.',
		'Esperando a {names} para finalizar el documento.',
		'Warten auf {names}, um das Dokument abzuschließen.'
	],
	'sign.loadingPreview': ['Loading preview…', 'Chargement de l’aperçu…', 'Cargando vista previa…', 'Vorschau wird geladen…'],
	'sign.previewUnavailable': [
		'Preview unavailable',
		'Aperçu indisponible',
		'Vista previa no disponible',
		'Vorschau nicht verfügbar'
	],
	'sign.previewUnavailableText': [
		'The document could not be rendered in the browser. You can still download it, review it, and fill the fields on the right.',
		'Le document n’a pas pu être affiché dans le navigateur. Vous pouvez le télécharger, le relire et remplir les champs à droite.',
		'No se pudo mostrar el documento en el navegador. Aún puedes descargarlo, revisarlo y rellenar los campos de la derecha.',
		'Das Dokument konnte im Browser nicht dargestellt werden. Du kannst es trotzdem herunterladen, prüfen und die Felder rechts ausfüllen.'
	],
	'sign.signatureBy': ['Signature by {name}', 'Signature de {name}', 'Firma de {name}', 'Unterschrift von {name}'],
	'sign.signatureAlt': ['Signature', 'Signature', 'Firma', 'Unterschrift'],
	'sign.signingAs': ['Signing as {name}', 'Signature en tant que {name}', 'Firmando como {name}', 'Unterschrift als {name}'],
	'sign.signatureRequired': [
		'A signature is required.',
		'Une signature est obligatoire.',
		'Se requiere una firma.',
		'Eine Unterschrift ist erforderlich.'
	],
	'sign.fieldRequired': [
		'This field is required.',
		'Ce champ est obligatoire.',
		'Este campo es obligatorio.',
		'Dieses Feld ist erforderlich.'
	],
	'sign.checkboxRequired': [
		'This box has to be ticked.',
		'Cette case doit être cochée.',
		'Hay que marcar esta casilla.',
		'Dieses Kästchen muss angekreuzt sein.'
	],
	'sign.useToday': [
		'Use today’s date',
		'Utiliser la date du jour',
		'Usar la fecha de hoy',
		'Heutiges Datum verwenden'
	],
	'sign.today': ['Today', 'Aujourd’hui', 'Hoy', 'Heute'],
	'sign.enterText': ['Enter text', 'Saisissez du texte', 'Introduce texto', 'Text eingeben'],
	'sign.missing.one': [
		'One required field is still empty.',
		'Un champ obligatoire est encore vide.',
		'Falta un campo obligatorio.',
		'Ein Pflichtfeld ist noch leer.'
	],
	'sign.missing.many': [
		'{count} required fields are still empty.',
		'{count} champs obligatoires sont encore vides.',
		'Faltan {count} campos obligatorios.',
		'{count} Pflichtfelder sind noch leer.'
	],
	'sign.missing.title': [
		'Fill the required fields first',
		'Remplissez d’abord les champs obligatoires',
		'Rellena primero los campos obligatorios',
		'Fülle zuerst die Pflichtfelder aus'
	],
	'sign.missing.list': ['Still empty: {names}.', 'Encore vides : {names}.', 'Aún vacíos: {names}.', 'Noch leer: {names}.'],
	'sign.error.title': ['Could not sign', 'Signature impossible', 'No se pudo firmar', 'Unterschrift nicht möglich'],
	'sign.submit': ['Sign & complete', 'Signer et terminer', 'Firmar y completar', 'Unterschreiben und abschließen'],
	'sign.decline': ['Decline', 'Refuser', 'Rechazar', 'Ablehnen'],
	'sign.declineFailed': [
		'Could not decline this document.',
		'Impossible de refuser ce document.',
		'No se pudo rechazar este documento.',
		'Dieses Dokument konnte nicht abgelehnt werden.'
	],
	'sign.confirm.title': ['Decline to sign?', 'Refuser de signer ?', '¿Rechazar la firma?', 'Unterschrift ablehnen?'],
	'sign.confirm.text': [
		'The sender is notified and the signing process stops here. This cannot be undone.',
		'L’expéditeur est prévenu et le processus de signature s’arrête ici. Cette action est irréversible.',
		'Se avisa al remitente y el proceso de firma se detiene aquí. No se puede deshacer.',
		'Der Absender wird benachrichtigt und der Signaturprozess endet hier. Das lässt sich nicht rückgängig machen.'
	],
	'sign.confirm.label': ['Decline', 'Refuser', 'Rechazar', 'Ablehnen'],
	'sign.confirm.cancel': ['Keep reviewing', 'Continuer la lecture', 'Seguir revisando', 'Weiter prüfen'],
	'verify.title': [
		'Verify a document — Plume',
		'Vérifier un document — Plume',
		'Verificar un documento — Plume',
		'Dokument prüfen — Plume'
	],
	'verify.description': [
		'Check whether a PDF was issued and signed through Plume.',
		'Vérifiez si un PDF a été émis et signé avec Plume.',
		'Comprueba si un PDF se emitió y firmó con Plume.',
		'Prüfe, ob ein PDF über Plume ausgestellt und unterschrieben wurde.'
	],
	'verify.login': ['Log in', 'Se connecter', 'Iniciar sesión', 'Anmelden'],
	'verify.dashboard': ['Dashboard', 'Tableau de bord', 'Panel', 'Übersicht'],
	'verify.heading': ['Verify a document', 'Vérifier un document', 'Verificar un documento', 'Dokument prüfen'],
	'verify.intro': [
		'Drop a PDF below. Plume re-computes its SHA-256 fingerprint and checks whether it matches a document we have on file. Nothing is stored.',
		'Déposez un PDF ci-dessous. Plume recalcule son empreinte SHA-256 et vérifie qu’elle correspond à un document enregistré. Rien n’est conservé.',
		'Suelta un PDF abajo. Plume vuelve a calcular su huella SHA-256 y comprueba si coincide con un documento registrado. No se guarda nada.',
		'Lege unten ein PDF ab. Plume berechnet den SHA-256-Fingerabdruck neu und prüft, ob er zu einem gespeicherten Dokument passt. Es wird nichts gespeichert.'
	],
	'verify.hintMax': ['Max 50 MB', 'Max. 50 Mo', 'Máx. 50 MB', 'Max. 50 MB'],
	'verify.tooLarge': [
		'That file is larger than 50 MB.',
		'Ce fichier dépasse 50 Mo.',
		'Ese archivo supera los 50 MB.',
		'Diese Datei ist größer als 50 MB.'
	],
	'verify.oneAtATime': [
		'Verify one document at a time.',
		'Vérifiez un document à la fois.',
		'Verifica un documento cada vez.',
		'Prüfe jeweils nur ein Dokument.'
	],
	'verify.onlyPdf': [
		'Only PDF files can be verified.',
		'Seuls les fichiers PDF peuvent être vérifiés.',
		'Solo se pueden verificar archivos PDF.',
		'Es können nur PDF-Dateien geprüft werden.'
	],
	'verify.failed': [
		'Failed to verify document',
		'Impossible de vérifier le document',
		'No se pudo verificar el documento',
		'Das Dokument konnte nicht geprüft werden'
	],
	'verify.errorTitle': ['Could not verify', 'Vérification impossible', 'No se pudo verificar', 'Prüfung nicht möglich'],
	'verify.verifying': ['Verifying…', 'Vérification…', 'Verificando…', 'Wird geprüft…'],
	'verify.submit': ['Verify document', 'Vérifier le document', 'Verificar documento', 'Dokument prüfen'],
	'verify.authentic': [
		'Authentic — issued by Plume',
		'Authentique, émis par Plume',
		'Auténtico, emitido por Plume',
		'Echt, von Plume ausgestellt'
	],
	'verify.matchSigned': [
		'This document matches our records. You uploaded the signed version.',
		'Ce document correspond à nos enregistrements. Vous avez importé la version signée.',
		'Este documento coincide con nuestros registros. Has subido la versión firmada.',
		'Dieses Dokument stimmt mit unseren Aufzeichnungen überein. Du hast die unterschriebene Version hochgeladen.'
	],
	'verify.matchOriginal': [
		'This document matches our records. You uploaded the original version.',
		'Ce document correspond à nos enregistrements. Vous avez importé la version originale.',
		'Este documento coincide con nuestros registros. Has subido la versión original.',
		'Dieses Dokument stimmt mit unseren Aufzeichnungen überein. Du hast die Originalversion hochgeladen.'
	],
	'verify.document': ['Document', 'Document', 'Documento', 'Dokument'],
	'verify.file': ['File', 'Fichier', 'Archivo', 'Datei'],
	'verify.uploaded': ['Uploaded', 'Importé', 'Subido', 'Hochgeladen'],
	'verify.completed': ['Completed', 'Terminé', 'Completado', 'Abgeschlossen'],
	'verify.signers': ['Signers', 'Signataires', 'Firmantes', 'Unterzeichner'],
	'verify.notRecognized': ['Not recognized', 'Non reconnu', 'No reconocido', 'Nicht erkannt'],
	'verify.notRecognizedText': [
		'Plume has no document matching this fingerprint. The file may have been altered after signing, was not issued through this server, or belongs to a different Plume instance.',
		'Plume n’a aucun document correspondant à cette empreinte. Le fichier a pu être modifié après la signature, n’a pas été émis par ce serveur, ou appartient à une autre instance de Plume.',
		'Plume no tiene ningún documento que coincida con esta huella. Puede que el archivo se haya modificado tras la firma, que no se emitiera en este servidor o que pertenezca a otra instancia de Plume.',
		'Plume hat kein Dokument mit diesem Fingerabdruck. Die Datei wurde möglicherweise nach der Unterschrift verändert, nicht über diesen Server ausgestellt oder gehört zu einer anderen Plume-Instanz.'
	],
	'verify.fingerprint': ['SHA-256 fingerprint', 'Empreinte SHA-256', 'Huella SHA-256', 'SHA-256-Fingerabdruck'],
	'verify.explain': [
		'Verification compares the SHA-256 hash of the uploaded file against hashes Plume recorded at upload and at completion. Any change to a single byte produces a different fingerprint.',
		'La vérification compare l’empreinte SHA-256 du fichier importé aux empreintes enregistrées par Plume à l’import et à la fin de la signature. Le moindre octet modifié produit une empreinte différente.',
		'La verificación compara el hash SHA-256 del archivo subido con los hashes que Plume registró al subirlo y al completarlo. Cualquier cambio en un solo byte produce una huella distinta.',
		'Die Prüfung vergleicht den SHA-256-Hash der hochgeladenen Datei mit den Hashes, die Plume beim Hochladen und beim Abschluss erfasst hat. Schon die Änderung eines einzigen Bytes ergibt einen anderen Fingerabdruck.'
	]
} satisfies Record<string, Row>;
