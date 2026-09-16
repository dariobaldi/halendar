import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';

/// Preferred tooltip placement relative to its anchor.
///
/// Flutter's [Tooltip] primitive only supports vertical placement; unlike
/// the upstream Tooltip (which also supports `left`/`right`), this wraps
/// [Tooltip.preferBelow] and Flutter will still flip the side automatically
/// when there isn't enough room.
enum LaTooltipPlacement { top, bottom }

/// A hover/long-press tooltip styled to match the La Suite numérique kit.
class LaTooltip extends StatelessWidget {
  const LaTooltip({
    super.key,
    required this.message,
    required this.child,
    this.placement = LaTooltipPlacement.bottom,
  });

  final String message;
  final Widget child;
  final LaTooltipPlacement placement;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    return Tooltip(
      message: message,
      preferBelow: placement == LaTooltipPlacement.bottom,
      verticalOffset: 12,
      waitDuration: const Duration(milliseconds: 400),
      showDuration: LaMotion.tooltip,
      padding: const EdgeInsets.symmetric(
        horizontal: LaSpacing.xs,
        vertical: LaSpacing.x3xs,
      ),
      decoration: BoxDecoration(
        color: colors.backgroundNeutralTertiary,
        borderRadius: BorderRadius.circular(LaRadius.md),
        border: Border.all(color: colors.borderNeutralTertiary),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.15),
            blurRadius: 5.4,
          ),
        ],
      ),
      textStyle: TextStyle(
        color: colors.contentNeutralTertiary,
        fontSize: LaFontSize.xs,
        fontFamily: LaFontFamily.base,
        fontFamilyFallback: LaFontFamily.fallback,
      ),
      child: child,
    );
  }
}
