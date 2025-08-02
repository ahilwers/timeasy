import 'package:flex_color_scheme/flex_color_scheme.dart';
import 'package:flutter/material.dart';

const FlexSchemeData timeasyTheme = FlexSchemeData(
  name: 'Timeasy theme',
  description: 'Custom theme for the timeasy app.',
  light: FlexSchemeColor(
    appBarColor: Colors.white,
    primary: Colors.black,
    primaryContainer: Color(0xff481e74),
    secondary: Color(0xff6276f9),
    secondaryContainer: Color(0xff2d3881),
  ),
  dark: FlexSchemeColor(
    appBarColor: Colors.black,
    primary: Colors.white,
    primaryContainer: Color(0xff481e74),
    secondary: Color(0xff6276f9),
    secondaryContainer: Color(0xff2d3881),
  ),
);

// Create theme data with consistent AppBar styling
ThemeData getLightTheme() {
  final baseTheme = FlexColorScheme.light(colors: timeasyTheme.light).toTheme;
  return baseTheme.copyWith(
    appBarTheme: AppBarTheme(
      backgroundColor: Colors.white,
      foregroundColor: Colors.black,
      iconTheme: IconThemeData(color: Colors.black),
      actionsIconTheme: IconThemeData(color: Colors.black),
      titleTextStyle: TextStyle(
        color: Colors.black,
        fontSize: 20,
        fontWeight: FontWeight.bold,
      ),
      elevation: 0,
    ),
  );
}

ThemeData getDarkTheme() {
  final baseTheme = FlexColorScheme.dark(colors: timeasyTheme.dark).toTheme;
  return baseTheme.copyWith(
    appBarTheme: AppBarTheme(
      backgroundColor: Colors.black,
      foregroundColor: Colors.white,
      iconTheme: IconThemeData(color: Colors.white),
      actionsIconTheme: IconThemeData(color: Colors.white),
      titleTextStyle: TextStyle(
        color: Colors.white,
        fontSize: 20,
        fontWeight: FontWeight.bold,
      ),
      elevation: 0,
    ),
  );
}
