// ignore: unused_import
import 'package:intl/intl.dart' as intl;

import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for German (`de`).
class AppLocalizationsDe extends AppLocalizations {
  AppLocalizationsDe([String locale = 'de']) : super(locale);

  @override
  String get weeklyOverview => 'Wochenübersicht';

  @override
  String get timeEntries => 'Zeiteinträge';

  @override
  String get projects => 'Projekte';

  @override
  String get loadingWeek => 'Lade Woche';

  @override
  String weekTitle(int weekNumber) {
    return '$weekNumber. Kalenderwoche';
  }

  @override
  String get weeklyHourSum => 'Summe:';

  @override
  String get monday => 'Montag';

  @override
  String get tuesday => 'Dienstag';

  @override
  String get wednesday => 'Mittwoch';

  @override
  String get thursday => 'Donnerstag';

  @override
  String get friday => 'Freitag';

  @override
  String get saturday => 'Samstag';

  @override
  String get sunday => 'Sonntag';

  @override
  String get times => 'Zeiten';

  @override
  String get loadingTimes => 'Lade Zeiten...';

  @override
  String get start => 'Start';

  @override
  String get end => 'Ende';

  @override
  String get hours => 'Stunden';

  @override
  String get tooltipTimeStart => 'Startzeit';

  @override
  String get tooltipTimeEnd => 'Endzeit';

  @override
  String get tooltipHours => 'Anzahl Stunden';

  @override
  String get loadingTimeEntry => 'Lade Zeiteintrag...';

  @override
  String get errorMissingStartTime => 'Bitte gib eine Startzeit an.';

  @override
  String get errorMissingEndTime => 'Bitte gib eine Endzeit an.';

  @override
  String get errorEndtimeNotAfterStartTime =>
      'Die Startzeit muss vor der Endzeit liegen.';

  @override
  String get save => 'Speichern';

  @override
  String get delete => 'Löschen';

  @override
  String get endDate => 'Enddatum';

  @override
  String get endTime => 'Endzeit';

  @override
  String get addTimeEntry => 'Zeit hinzufügen';

  @override
  String get editTimeEntry => 'Zeit bearbeiten';

  @override
  String get yes => 'Ja';

  @override
  String get no => 'Nein';

  @override
  String get deleteTimeEntryRequest =>
      'Möchtest du den Zeiteintrag wirklich löschen?';

  @override
  String get loadingProjects => 'Lade Projekte...';

  @override
  String get loadingProject => 'Lade Projekt...';

  @override
  String get projectName => 'Projektname';

  @override
  String get addProject => 'Projekt hinzufügen';

  @override
  String get editProject => 'Projekt bearbeiten';

  @override
  String get errorMissingProjectName =>
      'Bitte gib einen Namen für das Projekt an.';

  @override
  String get deleteProjectRequest =>
      'Möchtest du das Projekt wirklich löschen?';

  @override
  String get info => 'Info';

  @override
  String get export => 'Exportieren';

  @override
  String get timesExported => 'Zeiten erfolgreich exportiert';

  @override
  String get date => 'Datum';

  @override
  String get pause => 'Pause';

  @override
  String get projectColor => 'Projektfarbe';

  @override
  String get colorBlue => 'Blau';

  @override
  String get colorGreen => 'Grün';

  @override
  String get colorRed => 'Rot';

  @override
  String get colorOrange => 'Orange';

  @override
  String get colorPurple => 'Lila';

  @override
  String get colorCyan => 'Türkis';

  @override
  String get colorYellow => 'Gelb';

  @override
  String get colorPink => 'Rosa';

  @override
  String get colorGray => 'Grau';

  @override
  String get colorBrown => 'Braun';

  @override
  String get noProjectsFound => 'Keine Projekte gefunden';

  @override
  String get createNewProject => 'Neues Projekt erstellen';

  @override
  String get addNewProject => 'Erstelle ein neues Projekt';

  @override
  String get whatAreYouDoingNow => 'Was machst du gerade?';

  @override
  String get navHome => 'Start';

  @override
  String get navWeek => 'Woche';

  @override
  String get navTime => 'Zeit';

  @override
  String get navProjects => 'Projekte';

  @override
  String get appTitle => 'timeasy';

  @override
  String get entryDescription => 'Beschreibung';

  @override
  String get whatDidYouDo => 'Was hast du gemacht?';

  @override
  String get deleteTimeEntry => 'Zeiteintrag löschen';

  @override
  String get deleteTimeEntryConfirmation =>
      'Bist du sicher, dass du diesen Zeiteintrag löschen möchten?';

  @override
  String get cancel => 'Abbrechen';

  @override
  String get timeEntryDeleted => 'Zeiteintrag gelöscht';

  @override
  String get undo => 'Rückgängig';

  @override
  String get noTimeEntries => 'Keine Zeiteinträge für diesen Zeitraum gefunden';

  @override
  String get deleteProjectConfirmation =>
      'Bist du sicher, dass du dieses Projekt löschen möchtest?';

  @override
  String get projectDeleted => 'Projekt gelöscht';

