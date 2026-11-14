import type { Row } from '../row';

/** Strings of the public landing page. */
export const landing = {
	'landing.title': [
		'Plume · Document signing on your own server',
		'Plume · La signature de documents sur votre serveur',
		'Plume · Firma de documentos en tu propio servidor',
		'Plume · Dokumente signieren auf dem eigenen Server'
	],
	'landing.description': [
		'Upload a PDF, place the fields and send the link. Plume returns the signed copy with its proof.',
		"Importez un PDF, placez les champs et envoyez le lien. Plume vous rend la copie signée avec sa preuve.",
		'Sube un PDF, coloca los campos y envía el enlace. Plume te devuelve la copia firmada con su prueba.',
		'PDF hochladen, Felder platzieren, Link senden. Plume liefert die signierte Kopie samt Nachweis.'
	],
	'landing.heroTitle': [
		'Get your PDFs signed',
		'Faites signer vos PDF',
		'Haz firmar tus PDF',
		'PDFs unterschreiben lassen'
	],
	'landing.heroText': [
		'Upload a PDF, place the fields, send the link. You get back a signed copy.',
		'Importez un PDF, placez les champs, envoyez le lien. Vous recevez la copie signée.',
		'Sube un PDF, coloca los campos y envía el enlace. Recibes la copia firmada.',
		'PDF hochladen, Felder platzieren, Link senden. Zurück kommt die signierte Kopie.'
	],
	'landing.demo.label': [
		'A signature field is placed on a contract, sent, then signed.',
		'Un champ de signature est placé sur un contrat, envoyé, puis signé.',
		'Se coloca un campo de firma en un contrato, se envía y se firma.',
		'Ein Signaturfeld wird auf einen Vertrag gesetzt, versendet und unterschrieben.'
	],
	'landing.demo.document': [
		'Service agreement',
		'Contrat de prestation',
		'Contrato de servicios',
		'Dienstleistungsvertrag'
	],
	'landing.feature.fields.title': ['Fields on any page', 'Des champs sur chaque page', 'Campos en cualquier página', 'Felder auf jeder Seite'],
	'landing.feature.fields.text': [
		'Drop signature, text, date and checkbox fields where each signer needs them.',
		'Posez signature, texte, date et case à cocher là où chaque signataire en a besoin.',
		'Coloca firma, texto, fecha y casillas donde cada firmante los necesite.',
		'Signatur-, Text-, Datums- und Kontrollkästchenfelder genau dort, wo jeder Unterzeichner sie braucht.'
	],
	'landing.feature.order.title': ['Signing order', 'Ordre de signature', 'Orden de firma', 'Reihenfolge'],
	'landing.feature.order.text': [
		'Signers sign in parallel or one after another, and Plume reminds the late ones.',
		"Les signataires signent en parallèle ou l'un après l'autre, et Plume relance les retardataires.",
		'Los firmantes firman en paralelo o uno tras otro, y Plume recuerda a los rezagados.',
		'Parallel oder nacheinander unterschreiben, und Plume erinnert die Säumigen.'
	],
	'landing.feature.proof.title': ['Proof you can check', 'Une preuve vérifiable', 'Una prueba verificable', 'Prüfbarer Nachweis'],
	'landing.feature.proof.text': [
		'Each signed PDF comes with a certificate and an audit trail. Anyone can verify it by its hash.',
		"Chaque PDF signé est livré avec un certificat et un journal d'audit. Tout le monde peut le vérifier par son empreinte.",
		'Cada PDF firmado incluye un certificado y un registro de auditoría. Cualquiera puede verificarlo por su hash.',
		'Jedes signierte PDF kommt mit Zertifikat und Prüfprotokoll. Jeder kann es über seinen Hash prüfen.'
	],
	'landing.feature.hosted.title': ['Self-hosted', 'Auto-hébergé', 'Autoalojado', 'Selbst gehostet'],
	'landing.feature.hosted.text': [
		'Your PDFs stay on your server. No third-party signing service sees them.',
		'Vos PDF restent sur votre serveur. Aucun service de signature tiers ne les voit.',
		'Tus PDF se quedan en tu servidor. Ningún servicio de firma externo los ve.',
		'Deine PDFs bleiben auf deinem Server. Kein fremder Signaturdienst sieht sie.'
	],
	'landing.ctaTitle': ['Send your first document', 'Envoyez votre premier document', 'Envía tu primer documento', 'Sende dein erstes Dokument'],
	'landing.ctaText': [
		'Sign in with your Facile account.',
		'Connectez-vous avec votre compte Facile.',
		'Inicia sesión con tu cuenta de Facile.',
		'Melde dich mit deinem Facile-Konto an.'
	],
	'landing.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise']
} satisfies Record<string, Row>;
