import type { Row } from '../row';

/** Strings of the settings pages. */
export const settings = {
	'settings.heading': ['Settings', 'Paramètres', 'Ajustes', 'Einstellungen'],
	'settings.subtitle': [
		'Your account, this browser, and how Plume reaches the outside world.',
		'Votre compte, ce navigateur, et la façon dont Plume communique avec l’extérieur.',
		'Tu cuenta, este navegador y cómo se comunica Plume con el exterior.',
		'Dein Konto, dieser Browser und wie Plume mit der Außenwelt kommuniziert.'
	],
	'settings.sections': ['Settings sections', 'Sections des paramètres', 'Secciones de ajustes', 'Einstellungsbereiche'],
	'settings.tab.profile': ['Profile', 'Profil', 'Perfil', 'Profil'],
	'settings.tab.appearance': ['Appearance', 'Apparence', 'Apariencia', 'Darstellung'],
	'settings.tab.notifications': ['Notifications', 'Notifications', 'Notificaciones', 'Benachrichtigungen'],
	'settings.tab.api': ['API', 'API', 'API', 'API'],
	'settings.tab.integrations': ['Integrations', 'Intégrations', 'Integraciones', 'Integrationen'],
	'settings.tab.advanced': ['Advanced', 'Avancé', 'Avanzado', 'Erweitert'],
	'profile.title': ['Profile — Plume', 'Profil — Plume', 'Perfil — Plume', 'Profil — Plume'],
	'profile.account': ['Account', 'Compte', 'Cuenta', 'Konto'],
	'profile.memberSince': ['Member since', 'Membre depuis', 'Miembro desde', 'Mitglied seit'],
	'profile.photo': ['Photo', 'Photo', 'Foto', 'Foto'],
	'profile.photoSso': ['Facile account', 'Compte Facile', 'Cuenta de Facile', 'Facile-Konto'],
	'profile.photoNone': ['None yet', 'Aucune pour l’instant', 'Ninguna por ahora', 'Noch keines'],
	'profile.avatarSso': [
		'Your photo comes from your Facile account. Change it there and it shows here on your next page load.',
		'Votre photo vient de votre compte Facile. Modifiez-la là-bas et elle s’affichera ici au prochain chargement de page.',
		'Tu foto viene de tu cuenta de Facile. Cámbiala allí y se mostrará aquí al cargar la página de nuevo.',
		'Dein Foto stammt aus deinem Facile-Konto. Ändere es dort, dann erscheint es hier beim nächsten Laden der Seite.'
	],
	'profile.avatarNone': [
		'No photo on your Facile account yet — add one there and it appears here.',
		'Pas encore de photo sur votre compte Facile : ajoutez-en une là-bas et elle apparaîtra ici.',
		'Tu cuenta de Facile aún no tiene foto: añade una allí y aparecerá aquí.',
		'Dein Facile-Konto hat noch kein Foto — füge dort eines hinzu, dann erscheint es hier.'
	],
	'profile.loadFailed': [
		'Failed to load profile',
		'Impossible de charger le profil',
		'No se pudo cargar el perfil',
		'Das Profil konnte nicht geladen werden'
	],
	'profile.updated': ['Profile updated.', 'Profil mis à jour.', 'Perfil actualizado.', 'Profil aktualisiert.'],
	'profile.updateFailed': [
		'Failed to update profile',
		'Impossible de mettre à jour le profil',
		'No se pudo actualizar el perfil',
		'Das Profil konnte nicht aktualisiert werden'
	],
	'profile.identity': ['Identity', 'Identité', 'Identidad', 'Identität'],
	'profile.nameText': [
		'Shown on the documents you send and in the signing emails.',
		'Affiché sur les documents que vous envoyez et dans les emails de signature.',
		'Se muestra en los documentos que envías y en los correos de firma.',
		'Wird auf den Dokumenten, die du sendest, und in den Signatur-E-Mails angezeigt.'
	],
	'profile.namePlaceholder': ['Your name', 'Votre nom', 'Tu nombre', 'Dein Name'],
	'profile.emailText': [
		'Where signing notifications and test emails are delivered.',
		'Là où arrivent les notifications de signature et les emails de test.',
		'Donde llegan las notificaciones de firma y los correos de prueba.',
		'Hierhin werden Signaturbenachrichtigungen und Test-E-Mails zugestellt.'
	],
	'profile.emailPlaceholder': ['you@example.com', 'vous@exemple.fr', 'tu@ejemplo.es', 'du@beispiel.de'],
	'profile.session': ['Session', 'Session', 'Sesión', 'Sitzung'],
	'profile.sessionText': [
		'Sessions are per browser. Signing out here leaves your other devices alone.',
		'Les sessions sont propres à chaque navigateur. Se déconnecter ici n’affecte pas vos autres appareils.',
		'Las sesiones son por navegador. Cerrar sesión aquí no afecta a tus otros dispositivos.',
		'Sitzungen gelten pro Browser. Die Abmeldung hier lässt deine anderen Geräte unberührt.'
	],
	'profile.logout': ['Sign out', 'Se déconnecter', 'Cerrar sesión', 'Abmelden'],
	'profile.logoutText': [
		'Ends this session and returns you to the sign-in page. Documents already sent keep collecting signatures.',
		'Termine cette session et vous ramène à la page de connexion. Les documents déjà envoyés continuent de recueillir des signatures.',
		'Termina esta sesión y te lleva a la página de inicio de sesión. Los documentos ya enviados siguen recogiendo firmas.',
		'Beendet diese Sitzung und führt dich zur Anmeldeseite. Bereits gesendete Dokumente sammeln weiter Unterschriften.'
	],
	'appearance.title': ['Appearance — Plume', 'Apparence — Plume', 'Apariencia — Plume', 'Darstellung — Plume'],
	'appearance.theme': ['Theme', 'Thème', 'Tema', 'Design'],
	'appearance.themeText': [
		'Applied to this browser. Every device you sign in from keeps its own choice.',
		'Appliqué à ce navigateur. Chaque appareil depuis lequel vous vous connectez garde son propre choix.',
		'Se aplica a este navegador. Cada dispositivo desde el que inicias sesión conserva su propia elección.',
		'Gilt für diesen Browser. Jedes Gerät, mit dem du dich anmeldest, behält seine eigene Wahl.'
	],
	'appearance.scheme': ['Colour scheme', 'Jeu de couleurs', 'Esquema de color', 'Farbschema'],
	'appearance.schemeText': [
		'System follows your operating system setting and changes with it.',
		'« Système » suit le réglage de votre système d’exploitation et change avec lui.',
		'«Sistema» sigue el ajuste de tu sistema operativo y cambia con él.',
		'„System“ folgt der Einstellung deines Betriebssystems und ändert sich mit ihr.'
	],
	'appearance.system': ['System', 'Système', 'Sistema', 'System'],
	'appearance.light': ['Light', 'Clair', 'Claro', 'Hell'],
	'appearance.dark': ['Dark', 'Sombre', 'Oscuro', 'Dunkel'],
	'notifications.title': ['Notifications — Plume', 'Notifications — Plume', 'Notificaciones — Plume', 'Benachrichtigungen — Plume'],
	'notifications.reminders': ['Reminders', 'Rappels', 'Recordatorios', 'Erinnerungen'],
	'notifications.remindersText': [
		'Plume re-sends the signing invitation to anyone who has not signed yet. Sending needs a working SMTP relay.',
		'Plume renvoie l’invitation à signer à toute personne qui n’a pas encore signé. L’envoi nécessite un relais SMTP fonctionnel.',
		'Plume reenvía la invitación a firmar a quien aún no ha firmado. El envío requiere un relé SMTP que funcione.',
		'Plume sendet die Einladung zur Unterschrift erneut an alle, die noch nicht unterschrieben haben. Dafür ist ein funktionierendes SMTP-Relay nötig.'
	],
	'notifications.statusText': [
		'Reminders run in the background, one pass per interval.',
		'Les rappels s’exécutent en arrière-plan, un passage par intervalle.',
		'Los recordatorios se ejecutan en segundo plano, una pasada por intervalo.',
		'Erinnerungen laufen im Hintergrund, ein Durchlauf pro Intervall.'
	],
	'notifications.every': [
		{ one: 'Every {count} day', other: 'Every {count} days' },
		{ one: 'Tous les {count} jour', many: 'Tous les {count} jours', other: 'Tous les {count} jours' },
		{ one: 'Cada {count} día', many: 'Cada {count} días', other: 'Cada {count} días' },
		{ one: 'Alle {count} Tag', other: 'Alle {count} Tage' }
	],
	'notifications.interval': ['Reminder interval', 'Intervalle de rappel', 'Intervalo de recordatorio', 'Erinnerungsintervall'],
	'notifications.intervalText': [
		'In days, between 0 and 30. Zero disables automatic reminders entirely.',
		'En jours, entre 0 et 30. Zéro désactive complètement les rappels automatiques.',
		'En días, entre 0 y 30. Cero desactiva por completo los recordatorios automáticos.',
		'In Tagen, zwischen 0 und 30. Null schaltet die automatischen Erinnerungen ganz ab.'
	],
	'notifications.intervalHelp': [
		'0 to disable automatic reminders',
		'0 pour désactiver les rappels automatiques',
		'0 para desactivar los recordatorios automáticos',
		'0, um automatische Erinnerungen zu deaktivieren'
	],
	'notifications.intervalAria': [
		'Reminder interval in days',
		'Intervalle de rappel en jours',
		'Intervalo de recordatorio en días',
		'Erinnerungsintervall in Tagen'
	],
	'notifications.notLoaded': [
		'Profile not loaded',
		'Profil non chargé',
		'Perfil no cargado',
		'Profil nicht geladen'
	],
	'notifications.range': [
		'Interval must be between 0 and 30 days',
		'L’intervalle doit être compris entre 0 et 30 jours',
		'El intervalo debe estar entre 0 y 30 días',
		'Das Intervall muss zwischen 0 und 30 Tagen liegen'
	],
	'notifications.saved': [
		'Reminder settings saved',
		'Réglages des rappels enregistrés',
		'Ajustes de recordatorios guardados',
		'Erinnerungseinstellungen gespeichert'
	],
	'notifications.saveFailed': [
		'Failed to save reminder settings',
		'Impossible d’enregistrer les réglages des rappels',
		'No se pudieron guardar los ajustes de recordatorios',
		'Die Erinnerungseinstellungen konnten nicht gespeichert werden'
	],
	'api.title': ['API — Plume', 'API — Plume', 'API — Plume', 'API — Plume'],
	'api.reference': ['Reference', 'Référence', 'Referencia', 'Referenz'],
	'api.referenceText': [
		'Every route this instance serves, generated from the handlers themselves rather than kept by hand.',
		'Toutes les routes servies par cette instance, générées à partir des handlers eux-mêmes plutôt que maintenues à la main.',
		'Todas las rutas que sirve esta instancia, generadas a partir de los propios manejadores y no mantenidas a mano.',
		'Alle Routen dieser Instanz, aus den Handlern selbst erzeugt statt von Hand gepflegt.'
	],
	'api.openReference': ['Open the reference', 'Ouvrir la référence', 'Abrir la referencia', 'Referenz öffnen'],
	'api.basePath': ['Base path', 'Chemin de base', 'Ruta base', 'Basispfad'],
	'api.basePathText': [
		'Every endpoint is scoped under it, on this same origin.',
		'Chaque point d’accès se trouve sous ce chemin, sur cette même origine.',
		'Cada endpoint cuelga de ella, en este mismo origen.',
		'Jeder Endpunkt liegt darunter, auf demselben Origin.'
	],
	'api.openapi': ['OpenAPI document', 'Document OpenAPI', 'Documento OpenAPI', 'OpenAPI-Dokument'],
	'api.openapiText': [
		'The machine-readable schema behind the reference — feed it to a client generator.',
		'Le schéma lisible par machine derrière la référence : donnez-le à un générateur de client.',
		'El esquema legible por máquina detrás de la referencia: pásalo a un generador de clientes.',
		'Das maschinenlesbare Schema hinter der Referenz — übergib es an einen Client-Generator.'
	],
	'api.auth': ['Authentication', 'Authentification', 'Autenticación', 'Authentifizierung'],
	'api.authText': [
		'How a request proves who it is.',
		'Comment une requête prouve son identité.',
		'Cómo demuestra una petición quién es.',
		'Wie sich eine Anfrage ausweist.'
	],
	'api.scheme': ['Scheme', 'Schéma', 'Esquema', 'Schema'],
	'api.schemeText': [
		'Send the token as Authorization: Bearer <token>. Browser sessions and API tokens are accepted the same way.',
		'Envoyez le jeton sous la forme Authorization: Bearer <token>. Les sessions du navigateur et les jetons d’API sont acceptés de la même façon.',
		'Envía el token como Authorization: Bearer <token>. Las sesiones del navegador y los tokens de API se aceptan igual.',
		'Sende das Token als Authorization: Bearer <token>. Browsersitzungen und API-Tokens werden gleich behandelt.'
	],
	'api.tokens': ['API tokens', 'Jetons d’API', 'Tokens de API', 'API-Tokens'],
	'api.tokensText': [
		'Named machine credentials for other apps and scripts, such as Ardoise. A token acts as your account, has no expiry and survives a sign-out.',
		'Identifiants machine nommés pour d’autres apps et scripts, comme Ardoise. Un jeton agit comme votre compte, n’expire pas et survit à une déconnexion.',
		'Credenciales de máquina con nombre para otras apps y scripts, como Ardoise. Un token actúa como tu cuenta, no caduca y sobrevive al cierre de sesión.',
		'Benannte Maschinenzugänge für andere Apps und Skripte, etwa Ardoise. Ein Token handelt als dein Konto, läuft nie ab und übersteht eine Abmeldung.'
	],
	'api.generate': ['Generate a token', 'Générer un jeton', 'Generar un token', 'Token erzeugen'],
	'api.col.name': ['Name', 'Nom', 'Nombre', 'Name'],
	'api.col.created': ['Created', 'Création', 'Creado', 'Erstellt'],
	'api.col.lastUsed': ['Last used', 'Dernière utilisation', 'Último uso', 'Zuletzt verwendet'],
	'api.col.action': ['Action', 'Action', 'Acción', 'Aktion'],
	'api.revoke': ['Revoke', 'Révoquer', 'Revocar', 'Widerrufen'],
	'api.noToken': ['No token', 'Aucun jeton', 'Ningún token', 'Kein Token'],
	'api.noTokenText': [
		'A token is shown once, when you create it. Plume stores only a hash and cannot show it again.',
		'Un jeton n’est affiché qu’une fois, à sa création. Plume ne conserve qu’une empreinte et ne peut pas le réafficher.',
		'Un token se muestra una sola vez, al crearlo. Plume solo guarda un hash y no puede mostrarlo de nuevo.',
		'Ein Token wird nur einmal angezeigt, beim Erstellen. Plume speichert nur einen Hash und kann ihn nicht erneut zeigen.'
	],
	'api.noneGenerated': ['None generated', 'Aucun généré', 'Ninguno generado', 'Keines erzeugt'],
	'api.loadFailed': [
		'Failed to load tokens',
		'Impossible de charger les jetons',
		'No se pudieron cargar los tokens',
		'Die Tokens konnten nicht geladen werden'
	],
	'api.generated': ['Token generated', 'Jeton généré', 'Token generado', 'Token erzeugt'],
	'api.generateFailed': [
		'Failed to generate the token',
		'Impossible de générer le jeton',
		'No se pudo generar el token',
		'Das Token konnte nicht erzeugt werden'
	],
	'api.revoked': ['Token revoked', 'Jeton révoqué', 'Token revocado', 'Token widerrufen'],
	'api.revokeFailed': [
		'Failed to revoke the token',
		'Impossible de révoquer le jeton',
		'No se pudo revocar el token',
		'Das Token konnte nicht widerrufen werden'
	],
	'api.drawer.title': [
		'Generate an API token',
		'Générer un jeton d’API',
		'Generar un token de API',
		'API-Token erzeugen'
	],
	'api.drawer.text': [
		'Name it after the app or machine that will use it. The name is the only way to know later what you are revoking.',
		'Nommez-le d’après l’app ou la machine qui l’utilisera. Le nom est le seul moyen de savoir plus tard ce que vous révoquez.',
		'Ponle el nombre de la app o la máquina que lo usará. El nombre es la única forma de saber luego qué estás revocando.',
		'Benenne es nach der App oder dem Rechner, der es nutzt. Nur am Namen erkennst du später, was du widerrufst.'
	],
	'api.copyNow': ['Copy it now', 'Copiez-le maintenant', 'Cópialo ahora', 'Jetzt kopieren'],
	'api.copyNowText': [
		'This token is never shown again. Plume keeps only a hash, so once this drawer closes you will have to generate another.',
		'Ce jeton ne sera plus jamais affiché. Plume n’en garde qu’une empreinte : une fois ce panneau fermé, vous devrez en générer un autre.',
		'Este token no se volverá a mostrar. Plume solo guarda un hash, así que al cerrar este panel tendrás que generar otro.',
		'Dieses Token wird nie wieder angezeigt. Plume speichert nur einen Hash, nach dem Schließen musst du ein neues erzeugen.'
	],
	'api.tokenLabel': ['Token', 'Jeton', 'Token', 'Token'],
	'api.tokenName': ['Token name', 'Nom du jeton', 'Nombre del token', 'Tokenname'],
	'api.tokenNamePlaceholder': [
		'ardoise, ci, laptop',
		'ardoise, ci, portable',
		'ardoise, ci, portátil',
		'ardoise, ci, Laptop'
	],
	'api.tokenWarning': [
		'A token carries everything your account can do and has no scope or expiry. Revoke it as soon as the app that uses it is gone.',
		'Un jeton porte tout ce que votre compte peut faire et n’a ni portée ni expiration. Révoquez-le dès que l’app qui l’utilise disparaît.',
		'Un token lleva todo lo que puede hacer tu cuenta y no tiene alcance ni caducidad. Revócalo en cuanto desaparezca la app que lo usa.',
		'Ein Token kann alles, was dein Konto kann, und hat weder Geltungsbereich noch Ablauf. Widerrufe es, sobald die nutzende App wegfällt.'
	],
	'api.generateShort': ['Generate', 'Générer', 'Generar', 'Erzeugen'],
	'api.generating': ['Generating…', 'Génération…', 'Generando…', 'Wird erzeugt…'],
	'api.confirm.title': ['Revoke this token?', 'Révoquer ce jeton ?', '¿Revocar este token?', 'Dieses Token widerrufen?'],
	'api.confirm.text': [
		'Every app or script using {name} stops working at once, and a revoked token cannot be restored. You will have to generate a new one and paste it into each place that used the old one.',
		'Toute app ou tout script qui utilise {name} cesse de fonctionner immédiatement, et un jeton révoqué ne peut pas être rétabli. Il faudra en générer un nouveau et le coller partout où l’ancien servait.',
		'Toda app o script que use {name} deja de funcionar al instante, y un token revocado no se puede restaurar. Tendrás que generar uno nuevo y pegarlo en cada lugar donde se usaba el anterior.',
		'Jede App und jedes Skript, die {name} verwenden, funktionieren sofort nicht mehr, und ein widerrufenes Token lässt sich nicht wiederherstellen. Du musst ein neues erzeugen und überall dort einfügen, wo das alte verwendet wurde.'
	],
	'api.confirm.it': ['it', 'lui', 'él', 'es'],
	'integrations.title': ['Integrations — Plume', 'Intégrations — Plume', 'Integraciones — Plume', 'Integrationen — Plume'],
	'smtp.title': ['Email (SMTP)', 'Email (SMTP)', 'Correo (SMTP)', 'E-Mail (SMTP)'],
	'smtp.text': [
		'The relay Plume sends signing invitations, reminders and completion notices through.',
		'Le relais par lequel Plume envoie les invitations à signer, les rappels et les avis de finalisation.',
		'El relé por el que Plume envía las invitaciones a firmar, los recordatorios y los avisos de finalización.',
		'Das Relay, über das Plume Signatureinladungen, Erinnerungen und Abschlussmitteilungen versendet.'
	],
	'smtp.sendTest': ['Send a test', 'Envoyer un test', 'Enviar una prueba', 'Test senden'],
	'smtp.statusText': [
		'Saved credentials, not a live connection check.',
		'Identifiants enregistrés, pas un test de connexion en direct.',
		'Credenciales guardadas, no una comprobación de conexión en vivo.',
		'Gespeicherte Zugangsdaten, keine Live-Verbindungsprüfung.'
	],
	'smtp.configured': ['Configured — {host}', 'Configuré — {host}', 'Configurado — {host}', 'Konfiguriert — {host}'],
	'smtp.notConfigured': ['Not configured', 'Non configuré', 'Sin configurar', 'Nicht konfiguriert'],
	'smtp.host': ['Host', 'Hôte', 'Servidor', 'Host'],
	'smtp.hostText': [
		'The relay’s hostname, without a scheme.',
		'Le nom d’hôte du relais, sans schéma.',
		'El nombre de host del relé, sin esquema.',
		'Der Hostname des Relays, ohne Schema.'
	],
	'smtp.hostAria': ['SMTP host', 'Hôte SMTP', 'Servidor SMTP', 'SMTP-Host'],
	'smtp.port': ['Port', 'Port', 'Puerto', 'Port'],
	'smtp.portText': [
		'587 for STARTTLS, 465 for implicit TLS.',
		'587 pour STARTTLS, 465 pour TLS implicite.',
		'587 para STARTTLS, 465 para TLS implícito.',
		'587 für STARTTLS, 465 für implizites TLS.'
	],
	'smtp.portAria': ['SMTP port', 'Port SMTP', 'Puerto SMTP', 'SMTP-Port'],
	'smtp.username': ['Username', 'Nom d’utilisateur', 'Usuario', 'Benutzername'],
	'smtp.usernameText': [
		'Leave empty for a relay that takes no auth.',
		'Laissez vide pour un relais sans authentification.',
		'Déjalo vacío si el relé no usa autenticación.',
		'Leer lassen für ein Relay ohne Authentifizierung.'
	],
	'smtp.usernameAria': ['SMTP username', 'Utilisateur SMTP', 'Usuario SMTP', 'SMTP-Benutzername'],
	'smtp.password': ['Password', 'Mot de passe', 'Contraseña', 'Passwort'],
	'smtp.passwordKeep': [
		'Plume never sends a stored password back to the browser. Leave this empty to keep the one on file, or type a new one to replace it.',
		'Plume ne renvoie jamais un mot de passe enregistré au navigateur. Laissez vide pour conserver celui en place, ou saisissez-en un nouveau pour le remplacer.',
		'Plume nunca devuelve al navegador una contraseña guardada. Déjalo vacío para conservar la actual o escribe una nueva para sustituirla.',
		'Plume sendet ein gespeichertes Passwort nie an den Browser zurück. Leer lassen, um das vorhandene zu behalten, oder ein neues eingeben, um es zu ersetzen.'
	],
	'smtp.passwordNew': [
		'Stored server-side and never returned once saved.',
		'Conservé côté serveur et jamais renvoyé une fois enregistré.',
		'Se guarda en el servidor y no se devuelve una vez guardado.',
		'Wird serverseitig gespeichert und nach dem Speichern nie zurückgegeben.'
	],
	'smtp.fromEmail': ['From email', 'Email d’expédition', 'Correo del remitente', 'Absender-E-Mail'],
	'smtp.fromEmailText': [
		'The envelope sender every Plume email uses.',
		'L’expéditeur d’enveloppe utilisé par tous les emails de Plume.',
		'El remitente de sobre que usan todos los correos de Plume.',
		'Der Envelope-Absender, den alle Plume-E-Mails verwenden.'
	],
	'smtp.fromName': ['From name', 'Nom d’expéditeur', 'Nombre del remitente', 'Absendername'],
	'smtp.fromNameText': [
		'The display name recipients see.',
		'Le nom d’affichage que voient les destinataires.',
		'El nombre que ven los destinatarios.',
		'Der Anzeigename, den Empfänger sehen.'
	],
	'smtp.deleteConfig': [
		'Delete configuration',
		'Supprimer la configuration',
		'Eliminar configuración',
		'Konfiguration löschen'
	],
	'smtp.saved': [
		'SMTP configuration saved',
		'Configuration SMTP enregistrée',
		'Configuración SMTP guardada',
		'SMTP-Konfiguration gespeichert'
	],
	'smtp.saveFailed': [
		'Failed to save SMTP configuration',
		'Impossible d’enregistrer la configuration SMTP',
		'No se pudo guardar la configuración SMTP',
		'Die SMTP-Konfiguration konnte nicht gespeichert werden'
	],
	'smtp.removed': [
		'SMTP configuration removed',
		'Configuration SMTP supprimée',
		'Configuración SMTP eliminada',
		'SMTP-Konfiguration entfernt'
	],
	'smtp.removeFailed': [
		'Failed to delete SMTP configuration',
		'Impossible de supprimer la configuration SMTP',
		'No se pudo eliminar la configuración SMTP',
		'Die SMTP-Konfiguration konnte nicht gelöscht werden'
	],
	'smtp.noAddress': [
		'No email address found for current user',
		'Aucune adresse email trouvée pour l’utilisateur actuel',
		'No se encontró ninguna dirección de correo para el usuario actual',
		'Für den aktuellen Benutzer wurde keine E-Mail-Adresse gefunden'
	],
	'smtp.testSent': [
		'Test email sent to {email}',
		'Email de test envoyé à {email}',
		'Correo de prueba enviado a {email}',
		'Test-E-Mail an {email} gesendet'
	],
	'smtp.testFailed': [
		'Failed to send test email',
		'Impossible d’envoyer l’email de test',
		'No se pudo enviar el correo de prueba',
		'Die Test-E-Mail konnte nicht gesendet werden'
	],
	'smtp.confirm.title': [
		'Delete the SMTP configuration?',
		'Supprimer la configuration SMTP ?',
		'¿Eliminar la configuración SMTP?',
		'SMTP-Konfiguration löschen?'
	],
	'smtp.confirm.text': [
		'Plume stops sending signing invitations, reminders and completion notices until a relay is configured again. The stored password is destroyed and cannot be recovered.',
		'Plume cesse d’envoyer les invitations à signer, les rappels et les avis de finalisation tant qu’un relais n’est pas reconfiguré. Le mot de passe enregistré est détruit et ne peut pas être récupéré.',
		'Plume deja de enviar invitaciones a firmar, recordatorios y avisos de finalización hasta que se configure un relé de nuevo. La contraseña guardada se destruye y no se puede recuperar.',
		'Plume sendet keine Signatureinladungen, Erinnerungen und Abschlussmitteilungen mehr, bis wieder ein Relay konfiguriert ist. Das gespeicherte Passwort wird vernichtet und lässt sich nicht wiederherstellen.'
	],
	'webhooks.title': ['Webhooks', 'Webhooks', 'Webhooks', 'Webhooks'],
	'webhooks.text': [
		'Every document and signer event is POSTed to each enabled endpoint.',
		'Chaque événement de document et de signataire est envoyé en POST à chaque point d’accès activé.',
		'Cada evento de documento y de firmante se envía por POST a cada endpoint activado.',
		'Jedes Dokument- und Unterzeichner-Ereignis wird per POST an jeden aktivierten Endpunkt gesendet.'
	],
	'webhooks.add': ['Add webhook', 'Ajouter un webhook', 'Añadir webhook', 'Webhook hinzufügen'],
	'webhooks.emptyTitle': ['No endpoints yet', 'Aucun point d’accès pour l’instant', 'Aún no hay endpoints', 'Noch keine Endpunkte'],
	'webhooks.emptyText': [
		'Add one and Plume starts delivering signed events to it immediately.',
		'Ajoutez-en un et Plume commence aussitôt à lui livrer les événements.',
		'Añade uno y Plume empieza a entregarle eventos de inmediato.',
		'Füge einen hinzu, dann beginnt Plume sofort, ihm Ereignisse zuzustellen.'
	],
	'webhooks.lastDelivered': [
		'Last delivered {date}',
		'Dernière livraison : {date}',
		'Última entrega: {date}',
		'Zuletzt zugestellt: {date}'
	],
	'webhooks.neverDelivered': ['Never delivered', 'Jamais livré', 'Nunca entregado', 'Nie zugestellt'],
	'webhooks.active': ['Active', 'Actif', 'Activo', 'Aktiv'],
	'webhooks.disabled': ['Disabled', 'Désactivé', 'Desactivado', 'Deaktiviert'],
	'webhooks.disable': ['Disable {url}', 'Désactiver {url}', 'Desactivar {url}', '{url} deaktivieren'],
	'webhooks.enable': ['Enable {url}', 'Activer {url}', 'Activar {url}', '{url} aktivieren'],
	'webhooks.test': ['Test', 'Tester', 'Probar', 'Testen'],
	'webhooks.saved': ['Webhook saved', 'Webhook enregistré', 'Webhook guardado', 'Webhook gespeichert'],
	'webhooks.saveFailed': [
		'Failed to save webhook',
		'Impossible d’enregistrer le webhook',
		'No se pudo guardar el webhook',
		'Der Webhook konnte nicht gespeichert werden'
	],
	'webhooks.updateFailed': [
		'Failed to update webhook',
		'Impossible de mettre à jour le webhook',
		'No se pudo actualizar el webhook',
		'Der Webhook konnte nicht aktualisiert werden'
	],
	'webhooks.deleted': ['Webhook deleted', 'Webhook supprimé', 'Webhook eliminado', 'Webhook gelöscht'],
	'webhooks.deleteFailed': [
		'Failed to delete webhook',
		'Impossible de supprimer le webhook',
		'No se pudo eliminar el webhook',
		'Der Webhook konnte nicht gelöscht werden'
	],
	'webhooks.delivered': [
		'Test event delivered',
		'Événement de test livré',
		'Evento de prueba entregado',
		'Testereignis zugestellt'
	],
	'webhooks.deliveryFailed': [
		'Test delivery failed',
		'Échec de la livraison de test',
		'Falló la entrega de prueba',
		'Testzustellung fehlgeschlagen'
	],
	'webhooks.contract': ['Event contract', 'Contrat d’événements', 'Contrato de eventos', 'Ereignisvertrag'],
	'webhooks.contractText': [
		'What a receiver has to be ready for. Failed deliveries retry up to three times with backoff.',
		'Ce à quoi un récepteur doit être prêt. Les livraisons échouées sont retentées jusqu’à trois fois avec un délai croissant.',
		'Para qué debe estar preparado un receptor. Las entregas fallidas se reintentan hasta tres veces con espera progresiva.',
		'Worauf ein Empfänger vorbereitet sein muss. Fehlgeschlagene Zustellungen werden bis zu dreimal mit wachsendem Abstand wiederholt.'
	],
	'webhooks.eventTypes': ['Event types', 'Types d’événements', 'Tipos de eventos', 'Ereignistypen'],
	'webhooks.eventTypesText': [
		'Every endpoint receives all of them — filtering happens on your side.',
		'Chaque point d’accès les reçoit tous : le filtrage se fait de votre côté.',
		'Cada endpoint los recibe todos: el filtrado se hace por tu parte.',
		'Jeder Endpunkt erhält alle — gefiltert wird auf deiner Seite.'
	],
	'webhooks.signatureHeader': ['Signature header', 'En-tête de signature', 'Cabecera de firma', 'Signatur-Header'],
	'webhooks.signatureHeaderText': [
		'Each delivery is signed HMAC-SHA256 with the endpoint’s secret.',
		'Chaque livraison est signée en HMAC-SHA256 avec le secret du point d’accès.',
		'Cada entrega se firma con HMAC-SHA256 usando el secreto del endpoint.',
		'Jede Zustellung wird mit dem Geheimnis des Endpunkts per HMAC-SHA256 signiert.'
	],
	'webhooks.drawer.edit': ['Edit webhook', 'Modifier le webhook', 'Editar webhook', 'Webhook bearbeiten'],
	'webhooks.drawer.text': [
		'Plume POSTs a JSON event to this URL and signs the body with the shared secret.',
		'Plume envoie un événement JSON en POST à cette URL et signe le corps avec le secret partagé.',
		'Plume envía un evento JSON por POST a esta URL y firma el cuerpo con el secreto compartido.',
		'Plume sendet ein JSON-Ereignis per POST an diese URL und signiert den Body mit dem gemeinsamen Geheimnis.'
	],
	'webhooks.url': ['URL', 'URL', 'URL', 'URL'],
	'webhooks.secret': ['Signing secret', 'Secret de signature', 'Secreto de firma', 'Signaturgeheimnis'],
	'webhooks.secretKeep': [
		'The stored secret is never sent back. Leave this empty to keep it, or type a new one to replace it.',
		'Le secret enregistré n’est jamais renvoyé. Laissez vide pour le conserver, ou saisissez-en un nouveau pour le remplacer.',
		'El secreto guardado nunca se devuelve. Déjalo vacío para conservarlo o escribe uno nuevo para sustituirlo.',
		'Das gespeicherte Geheimnis wird nie zurückgesendet. Leer lassen, um es zu behalten, oder ein neues eingeben, um es zu ersetzen.'
	],
	'webhooks.secretNew': [
		'Used to compute the HMAC-SHA256 signature. Store the same value on the receiving end.',
		'Sert à calculer la signature HMAC-SHA256. Conservez la même valeur côté récepteur.',
		'Sirve para calcular la firma HMAC-SHA256. Guarda el mismo valor en el receptor.',
		'Dient zur Berechnung der HMAC-SHA256-Signatur. Hinterlege denselben Wert auf der Empfangsseite.'
	],
	'webhooks.confirm.title': ['Delete this webhook?', 'Supprimer ce webhook ?', '¿Eliminar este webhook?', 'Diesen Webhook löschen?'],
	'webhooks.confirm.text': [
		'{url} stops receiving events immediately and its signing secret is destroyed. Anything downstream that relies on these deliveries goes quiet.',
		'{url} cesse immédiatement de recevoir des événements et son secret de signature est détruit. Tout ce qui, en aval, dépend de ces livraisons se tait.',
		'{url} deja de recibir eventos de inmediato y su secreto de firma se destruye. Todo lo que dependa de estas entregas se queda en silencio.',
		'{url} erhält sofort keine Ereignisse mehr und sein Signaturgeheimnis wird vernichtet. Alles, was von diesen Zustellungen abhängt, verstummt.'
	],
	'webhooks.confirm.label': ['Delete webhook', 'Supprimer le webhook', 'Eliminar webhook', 'Webhook löschen'],
	'advanced.title': ['Advanced — Plume', 'Avancé — Plume', 'Avanzado — Plume', 'Erweitert — Plume'],
	'advanced.instance': ['Instance', 'Instance', 'Instancia', 'Instanz'],
	'advanced.instanceText': [
		'What this browser is talking to.',
		'Ce avec quoi ce navigateur communique.',
		'Con qué se comunica este navegador.',
		'Womit dieser Browser kommuniziert.'
	],
	'advanced.instanceUrl': ['Instance URL', 'URL de l’instance', 'URL de la instancia', 'Instanz-URL'],
	'advanced.basePathText': [
		'Every request is scoped under it.',
		'Chaque requête passe sous ce chemin.',
		'Cada petición cuelga de ella.',
		'Jede Anfrage liegt darunter.'
	],
	'advanced.account': ['Account', 'Compte', 'Cuenta', 'Konto'],
	'advanced.accountText': [
		'Facts an operator may ask you for.',
		'Des informations qu’un opérateur peut vous demander.',
		'Datos que un operador puede pedirte.',
		'Angaben, nach denen ein Betreiber fragen könnte.'
	],
	'advanced.userId': ['User ID', 'Identifiant utilisateur', 'ID de usuario', 'Benutzer-ID'],
	'advanced.avatarSource': ['Avatar source', 'Source de l’avatar', 'Origen del avatar', 'Avatarquelle'],
	'advanced.avatarSourceText': [
		'Where your profile picture comes from.',
		'D’où vient votre photo de profil.',
		'De dónde viene tu foto de perfil.',
		'Woher dein Profilbild stammt.'
	],
	'advanced.avatarLocal': ['local', 'local', 'local', 'lokal'],
	'advanced.memberSinceText': [
		'When this account was created.',
		'Quand ce compte a été créé.',
		'Cuándo se creó esta cuenta.',
		'Wann dieses Konto erstellt wurde.'
	],
	'advanced.resync': [
		'Re-read your Facile profile',
		'Relire votre profil Facile',
		'Volver a leer tu perfil de Facile',
		'Facile-Profil neu einlesen'
	],
	'advanced.resyncText': [
		'Pulls your name and photo from your Facile account again instead of waiting for the next automatic sync.',
		'Récupère à nouveau votre nom et votre photo depuis votre compte Facile au lieu d’attendre la prochaine synchronisation automatique.',
		'Vuelve a traer tu nombre y tu foto desde tu cuenta de Facile en lugar de esperar a la próxima sincronización automática.',
		'Holt Namen und Foto erneut aus deinem Facile-Konto, statt auf die nächste automatische Synchronisierung zu warten.'
	],
	'advanced.syncNow': ['Sync now', 'Synchroniser', 'Sincronizar ahora', 'Jetzt synchronisieren'],
	'advanced.synced': [
		'Profile re-read from your Facile account',
		'Profil relu depuis votre compte Facile',
		'Perfil leído de nuevo desde tu cuenta de Facile',
		'Profil erneut aus deinem Facile-Konto eingelesen'
	],
	'advanced.syncFailed': [
		'Failed to sync profile',
		'Impossible de synchroniser le profil',
		'No se pudo sincronizar el perfil',
		'Das Profil konnte nicht synchronisiert werden'
	],
	'advanced.danger': ['Danger zone', 'Zone de danger', 'Zona de peligro', 'Gefahrenbereich'],
	'advanced.dangerText': [
		'Irreversible from this browser. Your account, documents and signatures are untouched.',
		'Irréversible depuis ce navigateur. Votre compte, vos documents et vos signatures ne sont pas touchés.',
		'Irreversible desde este navegador. Tu cuenta, tus documentos y tus firmas no se ven afectados.',
		'Von diesem Browser aus nicht umkehrbar. Dein Konto, deine Dokumente und Unterschriften bleiben unberührt.'
	],
	'advanced.clear': ['Clear local data', 'Effacer les données locales', 'Borrar datos locales', 'Lokale Daten löschen'],
	'advanced.clearText': [
		'Removes the stored session and the theme preference, then signs you out. To delete your Facile account and the data it holds in Plume, do it from your Facile account (Registre).',
		'Supprime la session enregistrée et la préférence de thème, puis vous déconnecte. Pour supprimer votre compte Facile et les données qu’il contient dans Plume, faites-le depuis votre compte Facile (Registre).',
		'Elimina la sesión guardada y la preferencia de tema, y cierra tu sesión. Para eliminar tu cuenta de Facile y los datos que tiene en Plume, hazlo desde tu cuenta de Facile (Registre).',
		'Entfernt die gespeicherte Sitzung und die Designauswahl und meldet dich ab. Um dein Facile-Konto und die zugehörigen Daten in Plume zu löschen, nutze dein Facile-Konto (Registre).'
	],
	'advanced.cleared': ['Local data cleared.', 'Données locales effacées.', 'Datos locales borrados.', 'Lokale Daten gelöscht.'],
	'advanced.confirm.title': [
		'Clear local data?',
		'Effacer les données locales ?',
		'¿Borrar los datos locales?',
		'Lokale Daten löschen?'
	],
	'advanced.confirm.text': [
		'Every Plume preference stored in this browser is deleted and you are signed out — there is no undo. Your account, documents and signatures stay on the server.',
		'Toutes les préférences Plume enregistrées dans ce navigateur sont supprimées et vous êtes déconnecté, sans retour possible. Votre compte, vos documents et vos signatures restent sur le serveur.',
		'Se borran todas las preferencias de Plume guardadas en este navegador y se cierra tu sesión; no se puede deshacer. Tu cuenta, tus documentos y tus firmas permanecen en el servidor.',
		'Alle in diesem Browser gespeicherten Plume-Einstellungen werden gelöscht und du wirst abgemeldet — das lässt sich nicht rückgängig machen. Dein Konto, deine Dokumente und Unterschriften bleiben auf dem Server.'
	],
	'advanced.confirm.label': [
		'Clear and sign out',
		'Effacer et se déconnecter',
		'Borrar y cerrar sesión',
		'Löschen und abmelden'
	]
} satisfies Record<string, Row>;