  @override
  String get loginHint =>
      'Du kannst timeasy auch ohne Konto nutzen. Melde dich an, um deine Daten geräteübergreifend zu synchronisieren.';

  @override
  String get loginNow => 'Anmelden';

  @override
  String get freeUseHeading => 'Nutze timeasy kostenlos – ganz ohne Anmeldung.';

  @override
  String get freeUseDescription =>
      'Du kannst die App sofort verwenden und deine Zeit einfach erfassen.';

  @override
  String get subscriptionBenefits =>
      'Wenn du noch mehr herausholen willst: Mit einem Abo bekommst du automatische Synchronisation, Auswertungen, Exporte und viele weitere Funktionen, die dir den Alltag noch leichter machen.';

  @override
  String get subscriptionRequiredTitle => 'Abo erforderlich';

  @override
  String get subscriptionRequiredMessage =>
      'Die Synchronisation über Geräte hinweg ist Teil von timeasy Pro. Deine Daten bleiben auf diesem Gerät — abonniere timeasy Pro, um die Synchronisation zu aktivieren.';

  @override
  String subscriptionPriceMonthly(String price) {
    return '$price / Monat';
  }

  @override
  String subscriptionPriceYearly(String price) {
    return '$price / Jahr';
  }

  @override
  String get subscriptionRestore => 'Käufe wiederherstellen';

  @override
  String get subscriptionNoOfferings =>
      'Abonnements sind derzeit nicht verfügbar. Bitte versuche es später erneut.';

  @override
  String get subscriptionRetry => 'Erneut versuchen';

  @override
  String get subscriptionProcessing => 'Kauf wird verarbeitet…';

  @override
  String get subscriptionProcessingHint =>
      'Das kann ein paar Sekunden dauern. Bitte lass die App geöffnet.';

  @override
  String get subscriptionPurchaseTimeout =>
      'Der Kauf dauert länger als erwartet. Prüfe ihn gleich über „Käufe wiederherstellen“.';

  @override
  String get subscriptionPurchaseFailed =>
      'Der Kauf konnte nicht abgeschlossen werden. Bitte versuche es erneut.';

  @override
  String get subscriptionPurchasePending =>
      'Dein Kauf ist in Bearbeitung. timeasy Pro wird aktiviert, sobald die Zahlung abgeschlossen ist.';

  @override
  String get subscriptionRestoreFailed =>
      'Für dein Konto wurde kein aktives Abo gefunden.';

  @override
  String get subscriptionAutoRenewNote =>
      'Das Abo verlängert sich automatisch, bis du es im Store kündigst.';

  @override
  String get subscriptionRequiredLater => 'Vielleicht später';

  @override
  String get manageSubscription => 'Abo verwalten';

  @override
  String get noActiveSubscription => 'Kein aktives Abo gefunden.';

  @override
  String get subscriptionManagedOnWeb =>
      'Dein Abo wurde über die timeasy-Website abgeschlossen und wird dort verwaltet.';

  @override
  String get subscriptionTerms => 'Nutzungsbedingungen';

  @override
  String get subscriptionPrivacy => 'Datenschutz';

  @override
  String get signInPromo =>
      'Melde dich an und entdecke das volle Potenzial von timeasy!';

  @override
  String get signIn => 'Anmelden';

  @override
  String get register => 'Konto erstellen';

  @override
  String get settings => 'Einstellungen';

  @override
  String get accountSection => 'Konto';

  @override
  String get subscriptionSection => 'Abo';

  @override
  String get syncSection => 'Synchronisierung';

  @override
  String greeting(String name) {
    return 'Hallo $name!';
  }

  @override
  String get logout => 'Abmelden';

  @override
  String get internetConnected => 'Online';

  @override
  String get internetDisconnected => 'Offline';

  @override
  String lastSyncLabel(String date) {
    return 'Letzte Synchronisierung: $date';
  }

  @override
  String get lastSyncNever => 'Noch nie';

  @override
  String get lastSyncLoading => 'Wird geladen…';

  @override
  String get lastSyncError =>
      'Letzte Synchronisierung konnte nicht geladen werden';

  @override
  String syncError(String message) {
    return 'Synchronisierung fehlgeschlagen: $message';
  }

  @override
  String get syncNow => 'Jetzt synchronisieren';

  @override
  String get syncing => 'Synchronisierung läuft…';

  @override
  String get aboutSection => 'Über timeasy';

  @override
  String get rateApp => 'timeasy bewerten';

  @override
  String get rateAppDescription =>
      'Deine Bewertung hilft anderen, timeasy zu finden.';

  @override
  String get subscriptionInactive => 'Kein aktives Abo';

  @override
  String get subscriptionActive => 'Dein timeasy-Pro-Abo ist aktiv.';

  @override
  String get subscriptionActiveStore =>
      'Dein timeasy-Pro-Abo ist aktiv und wird über den Store verwaltet, in dem du es abgeschlossen hast.';

  @override
  String get projectActive => 'Aktiv';

  @override
  String get pending => 'ausstehend';

  @override
  String get resolvingIssue => 'Issue wird aufgelöst...';

  @override
  String get issueResolved => 'Issue aufgelöst';
}
