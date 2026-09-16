import 'package:flutter/material.dart';

import '../../tokens/tokens.dart';
import 'la_field_types.dart';

/// Builds the shared [InputDecoration] used by [LaTextField] and
/// [LaTextArea] for a given [LaFieldVariant]/[LaFieldState], centralizing
/// the border/color/radius rules so both controls stay visually consistent.
InputDecoration buildLaInputDecoration({
  required LaColors colors,
  required LaFieldVariant variant,
  required LaFieldState state,
  String? label,
  String? placeholder,
  String? helperText,
  Widget? icon,
  Widget? suffixIcon,
  bool filled = true,
}) {
  final isFloating = variant == LaFieldVariant.floating;

  Color enabledBorderColor = colors.borderNeutralTertiary;
  Color focusedBorderColor = colors.borderBrandPrimary;
  Color? helperColor = colors.contentNeutralSecondary;

  switch (state) {
    case LaFieldState.error:
      enabledBorderColor = colors.borderErrorPrimary;
      focusedBorderColor = colors.borderErrorPrimary;
      helperColor = colors.contentErrorSecondary;
    case LaFieldState.success:
      enabledBorderColor = colors.borderSuccessPrimary;
      focusedBorderColor = colors.borderSuccessPrimary;
      helperColor = colors.contentSuccessSecondary;
    case LaFieldState.normal:
      break;
  }

  OutlineInputBorder border(Color color, double radius) => OutlineInputBorder(
    borderRadius: BorderRadius.circular(radius),
    borderSide: BorderSide(color: color),
  );

  return InputDecoration(
    isDense: true,
    filled: filled,
    fillColor: colors.surfaceSecondary,
    labelText: isFloating ? label : null,
    hintText: isFloating ? null : placeholder,
    helperText: helperText,
    helperStyle: TextStyle(color: helperColor, fontSize: LaFontSize.xs),
    helperMaxLines: 3,
    floatingLabelBehavior: FloatingLabelBehavior.auto,
    labelStyle: TextStyle(
      color: colors.contentNeutralTertiary,
      fontSize: LaFontSize.sm,
    ),
    floatingLabelStyle: TextStyle(
      color: colors.contentBrandPrimary,
      fontSize: LaFontSize.sm,
    ),
    hintStyle: TextStyle(
      color: colors.contentNeutralTertiary,
      fontSize: LaFontSize.sm,
    ),
    prefixIcon: icon == null
        ? null
        : Padding(
            padding: const EdgeInsets.only(left: LaSpacing.xs),
            child: icon,
          ),
    prefixIconConstraints: const BoxConstraints(minWidth: 40, minHeight: 24),
    suffixIcon: suffixIcon,
    contentPadding: const EdgeInsets.symmetric(
      horizontal: LaSpacing.sm,
      vertical: LaSpacing.sm,
    ),
    border: border(enabledBorderColor, LaRadius.md),
    enabledBorder: border(enabledBorderColor, LaRadius.md),
    focusedBorder: border(focusedBorderColor, LaRadius.sm),
    disabledBorder: border(colors.borderDisabledPrimary, LaRadius.md),
    errorBorder: border(colors.borderErrorPrimary, LaRadius.md),
    focusedErrorBorder: border(colors.borderErrorPrimary, LaRadius.sm),
  );
}
