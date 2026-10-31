// ignore: unused_import
import 'package:intl/intl.dart' as intl;

import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get weeklyOverview => 'Weekly Overview';

  @override
  String get timeEntries => 'Time Entries';

  @override
  String get projects => 'Projects';

  @override
  String get loadingWeek => 'Loading week';

  @override
  String weekTitle(int weekNumber) {
    return 'Week $weekNumber';
  }

  @override
  String get weeklyHourSum => 'Sum:';

  @override
  String get monday => 'Monday';

  @override
  String get tuesday => 'Tuesday';

  @override
  String get wednesday => 'Wednesday';

  @override
  String get thursday => 'Thursday';

  @override
  String get friday => 'Friday';

  @override
  String get saturday => 'Saturday';

  @override
  String get sunday => 'Sunday';

  @override
  String get times => 'Times';

  @override
  String get loadingTimes => 'Loading times...';

  @override
  String get start => 'Start';

  @override
  String get end => 'End';

  @override
  String get hours => 'Hours';

  @override
  String get tooltipTimeStart => 'The start time';

  @override
  String get tooltipTimeEnd => 'The end time';

  @override
  String get tooltipHours => 'Amount of hours';

  @override
  String get loadingTimeEntry => 'Loading time entry...';

  @override
  String get errorMissingStartTime => 'Please specify a start time.';

  @override
  String get errorMissingEndTime => 'Please specify an end time.';

  @override
  String get errorEndtimeNotAfterStartTime =>
      'The start time must be before the end time.';

  @override
  String get save => 'Save';

  @override
  String get delete => 'Delete';

  @override
  String get endDate => 'End date';

  @override
  String get endTime => 'End time';

  @override
  String get addTimeEntry => 'Add Time';

  @override
  String get editTimeEntry => 'Edit Time';

  @override
  String get yes => 'Yes';

  @override
  String get no => 'No';

  @override
  String get deleteTimeEntryRequest =>
      'Do you really want to delete this time entry?';

  @override
  String get loadingProjects => 'Loading projects...';

  @override
  String get loadingProject => 'Loading project...';

  @override
  String get projectName => 'Project name';

  @override
  String get addProject => 'Add Project';

  @override
  String get editProject => 'Edit Project';

  @override
  String get errorMissingProjectName =>
      'Please specify a name for your project.';

  @override
  String get deleteProjectRequest =>
      'Do your really want to delete the project?';

  @override
  String get info => 'Info';

  @override
  String get export => 'Export';

  @override
  String get timesExported => 'Times exported successfully';

  @override
  String get date => 'Date';

  @override
  String get pause => 'Break';

  @override
  String get projectColor => 'Project color';

  @override
  String get colorBlue => 'Blue';

  @override
  String get colorGreen => 'Green';

  @override
  String get colorRed => 'Red';

  @override
  String get colorOrange => 'Orange';

  @override
  String get colorPurple => 'Purple';

  @override
  String get colorCyan => 'Cyan';

  @override
  String get colorYellow => 'Yellow';

  @override
  String get colorPink => 'Pink';

  @override
  String get colorGray => 'Gray';

  @override
  String get colorBrown => 'Brown';

  @override
  String get noProjectsFound => 'No projects found';

  @override
  String get createNewProject => 'Create New Project';

  @override
  String get addNewProject => 'Create a new project';

  @override
  String get whatAreYouDoingNow => 'What are you doing right now?';

  @override
  String get navHome => 'Home';

  @override
  String get navWeek => 'Week';

  @override
  String get navTime => 'Time';

  @override
  String get navProjects => 'Projects';

  @override
  String get appTitle => 'timeasy';

  @override
  String get entryDescription => 'Description';

  @override
  String get whatDidYouDo => 'What did you do?';

  @override
  String get deleteTimeEntry => 'Delete Time Entry';

  @override
  String get deleteTimeEntryConfirmation =>
      'Are you sure you want to delete this time entry?';

  @override
  String get cancel => 'Cancel';

  @override
  String get timeEntryDeleted => 'Time entry deleted';

  @override
  String get undo => 'Undo';

  @override
  String get noTimeEntries => 'No time entries found for this period';

  @override
  String get deleteProjectConfirmation =>
      'Are you sure you want to delete this project?';

  @override
  String get projectDeleted => 'Project deleted';

  @override
  String get loginHint =>
      'You can use timeasy without an account. Sign in to sync your data across devices.';

  @override
  String get loginNow => 'Sign in';

  @override
  String get freeUseHeading => 'Use timeasy for free – no sign-up needed.';

  @override
  String get freeUseDescription =>
      'You can start tracking your time right away, completely free of charge.';

  @override
  String get subscriptionBenefits =>
      'Want to get even more out of it? A subscription gives you automatic sync, powerful insights, exports, and many other features to make your day easier.';

  @override
  String get subscriptionRequiredTitle => 'Subscription required';

  @override
  String get subscriptionRequiredMessage =>
      'Syncing across devices is part of timeasy Pro. Your data stays on this device — subscribe to timeasy Pro to enable sync.';

  @override
  String subscriptionPriceMonthly(String price) {
    return '$price / month';
  }

  @override
  String subscriptionPriceYearly(String price) {
    return '$price / year';
  }

  @override
  String get subscriptionRestore => 'Restore purchases';

  @override
  String get subscriptionNoOfferings =>
      'Subscriptions are currently unavailable. Please try again later.';

  @override
  String get subscriptionRetry => 'Try again';

  @override
  String get subscriptionProcessing => 'Processing purchase…';

  @override
  String get subscriptionProcessingHint =>
      'This can take a few seconds. Please keep the app open.';

  @override
  String get subscriptionPurchaseTimeout =>
      'The purchase is taking longer than expected. Please check again in a moment with \"Restore purchases\".';

  @override
  String get subscriptionPurchaseFailed =>
      'The purchase could not be completed. Please try again.';

  @override
  String get subscriptionPurchasePending =>
      'Your purchase is pending. timeasy Pro will be activated as soon as the payment is completed.';

  @override
  String get subscriptionRestoreFailed =>
      'No active subscription was found for your account.';

  @override
  String get subscriptionAutoRenewNote =>
      'The subscription renews automatically until you cancel it in the store.';

  @override
  String get subscriptionRequiredLater => 'Maybe later';

  @override
  String get manageSubscription => 'Manage subscription';

  @override
  String get noActiveSubscription => 'No active subscription found.';

  @override
  String get subscriptionManagedOnWeb =>
      'Your subscription was purchased on the timeasy website and is managed there.';

  @override
  String get subscriptionTerms => 'Terms of Use';

  @override
  String get subscriptionPrivacy => 'Privacy Policy';

  @override
  String get signInPromo => 'Sign in and unlock the full power of timeasy!';

  @override
  String get signIn => 'Sign in';

  @override
  String get register => 'Create account';

  @override
  String get settings => 'Settings';

  @override
  String get accountSection => 'Account';

  @override
  String get subscriptionSection => 'Subscription';

  @override
  String get syncSection => 'Synchronization';

  @override
  String greeting(String name) {
    return 'Hello, $name!';
  }

  @override
  String get logout => 'Log out';

  @override
  String get internetConnected => 'Online';

  @override
  String get internetDisconnected => 'Offline';

  @override
  String lastSyncLabel(String date) {
    return 'Last sync: $date';
  }

  @override
  String get lastSyncNever => 'Never';

  @override
  String get lastSyncLoading => 'Loading…';

  @override
  String get lastSyncError => 'Could not load the last sync time';

  @override
  String syncError(String message) {
    return 'Sync failed: $message';
  }

  @override
  String get syncNow => 'Sync now';

  @override
  String get syncing => 'Syncing…';

  @override
  String get aboutSection => 'About timeasy';

  @override
  String get rateApp => 'Rate timeasy';

  @override
  String get rateAppDescription => 'Your rating helps others discover timeasy.';

  @override
  String get subscriptionInactive => 'No active subscription';

  @override
  String get subscriptionActive => 'timeasy Pro is active.';

  @override
  String get subscriptionActiveStore =>
      'timeasy Pro is active. The subscription is managed by the store you purchased it from.';

  @override
  String get projectActive => 'Active';

  @override
  String get pending => 'pending';

  @override
  String get resolvingIssue => 'Resolving issue...';

  @override
  String get issueResolved => 'Issue resolved';
}
