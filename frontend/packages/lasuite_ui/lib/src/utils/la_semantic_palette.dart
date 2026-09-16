import 'package:flutter/widgets.dart';

import '../tokens/la_colors.dart';

/// The semantic color categories shared by most components (button colors,
/// badge/alert types, field validation states, ...).
enum LaSemanticCategory { brand, neutral, info, success, warning, error }

/// A resolved bundle of the `background` / `content` / `border` tones for
/// one [LaSemanticCategory], read off a [LaColors] instance.
///
/// Centralizes the "pick the right generated LaColors field for this
/// category" switch so components (button, badge, alert, ...) don't each
/// reimplement it.
@immutable
class LaSemanticPalette {
  const LaSemanticPalette({
    required this.backgroundPrimary,
    required this.backgroundPrimaryHover,
    required this.backgroundSecondary,
    required this.backgroundSecondaryHover,
    required this.backgroundTertiary,
    required this.backgroundTertiaryHover,
    required this.contentPrimary,
    required this.contentSecondary,
    required this.contentTertiary,
    required this.contentOn,
    required this.borderPrimary,
    required this.borderSecondary,
    required this.borderTertiary,
  });

  factory LaSemanticPalette.of(LaColors c, LaSemanticCategory category) {
    switch (category) {
      case LaSemanticCategory.brand:
        return LaSemanticPalette(
          backgroundPrimary: c.backgroundBrandPrimary,
          backgroundPrimaryHover: c.backgroundBrandPrimaryHover,
          backgroundSecondary: c.backgroundBrandSecondary,
          backgroundSecondaryHover: c.backgroundBrandSecondaryHover,
          backgroundTertiary: c.backgroundBrandTertiary,
          backgroundTertiaryHover: c.backgroundBrandTertiaryHover,
          contentPrimary: c.contentBrandPrimary,
          contentSecondary: c.contentBrandSecondary,
          contentTertiary: c.contentBrandTertiary,
          contentOn: c.contentOnBrand,
          borderPrimary: c.borderBrandPrimary,
          borderSecondary: c.borderBrandSecondary,
          borderTertiary: c.borderBrandTertiary,
        );
      case LaSemanticCategory.neutral:
        return LaSemanticPalette(
          backgroundPrimary: c.backgroundNeutralPrimary,
          backgroundPrimaryHover: c.backgroundNeutralPrimaryHover,
          backgroundSecondary: c.backgroundNeutralSecondary,
          backgroundSecondaryHover: c.backgroundNeutralSecondaryHover,
          backgroundTertiary: c.backgroundNeutralTertiary,
          backgroundTertiaryHover: c.backgroundNeutralTertiaryHover,
          contentPrimary: c.contentNeutralPrimary,
          contentSecondary: c.contentNeutralSecondary,
          contentTertiary: c.contentNeutralTertiary,
          contentOn: c.contentOnNeutral,
          borderPrimary: c.borderNeutralPrimary,
          borderSecondary: c.borderNeutralSecondary,
          borderTertiary: c.borderNeutralTertiary,
        );
      case LaSemanticCategory.info:
        return LaSemanticPalette(
          backgroundPrimary: c.backgroundInfoPrimary,
          backgroundPrimaryHover: c.backgroundInfoPrimaryHover,
          backgroundSecondary: c.backgroundInfoSecondary,
          backgroundSecondaryHover: c.backgroundInfoSecondaryHover,
          backgroundTertiary: c.backgroundInfoTertiary,
          backgroundTertiaryHover: c.backgroundInfoTertiaryHover,
          contentPrimary: c.contentInfoPrimary,
          contentSecondary: c.contentInfoSecondary,
          contentTertiary: c.contentInfoTertiary,
          contentOn: c.contentOnInfo,
          borderPrimary: c.borderInfoPrimary,
          borderSecondary: c.borderInfoSecondary,
          borderTertiary: c.borderInfoTertiary,
        );
      case LaSemanticCategory.success:
        return LaSemanticPalette(
          backgroundPrimary: c.backgroundSuccessPrimary,
          backgroundPrimaryHover: c.backgroundSuccessPrimaryHover,
          backgroundSecondary: c.backgroundSuccessSecondary,
          backgroundSecondaryHover: c.backgroundSuccessSecondaryHover,
          backgroundTertiary: c.backgroundSuccessTertiary,
          backgroundTertiaryHover: c.backgroundSuccessTertiaryHover,
          contentPrimary: c.contentSuccessPrimary,
          contentSecondary: c.contentSuccessSecondary,
          contentTertiary: c.contentSuccessTertiary,
          contentOn: c.contentOnSuccess,
          borderPrimary: c.borderSuccessPrimary,
          borderSecondary: c.borderSuccessSecondary,
          borderTertiary: c.borderSuccessTertiary,
        );
      case LaSemanticCategory.warning:
        return LaSemanticPalette(
          backgroundPrimary: c.backgroundWarningPrimary,
          backgroundPrimaryHover: c.backgroundWarningPrimaryHover,
          backgroundSecondary: c.backgroundWarningSecondary,
          backgroundSecondaryHover: c.backgroundWarningSecondaryHover,
          backgroundTertiary: c.backgroundWarningTertiary,
          backgroundTertiaryHover: c.backgroundWarningTertiaryHover,
          contentPrimary: c.contentWarningPrimary,
          contentSecondary: c.contentWarningSecondary,
          contentTertiary: c.contentWarningTertiary,
          contentOn: c.contentOnWarning,
          borderPrimary: c.borderWarningPrimary,
          borderSecondary: c.borderWarningSecondary,
          borderTertiary: c.borderWarningTertiary,
        );
      case LaSemanticCategory.error:
        return LaSemanticPalette(
          backgroundPrimary: c.backgroundErrorPrimary,
          backgroundPrimaryHover: c.backgroundErrorPrimaryHover,
          backgroundSecondary: c.backgroundErrorSecondary,
          backgroundSecondaryHover: c.backgroundErrorSecondaryHover,
          backgroundTertiary: c.backgroundErrorTertiary,
          backgroundTertiaryHover: c.backgroundErrorTertiaryHover,
          contentPrimary: c.contentErrorPrimary,
          contentSecondary: c.contentErrorSecondary,
          contentTertiary: c.contentErrorTertiary,
          contentOn: c.contentOnError,
          borderPrimary: c.borderErrorPrimary,
          borderSecondary: c.borderErrorSecondary,
          borderTertiary: c.borderErrorTertiary,
        );
    }
  }

  final Color backgroundPrimary;
  final Color backgroundPrimaryHover;
  final Color backgroundSecondary;
  final Color backgroundSecondaryHover;
  final Color backgroundTertiary;
  final Color backgroundTertiaryHover;
  final Color contentPrimary;
  final Color contentSecondary;
  final Color contentTertiary;
  final Color contentOn;
  final Color borderPrimary;
  final Color borderSecondary;
  final Color borderTertiary;
}
