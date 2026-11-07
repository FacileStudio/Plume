import type { Row } from '../row';

/** Strings of the space pages. */
export const spaces = {
	'spaces.title': ['Spaces — Plume', 'Espaces — Plume', 'Espacios — Plume', 'Bereiche — Plume'],
	'spaces.heading': ['Spaces', 'Espaces', 'Espacios', 'Bereiche'],
	'spaces.subtitle': [
		'Shared workspaces. Everyone in a space sees its documents and clients.',
		'Des espaces de travail partagés. Chaque membre d’un espace voit ses documents et ses clients.',
		'Espacios de trabajo compartidos. Todas las personas de un espacio ven sus documentos y clientes.',
		'Gemeinsame Arbeitsbereiche. Alle in einem Bereich sehen dessen Dokumente und Kunden.'
	],
	'spaces.subtitleRegistre': [
		'Spaces and their members are managed in Registre, for every Facile app at once.',
		'Les espaces et leurs membres sont gérés dans Registre, pour toutes les apps Facile à la fois.',
		'Los espacios y sus miembros se gestionan en Registre, para todas las apps de Facile a la vez.',
		'Bereiche und ihre Mitglieder werden in Registre verwaltet, für alle Facile-Apps zugleich.'
	],
	'spaces.new': ['New space', 'Nouvel espace', 'Nuevo espacio', 'Neuer Bereich'],
	'spaces.empty.title': ['No spaces yet', 'Aucun espace pour l’instant', 'Aún no hay espacios', 'Noch keine Bereiche'],
	'spaces.empty.registre': [
		'Create one in Registre and it shows up here within seconds.',
		'Créez-en un dans Registre, il apparaît ici en quelques secondes.',
		'Crea uno en Registre y aparecerá aquí en unos segundos.',
		'Erstelle einen in Registre, er erscheint hier innerhalb weniger Sekunden.'
	],
	'spaces.empty.local': [
		'Create one to collaborate with your team on the same documents.',
		'Créez-en un pour collaborer avec votre équipe sur les mêmes documents.',
		'Crea uno para colaborar con tu equipo en los mismos documentos.',
		'Erstelle einen, um mit deinem Team an denselben Dokumenten zu arbeiten.'
	],
	'spaces.col.space': ['Space', 'Espace', 'Espacio', 'Bereich'],
	'spaces.col.role': ['Your role', 'Votre rôle', 'Tu rol', 'Deine Rolle'],
	'spaces.leave': ['Leave', 'Quitter', 'Salir', 'Verlassen'],
	'spaces.leaveAria': ['Leave {name}', 'Quitter {name}', 'Salir de {name}', '{name} verlassen'],
	'spaces.left': ['You left the space', 'Vous avez quitté l’espace', 'Has salido del espacio', 'Du hast den Bereich verlassen'],
	'spaces.leaveFailed': [
		'Failed to leave space',
		'Impossible de quitter l’espace',
		'No se pudo salir del espacio',
		'Der Bereich konnte nicht verlassen werden'
	],
	'spaces.confirm.title': ['Leave {name}?', 'Quitter {name} ?', '¿Salir de {name}?', '{name} verlassen?'],
	'spaces.confirm.thisSpace': ['this space', 'cet espace', 'este espacio', 'diesen Bereich'],
	'spaces.confirm.text': [
		'You lose access to every document in this space. An owner can invite you back.',
		'Vous perdez l’accès à tous les documents de cet espace. Un propriétaire peut vous y réinviter.',
		'Pierdes el acceso a todos los documentos de este espacio. Un propietario puede volver a invitarte.',
		'Du verlierst den Zugriff auf alle Dokumente in diesem Bereich. Ein Eigentümer kann dich wieder einladen.'
	],
	'spaces.confirm.label': ['Leave space', 'Quitter l’espace', 'Salir del espacio', 'Bereich verlassen'],
	'spaces.confirm.cancel': ['Stay', 'Rester', 'Quedarme', 'Bleiben'],
	'spaceNew.title': ['New Space — Plume', 'Nouvel espace — Plume', 'Nuevo espacio — Plume', 'Neuer Bereich — Plume'],
	'spaceNew.back': ['Back to spaces', 'Retour aux espaces', 'Volver a los espacios', 'Zurück zu den Bereichen'],
	'spaceNew.subtitle': [
		'A space is a shared workspace. You start as its owner and can invite people afterwards.',
		'Un espace est un espace de travail partagé. Vous en êtes d’abord le propriétaire et pourrez inviter des personnes ensuite.',
		'Un espacio es un espacio de trabajo compartido. Empiezas como propietario y puedes invitar a otras personas después.',
		'Ein Bereich ist ein gemeinsamer Arbeitsbereich. Du beginnst als Eigentümer und kannst danach Leute einladen.'
	],
	'spaceNew.namePlaceholder': ['My team', 'Mon équipe', 'Mi equipo', 'Mein Team'],
	'spaceNew.descriptionHelp': [
		'Optional — what the space is for.',
		'Facultatif : à quoi sert l’espace.',
		'Opcional: para qué sirve el espacio.',
		'Optional: wofür der Bereich gedacht ist.'
	],
	'spaceNew.descriptionPlaceholder': [
		'Client contracts and NDAs.',
		'Contrats clients et NDA.',
		'Contratos de clientes y NDA.',
		'Kundenverträge und NDAs.'
	],
	'spaceNew.created': ['Space created', 'Espace créé', 'Espacio creado', 'Bereich erstellt'],
	'spaceNew.failed': [
		'Failed to create space',
		'Impossible de créer l’espace',
		'No se pudo crear el espacio',
		'Der Bereich konnte nicht erstellt werden'
	],
	'spaceNew.submit': ['Create space', 'Créer l’espace', 'Crear espacio', 'Bereich erstellen'],
	'space.titleFallback': ['Space', 'Espace', 'Espacio', 'Bereich'],
	'space.noDescription': [
		'No description yet.',
		'Pas encore de description.',
		'Aún no hay descripción.',
		'Noch keine Beschreibung.'
	],
	'space.membersCount': ['Members ({count})', 'Membres ({count})', 'Miembros ({count})', 'Mitglieder ({count})'],
	'space.settings': ['Settings', 'Paramètres', 'Ajustes', 'Einstellungen'],
	'space.viewAllMembers': [
		'View all members',
		'Voir tous les membres',
		'Ver todos los miembros',
		'Alle Mitglieder anzeigen'
	],
	'space.membersText': [
		'The people who can see everything in this space.',
		'Les personnes qui voient tout ce qui se trouve dans cet espace.',
		'Las personas que pueden ver todo lo que hay en este espacio.',
		'Die Personen, die alles in diesem Bereich sehen können.'
	],
	'space.noMembers': ['No one here yet', 'Personne pour l’instant', 'Aún no hay nadie', 'Noch niemand hier'],
	'space.noMembersText': [
		'Invite a teammate and they will see this space as soon as they accept.',
		'Invitez un collègue : il verra cet espace dès qu’il acceptera.',
		'Invita a un compañero y verá este espacio en cuanto acepte.',
		'Lade eine Kollegin oder einen Kollegen ein, dann sieht sie oder er diesen Bereich, sobald die Einladung angenommen ist.'
	],
	'members.title': [
		'Members — {space} — Plume',
		'Membres — {space} — Plume',
		'Miembros — {space} — Plume',
		'Mitglieder — {space} — Plume'
	],
	'members.count': [
		{ one: '{count} member in this space.', other: '{count} members in this space.' },
		{
			one: '{count} membre dans cet espace.',
			many: '{count} membres dans cet espace.',
			other: '{count} membres dans cet espace.'
		},
		{
			one: '{count} miembro en este espacio.',
			many: '{count} miembros en este espacio.',
			other: '{count} miembros en este espacio.'
		},
		{ one: '{count} Mitglied in diesem Bereich.', other: '{count} Mitglieder in diesem Bereich.' }
	],
	'members.managedInRegistre': [
		'Members and roles are managed in Registre, for every Facile app at once.',
		'Les membres et les rôles sont gérés dans Registre, pour toutes les apps Facile à la fois.',
		'Los miembros y los roles se gestionan en Registre, para todas las apps de Facile a la vez.',
		'Mitglieder und Rollen werden in Registre verwaltet, für alle Facile-Apps zugleich.'
	],
	'members.ownersStay': [
		'Owners cannot be removed.',
		'Les propriétaires ne peuvent pas être retirés.',
		'Los propietarios no se pueden quitar.',
		'Eigentümer können nicht entfernt werden.'
	],
	'members.addTitle': ['Add a member', 'Ajouter un membre', 'Añadir un miembro', 'Mitglied hinzufügen'],
	'members.addText': [
		'They get access to every document in this space straight away.',
		'Elle ou il accède tout de suite à tous les documents de cet espace.',
		'Tendrá acceso de inmediato a todos los documentos de este espacio.',
		'Die Person erhält sofort Zugriff auf alle Dokumente in diesem Bereich.'
	],
	'members.emailPlaceholder': [
		'colleague@example.com',
		'collegue@exemple.fr',
		'companero@ejemplo.es',
		'kollege@beispiel.de'
	],
	'members.role': ['Role', 'Rôle', 'Rol', 'Rolle'],
	'members.roleFor': ['Role for {name}', 'Rôle de {name}', 'Rol de {name}', 'Rolle von {name}'],
	'members.removeAria': ['Remove {name}', 'Retirer {name}', 'Quitar a {name}', '{name} entfernen'],
	'members.added': ['Member added', 'Membre ajouté', 'Miembro añadido', 'Mitglied hinzugefügt'],
	'members.addFailed': [
		'Failed to add member',
		'Impossible d’ajouter le membre',
		'No se pudo añadir el miembro',
		'Das Mitglied konnte nicht hinzugefügt werden'
	],
	'members.roleUpdated': ['Role updated', 'Rôle mis à jour', 'Rol actualizado', 'Rolle aktualisiert'],
	'members.roleFailed': [
		'Failed to update role',
		'Impossible de mettre à jour le rôle',
		'No se pudo actualizar el rol',
		'Die Rolle konnte nicht aktualisiert werden'
	],
	'members.removed': ['Member removed', 'Membre retiré', 'Miembro quitado', 'Mitglied entfernt'],
	'members.removeFailed': [
		'Failed to remove member',
		'Impossible de retirer le membre',
		'No se pudo quitar al miembro',
		'Das Mitglied konnte nicht entfernt werden'
	],
	'members.people': ['People', 'Personnes', 'Personas', 'Personen'],
	'members.peopleManage': [
		'Change a role or revoke access.',
		'Modifiez un rôle ou révoquez un accès.',
		'Cambia un rol o revoca un acceso.',
		'Ändere eine Rolle oder entziehe den Zugriff.'
	],
	'members.peopleView': [
		'Everyone with access to this space.',
		'Toutes les personnes qui ont accès à cet espace.',
		'Todas las personas con acceso a este espacio.',
		'Alle mit Zugriff auf diesen Bereich.'
	],
	'members.emptyManage': [
		'Add a teammate above and they will show up in this list.',
		'Ajoutez un collègue ci-dessus et il apparaîtra dans cette liste.',
		'Añade a un compañero arriba y aparecerá en esta lista.',
		'Füge oben eine Kollegin oder einen Kollegen hinzu, dann erscheint sie oder er in dieser Liste.'
	],
	'members.emptyView': [
		'People added in Registre show up here within seconds.',
		'Les personnes ajoutées dans Registre apparaissent ici en quelques secondes.',
		'Las personas añadidas en Registre aparecen aquí en unos segundos.',
		'In Registre hinzugefügte Personen erscheinen hier innerhalb weniger Sekunden.'
	],
	'members.thisMember': ['this member', 'ce membre', 'este miembro', 'dieses Mitglied'],
	'members.confirm.title': ['Remove {name}?', 'Retirer {name} ?', '¿Quitar a {name}?', '{name} entfernen?'],
	'members.confirm.text': [
		'They lose access to every document in this space immediately.',
		'Elle ou il perd immédiatement l’accès à tous les documents de cet espace.',
		'Pierde de inmediato el acceso a todos los documentos de este espacio.',
		'Die Person verliert sofort den Zugriff auf alle Dokumente in diesem Bereich.'
	],
	'members.confirm.label': ['Remove member', 'Retirer le membre', 'Quitar miembro', 'Mitglied entfernen'],
	'members.confirm.cancel': ['Keep access', 'Conserver l’accès', 'Mantener el acceso', 'Zugriff behalten'],
	'spaceSettings.title': [
		'Settings — {space} — Plume',
		'Paramètres — {space} — Plume',
		'Ajustes — {space} — Plume',
		'Einstellungen — {space} — Plume'
	],
	'spaceSettings.heading': ['Space settings', 'Paramètres de l’espace', 'Ajustes del espacio', 'Bereichseinstellungen'],
	'spaceSettings.general': ['General', 'Général', 'General', 'Allgemein'],
	'spaceSettings.generalText': [
		'How this space is named and described.',
		'Le nom et la description de cet espace.',
		'Cómo se nombra y describe este espacio.',
		'Name und Beschreibung dieses Bereichs.'
	],
	'spaceSettings.updated': ['Space updated', 'Espace mis à jour', 'Espacio actualizado', 'Bereich aktualisiert'],
	'spaceSettings.updateFailed': [
		'Failed to update space',
		'Impossible de mettre à jour l’espace',
		'No se pudo actualizar el espacio',
		'Der Bereich konnte nicht aktualisiert werden'
	],
	'spaceSettings.save': ['Save changes', 'Enregistrer les modifications', 'Guardar cambios', 'Änderungen speichern'],
	'spaceSettings.danger': ['Danger zone', 'Zone de danger', 'Zona de peligro', 'Gefahrenbereich'],
	'spaceSettings.dangerText': [
		'Irreversible, and nobody can undo it for you.',
		'Irréversible, et personne ne peut l’annuler pour vous.',
		'Irreversible, y nadie puede deshacerlo por ti.',
		'Nicht umkehrbar, und niemand kann es für dich rückgängig machen.'
	],
	'spaceSettings.deleteLabel': ['Delete this space', 'Supprimer cet espace', 'Eliminar este espacio', 'Diesen Bereich löschen'],
	'spaceSettings.deleteText': [
		'Every member loses access. Documents in the space go with it.',
		'Tous les membres perdent l’accès. Les documents de l’espace partent avec lui.',
		'Todos los miembros pierden el acceso. Los documentos del espacio se van con él.',
		'Alle Mitglieder verlieren den Zugriff. Die Dokumente im Bereich werden mit gelöscht.'
	],
	'spaceSettings.deleteButton': ['Delete space', 'Supprimer l’espace', 'Eliminar espacio', 'Bereich löschen'],
	'spaceSettings.deleted': ['Space deleted', 'Espace supprimé', 'Espacio eliminado', 'Bereich gelöscht'],
	'spaceSettings.deleteFailed': [
		'Failed to delete space',
		'Impossible de supprimer l’espace',
		'No se pudo eliminar el espacio',
		'Der Bereich konnte nicht gelöscht werden'
	],
	'spaceSettings.confirm.title': ['Delete {name}?', 'Supprimer {name} ?', '¿Eliminar {name}?', '{name} löschen?'],
	'spaceSettings.confirm.text': [
		'This cannot be undone. All members lose access immediately.',
		'Cette action est irréversible. Tous les membres perdent immédiatement l’accès.',
		'No se puede deshacer. Todos los miembros pierden el acceso de inmediato.',
		'Das lässt sich nicht rückgängig machen. Alle Mitglieder verlieren sofort den Zugriff.'
	],
	'spaceSettings.confirm.cancel': ['Keep space', 'Conserver l’espace', 'Conservar espacio', 'Bereich behalten']
} satisfies Record<string, Row>;
