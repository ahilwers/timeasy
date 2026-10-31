import 'dart:io';

import 'package:in_app_review/in_app_review.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:timeasy/environment.dart';
import 'package:timeasy/utils/app_logger.dart';

/// Asks for a store rating once the app has actually been used for a while.
///
/// The prompt itself belongs to the operating system: [InAppReview.requestReview]
/// only asks iOS/Android to show it. The system decides whether anything
/// appears (quota, debug or TestFlight builds, disabled in the settings) and
/// never reports back, so nothing here can depend on the prompt being shown.
/// That is why the settings view additionally offers [openStoreListing] as a
/// guaranteed manual way to rate the app.
class ReviewService {
  static const _firstEntryKey = "reviewFirstEntryAt";
  static const _savedEntryCountKey = "reviewSavedEntryCount";
  static const _promptedVersionKey = "reviewPromptedVersion";

  /// The prompt is only requested after this many saved time entries ...
  static const int minSavedEntries = 10;

  /// ... and after this many days of use. Asking earlier would be wasted: the
  /// iOS prompt is limited to three appearances per year and is also counted
  /// when the user dismisses it.
  static const int minDaysOfUse = 7;

  /// Counts a newly saved time entry and asks for a review once the app has
  /// been used long enough. Only new entries belong here, not edits.
  Future<void> registerSavedTimeEntry() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final savedEntries = (prefs.getInt(_savedEntryCountKey) ?? 0) + 1;
      await prefs.setInt(_savedEntryCountKey, savedEntries);

      final firstEntryAt = prefs.getInt(_firstEntryKey);
      if (firstEntryAt == null) {
        await prefs.setInt(
            _firstEntryKey, DateTime.now().millisecondsSinceEpoch);
        return;
      }

      if (savedEntries < minSavedEntries) {
        return;
      }

      final daysOfUse = DateTime.now()
          .difference(DateTime.fromMillisecondsSinceEpoch(firstEntryAt))
          .inDays;
      if (daysOfUse < minDaysOfUse) {
        return;
      }

      final version = (await PackageInfo.fromPlatform()).version;
      if (prefs.getString(_promptedVersionKey) == version) {
        return;
      }
      // Remember the version before asking: the request has no result, so this
      // is the only way to avoid asking again in the same version.
      await prefs.setString(_promptedVersionKey, version);
      await _requestReview();
    } catch (e, stackTrace) {
      // Rating is a nice-to-have; never let it interfere with saving an entry.
      AppLogger.w("Could not check for a store review",
          error: e,
          stackTrace: stackTrace,
          method: "ReviewService.registerSavedTimeEntry");
    }
  }

  /// Opens the store page with the rating action. Unlike [InAppReview.requestReview]
  /// this always works, including when the system refuses to show its prompt.
  Future<void> openStoreListing() async {
    if (!_isSupported) {
      return;
    }
    try {
      await InAppReview.instance
          .openStoreListing(appStoreId: Environment.appStoreId);
    } catch (e, stackTrace) {
      AppLogger.w("Could not open the store listing",
          error: e,
          stackTrace: stackTrace,
          method: "ReviewService.openStoreListing");
    }
  }

  Future<void> _requestReview() async {
    if (!_isSupported) {
      return;
    }
    try {
      final inAppReview = InAppReview.instance;
      if (await inAppReview.isAvailable()) {
        await inAppReview.requestReview();
      }
    } catch (e, stackTrace) {
      AppLogger.w("Could not request a store review",
          error: e,
          stackTrace: stackTrace,
          method: "ReviewService._requestReview");
    }
  }

  /// Without the App Store or Play Store there is nothing to rate; this also
  /// keeps the plugin away from desktop and web targets.
  bool get _isSupported => Platform.isIOS || Platform.isAndroid;
}
