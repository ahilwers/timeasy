## Flutter wrapper
-keep class io.flutter.app.** { *; }
-keep class io.flutter.plugin.**  { *; }
-keep class io.flutter.util.**  { *; }
-keep class io.flutter.view.**  { *; }
-keep class io.flutter.**  { *; }
-keep class io.flutter.plugins.**  { *; }
## Keep Google Play Core classes referenced by Flutter's deferred components
-keep class com.google.android.play.core.** { *; }
-dontwarn com.google.android.play.**
