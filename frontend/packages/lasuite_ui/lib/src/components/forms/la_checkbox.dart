import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';
import 'la_field_types.dart';

/// A checkbox with label, indeterminate state, and validation-state
/// support, matching the upstream Checkbox.
class LaCheckbox extends StatelessWidget {
  const LaCheckbox({
    super.key,
    required this.value,
    required this.onChanged,
    this.label,
    this.indeterminate = false,
    this.state = LaFieldState.normal,
  });

  final bool value;
  final ValueChanged<bool>? onChanged;
  final Widget? label;
  final bool indeterminate;
  final LaFieldState state;

  bool get _enabled => onChanged != null;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final filled = value || indeterminate;

    Color borderColor = colors.borderNeutralTertiary;
    if (state == LaFieldState.error) {
      borderColor = colors.borderErrorPrimary;
    }
    if (state == LaFieldState.success) {
      borderColor = colors.borderSuccessPrimary;
    }

    final boxColor = !_enabled
        ? (filled ? colors.backgroundDisabledPrimary : Colors.transparent)
        : (filled ? colors.backgroundBrandPrimary : Colors.transparent);

    final box = AnimatedContainer(
      duration: LaMotion.duration,
      curve: LaMotion.easeOut,
      height: 20,
      width: 20,
      // Safe here (unlike LaButton): both width and height are already
      // tight, so `alignment` centers the checkmark without expanding the
      // box to fill any extra bounded space.
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: boxColor,
        borderRadius: BorderRadius.circular(LaRadius.xs),
        border: filled
            ? null
            : Border.all(
                color: _enabled ? borderColor : colors.borderDisabledPrimary,
                width: 1.5,
              ),
      ),
      child: filled
          ? Icon(
              indeterminate ? Icons.remove : Icons.check,
              size: 14,
              color: colors.contentContextualPrimary,
            )
          : null,
    );

    return Semantics(
      checked: value,
      child: MouseRegion(
        cursor: _enabled ? SystemMouseCursors.click : SystemMouseCursors.basic,
        child: GestureDetector(
          onTap: _enabled ? () => onChanged!(!value) : null,
          child: Row(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              box,
              if (label != null) ...[
                const SizedBox(width: LaSpacing.x2xs),
                DefaultTextStyle.merge(
                  style: TextStyle(
                    color: _enabled
                        ? colors.contentNeutralPrimary
                        : colors.contentDisabledPrimary,
                    fontSize: LaFontSize.sm,
                    fontWeight: LaFontWeight.medium,
                    fontFamily: LaFontFamily.base,
                    fontFamilyFallback: LaFontFamily.fallback,
                  ),
                  child: label!,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
