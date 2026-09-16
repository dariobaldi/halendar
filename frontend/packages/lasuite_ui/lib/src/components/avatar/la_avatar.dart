import 'package:flutter/material.dart';

import '../../tokens/tokens.dart';

/// Avatar size preset.
enum LaAvatarSize { small, medium, large }

/// A circular initials avatar, colored from the accent palette.
///
/// Pass [hue] to pick a specific [LaAccentHue], or leave it unset to derive
/// a stable color from [name] (so the same person always gets the same
/// color).
class LaAvatar extends StatelessWidget {
  const LaAvatar({
    super.key,
    required this.name,
    this.size = LaAvatarSize.medium,
    this.hue,
    this.image,
  });

  final String name;
  final LaAvatarSize size;
  final LaAccentHue? hue;
  final ImageProvider? image;

  double get _dimension => switch (size) {
    LaAvatarSize.small => 24,
    LaAvatarSize.medium => 32,
    LaAvatarSize.large => 48,
  };

  double get _fontSize => switch (size) {
    LaAvatarSize.small => LaFontSize.t,
    LaAvatarSize.medium => LaFontSize.xs,
    LaAvatarSize.large => LaFontSize.md,
  };

  String get _initials {
    final parts = name
        .trim()
        .split(RegExp(r'\s+'))
        .where((p) => p.isNotEmpty)
        .toList();
    if (parts.isEmpty) return '?';
    if (parts.length == 1) return parts.first.substring(0, 1).toUpperCase();
    return (parts.first.substring(0, 1) + parts.last.substring(0, 1))
        .toUpperCase();
  }

  LaAccentHue get _resolvedHue =>
      hue ??
      LaAccentHue.values[name.hashCode.abs() % LaAccentHue.values.length];

  @override
  Widget build(BuildContext context) {
    final brightness = Theme.of(context).brightness;
    final triples = brightness == Brightness.dark
        ? LaAccentColors.dark
        : LaAccentColors.light;
    final triple = triples[_resolvedHue]!;

    return CircleAvatar(
      radius: _dimension / 2,
      backgroundColor: triple.tertiary,
      backgroundImage: image,
      child: image != null
          ? null
          : Text(
              _initials,
              style: TextStyle(
                color: triple.primary,
                fontSize: _fontSize,
                fontWeight: LaFontWeight.bold,
                fontFamily: LaFontFamily.base,
                fontFamilyFallback: LaFontFamily.fallback,
              ),
            ),
    );
  }
}
