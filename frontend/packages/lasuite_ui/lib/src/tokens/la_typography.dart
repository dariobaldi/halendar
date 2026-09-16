import 'package:flutter/widgets.dart';

/// Font size scale (design token `globals.font.sizes`), in logical pixels.
abstract final class LaFontSize {
  static const double t = 11; // 0.6875rem
  static const double xs = 12; // 0.75rem
  static const double sm = 14; // 0.875rem
  static const double ml = 15; // 0.938rem
  static const double md = 16; // 1rem
  static const double lg = 18; // 1.125rem
  static const double xl = 20; // 1.25rem

  static const double h6 = 18; // 1.125rem
  static const double h5 = 20; // 1.25rem
  static const double h4 = 22; // 1.375rem
  static const double h3 = 24; // 1.5rem
  static const double h2 = 28; // 1.75rem
  static const double h1 = 32; // 2rem

  static const double displaySm = 48; // xs-alt
  static const double displayMd = 56; // sm-alt
  static const double displayLg = 64; // md-alt
  static const double displayXl = 72; // lg-alt
  static const double displayXxl = 80; // xl-alt
}

/// Font weight scale (design token `globals.font.weights`).
abstract final class LaFontWeight {
  static const FontWeight thin = FontWeight.w100;
  static const FontWeight light = FontWeight.w300;
  static const FontWeight regular = FontWeight.w400;
  static const FontWeight medium = FontWeight.w500;
  static const FontWeight bold = FontWeight.w600;
  static const FontWeight extrabold = FontWeight.w800;
}

/// Font family tokens (design token `globals.font.families`).
///
/// The upstream kit ships "Hanken Grotesk" (with Inter and Roboto Flex
/// Variable as fallbacks). This package does not bundle any font files (to
/// avoid extra binary weight and font licensing decisions for consumers) —
/// [base] names the intended family so an app that has added the font via
/// its own `pubspec.yaml`/`google_fonts` picks it up automatically; otherwise
/// Flutter falls back to the platform default.
abstract final class LaFontFamily {
  static const String base = 'Hanken Grotesk';
  static const List<String> fallback = ['Inter', 'Roboto', 'sans-serif'];
}

/// Named text style presets built from the type-scale tokens above.
///
/// These carry size/weight/height/family only — no color — so they compose
/// cleanly with [LaColors] (e.g. `LaTextStyles.h1.copyWith(color: ...)`) and
/// with Flutter's [DefaultTextStyle]/[Theme] inheritance.
abstract final class LaTextStyles {
  static const TextStyle _base = TextStyle(
    fontFamily: LaFontFamily.base,
    fontFamilyFallback: LaFontFamily.fallback,
  );

  static final TextStyle h1 = _base.copyWith(
    fontSize: LaFontSize.h1,
    fontWeight: LaFontWeight.bold,
    height: 1.25,
  );
  static final TextStyle h2 = _base.copyWith(
    fontSize: LaFontSize.h2,
    fontWeight: LaFontWeight.bold,
    height: 1.28,
  );
  static final TextStyle h3 = _base.copyWith(
    fontSize: LaFontSize.h3,
    fontWeight: LaFontWeight.bold,
    height: 1.3,
  );
  static final TextStyle h4 = _base.copyWith(
    fontSize: LaFontSize.h4,
    fontWeight: LaFontWeight.bold,
    height: 1.3,
  );
  static final TextStyle h5 = _base.copyWith(
    fontSize: LaFontSize.h5,
    fontWeight: LaFontWeight.bold,
    height: 1.35,
  );
  static final TextStyle h6 = _base.copyWith(
    fontSize: LaFontSize.h6,
    fontWeight: LaFontWeight.bold,
    height: 1.35,
  );

  static final TextStyle bodyLg = _base.copyWith(
    fontSize: LaFontSize.lg,
    fontWeight: LaFontWeight.regular,
    height: 1.5,
  );
  static final TextStyle bodyMd = _base.copyWith(
    fontSize: LaFontSize.md,
    fontWeight: LaFontWeight.regular,
    height: 1.5,
  );
  static final TextStyle bodySm = _base.copyWith(
    fontSize: LaFontSize.sm,
    fontWeight: LaFontWeight.regular,
    height: 1.45,
  );
  static final TextStyle caption = _base.copyWith(
    fontSize: LaFontSize.xs,
    fontWeight: LaFontWeight.regular,
    height: 1.4,
  );

  static final TextStyle labelLg = _base.copyWith(
    fontSize: LaFontSize.md,
    fontWeight: LaFontWeight.medium,
    height: 1.4,
  );
  static final TextStyle labelMd = _base.copyWith(
    fontSize: LaFontSize.sm,
    fontWeight: LaFontWeight.medium,
    height: 1.35,
  );
  static final TextStyle labelSm = _base.copyWith(
    fontSize: LaFontSize.xs,
    fontWeight: LaFontWeight.medium,
    height: 1.3,
  );
}
