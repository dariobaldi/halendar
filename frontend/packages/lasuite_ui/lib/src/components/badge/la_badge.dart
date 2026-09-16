import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';
import '../../utils/la_semantic_palette.dart';

/// Badge color/semantic type, mirroring the upstream `BadgeType`.
enum LaBadgeType { accent, neutral, danger, success, warning, info }

/// A small pill-shaped status/category label.
class LaBadge extends StatelessWidget {
  const LaBadge({
    super.key,
    required this.label,
    this.type = LaBadgeType.accent,
    this.uppercased = false,
  });

  final String label;
  final LaBadgeType type;
  final bool uppercased;

  @override
  Widget build(BuildContext context) {
    final palette = LaSemanticPalette.of(context.laColors, _category);
    return DecoratedBox(
      decoration: BoxDecoration(
        color: palette.backgroundSecondary,
        borderRadius: BorderRadius.circular(LaRadius.lg),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(
          horizontal: LaSpacing.xs,
          vertical: LaSpacing.x2xs,
        ),
        child: Text(
          uppercased ? label.toUpperCase() : label,
          style: TextStyle(
            color: palette.contentSecondary,
            fontSize: LaFontSize.xs,
            fontWeight: LaFontWeight.medium,
            fontFamily: LaFontFamily.base,
            fontFamilyFallback: LaFontFamily.fallback,
            height: 1.3,
          ),
        ),
      ),
    );
  }

  LaSemanticCategory get _category => switch (type) {
    LaBadgeType.accent => LaSemanticCategory.brand,
    LaBadgeType.neutral => LaSemanticCategory.neutral,
    LaBadgeType.danger => LaSemanticCategory.error,
    LaBadgeType.success => LaSemanticCategory.success,
    LaBadgeType.warning => LaSemanticCategory.warning,
    LaBadgeType.info => LaSemanticCategory.info,
  };
}
