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
