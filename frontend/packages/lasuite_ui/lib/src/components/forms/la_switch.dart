import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';

/// Which side of the track [LaSwitch.label] is rendered on.
enum LaLabelSide { left, right }

/// A toggle switch with an optional label.
class LaSwitch extends StatelessWidget {
  const LaSwitch({
    super.key,
    required this.value,
    required this.onChanged,
    this.label,
    this.labelSide = LaLabelSide.left,
  });

  final bool value;
  final ValueChanged<bool>? onChanged;
  final String? label;
  final LaLabelSide labelSide;

  bool get _enabled => onChanged != null;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;

    final trackColor = !_enabled
        ? (value
              ? colors.backgroundDisabledPrimary
              : colors.backgroundDisabledSecondary)
        : (value
              ? colors.backgroundBrandPrimary
              : colors.backgroundNeutralTertiary);
    final handleColor = !_enabled
        ? colors.surfaceTertiary
        : colors.surfacePrimary;

    final track = MouseRegion(
      cursor: _enabled ? SystemMouseCursors.click : SystemMouseCursors.basic,
      child: GestureDetector(
        onTap: _enabled ? () => onChanged!(!value) : null,
        child: AnimatedContainer(
          duration: LaMotion.duration,
          curve: LaMotion.easeOut,
          height: 22,
          width: 38,
          padding: const EdgeInsets.all(2),
          decoration: BoxDecoration(
            color: trackColor,
            borderRadius: BorderRadius.circular(LaRadius.full),
          ),
          alignment: value ? Alignment.centerRight : Alignment.centerLeft,
          child: Container(
            height: 18,
            width: 18,
            decoration: BoxDecoration(
              color: handleColor,
              shape: BoxShape.circle,
            ),
          ),
        ),
      ),
    );

    if (label == null) return track;

    final text = Text(
      label!,
      style: TextStyle(
        color: _enabled
            ? colors.contentNeutralPrimary
            : colors.contentDisabledPrimary,
        fontSize: LaFontSize.sm,
        fontWeight: LaFontWeight.medium,
        fontFamily: LaFontFamily.base,
        fontFamilyFallback: LaFontFamily.fallback,
      ),
    );

    final children = labelSide == LaLabelSide.left
        ? [text, const SizedBox(width: LaSpacing.x2xs), track]
        : [track, const SizedBox(width: LaSpacing.x2xs), text];

    return Row(mainAxisSize: MainAxisSize.min, children: children);
  }
}
