import 'package:flutter/material.dart';

import '../tokens/tokens.dart';

/// Builds [ThemeData] for the La Suite numérique design system and exposes
/// the [LaColors] semantic token set as a [ThemeExtension].
///
/// Use [LaSuiteTheme.light] / [LaSuiteTheme.dark] directly as
/// `MaterialApp.theme` / `MaterialApp.darkTheme`, or merge the
/// [LaColors] extension into an existing [ThemeData] with
/// `theme.copyWith(extensions: [LaColors.light])`.
abstract final class LaSuiteTheme {
  static ThemeData light() => _build(LaColors.light, Brightness.light);

  static ThemeData dark() => _build(LaColors.dark, Brightness.dark);

  static ThemeData _build(LaColors colors, Brightness brightness) {
    final colorScheme = ColorScheme(
      brightness: brightness,
      primary: colors.backgroundBrandPrimary,
      onPrimary: colors.contentOnBrand,
      secondary: colors.backgroundBrandSecondary,
      onSecondary: colors.contentBrandPrimary,
      error: colors.backgroundErrorPrimary,
      onError: colors.contentOnError,
      surface: colors.surfacePrimary,
      onSurface: colors.contentNeutralPrimary,
    );

    final textTheme = TextTheme(
      displayLarge: LaTextStyles.h1.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      displayMedium: LaTextStyles.h2.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      displaySmall: LaTextStyles.h3.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      headlineLarge: LaTextStyles.h3.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      headlineMedium: LaTextStyles.h4.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      headlineSmall: LaTextStyles.h5.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      titleLarge: LaTextStyles.h6.copyWith(color: colors.contentNeutralPrimary),
      titleMedium: LaTextStyles.labelLg.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      titleSmall: LaTextStyles.labelMd.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      bodyLarge: LaTextStyles.bodyLg.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      bodyMedium: LaTextStyles.bodyMd.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      bodySmall: LaTextStyles.bodySm.copyWith(
        color: colors.contentNeutralSecondary,
      ),
      labelLarge: LaTextStyles.labelLg.copyWith(
        color: colors.contentNeutralPrimary,
      ),
      labelMedium: LaTextStyles.labelMd.copyWith(
        color: colors.contentNeutralSecondary,
      ),
      labelSmall: LaTextStyles.labelSm.copyWith(
        color: colors.contentNeutralTertiary,
      ),
    );

    return ThemeData(
      useMaterial3: true,
      brightness: brightness,
      colorScheme: colorScheme,
      scaffoldBackgroundColor: colors.surfaceTertiary,
      canvasColor: colors.surfacePrimary,
      dividerColor: colors.borderSurfacePrimary,
      fontFamily: LaFontFamily.base,
      fontFamilyFallback: LaFontFamily.fallback,
      textTheme: textTheme,
      splashFactory: NoSplash.splashFactory,
      highlightColor: Colors.transparent,
      extensions: [colors],
    );
  }
}

/// Convenience accessors for the La Suite numérique design tokens from a
/// [BuildContext].
extension LaSuiteThemeContext on BuildContext {
  /// The current [LaColors] semantic token set. Falls back to the light
  /// palette if no [LaSuiteTheme] was installed (e.g. in isolated tests).
  LaColors get laColors =>
      Theme.of(this).extension<LaColors>() ?? LaColors.light;
}
