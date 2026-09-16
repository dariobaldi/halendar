// GENERATED FILE. Derived from the La Suite numérique UI Kit
// (https://github.com/suitenumerique/ui-kit) design tokens (light/dark
// "contextuals": background / content / border semantic color maps).
// See ATTRIBUTION.md. Do not hand-edit; regenerate instead.

import 'package:flutter/material.dart';

/// Semantic design-token colors for the La Suite numérique theme.
///
/// Registered as a [ThemeExtension] so widgets read it via
/// `Theme.of(context).extension<LaColors>()!` (or the `context.laColors`
/// helper). Mirrors the upstream `contextuals.background` /
/// `contextuals.content` / `contextuals.border` token groups:
/// backgroundX / contentX / borderX, each with primary/secondary/tertiary
/// (+ hover variants for backgrounds) across the brand, neutral, info,
/// success, warning and error semantic categories.
@immutable
class LaColors extends ThemeExtension<LaColors> {
  const LaColors({
    required this.surfacePrimary,
    required this.surfaceSecondary,
    required this.surfaceTertiary,
    required this.overlayPrimary,
    required this.overlayPrimaryHover,
    required this.overlayOnSurface,
    required this.contextualPrimary,
    required this.contextualPrimaryHover,
    required this.backgroundBrandPrimary,
    required this.backgroundBrandPrimaryHover,
    required this.backgroundBrandSecondary,
    required this.backgroundBrandSecondaryHover,
    required this.backgroundBrandTertiary,
    required this.backgroundBrandTertiaryHover,
    required this.backgroundNeutralPrimary,
    required this.backgroundNeutralPrimaryHover,
    required this.backgroundNeutralSecondary,
    required this.backgroundNeutralSecondaryHover,
    required this.backgroundNeutralTertiary,
    required this.backgroundNeutralTertiaryHover,
    required this.backgroundInfoPrimary,
    required this.backgroundInfoPrimaryHover,
    required this.backgroundInfoSecondary,
    required this.backgroundInfoSecondaryHover,
    required this.backgroundInfoTertiary,
    required this.backgroundInfoTertiaryHover,
    required this.backgroundSuccessPrimary,
    required this.backgroundSuccessPrimaryHover,
    required this.backgroundSuccessSecondary,
    required this.backgroundSuccessSecondaryHover,
    required this.backgroundSuccessTertiary,
    required this.backgroundSuccessTertiaryHover,
    required this.backgroundWarningPrimary,
    required this.backgroundWarningPrimaryHover,
    required this.backgroundWarningSecondary,
    required this.backgroundWarningSecondaryHover,
    required this.backgroundWarningTertiary,
    required this.backgroundWarningTertiaryHover,
    required this.backgroundErrorPrimary,
    required this.backgroundErrorPrimaryHover,
    required this.backgroundErrorSecondary,
    required this.backgroundErrorSecondaryHover,
    required this.backgroundErrorTertiary,
    required this.backgroundErrorTertiaryHover,
    required this.backgroundDisabledPrimary,
    required this.backgroundDisabledSecondary,
    required this.contentContextualPrimary,
    required this.contentOverlayPrimary,
    required this.contentOverlaySecondary,
    required this.contentBrandPrimary,
    required this.contentBrandSecondary,
    required this.contentBrandTertiary,
    required this.contentOnBrand,
    required this.contentNeutralPrimary,
    required this.contentNeutralSecondary,
    required this.contentNeutralTertiary,
    required this.contentOnNeutral,
    required this.contentInfoPrimary,
    required this.contentInfoSecondary,
    required this.contentInfoTertiary,
    required this.contentOnInfo,
    required this.contentSuccessPrimary,
    required this.contentSuccessSecondary,
    required this.contentSuccessTertiary,
    required this.contentOnSuccess,
    required this.contentWarningPrimary,
    required this.contentWarningSecondary,
    required this.contentWarningTertiary,
    required this.contentOnWarning,
    required this.contentErrorPrimary,
    required this.contentErrorSecondary,
    required this.contentErrorTertiary,
    required this.contentOnError,
    required this.contentDisabledPrimary,
    required this.contentDisabledSecondary,
    required this.logoPrimary,
    required this.borderSurfacePrimary,
    required this.borderContextualPrimary,
    required this.borderOverlayPrimary,
    required this.borderBrandPrimary,
    required this.borderBrandSecondary,
    required this.borderBrandTertiary,
    required this.borderNeutralPrimary,
    required this.borderNeutralSecondary,
    required this.borderNeutralTertiary,
    required this.borderInfoPrimary,
    required this.borderInfoSecondary,
    required this.borderInfoTertiary,
    required this.borderSuccessPrimary,
    required this.borderSuccessSecondary,
    required this.borderSuccessTertiary,
    required this.borderWarningPrimary,
    required this.borderWarningSecondary,
    required this.borderWarningTertiary,
    required this.borderErrorPrimary,
    required this.borderErrorSecondary,
    required this.borderErrorTertiary,
    required this.borderDisabledPrimary,
  });

