import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';
import 'la_field_types.dart';
import 'la_input_decoration.dart';

/// A multi-line text input sharing [LaTextField]'s label variants and
/// validation states.
class LaTextArea extends StatelessWidget {
  const LaTextArea({
    super.key,
    this.controller,
    this.label,
    this.variant = LaFieldVariant.floating,
    this.placeholder,
    this.charCounter = false,
    this.charCounterMax,
    this.state = LaFieldState.normal,
    this.helperText,
    this.enabled = true,
    this.minLines = 3,
    this.maxLines = 6,
    this.onChanged,
    this.focusNode,
  });

  final TextEditingController? controller;
  final String? label;
  final LaFieldVariant variant;
  final String? placeholder;
  final bool charCounter;
  final int? charCounterMax;
  final LaFieldState state;
  final String? helperText;
  final bool enabled;
  final int minLines;
  final int maxLines;
  final ValueChanged<String>? onChanged;
  final FocusNode? focusNode;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;

    final field = TextField(
      controller: controller,
      focusNode: focusNode,
      enabled: enabled,
      minLines: minLines,
      maxLines: maxLines,
      onChanged: onChanged,
      maxLength: charCounter ? charCounterMax : null,
      buildCounter: charCounter
          ? (
              context, {
              required currentLength,
              required isFocused,
              maxLength,
            }) => Text(
              maxLength != null
                  ? '$currentLength/$maxLength'
                  : '$currentLength',
              style: TextStyle(
                color: colors.contentNeutralTertiary,
                fontSize: LaFontSize.xs,
              ),
            )
          : (
              context, {
              required currentLength,
              required isFocused,
              maxLength,
            }) => null,
      style: TextStyle(
        color: colors.contentNeutralPrimary,
        fontSize: LaFontSize.sm,
        fontFamily: LaFontFamily.base,
        fontFamilyFallback: LaFontFamily.fallback,
      ),
      decoration: buildLaInputDecoration(
        colors: colors,
        variant: variant,
        state: state,
        label: label,
        placeholder: placeholder,
        helperText: helperText,
      ),
    );

    if (variant == LaFieldVariant.floating || label == null) return field;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(bottom: LaSpacing.x2xs),
          child: Text(
            label!,
            style: TextStyle(
              color: colors.contentNeutralPrimary,
              fontSize: LaFontSize.sm,
              fontWeight: LaFontWeight.medium,
            ),
          ),
        ),
        field,
      ],
    );
  }
}
