import 'package:flutter/material.dart';

/// App-wide light/dark mode override, toggled from a button visible on
/// every screen. A single instance is shared app-wide (there is no
/// per-user session state to scope it to) rather than threaded through
/// every screen's constructor.
class ThemeController extends ValueNotifier<ThemeMode> {
  ThemeController._() : super(ThemeMode.light);

  static final ThemeController instance = ThemeController._();

  void toggle() {
    value = value == ThemeMode.dark ? ThemeMode.light : ThemeMode.dark;
  }
}