  final Color surfacePrimary;
  final Color surfaceSecondary;
  final Color surfaceTertiary;
  final Color overlayPrimary;
  final Color overlayPrimaryHover;
  final Color overlayOnSurface;
  final Color contextualPrimary;
  final Color contextualPrimaryHover;
  final Color backgroundBrandPrimary;
  final Color backgroundBrandPrimaryHover;
  final Color backgroundBrandSecondary;
  final Color backgroundBrandSecondaryHover;
  final Color backgroundBrandTertiary;
  final Color backgroundBrandTertiaryHover;
  final Color backgroundNeutralPrimary;
  final Color backgroundNeutralPrimaryHover;
  final Color backgroundNeutralSecondary;
  final Color backgroundNeutralSecondaryHover;
  final Color backgroundNeutralTertiary;
  final Color backgroundNeutralTertiaryHover;
  final Color backgroundInfoPrimary;
  final Color backgroundInfoPrimaryHover;
  final Color backgroundInfoSecondary;
  final Color backgroundInfoSecondaryHover;
  final Color backgroundInfoTertiary;
  final Color backgroundInfoTertiaryHover;
  final Color backgroundSuccessPrimary;
  final Color backgroundSuccessPrimaryHover;
  final Color backgroundSuccessSecondary;
  final Color backgroundSuccessSecondaryHover;
  final Color backgroundSuccessTertiary;
  final Color backgroundSuccessTertiaryHover;
  final Color backgroundWarningPrimary;
  final Color backgroundWarningPrimaryHover;
  final Color backgroundWarningSecondary;
  final Color backgroundWarningSecondaryHover;
  final Color backgroundWarningTertiary;
  final Color backgroundWarningTertiaryHover;
  final Color backgroundErrorPrimary;
  final Color backgroundErrorPrimaryHover;
  final Color backgroundErrorSecondary;
  final Color backgroundErrorSecondaryHover;
  final Color backgroundErrorTertiary;
  final Color backgroundErrorTertiaryHover;
  final Color backgroundDisabledPrimary;
  final Color backgroundDisabledSecondary;
  final Color contentContextualPrimary;
  final Color contentOverlayPrimary;
  final Color contentOverlaySecondary;
  final Color contentBrandPrimary;
  final Color contentBrandSecondary;
  final Color contentBrandTertiary;
  final Color contentOnBrand;
  final Color contentNeutralPrimary;
  final Color contentNeutralSecondary;
  final Color contentNeutralTertiary;
  final Color contentOnNeutral;
  final Color contentInfoPrimary;
  final Color contentInfoSecondary;
  final Color contentInfoTertiary;
  final Color contentOnInfo;
  final Color contentSuccessPrimary;
  final Color contentSuccessSecondary;
  final Color contentSuccessTertiary;
  final Color contentOnSuccess;
  final Color contentWarningPrimary;
  final Color contentWarningSecondary;
  final Color contentWarningTertiary;
  final Color contentOnWarning;
  final Color contentErrorPrimary;
  final Color contentErrorSecondary;
  final Color contentErrorTertiary;
  final Color contentOnError;
  final Color contentDisabledPrimary;
  final Color contentDisabledSecondary;
  final Color logoPrimary;
  final Color borderSurfacePrimary;
  final Color borderContextualPrimary;
  final Color borderOverlayPrimary;
  final Color borderBrandPrimary;
  final Color borderBrandSecondary;
  final Color borderBrandTertiary;
  final Color borderNeutralPrimary;
  final Color borderNeutralSecondary;
  final Color borderNeutralTertiary;
  final Color borderInfoPrimary;
  final Color borderInfoSecondary;
  final Color borderInfoTertiary;
  final Color borderSuccessPrimary;
  final Color borderSuccessSecondary;
  final Color borderSuccessTertiary;
  final Color borderWarningPrimary;
  final Color borderWarningSecondary;
  final Color borderWarningTertiary;
  final Color borderErrorPrimary;
  final Color borderErrorSecondary;
  final Color borderErrorTertiary;
  final Color borderDisabledPrimary;

  static const LaColors light = LaColors(
    surfacePrimary: Color(0xFFFFFFFF),
    surfaceSecondary: Color(0xFFFFFFFF),
    surfaceTertiary: Color(0xFFF8F8F9),
    overlayPrimary: Color(0x0D1B1B23),
    overlayPrimaryHover: Color(0x1A1B1B23),
    overlayOnSurface: Color(0xA6F8F8F9),
    contextualPrimary: Color(0x0D1B1B23),
    contextualPrimaryHover: Color(0x1A1B1B23),
    backgroundBrandPrimary: Color(0xFF5E5CD0),
    backgroundBrandPrimaryHover: Color(0xFF4844AD),
    backgroundBrandSecondary: Color(0xFFDDE2F5),
    backgroundBrandSecondaryHover: Color(0xFFCED3F1),
    backgroundBrandTertiary: Color(0xFFEEF1FA),
    backgroundBrandTertiaryHover: Color(0xFFDDE2F5),
    backgroundNeutralPrimary: Color(0xFF69697D),
    backgroundNeutralPrimaryHover: Color(0xFF515164),
    backgroundNeutralSecondary: Color(0xFFE2E2EA),
    backgroundNeutralSecondaryHover: Color(0xFFD3D4E0),
    backgroundNeutralTertiary: Color(0xFFF0F0F3),
    backgroundNeutralTertiaryHover: Color(0xFFE2E2EA),
    backgroundInfoPrimary: Color(0xFF0069CF),
    backgroundInfoPrimaryHover: Color(0xFF0D4EAA),
    backgroundInfoSecondary: Color(0xFFD5E4F3),
    backgroundInfoSecondaryHover: Color(0xFFBFD7F0),
    backgroundInfoTertiary: Color(0xFFEAF2F9),
    backgroundInfoTertiaryHover: Color(0xFFD5E4F3),
    backgroundSuccessPrimary: Color(0xFF027B3E),
    backgroundSuccessPrimaryHover: Color(0xFF006024),
    backgroundSuccessSecondary: Color(0xFFCFE4D4),
    backgroundSuccessSecondaryHover: Color(0xFFBAD9C1),
    backgroundSuccessTertiary: Color(0xFFE8F1EA),
    backgroundSuccessTertiaryHover: Color(0xFFCFE4D4),
    backgroundWarningPrimary: Color(0xFFBC4200),
    backgroundWarningPrimaryHover: Color(0xFF9E2300),
    backgroundWarningSecondary: Color(0xFFF1E0D3),
    backgroundWarningSecondaryHover: Color(0xFFECD0BC),
    backgroundWarningTertiary: Color(0xFFF8F0E9),
    backgroundWarningTertiaryHover: Color(0xFFF1E0D3),
    backgroundErrorPrimary: Color(0xFFD7010E),
    backgroundErrorPrimaryHover: Color(0xFFAA0000),
    backgroundErrorSecondary: Color(0xFFF4DFD9),
    backgroundErrorSecondaryHover: Color(0xFFF0CEC6),
    backgroundErrorTertiary: Color(0xFFF9EFEC),
    backgroundErrorTertiaryHover: Color(0xFFF4DFD9),
    backgroundDisabledPrimary: Color(0xFFE2E2EA),
    backgroundDisabledSecondary: Color(0xFFF0F0F3),
    contentContextualPrimary: Color(0xF2F8F8F9),
    contentOverlayPrimary: Color(0xF2F8F8F9),
    contentOverlaySecondary: Color(0xCCF8F8F9),
    contentBrandPrimary: Color(0xFF3E3B98),
    contentBrandSecondary: Color(0xFF534FC2),
    contentBrandTertiary: Color(0xFF5E5CD0),
    contentOnBrand: Color(0xFFEEF1FA),
    contentNeutralPrimary: Color(0xFF25252F),
    contentNeutralSecondary: Color(0xFF5D5D70),
    contentNeutralTertiary: Color(0xFF69697D),
    contentOnNeutral: Color(0xFFF0F0F3),
    contentInfoPrimary: Color(0xFF124394),
    contentInfoSecondary: Color(0xFF005BC0),
    contentInfoTertiary: Color(0xFF0069CF),
    contentOnInfo: Color(0xFFEAF2F9),
    contentSuccessPrimary: Color(0xFF005317),
    contentSuccessSecondary: Color(0xFF016D31),
    contentSuccessTertiary: Color(0xFF027B3E),
    contentOnSuccess: Color(0xFFE8F1EA),
    contentWarningPrimary: Color(0xFF882011),
    contentWarningSecondary: Color(0xFFAD3300),
    contentWarningTertiary: Color(0xFFBC4200),
    contentOnWarning: Color(0xFFF8F0E9),
    contentErrorPrimary: Color(0xFF910C06),
    contentErrorSecondary: Color(0xFFC00100),
    contentErrorTertiary: Color(0xFFD7010E),
    contentOnError: Color(0xFFF9EFEC),
    contentDisabledPrimary: Color(0xFFA9A9BF),
    contentDisabledSecondary: Color(0x80F8F8F9),
    logoPrimary: Color(0xFF4844AD),
    borderSurfacePrimary: Color(0xFFE2E2EA),
    borderContextualPrimary: Color(0x33F8F8F9),
    borderOverlayPrimary: Color(0x33F8F8F9),
    borderBrandPrimary: Color(0xFF5E5CD0),
    borderBrandSecondary: Color(0xFFA0A5F6),
    borderBrandTertiary: Color(0xFFCED3F1),
    borderNeutralPrimary: Color(0xFF69697D),
    borderNeutralSecondary: Color(0xFFA9A9BF),
    borderNeutralTertiary: Color(0xFFD3D4E0),
    borderInfoPrimary: Color(0xFF0069CF),
    borderInfoSecondary: Color(0xFF6EB0F2),
    borderInfoTertiary: Color(0xFFBFD7F0),
    borderSuccessPrimary: Color(0xFF027B3E),
    borderSuccessSecondary: Color(0xFF6CBA83),
    borderSuccessTertiary: Color(0xFFBAD9C1),
    borderWarningPrimary: Color(0xFFBC4200),
    borderWarningSecondary: Color(0xFFEB9970),
    borderWarningTertiary: Color(0xFFECD0BC),
    borderErrorPrimary: Color(0xFFD7010E),
    borderErrorSecondary: Color(0xFFEF9486),
    borderErrorTertiary: Color(0xFFF0CEC6),
    borderDisabledPrimary: Color(0xFFE2E2EA),
  );

