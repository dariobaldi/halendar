import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';
import 'la_field_types.dart';
import 'la_input_decoration.dart';

export 'la_field_types.dart' show LaFieldState, LaFieldVariant;

/// A single-line text input, matching the upstream Input's floating /
/// classic / inline label variants and validation states.
class LaTextField extends StatelessWidget {
  const LaTextField({
    super.key,
    this.controller,
    this.label,
    this.labelDescription,
    this.variant = LaFieldVariant.floating,
    this.placeholder,
    this.icon,
    this.rightIcon,
    this.charCounter = false,
    this.charCounterMax,
    this.state = LaFieldState.normal,
    this.helperText,
    this.enabled = true,
    this.obscureText = false,
    this.keyboardType,
    this.textInputAction,
    this.onChanged,
    this.onSubmitted,
    this.focusNode,
    this.autofocus = false,
  });

  final TextEditingController? controller;
  final String? label;

  /// Secondary text under the label. Only shown by the "classic"/"inline"
  /// variants — the floating label has no room for a second line.
  final String? labelDescription;
  final LaFieldVariant variant;
  final String? placeholder;
  final Widget? icon;
  final Widget? rightIcon;
  final bool charCounter;
  final int? charCounterMax;
  final LaFieldState state;
  final String? helperText;
  final bool enabled;
  final bool obscureText;
  final TextInputType? keyboardType;
  final TextInputAction? textInputAction;
  final ValueChanged<String>? onChanged;
  final ValueChanged<String>? onSubmitted;
  final FocusNode? focusNode;
  final bool autofocus;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;

    final field = TextField(
      controller: controller,
      focusNode: focusNode,
      autofocus: autofocus,
      enabled: enabled,
      obscureText: obscureText,
      keyboardType: keyboardType,
      textInputAction: textInputAction,
      onChanged: onChanged,
      onSubmitted: onSubmitted,
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
        icon: icon,
        suffixIcon: rightIcon,
      ),
    );

    if (variant == LaFieldVariant.floating) return field;

    final labelRow = label == null
        ? null
        : Padding(
            padding: const EdgeInsets.only(bottom: LaSpacing.x2xs),
            child: Text(
              label!,
              style: TextStyle(
                color: colors.contentNeutralPrimary,
                fontSize: LaFontSize.sm,
                fontWeight: LaFontWeight.medium,
                fontFamily: LaFontFamily.base,
                fontFamilyFallback: LaFontFamily.fallback,
              ),
            ),
          );

    final descriptionRow = labelDescription == null
        ? null
        : Padding(
            padding: const EdgeInsets.only(bottom: LaSpacing.x2xs),
            child: Text(
              labelDescription!,
              style: TextStyle(
                color: colors.contentNeutralSecondary,
                fontSize: LaFontSize.xs,
              ),
            ),
          );

    if (variant == LaFieldVariant.classic) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [?labelRow, ?descriptionRow, field],
      );
    }

    // inline
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 140,
          child: Padding(
            padding: const EdgeInsets.only(top: LaSpacing.sm),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [?labelRow, ?descriptionRow],
            ),
          ),
        ),
        const SizedBox(width: LaSpacing.xs),
        Expanded(child: field),
      ],
    );
  }
}
