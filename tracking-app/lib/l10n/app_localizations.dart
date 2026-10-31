import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_de.dart';
import 'app_localizations_en.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('de'),
    Locale('en'),
  ];

  /// The menu item for accessing and the headline of the weekly overview.
  ///
  /// In en, this message translates to:
  /// **'Weekly Overview'**
  String get weeklyOverview;

  /// The menu item for accessing the time entries.
  ///
  /// In en, this message translates to:
  /// **'Time Entries'**
  String get timeEntries;

  /// The menu item for accessing the projects.
  ///
  /// In en, this message translates to:
  /// **'Projects'**
  String get projects;

  /// Displayed while loading statistics for a specific week.
  ///
  /// In en, this message translates to:
  /// **'Loading week'**
  String get loadingWeek;

  /// Displayed while loading statistics for a specific week.
  ///
  /// In en, this message translates to:
  /// **'Week {weekNumber}'**
  String weekTitle(int weekNumber);

  /// Sum that is displayed below the weekly statistics.
  ///
  /// In en, this message translates to:
  /// **'Sum:'**
  String get weeklyHourSum;

  /// No description provided for @monday.
  ///
  /// In en, this message translates to:
  /// **'Monday'**
  String get monday;

  /// No description provided for @tuesday.
  ///
  /// In en, this message translates to:
  /// **'Tuesday'**
  String get tuesday;

  /// No description provided for @wednesday.
  ///
  /// In en, this message translates to:
  /// **'Wednesday'**
  String get wednesday;

  /// No description provided for @thursday.
  ///
  /// In en, this message translates to:
  /// **'Thursday'**
  String get thursday;

  /// No description provided for @friday.
  ///
  /// In en, this message translates to:
  /// **'Friday'**
  String get friday;

  /// No description provided for @saturday.
  ///
  /// In en, this message translates to:
  /// **'Saturday'**
  String get saturday;

  /// No description provided for @sunday.
  ///
  /// In en, this message translates to:
  /// **'Sunday'**
  String get sunday;

  /// No description provided for @times.
  ///
  /// In en, this message translates to:
  /// **'Times'**
  String get times;

  /// Displayed while loading the time entries.
  ///
  /// In en, this message translates to:
  /// **'Loading times...'**
  String get loadingTimes;

  /// No description provided for @start.
  ///
  /// In en, this message translates to:
  /// **'Start'**
  String get start;

  /// No description provided for @end.
  ///
  /// In en, this message translates to:
  /// **'End'**
  String get end;

  /// No description provided for @hours.
  ///
  /// In en, this message translates to:
  /// **'Hours'**
  String get hours;

  /// No description provided for @tooltipTimeStart.
  ///
  /// In en, this message translates to:
  /// **'The start time'**
  String get tooltipTimeStart;

  /// No description provided for @tooltipTimeEnd.
  ///
  /// In en, this message translates to:
  /// **'The end time'**
  String get tooltipTimeEnd;

  /// No description provided for @tooltipHours.
  ///
  /// In en, this message translates to:
  /// **'Amount of hours'**
  String get tooltipHours;

  /// Displayed while loading one time entry from the database.
  ///
  /// In en, this message translates to:
  /// **'Loading time entry...'**
  String get loadingTimeEntry;

  /// No description provided for @errorMissingStartTime.
  ///
  /// In en, this message translates to:
  /// **'Please specify a start time.'**
  String get errorMissingStartTime;

  /// No description provided for @errorMissingEndTime.
  ///
  /// In en, this message translates to:
  /// **'Please specify an end time.'**
  String get errorMissingEndTime;

  /// No description provided for @errorEndtimeNotAfterStartTime.
  ///
  /// In en, this message translates to:
  /// **'The start time must be before the end time.'**
  String get errorEndtimeNotAfterStartTime;

  /// No description provided for @save.
  ///
  /// In en, this message translates to:
  /// **'Save'**
  String get save;

  /// No description provided for @delete.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get delete;

  /// No description provided for @endDate.
  ///
  /// In en, this message translates to:
  /// **'End date'**
  String get endDate;

  /// No description provided for @endTime.
  ///
  /// In en, this message translates to:
  /// **'End time'**
  String get endTime;

  /// No description provided for @addTimeEntry.
  ///
  /// In en, this message translates to:
  /// **'Add Time'**
  String get addTimeEntry;

  /// No description provided for @editTimeEntry.
  ///
  /// In en, this message translates to:
  /// **'Edit Time'**
  String get editTimeEntry;

  /// No description provided for @yes.
  ///
  /// In en, this message translates to:
  /// **'Yes'**
  String get yes;

  /// No description provided for @no.
  ///
  /// In en, this message translates to:
  /// **'No'**
  String get no;

  /// No description provided for @deleteTimeEntryRequest.
  ///
  /// In en, this message translates to:
  /// **'Do you really want to delete this time entry?'**
  String get deleteTimeEntryRequest;

  /// Displayed while loading the project list.
  ///
  /// In en, this message translates to:
  /// **'Loading projects...'**
  String get loadingProjects;

  /// Displayed while loading a single project.
  ///
  /// In en, this message translates to:
  /// **'Loading project...'**
  String get loadingProject;

  /// No description provided for @projectName.
  ///
  /// In en, this message translates to:
  /// **'Project name'**
  String get projectName;

  /// No description provided for @addProject.
  ///
  /// In en, this message translates to:
  /// **'Add Project'**
  String get addProject;

  /// No description provided for @editProject.
  ///
  /// In en, this message translates to:
  /// **'Edit Project'**
  String get editProject;

  /// No description provided for @errorMissingProjectName.
  ///
  /// In en, this message translates to:
  /// **'Please specify a name for your project.'**
  String get errorMissingProjectName;

  /// No description provided for @deleteProjectRequest.
  ///
  /// In en, this message translates to:
  /// **'Do your really want to delete the project?'**
  String get deleteProjectRequest;

  /// No description provided for @info.
  ///
  /// In en, this message translates to:
  /// **'Info'**
  String get info;

  /// No description provided for @export.
  ///
  /// In en, this message translates to:
  /// **'Export'**
  String get export;

  /// No description provided for @timesExported.
  ///
  /// In en, this message translates to:
  /// **'Times exported successfully'**
  String get timesExported;

  /// No description provided for @date.
  ///
  /// In en, this message translates to:
  /// **'Date'**
  String get date;

  /// No description provided for @pause.
  ///
  /// In en, this message translates to:
  /// **'Break'**
  String get pause;

  /// Label for the project color selection
  ///
  /// In en, this message translates to:
  /// **'Project color'**
  String get projectColor;

  /// No description provided for @colorBlue.
  ///
  /// In en, this message translates to:
  /// **'Blue'**
  String get colorBlue;

  /// No description provided for @colorGreen.
  ///
  /// In en, this message translates to:
  /// **'Green'**
  String get colorGreen;

  /// No description provided for @colorRed.
  ///
  /// In en, this message translates to:
  /// **'Red'**
  String get colorRed;

  /// No description provided for @colorOrange.
  ///
  /// In en, this message translates to:
  /// **'Orange'**
  String get colorOrange;

  /// No description provided for @colorPurple.
  ///
  /// In en, this message translates to:
  /// **'Purple'**
  String get colorPurple;

  /// No description provided for @colorCyan.
  ///
  /// In en, this message translates to:
  /// **'Cyan'**
  String get colorCyan;

  /// No description provided for @colorYellow.
  ///
  /// In en, this message translates to:
  /// **'Yellow'**
  String get colorYellow;

  /// No description provided for @colorPink.
  ///
  /// In en, this message translates to:
  /// **'Pink'**
  String get colorPink;

  /// No description provided for @colorGray.
  ///
  /// In en, this message translates to:
  /// **'Gray'**
  String get colorGray;

  /// No description provided for @colorBrown.
  ///
  /// In en, this message translates to:
  /// **'Brown'**
  String get colorBrown;

  /// Displayed when no projects are available
  ///
  /// In en, this message translates to:
  /// **'No projects found'**
  String get noProjectsFound;

  /// Button text to create a new project
  ///
  /// In en, this message translates to:
  /// **'Create New Project'**
  String get createNewProject;

  /// Text shown on the add new project page
  ///
  /// In en, this message translates to:
  /// **'Create a new project'**
  String get addNewProject;

  /// Hint text for the time entry description field
  ///
  /// In en, this message translates to:
  /// **'What are you doing right now?'**
  String get whatAreYouDoingNow;

  /// Navigation label for the home tab
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get navHome;

  /// Navigation label for the weekly overview tab
  ///
  /// In en, this message translates to:
  /// **'Week'**
  String get navWeek;

  /// Navigation label for the time entries tab
  ///
  /// In en, this message translates to:
  /// **'Time'**
  String get navTime;

  /// Navigation label for the projects tab
  ///
  /// In en, this message translates to:
  /// **'Projects'**
  String get navProjects;

  /// The title of the application
  ///
  /// In en, this message translates to:
  /// **'timeasy'**
  String get appTitle;

  /// Label for the time entry description field
  ///
  /// In en, this message translates to:
  /// **'Description'**
  String get entryDescription;

  /// Hint text for the time entry description field in edit view
  ///
  /// In en, this message translates to:
  /// **'What did you do?'**
  String get whatDidYouDo;

  /// Title for the delete time entry confirmation dialog
  ///
  /// In en, this message translates to:
  /// **'Delete Time Entry'**
  String get deleteTimeEntry;

  /// Confirmation message for deleting a time entry
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to delete this time entry?'**
  String get deleteTimeEntryConfirmation;

  /// Cancel button text
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// Message shown when a time entry is deleted
  ///
  /// In en, this message translates to:
  /// **'Time entry deleted'**
  String get timeEntryDeleted;

  /// Undo button text
  ///
  /// In en, this message translates to:
  /// **'Undo'**
  String get undo;

  /// Message shown when there are no time entries for a selected period
  ///
  /// In en, this message translates to:
  /// **'No time entries found for this period'**
  String get noTimeEntries;

  /// Confirmation message for deleting a project
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to delete this project?'**
  String get deleteProjectConfirmation;

  /// Message shown when a project is deleted
  ///
  /// In en, this message translates to:
  /// **'Project deleted'**
  String get projectDeleted;

  /// Hint shown on the home screen when the user is not logged in: local use is possible, an account enables sync
  ///
  /// In en, this message translates to:
  /// **'You can use timeasy without an account. Sign in to sync your data across devices.'**
  String get loginHint;

  /// Text for the login link
  ///
  /// In en, this message translates to:
  /// **'Sign in'**
  String get loginNow;

  /// Heading for the promotional text in settings view
  ///
  /// In en, this message translates to:
  /// **'Use timeasy for free – no sign-up needed.'**
  String get freeUseHeading;

  /// Description for free usage in settings view
  ///
  /// In en, this message translates to:
  /// **'You can start tracking your time right away, completely free of charge.'**
  String get freeUseDescription;

  /// Description of subscription benefits in settings view
  ///
  /// In en, this message translates to:
  /// **'Want to get even more out of it? A subscription gives you automatic sync, powerful insights, exports, and many other features to make your day easier.'**
  String get subscriptionBenefits;

  /// Title of the paywall dialog when sync requires a subscription
  ///
  /// In en, this message translates to:
  /// **'Subscription required'**
  String get subscriptionRequiredTitle;

  /// Message of the paywall dialog shown when the server responds with 402 during sync
  ///
  /// In en, this message translates to:
  /// **'Syncing across devices is part of timeasy Pro. Your data stays on this device — subscribe to timeasy Pro to enable sync.'**
  String get subscriptionRequiredMessage;

  /// Monthly subscription button in the paywall dialog including the localized price
  ///
  /// In en, this message translates to:
  /// **'{price} / month'**
  String subscriptionPriceMonthly(String price);

  /// Yearly subscription button in the paywall dialog including the localized price
  ///
  /// In en, this message translates to:
  /// **'{price} / year'**
  String subscriptionPriceYearly(String price);

  /// Button to restore previous in-app purchases (required by the App Store)
  ///
  /// In en, this message translates to:
  /// **'Restore purchases'**
  String get subscriptionRestore;

  /// Shown in the paywall dialog when no subscription offers can be loaded
  ///
  /// In en, this message translates to:
  /// **'Subscriptions are currently unavailable. Please try again later.'**
  String get subscriptionNoOfferings;

  /// Button in the paywall dialog that reloads the subscription offers
  ///
  /// In en, this message translates to:
  /// **'Try again'**
  String get subscriptionRetry;

  /// Shown in the paywall dialog while the store transaction is processed
  ///
  /// In en, this message translates to:
  /// **'Processing purchase…'**
  String get subscriptionProcessing;

  /// Hint below the spinner while a purchase is processed
  ///
  /// In en, this message translates to:
  /// **'This can take a few seconds. Please keep the app open.'**
  String get subscriptionProcessingHint;

  /// Shown in the paywall dialog when the purchase exceeds the timeout
  ///
  /// In en, this message translates to:
  /// **'The purchase is taking longer than expected. Please check again in a moment with \"Restore purchases\".'**
  String get subscriptionPurchaseTimeout;

  /// Error shown in the paywall dialog when a purchase fails
  ///
  /// In en, this message translates to:
  /// **'The purchase could not be completed. Please try again.'**
  String get subscriptionPurchaseFailed;

  /// Shown in the paywall dialog when a purchase is pending (e.g. cash payment methods on Google Play)
  ///
  /// In en, this message translates to:
  /// **'Your purchase is pending. timeasy Pro will be activated as soon as the payment is completed.'**
  String get subscriptionPurchasePending;

  /// Error shown when restoring purchases did not find an active subscription
  ///
  /// In en, this message translates to:
  /// **'No active subscription was found for your account.'**
  String get subscriptionRestoreFailed;

  /// Auto-renewal note in the paywall dialog (store compliance)
  ///
  /// In en, this message translates to:
  /// **'The subscription renews automatically until you cancel it in the store.'**
  String get subscriptionAutoRenewNote;

  /// Button to dismiss the paywall dialog
  ///
  /// In en, this message translates to:
  /// **'Maybe later'**
  String get subscriptionRequiredLater;

  /// Button in the settings view that opens the store's subscription management
  ///
  /// In en, this message translates to:
  /// **'Manage subscription'**
  String get manageSubscription;

  /// Shown when no subscription management URL is available
  ///
  /// In en, this message translates to:
  /// **'No active subscription found.'**
  String get noActiveSubscription;

  /// Neutral note in the settings view for subscriptions purchased on timeasy.org; deliberately without a link (store compliance)
  ///
  /// In en, this message translates to:
  /// **'Your subscription was purchased on the timeasy website and is managed there.'**
  String get subscriptionManagedOnWeb;

  /// Link to the terms of use in the paywall dialog
  ///
  /// In en, this message translates to:
  /// **'Terms of Use'**
  String get subscriptionTerms;

  /// Link to the privacy policy in the paywall dialog
  ///
  /// In en, this message translates to:
  /// **'Privacy Policy'**
  String get subscriptionPrivacy;

  /// Call to action for signing in
  ///
  /// In en, this message translates to:
  /// **'Sign in and unlock the full power of timeasy!'**
  String get signInPromo;

  /// Button that opens the login flow for existing users
  ///
  /// In en, this message translates to:
  /// **'Sign in'**
  String get signIn;

  /// Button that opens the Keycloak registration flow for new users; an account is required before an in-app purchase can be made
  ///
  /// In en, this message translates to:
  /// **'Create account'**
  String get register;

  /// Title for the settings screen
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settings;

  /// Title of the account card in the settings screen
  ///
  /// In en, this message translates to:
  /// **'Account'**
  String get accountSection;

  /// Title of the subscription card in the settings screen
  ///
  /// In en, this message translates to:
  /// **'Subscription'**
  String get subscriptionSection;

  /// Title of the synchronization card in the settings screen
  ///
  /// In en, this message translates to:
  /// **'Synchronization'**
  String get syncSection;

  /// Greeting of the signed-in user
  ///
  /// In en, this message translates to:
  /// **'Hello, {name}!'**
  String greeting(String name);

  /// Button that signs the user out
  ///
  /// In en, this message translates to:
  /// **'Log out'**
  String get logout;

  /// Shown when the device has an internet connection
  ///
  /// In en, this message translates to:
  /// **'Online'**
  String get internetConnected;

  /// Shown when the device has no internet connection
  ///
  /// In en, this message translates to:
  /// **'Offline'**
  String get internetDisconnected;

  /// Last successful synchronization including the formatted date
  ///
  /// In en, this message translates to:
  /// **'Last sync: {date}'**
  String lastSyncLabel(String date);

  /// Shown when the account has never been synchronized
  ///
  /// In en, this message translates to:
  /// **'Never'**
  String get lastSyncNever;

  /// Shown while the last sync time is loaded
  ///
  /// In en, this message translates to:
  /// **'Loading…'**
  String get lastSyncLoading;

  /// Shown when the last sync time cannot be read
  ///
  /// In en, this message translates to:
  /// **'Could not load the last sync time'**
  String get lastSyncError;

  /// Shown when the last synchronization failed
  ///
  /// In en, this message translates to:
  /// **'Sync failed: {message}'**
  String syncError(String message);

  /// Button that triggers a synchronization immediately
  ///
  /// In en, this message translates to:
  /// **'Sync now'**
  String get syncNow;

  /// Shown while a synchronization is running
  ///
  /// In en, this message translates to:
  /// **'Syncing…'**
  String get syncing;

  /// Settings section with app related actions
  ///
  /// In en, this message translates to:
  /// **'About timeasy'**
  String get aboutSection;

  /// Settings entry that opens the store page to rate the app
  ///
  /// In en, this message translates to:
  /// **'Rate timeasy'**
  String get rateApp;

  /// Explanation below the rate app entry
  ///
  /// In en, this message translates to:
  /// **'Your rating helps others discover timeasy.'**
  String get rateAppDescription;

  /// Subscription status when no subscription is active
  ///
  /// In en, this message translates to:
  /// **'No active subscription'**
  String get subscriptionInactive;

  /// Subscription status when the subscription is active but its purchase origin is unknown
  ///
  /// In en, this message translates to:
  /// **'timeasy Pro is active.'**
  String get subscriptionActive;

  /// Subscription status for subscriptions purchased through the App Store or Google Play
  ///
  /// In en, this message translates to:
  /// **'timeasy Pro is active. The subscription is managed by the store you purchased it from.'**
  String get subscriptionActiveStore;

  /// Label for the project active status toggle
  ///
  /// In en, this message translates to:
  /// **'Active'**
  String get projectActive;

  /// Status for pending issue references
  ///
  /// In en, this message translates to:
  /// **'pending'**
  String get pending;

  /// Status message while resolving an issue reference
  ///
  /// In en, this message translates to:
  /// **'Resolving issue...'**
  String get resolvingIssue;

  /// Status message when an issue is successfully resolved
  ///
  /// In en, this message translates to:
  /// **'Issue resolved'**
  String get issueResolved;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['de', 'en'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'de':
      return AppLocalizationsDe();
    case 'en':
      return AppLocalizationsEn();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