  static const LaColors dark = LaColors(
    surfacePrimary: Color(0xFF2F303D),
    surfaceSecondary: Color(0xFF25252F),
    surfaceTertiary: Color(0xFF1B1B23),
    overlayPrimary: Color(0x0DF8F8F9),
    overlayPrimaryHover: Color(0x1AF8F8F9),
    overlayOnSurface: Color(0xA61B1B23),
    contextualPrimary: Color(0x0DF8F8F9),
    contextualPrimaryHover: Color(0x1AF8F8F9),
    backgroundBrandPrimary: Color(0xFF5E5CD0),
    backgroundBrandPrimaryHover: Color(0xFF4844AD),
    backgroundBrandSecondary: Color(0xFF3E3B98),
    backgroundBrandSecondaryHover: Color(0xFF36347D),
    backgroundBrandTertiary: Color(0xFF36347D),
    backgroundBrandTertiaryHover: Color(0xFF2D2F5F),
    backgroundNeutralPrimary: Color(0xFF69697D),
    backgroundNeutralPrimaryHover: Color(0xFF515164),
    backgroundNeutralSecondary: Color(0xFF454558),
    backgroundNeutralSecondaryHover: Color(0xFF3A3A4C),
    backgroundNeutralTertiary: Color(0xFF3A3A4C),
    backgroundNeutralTertiaryHover: Color(0xFF2F303D),
    backgroundInfoPrimary: Color(0xFF0069CF),
    backgroundInfoPrimaryHover: Color(0xFF0D4EAA),
    backgroundInfoSecondary: Color(0xFF124394),
    backgroundInfoSecondaryHover: Color(0xFF163878),
    backgroundInfoTertiary: Color(0xFF163878),
    backgroundInfoTertiaryHover: Color(0xFF192F5A),
    backgroundSuccessPrimary: Color(0xFF027B3E),
    backgroundSuccessPrimaryHover: Color(0xFF006024),
    backgroundSuccessSecondary: Color(0xFF005317),
    backgroundSuccessSecondaryHover: Color(0xFF0D4511),
    backgroundSuccessTertiary: Color(0xFF0D4511),
    backgroundSuccessTertiaryHover: Color(0xFF11380E),
    backgroundWarningPrimary: Color(0xFFBC4200),
    backgroundWarningPrimaryHover: Color(0xFF9E2300),
    backgroundWarningSecondary: Color(0xFF882011),
    backgroundWarningSecondaryHover: Color(0xFF731E16),
    backgroundWarningTertiary: Color(0xFF731E16),
    backgroundWarningTertiaryHover: Color(0xFF58201A),
    backgroundErrorPrimary: Color(0xFFD7010E),
    backgroundErrorPrimaryHover: Color(0xFFAA0000),
    backgroundErrorSecondary: Color(0xFF910C06),
    backgroundErrorSecondaryHover: Color(0xFF731E16),
    backgroundErrorTertiary: Color(0xFF731E16),
    backgroundErrorTertiaryHover: Color(0xFF58201A),
    backgroundDisabledPrimary: Color(0xFF3A3A4C),
    backgroundDisabledSecondary: Color(0xFF2F303D),
    contentContextualPrimary: Color(0xD91B1B23),
    contentOverlayPrimary: Color(0xD91B1B23),
    contentOverlaySecondary: Color(0xB21B1B23),
    contentBrandPrimary: Color(0xFFEEF1FA),
    contentBrandSecondary: Color(0xFFDDE2F5),
    contentBrandTertiary: Color(0xFFAFB5F1),
    contentOnBrand: Color(0xFFEEF1FA),
    contentNeutralPrimary: Color(0xFFF0F0F3),
    contentNeutralSecondary: Color(0xFFE2E2EA),
    contentNeutralTertiary: Color(0xFFB7B7CB),
    contentOnNeutral: Color(0xFFF0F0F3),
    contentInfoPrimary: Color(0xFFEAF2F9),
    contentInfoSecondary: Color(0xFFD5E4F3),
    contentInfoTertiary: Color(0xFF8DBDEF),
    contentOnInfo: Color(0xFFEAF2F9),
    contentSuccessPrimary: Color(0xFFE8F1EA),
    contentSuccessSecondary: Color(0xFFCFE4D4),
    contentSuccessTertiary: Color(0xFF86C597),
    contentOnSuccess: Color(0xFFE8F1EA),
    contentWarningPrimary: Color(0xFFF8F0E9),
    contentWarningSecondary: Color(0xFFF1E0D3),
    contentWarningTertiary: Color(0xFFE8AE8A),
    contentOnWarning: Color(0xFFF8F0E9),
    contentErrorPrimary: Color(0xFFF9EFEC),
    contentErrorSecondary: Color(0xFFF4DFD9),
    contentErrorTertiary: Color(0xFFEEA99D),
    contentOnError: Color(0xFFF9EFEC),
    contentDisabledPrimary: Color(0xFF5D5D70),
    contentDisabledSecondary: Color(0x4D1B1B23),
    logoPrimary: Color(0xFFBEC5F0),
    borderSurfacePrimary: Color(0xFF3A3A4C),
    borderContextualPrimary: Color(0x331B1B23),
    borderOverlayPrimary: Color(0x331B1B23),
    borderBrandPrimary: Color(0xFF7576EE),
    borderBrandSecondary: Color(0xFF534FC2),
    borderBrandTertiary: Color(0xFF3E3B98),
    borderNeutralPrimary: Color(0xFF828297),
    borderNeutralSecondary: Color(0xFF5D5D70),
    borderNeutralTertiary: Color(0xFF454558),
    borderInfoPrimary: Color(0xFF1185ED),
    borderInfoSecondary: Color(0xFF005BC0),
    borderInfoTertiary: Color(0xFF124394),
    borderSuccessPrimary: Color(0xFF309556),
    borderSuccessSecondary: Color(0xFF016D31),
    borderSuccessTertiary: Color(0xFF005317),
    borderWarningPrimary: Color(0xFFDA5E18),
    borderWarningSecondary: Color(0xFFAD3300),
    borderWarningTertiary: Color(0xFF882011),
    borderErrorPrimary: Color(0xFFF0463D),
    borderErrorSecondary: Color(0xFFC00100),
    borderErrorTertiary: Color(0xFF910C06),
    borderDisabledPrimary: Color(0xFF2F303D),
  );

