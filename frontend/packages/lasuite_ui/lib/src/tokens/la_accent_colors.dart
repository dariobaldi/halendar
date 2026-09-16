// GENERATED FILE. Derived from the La Suite numérique UI Kit
// (https://github.com/suitenumerique/ui-kit) design tokens
// ("contextuals.background.palette"). See ATTRIBUTION.md.
// Do not hand-edit; regenerate instead.

import 'package:flutter/widgets.dart';

/// One of the 11 named accent hues available for avatars, tags and
/// data-visualization color rotation.
enum LaAccentHue {
  brand,
  red,
  orange,
  brown,
  yellow,
  green,
  blue1,
  blue2,
  purple,
  pink,
  gray,
}

/// A trio of tones (background-strength colors) for one accent hue.
@immutable
class LaAccentTriple {
  const LaAccentTriple({
    required this.primary,
    required this.secondary,
    required this.tertiary,
  });

  final Color primary;
  final Color secondary;
  final Color tertiary;
}

/// Named accent color triples, keyed by [LaAccentHue], for light and dark.
abstract final class LaAccentColors {
  static const Map<LaAccentHue, LaAccentTriple> light = {
    LaAccentHue.brand: LaAccentTriple(
      primary: Color(0xFF6969DF),
      secondary: Color(0xFF8F94FD),
      tertiary: Color(0xFFCED3F1),
    ),
    LaAccentHue.red: LaAccentTriple(
      primary: Color(0xFFDA3B49),
      secondary: Color(0xFFF37B7E),
      tertiary: Color(0xFFF1CDCB),
    ),
    LaAccentHue.orange: LaAccentTriple(
      primary: Color(0xFFB95D33),
      secondary: Color(0xFFE5845A),
      tertiary: Color(0xFFECD0BD),
    ),
    LaAccentHue.brown: LaAccentTriple(
      primary: Color(0xFF8F7158),
      secondary: Color(0xFFB8987E),
      tertiary: Color(0xFFEBD0BA),
    ),
    LaAccentHue.yellow: LaAccentTriple(
      primary: Color(0xFF9D6E00),
      secondary: Color(0xFFC2972E),
      tertiary: Color(0xFFE1D4B7),
    ),
    LaAccentHue.green: LaAccentTriple(
      primary: Color(0xFF008948),
      secondary: Color(0xFF45B173),
      tertiary: Color(0xFFB8D8C1),
    ),
    LaAccentHue.blue1: LaAccentTriple(
      primary: Color(0xFF4279B9),
      secondary: Color(0xFF68A1E4),
      tertiary: Color(0xFFC1D7F0),
    ),
    LaAccentHue.blue2: LaAccentTriple(
      primary: Color(0xFF00848F),
      secondary: Color(0xFF00AFBA),
      tertiary: Color(0xFFB2DCE0),
    ),
    LaAccentHue.purple: LaAccentTriple(
      primary: Color(0xFF9961AF),
      secondary: Color(0xFFC188D9),
      tertiary: Color(0xFFE7D1E7),
    ),
    LaAccentHue.pink: LaAccentTriple(
      primary: Color(0xFFAA5F80),
      secondary: Color(0xFFD685A8),
      tertiary: Color(0xFFEACEDF),
    ),
    LaAccentHue.gray: LaAccentTriple(
      primary: Color(0xFF75758A),
      secondary: Color(0xFF9C9CB2),
      tertiary: Color(0xFFD3D4E0),
    ),
  };
  static const Map<LaAccentHue, LaAccentTriple> dark = {
    LaAccentHue.brand: LaAccentTriple(
      primary: Color(0xFF8F94FD),
      secondary: Color(0xFF7576EE),
      tertiary: Color(0xFF5E5CD0),
    ),
    LaAccentHue.red: LaAccentTriple(
      primary: Color(0xFFF37B7E),
      secondary: Color(0xFFE94A55),
      tertiary: Color(0xFFCA2A3C),
    ),
    LaAccentHue.orange: LaAccentTriple(
      primary: Color(0xFFE5845A),
      secondary: Color(0xFFC86A40),
      tertiary: Color(0xFFAB5025),
    ),
    LaAccentHue.brown: LaAccentTriple(
      primary: Color(0xFFB8987E),
      secondary: Color(0xFF9D7E65),
      tertiary: Color(0xFF82654C),
    ),
    LaAccentHue.yellow: LaAccentTriple(
      primary: Color(0xFFC2972E),
      secondary: Color(0xFFAB7B00),
      tertiary: Color(0xFF916100),
    ),
    LaAccentHue.green: LaAccentTriple(
      primary: Color(0xFF45B173),
      secondary: Color(0xFF029755),
      tertiary: Color(0xFF017B3B),
    ),
    LaAccentHue.blue1: LaAccentTriple(
      primary: Color(0xFF68A1E4),
      secondary: Color(0xFF4E86C7),
      tertiary: Color(0xFF356CAC),
    ),
    LaAccentHue.blue2: LaAccentTriple(
      primary: Color(0xFF00AFBA),
      secondary: Color(0xFF00929D),
      tertiary: Color(0xFF007682),
    ),
    LaAccentHue.purple: LaAccentTriple(
      primary: Color(0xFFC188D9),
      secondary: Color(0xFFA66EBD),
      tertiary: Color(0xFF8B55A1),
    ),
    LaAccentHue.pink: LaAccentTriple(
      primary: Color(0xFFD685A8),
      secondary: Color(0xFFB86C8D),
      tertiary: Color(0xFF9C5374),
    ),
    LaAccentHue.gray: LaAccentTriple(
      primary: Color(0xFF9C9CB2),
      secondary: Color(0xFF828297),
      tertiary: Color(0xFF69697D),
    ),
  };
}
