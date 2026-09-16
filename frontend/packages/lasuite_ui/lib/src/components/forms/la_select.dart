import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';
import 'la_field_types.dart';
import 'la_input_decoration.dart';

/// One option in a [LaSelect].
class LaSelectItem<T> {
  const LaSelectItem({required this.value, required this.label, this.icon});

  final T value;
  final String label;
  final IconData? icon;
}

/// A single-select dropdown, styled to match [LaTextField].
///
/// This is a simplified, non-searchable single-select — the upstream kit's
/// searchable and multi-select variants aren't reproduced (see README).
class LaSelect<T> extends StatelessWidget {
  const LaSelect({
    super.key,
    required this.items,
    required this.value,
    required this.onChanged,
    this.label,
    this.variant = LaFieldVariant.floating,
    this.placeholder,
    this.state = LaFieldState.normal,
    this.helperText,
    this.enabled = true,
  });

  final List<LaSelectItem<T>> items;
  final T? value;
  final ValueChanged<T?>? onChanged;
  final String? label;
  final LaFieldVariant variant;
  final String? placeholder;
  final LaFieldState state;
  final String? helperText;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;

    final dropdown = DropdownButtonHideUnderline(
      child: DropdownButton<T>(
        value: value,
        isDense: true,
        isExpanded: true,
        icon: Icon(
          Icons.expand_more,
          color: colors.contentNeutralTertiary,
          size: 20,
        ),
        hint: variant == LaFieldVariant.floating
            ? null
            : Text(placeholder ?? ''),
        onChanged: enabled ? onChanged : null,
        dropdownColor: colors.surfacePrimary,
        borderRadius: BorderRadius.circular(LaRadius.md),
        style: TextStyle(
          color: colors.contentNeutralPrimary,
          fontSize: LaFontSize.sm,
          fontFamily: LaFontFamily.base,
          fontFamilyFallback: LaFontFamily.fallback,
        ),
        items: [
          for (final item in items)
            DropdownMenuItem<T>(
              value: item.value,
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (item.icon != null) ...[
                    Icon(
                      item.icon,
                      size: 18,
                      color: colors.contentNeutralSecondary,
                    ),
                    const SizedBox(width: LaSpacing.x2xs),
                  ],
                  Text(item.label),
                ],
              ),
            ),
        ],
      ),
    );

    final field = InputDecorator(
      decoration: buildLaInputDecoration(
        colors: colors,
        variant: variant,
        state: state,
        label: label,
        placeholder: placeholder,
        helperText: helperText,
        filled: true,
      ),
      isEmpty: value == null,
      child: dropdown,
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