  @override
  LaColors copyWith({
    Color? surfacePrimary,
    Color? surfaceSecondary,
    Color? surfaceTertiary,
    Color? overlayPrimary,
    Color? overlayPrimaryHover,
    Color? overlayOnSurface,
    Color? contextualPrimary,
    Color? contextualPrimaryHover,
    Color? backgroundBrandPrimary,
    Color? backgroundBrandPrimaryHover,
    Color? backgroundBrandSecondary,
    Color? backgroundBrandSecondaryHover,
    Color? backgroundBrandTertiary,
    Color? backgroundBrandTertiaryHover,
    Color? backgroundNeutralPrimary,
    Color? backgroundNeutralPrimaryHover,
    Color? backgroundNeutralSecondary,
    Color? backgroundNeutralSecondaryHover,
    Color? backgroundNeutralTertiary,
    Color? backgroundNeutralTertiaryHover,
    Color? backgroundInfoPrimary,
    Color? backgroundInfoPrimaryHover,
    Color? backgroundInfoSecondary,
    Color? backgroundInfoSecondaryHover,
    Color? backgroundInfoTertiary,
    Color? backgroundInfoTertiaryHover,
    Color? backgroundSuccessPrimary,
    Color? backgroundSuccessPrimaryHover,
    Color? backgroundSuccessSecondary,
    Color? backgroundSuccessSecondaryHover,
    Color? backgroundSuccessTertiary,
    Color? backgroundSuccessTertiaryHover,
    Color? backgroundWarningPrimary,
    Color? backgroundWarningPrimaryHover,
    Color? backgroundWarningSecondary,
    Color? backgroundWarningSecondaryHover,
    Color? backgroundWarningTertiary,
    Color? backgroundWarningTertiaryHover,
    Color? backgroundErrorPrimary,
    Color? backgroundErrorPrimaryHover,
    Color? backgroundErrorSecondary,
    Color? backgroundErrorSecondaryHover,
    Color? backgroundErrorTertiary,
    Color? backgroundErrorTertiaryHover,
    Color? backgroundDisabledPrimary,
    Color? backgroundDisabledSecondary,
    Color? contentContextualPrimary,
    Color? contentOverlayPrimary,
    Color? contentOverlaySecondary,
    Color? contentBrandPrimary,
    Color? contentBrandSecondary,
    Color? contentBrandTertiary,
    Color? contentOnBrand,
    Color? contentNeutralPrimary,
    Color? contentNeutralSecondary,
    Color? contentNeutralTertiary,
    Color? contentOnNeutral,
    Color? contentInfoPrimary,
    Color? contentInfoSecondary,
    Color? contentInfoTertiary,
    Color? contentOnInfo,
    Color? contentSuccessPrimary,
    Color? contentSuccessSecondary,
    Color? contentSuccessTertiary,
    Color? contentOnSuccess,
    Color? contentWarningPrimary,
    Color? contentWarningSecondary,
    Color? contentWarningTertiary,
    Color? contentOnWarning,
    Color? contentErrorPrimary,
    Color? contentErrorSecondary,
    Color? contentErrorTertiary,
    Color? contentOnError,
    Color? contentDisabledPrimary,
    Color? contentDisabledSecondary,
    Color? logoPrimary,
    Color? borderSurfacePrimary,
    Color? borderContextualPrimary,
    Color? borderOverlayPrimary,
    Color? borderBrandPrimary,
    Color? borderBrandSecondary,
    Color? borderBrandTertiary,
    Color? borderNeutralPrimary,
    Color? borderNeutralSecondary,
    Color? borderNeutralTertiary,
    Color? borderInfoPrimary,
    Color? borderInfoSecondary,
    Color? borderInfoTertiary,
    Color? borderSuccessPrimary,
    Color? borderSuccessSecondary,
    Color? borderSuccessTertiary,
    Color? borderWarningPrimary,
    Color? borderWarningSecondary,
    Color? borderWarningTertiary,
    Color? borderErrorPrimary,
    Color? borderErrorSecondary,
    Color? borderErrorTertiary,
    Color? borderDisabledPrimary,
  }) {
    return LaColors(
      surfacePrimary: surfacePrimary ?? this.surfacePrimary,
      surfaceSecondary: surfaceSecondary ?? this.surfaceSecondary,
      surfaceTertiary: surfaceTertiary ?? this.surfaceTertiary,
      overlayPrimary: overlayPrimary ?? this.overlayPrimary,
      overlayPrimaryHover: overlayPrimaryHover ?? this.overlayPrimaryHover,
      overlayOnSurface: overlayOnSurface ?? this.overlayOnSurface,
      contextualPrimary: contextualPrimary ?? this.contextualPrimary,
      contextualPrimaryHover:
          contextualPrimaryHover ?? this.contextualPrimaryHover,
      backgroundBrandPrimary:
          backgroundBrandPrimary ?? this.backgroundBrandPrimary,
      backgroundBrandPrimaryHover:
          backgroundBrandPrimaryHover ?? this.backgroundBrandPrimaryHover,
      backgroundBrandSecondary:
          backgroundBrandSecondary ?? this.backgroundBrandSecondary,
      backgroundBrandSecondaryHover:
          backgroundBrandSecondaryHover ?? this.backgroundBrandSecondaryHover,
      backgroundBrandTertiary:
          backgroundBrandTertiary ?? this.backgroundBrandTertiary,
      backgroundBrandTertiaryHover:
          backgroundBrandTertiaryHover ?? this.backgroundBrandTertiaryHover,
      backgroundNeutralPrimary:
          backgroundNeutralPrimary ?? this.backgroundNeutralPrimary,
      backgroundNeutralPrimaryHover:
          backgroundNeutralPrimaryHover ?? this.backgroundNeutralPrimaryHover,
      backgroundNeutralSecondary:
          backgroundNeutralSecondary ?? this.backgroundNeutralSecondary,
      backgroundNeutralSecondaryHover:
          backgroundNeutralSecondaryHover ??
          this.backgroundNeutralSecondaryHover,
      backgroundNeutralTertiary:
          backgroundNeutralTertiary ?? this.backgroundNeutralTertiary,
      backgroundNeutralTertiaryHover:
          backgroundNeutralTertiaryHover ?? this.backgroundNeutralTertiaryHover,
      backgroundInfoPrimary:
          backgroundInfoPrimary ?? this.backgroundInfoPrimary,
      backgroundInfoPrimaryHover:
          backgroundInfoPrimaryHover ?? this.backgroundInfoPrimaryHover,
      backgroundInfoSecondary:
          backgroundInfoSecondary ?? this.backgroundInfoSecondary,
      backgroundInfoSecondaryHover:
          backgroundInfoSecondaryHover ?? this.backgroundInfoSecondaryHover,
      backgroundInfoTertiary:
          backgroundInfoTertiary ?? this.backgroundInfoTertiary,
      backgroundInfoTertiaryHover:
          backgroundInfoTertiaryHover ?? this.backgroundInfoTertiaryHover,
      backgroundSuccessPrimary:
          backgroundSuccessPrimary ?? this.backgroundSuccessPrimary,
      backgroundSuccessPrimaryHover:
          backgroundSuccessPrimaryHover ?? this.backgroundSuccessPrimaryHover,
      backgroundSuccessSecondary:
          backgroundSuccessSecondary ?? this.backgroundSuccessSecondary,
      backgroundSuccessSecondaryHover:
          backgroundSuccessSecondaryHover ??
          this.backgroundSuccessSecondaryHover,
      backgroundSuccessTertiary:
          backgroundSuccessTertiary ?? this.backgroundSuccessTertiary,
      backgroundSuccessTertiaryHover:
          backgroundSuccessTertiaryHover ?? this.backgroundSuccessTertiaryHover,
      backgroundWarningPrimary:
          backgroundWarningPrimary ?? this.backgroundWarningPrimary,
      backgroundWarningPrimaryHover:
          backgroundWarningPrimaryHover ?? this.backgroundWarningPrimaryHover,
      backgroundWarningSecondary:
          backgroundWarningSecondary ?? this.backgroundWarningSecondary,
      backgroundWarningSecondaryHover:
          backgroundWarningSecondaryHover ??
          this.backgroundWarningSecondaryHover,
      backgroundWarningTertiary:
          backgroundWarningTertiary ?? this.backgroundWarningTertiary,
      backgroundWarningTertiaryHover:
          backgroundWarningTertiaryHover ?? this.backgroundWarningTertiaryHover,
      backgroundErrorPrimary:
          backgroundErrorPrimary ?? this.backgroundErrorPrimary,
      backgroundErrorPrimaryHover:
          backgroundErrorPrimaryHover ?? this.backgroundErrorPrimaryHover,
      backgroundErrorSecondary:
          backgroundErrorSecondary ?? this.backgroundErrorSecondary,
      backgroundErrorSecondaryHover:
          backgroundErrorSecondaryHover ?? this.backgroundErrorSecondaryHover,
      backgroundErrorTertiary:
          backgroundErrorTertiary ?? this.backgroundErrorTertiary,
      backgroundErrorTertiaryHover:
          backgroundErrorTertiaryHover ?? this.backgroundErrorTertiaryHover,
      backgroundDisabledPrimary:
          backgroundDisabledPrimary ?? this.backgroundDisabledPrimary,
      backgroundDisabledSecondary:
          backgroundDisabledSecondary ?? this.backgroundDisabledSecondary,
      contentContextualPrimary:
          contentContextualPrimary ?? this.contentContextualPrimary,
      contentOverlayPrimary:
          contentOverlayPrimary ?? this.contentOverlayPrimary,
      contentOverlaySecondary:
          contentOverlaySecondary ?? this.contentOverlaySecondary,
      contentBrandPrimary: contentBrandPrimary ?? this.contentBrandPrimary,
      contentBrandSecondary:
          contentBrandSecondary ?? this.contentBrandSecondary,
      contentBrandTertiary: contentBrandTertiary ?? this.contentBrandTertiary,
      contentOnBrand: contentOnBrand ?? this.contentOnBrand,
      contentNeutralPrimary:
          contentNeutralPrimary ?? this.contentNeutralPrimary,
      contentNeutralSecondary:
          contentNeutralSecondary ?? this.contentNeutralSecondary,
      contentNeutralTertiary:
          contentNeutralTertiary ?? this.contentNeutralTertiary,
      contentOnNeutral: contentOnNeutral ?? this.contentOnNeutral,
      contentInfoPrimary: contentInfoPrimary ?? this.contentInfoPrimary,
      contentInfoSecondary: contentInfoSecondary ?? this.contentInfoSecondary,
      contentInfoTertiary: contentInfoTertiary ?? this.contentInfoTertiary,
      contentOnInfo: contentOnInfo ?? this.contentOnInfo,
      contentSuccessPrimary:
          contentSuccessPrimary ?? this.contentSuccessPrimary,
      contentSuccessSecondary:
          contentSuccessSecondary ?? this.contentSuccessSecondary,
      contentSuccessTertiary:
          contentSuccessTertiary ?? this.contentSuccessTertiary,
      contentOnSuccess: contentOnSuccess ?? this.contentOnSuccess,
      contentWarningPrimary:
          contentWarningPrimary ?? this.contentWarningPrimary,
      contentWarningSecondary:
          contentWarningSecondary ?? this.contentWarningSecondary,
      contentWarningTertiary:
          contentWarningTertiary ?? this.contentWarningTertiary,
      contentOnWarning: contentOnWarning ?? this.contentOnWarning,
      contentErrorPrimary: contentErrorPrimary ?? this.contentErrorPrimary,
      contentErrorSecondary:
          contentErrorSecondary ?? this.contentErrorSecondary,
      contentErrorTertiary: contentErrorTertiary ?? this.contentErrorTertiary,
      contentOnError: contentOnError ?? this.contentOnError,
      contentDisabledPrimary:
          contentDisabledPrimary ?? this.contentDisabledPrimary,
      contentDisabledSecondary:
          contentDisabledSecondary ?? this.contentDisabledSecondary,
      logoPrimary: logoPrimary ?? this.logoPrimary,
      borderSurfacePrimary: borderSurfacePrimary ?? this.borderSurfacePrimary,
      borderContextualPrimary:
          borderContextualPrimary ?? this.borderContextualPrimary,
      borderOverlayPrimary: borderOverlayPrimary ?? this.borderOverlayPrimary,
      borderBrandPrimary: borderBrandPrimary ?? this.borderBrandPrimary,
      borderBrandSecondary: borderBrandSecondary ?? this.borderBrandSecondary,
      borderBrandTertiary: borderBrandTertiary ?? this.borderBrandTertiary,
      borderNeutralPrimary: borderNeutralPrimary ?? this.borderNeutralPrimary,
      borderNeutralSecondary:
          borderNeutralSecondary ?? this.borderNeutralSecondary,
      borderNeutralTertiary:
          borderNeutralTertiary ?? this.borderNeutralTertiary,
      borderInfoPrimary: borderInfoPrimary ?? this.borderInfoPrimary,
      borderInfoSecondary: borderInfoSecondary ?? this.borderInfoSecondary,
      borderInfoTertiary: borderInfoTertiary ?? this.borderInfoTertiary,
      borderSuccessPrimary: borderSuccessPrimary ?? this.borderSuccessPrimary,
      borderSuccessSecondary:
          borderSuccessSecondary ?? this.borderSuccessSecondary,
      borderSuccessTertiary:
          borderSuccessTertiary ?? this.borderSuccessTertiary,
      borderWarningPrimary: borderWarningPrimary ?? this.borderWarningPrimary,
      borderWarningSecondary:
          borderWarningSecondary ?? this.borderWarningSecondary,
      borderWarningTertiary:
          borderWarningTertiary ?? this.borderWarningTertiary,
      borderErrorPrimary: borderErrorPrimary ?? this.borderErrorPrimary,
      borderErrorSecondary: borderErrorSecondary ?? this.borderErrorSecondary,
      borderErrorTertiary: borderErrorTertiary ?? this.borderErrorTertiary,
      borderDisabledPrimary:
          borderDisabledPrimary ?? this.borderDisabledPrimary,
    );
  }

