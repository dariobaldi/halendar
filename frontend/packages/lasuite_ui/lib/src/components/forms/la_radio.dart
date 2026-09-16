import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';

/// A single radio button + label. Typically used inside [LaRadioGroup].
class LaRadio<T> extends StatelessWidget {
  const LaRadio({
    super.key,
    required this.value,
    required this.groupValue,
    required this.onChanged,
    this.label,
  });

  final T value;
  final T? groupValue;
  final ValueChanged<T>? onChanged;
  final Widget? label;

  bool get _enabled => onChanged != null;
  bool get _selected => value == groupValue;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final borderColor = _enabled
        ? colors.borderNeutralTertiary
        : colors.borderDisabledPrimary;
    final dotColor = _enabled
        ? colors.backgroundBrandPrimary
        : colors.backgroundDisabledPrimary;

    return Semantics(
      inMutuallyExclusiveGroup: true,
      checked: _selected,
      child: MouseRegion(
        cursor: _enabled ? SystemMouseCursors.click : SystemMouseCursors.basic,
        child: GestureDetector(
          onTap: _enabled ? () => onChanged!(value) : null,
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              AnimatedContainer(
                duration: LaMotion.duration,
                curve: LaMotion.easeOut,
                height: 20,
                width: 20,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: _selected ? dotColor : borderColor,
                    width: 1.5,
                  ),
                ),
                alignment: Alignment.center,
                child: AnimatedContainer(
                  duration: LaMotion.duration,
                  height: _selected ? 10 : 0,
                  width: _selected ? 10 : 0,
                  decoration: BoxDecoration(
                    color: dotColor,
                    shape: BoxShape.circle,
                  ),
                ),
              ),
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

/// A vertical group of [LaRadio] options sharing one selected value.
class LaRadioGroup<T> extends StatelessWidget {
  const LaRadioGroup({
    super.key,
    required this.options,
    required this.groupValue,
    required this.onChanged,
    this.labelBuilder,
    this.spacing = LaSpacing.x2xs,
  });

  final List<T> options;
  final T? groupValue;
  final ValueChanged<T>? onChanged;
  final Widget Function(T value)? labelBuilder;
  final double spacing;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (var i = 0; i < options.length; i++) ...[
          if (i > 0) SizedBox(height: spacing),
          LaRadio<T>(
            value: options[i],
            groupValue: groupValue,
            onChanged: onChanged,
            label: labelBuilder == null
                ? Text('${options[i]}')
                : labelBuilder!(options[i]),
          ),
        ],
      ],
    );
  }
}