  @override
  LaColors lerp(ThemeExtension<LaColors>? other, double t) {
    if (other is! LaColors) return this;
    return LaColors(
      surfacePrimary: Color.lerp(surfacePrimary, other.surfacePrimary, t)!,
      surfaceSecondary: Color.lerp(
        surfaceSecondary,
        other.surfaceSecondary,
        t,
      )!,
      surfaceTertiary: Color.lerp(surfaceTertiary, other.surfaceTertiary, t)!,
      overlayPrimary: Color.lerp(overlayPrimary, other.overlayPrimary, t)!,
      overlayPrimaryHover: Color.lerp(
        overlayPrimaryHover,
        other.overlayPrimaryHover,
        t,
      )!,
      overlayOnSurface: Color.lerp(
        overlayOnSurface,
        other.overlayOnSurface,
        t,
      )!,
      contextualPrimary: Color.lerp(
        contextualPrimary,
        other.contextualPrimary,
        t,
      )!,
      contextualPrimaryHover: Color.lerp(
        contextualPrimaryHover,
        other.contextualPrimaryHover,
        t,
      )!,
      backgroundBrandPrimary: Color.lerp(
        backgroundBrandPrimary,
        other.backgroundBrandPrimary,
        t,
      )!,
      backgroundBrandPrimaryHover: Color.lerp(
        backgroundBrandPrimaryHover,
        other.backgroundBrandPrimaryHover,
        t,
      )!,
      backgroundBrandSecondary: Color.lerp(
        backgroundBrandSecondary,
        other.backgroundBrandSecondary,
        t,
      )!,
      backgroundBrandSecondaryHover: Color.lerp(
        backgroundBrandSecondaryHover,
        other.backgroundBrandSecondaryHover,
        t,
      )!,
      backgroundBrandTertiary: Color.lerp(
        backgroundBrandTertiary,
        other.backgroundBrandTertiary,
        t,
      )!,
      backgroundBrandTertiaryHover: Color.lerp(
        backgroundBrandTertiaryHover,
        other.backgroundBrandTertiaryHover,
        t,
      )!,
      backgroundNeutralPrimary: Color.lerp(
        backgroundNeutralPrimary,
        other.backgroundNeutralPrimary,
        t,
      )!,
      backgroundNeutralPrimaryHover: Color.lerp(
        backgroundNeutralPrimaryHover,
        other.backgroundNeutralPrimaryHover,
        t,
      )!,
      backgroundNeutralSecondary: Color.lerp(
        backgroundNeutralSecondary,
        other.backgroundNeutralSecondary,
        t,
      )!,
      backgroundNeutralSecondaryHover: Color.lerp(
        backgroundNeutralSecondaryHover,
        other.backgroundNeutralSecondaryHover,
        t,
      )!,
      backgroundNeutralTertiary: Color.lerp(
        backgroundNeutralTertiary,
        other.backgroundNeutralTertiary,
        t,
      )!,
      backgroundNeutralTertiaryHover: Color.lerp(
        backgroundNeutralTertiaryHover,
        other.backgroundNeutralTertiaryHover,
        t,
      )!,
      backgroundInfoPrimary: Color.lerp(
        backgroundInfoPrimary,
        other.backgroundInfoPrimary,
        t,
      )!,
      backgroundInfoPrimaryHover: Color.lerp(
        backgroundInfoPrimaryHover,
        other.backgroundInfoPrimaryHover,
        t,
      )!,
      backgroundInfoSecondary: Color.lerp(
        backgroundInfoSecondary,
        other.backgroundInfoSecondary,
        t,
      )!,
      backgroundInfoSecondaryHover: Color.lerp(
        backgroundInfoSecondaryHover,
        other.backgroundInfoSecondaryHover,
        t,
      )!,
      backgroundInfoTertiary: Color.lerp(
        backgroundInfoTertiary,
        other.backgroundInfoTertiary,
        t,
      )!,
      backgroundInfoTertiaryHover: Color.lerp(
        backgroundInfoTertiaryHover,
        other.backgroundInfoTertiaryHover,
        t,
      )!,
      backgroundSuccessPrimary: Color.lerp(
        backgroundSuccessPrimary,
        other.backgroundSuccessPrimary,
        t,
      )!,
      backgroundSuccessPrimaryHover: Color.lerp(
        backgroundSuccessPrimaryHover,
        other.backgroundSuccessPrimaryHover,
        t,
      )!,
      backgroundSuccessSecondary: Color.lerp(
        backgroundSuccessSecondary,
        other.backgroundSuccessSecondary,
        t,
      )!,
      backgroundSuccessSecondaryHover: Color.lerp(
        backgroundSuccessSecondaryHover,
        other.backgroundSuccessSecondaryHover,
        t,
      )!,
      backgroundSuccessTertiary: Color.lerp(
        backgroundSuccessTertiary,
        other.backgroundSuccessTertiary,
        t,
      )!,
      backgroundSuccessTertiaryHover: Color.lerp(
        backgroundSuccessTertiaryHover,
        other.backgroundSuccessTertiaryHover,
        t,
      )!,
      backgroundWarningPrimary: Color.lerp(
        backgroundWarningPrimary,
        other.backgroundWarningPrimary,
        t,
      )!,
      backgroundWarningPrimaryHover: Color.lerp(
        backgroundWarningPrimaryHover,
        other.backgroundWarningPrimaryHover,
        t,
      )!,
      backgroundWarningSecondary: Color.lerp(
        backgroundWarningSecondary,
        other.backgroundWarningSecondary,
        t,
      )!,
      backgroundWarningSecondaryHover: Color.lerp(
        backgroundWarningSecondaryHover,
        other.backgroundWarningSecondaryHover,
        t,
      )!,
      backgroundWarningTertiary: Color.lerp(
        backgroundWarningTertiary,
        other.backgroundWarningTertiary,
        t,
      )!,
      backgroundWarningTertiaryHover: Color.lerp(
        backgroundWarningTertiaryHover,
        other.backgroundWarningTertiaryHover,
        t,
      )!,
      backgroundErrorPrimary: Color.lerp(
        backgroundErrorPrimary,
        other.backgroundErrorPrimary,
        t,
      )!,
      backgroundErrorPrimaryHover: Color.lerp(
        backgroundErrorPrimaryHover,
        other.backgroundErrorPrimaryHover,
        t,
      )!,
      backgroundErrorSecondary: Color.lerp(
        backgroundErrorSecondary,
        other.backgroundErrorSecondary,
        t,
      )!,
      backgroundErrorSecondaryHover: Color.lerp(
        backgroundErrorSecondaryHover,
        other.backgroundErrorSecondaryHover,
        t,
      )!,
      backgroundErrorTertiary: Color.lerp(
        backgroundErrorTertiary,
        other.backgroundErrorTertiary,
        t,
      )!,
      backgroundErrorTertiaryHover: Color.lerp(
        backgroundErrorTertiaryHover,
        other.backgroundErrorTertiaryHover,
        t,
      )!,
      backgroundDisabledPrimary: Color.lerp(
        backgroundDisabledPrimary,
        other.backgroundDisabledPrimary,
        t,
      )!,
      backgroundDisabledSecondary: Color.lerp(
        backgroundDisabledSecondary,
        other.backgroundDisabledSecondary,
        t,
      )!,
      contentContextualPrimary: Color.lerp(
        contentContextualPrimary,
        other.contentContextualPrimary,
        t,
      )!,
      contentOverlayPrimary: Color.lerp(
        contentOverlayPrimary,
        other.contentOverlayPrimary,
        t,
      )!,
      contentOverlaySecondary: Color.lerp(
        contentOverlaySecondary,
        other.contentOverlaySecondary,
        t,
      )!,
      contentBrandPrimary: Color.lerp(
        contentBrandPrimary,
        other.contentBrandPrimary,
        t,
      )!,
      contentBrandSecondary: Color.lerp(
        contentBrandSecondary,
        other.contentBrandSecondary,
        t,
      )!,
      contentBrandTertiary: Color.lerp(
        contentBrandTertiary,
        other.contentBrandTertiary,
        t,
      )!,
      contentOnBrand: Color.lerp(contentOnBrand, other.contentOnBrand, t)!,
      contentNeutralPrimary: Color.lerp(
        contentNeutralPrimary,
        other.contentNeutralPrimary,
        t,
      )!,
      contentNeutralSecondary: Color.lerp(
        contentNeutralSecondary,
        other.contentNeutralSecondary,
        t,
      )!,
      contentNeutralTertiary: Color.lerp(
        contentNeutralTertiary,
        other.contentNeutralTertiary,
        t,
      )!,
      contentOnNeutral: Color.lerp(
        contentOnNeutral,
        other.contentOnNeutral,
        t,
      )!,
      contentInfoPrimary: Color.lerp(
        contentInfoPrimary,
        other.contentInfoPrimary,
        t,
      )!,
      contentInfoSecondary: Color.lerp(
        contentInfoSecondary,
        other.contentInfoSecondary,
        t,
      )!,
      contentInfoTertiary: Color.lerp(
        contentInfoTertiary,
        other.contentInfoTertiary,
        t,
      )!,
      contentOnInfo: Color.lerp(contentOnInfo, other.contentOnInfo, t)!,
      contentSuccessPrimary: Color.lerp(
        contentSuccessPrimary,
        other.contentSuccessPrimary,
        t,
      )!,
      contentSuccessSecondary: Color.lerp(
        contentSuccessSecondary,
        other.contentSuccessSecondary,
        t,
      )!,
      contentSuccessTertiary: Color.lerp(
        contentSuccessTertiary,
        other.contentSuccessTertiary,
        t,
      )!,
      contentOnSuccess: Color.lerp(
        contentOnSuccess,
        other.contentOnSuccess,
        t,
      )!,
      contentWarningPrimary: Color.lerp(
        contentWarningPrimary,
        other.contentWarningPrimary,
        t,
      )!,
      contentWarningSecondary: Color.lerp(
        contentWarningSecondary,
        other.contentWarningSecondary,
        t,
      )!,
      contentWarningTertiary: Color.lerp(
        contentWarningTertiary,
        other.contentWarningTertiary,
        t,
      )!,
      contentOnWarning: Color.lerp(
        contentOnWarning,
        other.contentOnWarning,
        t,
      )!,
      contentErrorPrimary: Color.lerp(
        contentErrorPrimary,
        other.contentErrorPrimary,
        t,
      )!,
      contentErrorSecondary: Color.lerp(
        contentErrorSecondary,
        other.contentErrorSecondary,
        t,
      )!,
      contentErrorTertiary: Color.lerp(
        contentErrorTertiary,
        other.contentErrorTertiary,
        t,
      )!,
      contentOnError: Color.lerp(contentOnError, other.contentOnError, t)!,
      contentDisabledPrimary: Color.lerp(
        contentDisabledPrimary,
        other.contentDisabledPrimary,
        t,
      )!,
      contentDisabledSecondary: Color.lerp(
        contentDisabledSecondary,
        other.contentDisabledSecondary,
        t,
      )!,
      logoPrimary: Color.lerp(logoPrimary, other.logoPrimary, t)!,
      borderSurfacePrimary: Color.lerp(
        borderSurfacePrimary,
        other.borderSurfacePrimary,
        t,
      )!,
      borderContextualPrimary: Color.lerp(
        borderContextualPrimary,
        other.borderContextualPrimary,
        t,
      )!,
      borderOverlayPrimary: Color.lerp(
        borderOverlayPrimary,
        other.borderOverlayPrimary,
        t,
      )!,
      borderBrandPrimary: Color.lerp(
        borderBrandPrimary,
        other.borderBrandPrimary,
        t,
      )!,
      borderBrandSecondary: Color.lerp(
        borderBrandSecondary,
        other.borderBrandSecondary,
        t,
      )!,
      borderBrandTertiary: Color.lerp(
        borderBrandTertiary,
        other.borderBrandTertiary,
        t,
      )!,
      borderNeutralPrimary: Color.lerp(
        borderNeutralPrimary,
        other.borderNeutralPrimary,
        t,
      )!,
      borderNeutralSecondary: Color.lerp(
        borderNeutralSecondary,
        other.borderNeutralSecondary,
        t,
      )!,
      borderNeutralTertiary: Color.lerp(
        borderNeutralTertiary,
        other.borderNeutralTertiary,
        t,
      )!,
      borderInfoPrimary: Color.lerp(
        borderInfoPrimary,
        other.borderInfoPrimary,
        t,
      )!,
      borderInfoSecondary: Color.lerp(
        borderInfoSecondary,
        other.borderInfoSecondary,
        t,
      )!,
      borderInfoTertiary: Color.lerp(
        borderInfoTertiary,
        other.borderInfoTertiary,
        t,
      )!,
      borderSuccessPrimary: Color.lerp(
        borderSuccessPrimary,
        other.borderSuccessPrimary,
        t,
      )!,
      borderSuccessSecondary: Color.lerp(
        borderSuccessSecondary,
        other.borderSuccessSecondary,
        t,
      )!,
      borderSuccessTertiary: Color.lerp(
        borderSuccessTertiary,
        other.borderSuccessTertiary,
        t,
      )!,
      borderWarningPrimary: Color.lerp(
        borderWarningPrimary,
        other.borderWarningPrimary,
        t,
      )!,
      borderWarningSecondary: Color.lerp(
        borderWarningSecondary,
        other.borderWarningSecondary,
        t,
      )!,
      borderWarningTertiary: Color.lerp(
        borderWarningTertiary,
        other.borderWarningTertiary,
        t,
      )!,
      borderErrorPrimary: Color.lerp(
        borderErrorPrimary,
        other.borderErrorPrimary,
        t,
      )!,
      borderErrorSecondary: Color.lerp(
        borderErrorSecondary,
        other.borderErrorSecondary,
        t,
      )!,
      borderErrorTertiary: Color.lerp(
        borderErrorTertiary,
        other.borderErrorTertiary,
        t,
      )!,
      borderDisabledPrimary: Color.lerp(
        borderDisabledPrimary,
        other.borderDisabledPrimary,
        t,
      )!,
    );
  }
}
