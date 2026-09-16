import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';
import '../../utils/la_semantic_palette.dart';
import '../../utils/la_variant.dart';
import '../button/la_button.dart';

IconData _iconFor(LaVariant type) => switch (type) {
  LaVariant.info => Icons.info_outline,
  LaVariant.success => Icons.check_circle_outline,
  LaVariant.warning => Icons.error_outline,
  LaVariant.error => Icons.cancel_outlined,
  LaVariant.neutral => Icons.circle_outlined,
};

LaSemanticCategory _categoryFor(LaVariant type) => switch (type) {
  LaVariant.info => LaSemanticCategory.info,
  LaVariant.success => LaSemanticCategory.success,
  LaVariant.warning => LaSemanticCategory.warning,
  LaVariant.error => LaSemanticCategory.error,
  LaVariant.neutral => LaSemanticCategory.neutral,
};

/// An inline banner used to surface contextual feedback, with an accent
/// left border and background tinted by [type].
///
/// Supports an optional expandable "additional" section, up to two action
/// buttons, and a dismiss control — mirroring the upstream Alert's
/// one-line / additional / additional-expandable variants.
class LaAlert extends StatefulWidget {
  const LaAlert({
    super.key,
    required this.message,
    this.type = LaVariant.info,
    this.icon,
    this.additional,
    this.expandable = false,
    this.canClose = false,
    this.onClose,
    this.primaryLabel,
    this.onPrimaryPressed,
    this.secondaryLabel,
    this.onSecondaryPressed,
  });

  final String message;
  final LaVariant type;

  /// Overrides the default type icon. Pass `SizedBox.shrink()` to hide it.
  final Widget? icon;

  /// Extra content shown below [message] (collapsed behind a toggle when
  /// [expandable] is true).
  final String? additional;
  final bool expandable;

  final bool canClose;
  final VoidCallback? onClose;

  final String? primaryLabel;
  final VoidCallback? onPrimaryPressed;
  final String? secondaryLabel;
  final VoidCallback? onSecondaryPressed;

  @override
  State<LaAlert> createState() => _LaAlertState();
}

class _LaAlertState extends State<LaAlert> {
  bool _closed = false;
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    if (_closed) return const SizedBox.shrink();

    final palette = LaSemanticPalette.of(
      context.laColors,
      _categoryFor(widget.type),
    );
    final hasActions =
        widget.primaryLabel != null || widget.secondaryLabel != null;

    return DecoratedBox(
      decoration: BoxDecoration(
        color: palette.backgroundTertiary,
        borderRadius: BorderRadius.circular(LaRadius.sm),
        border: Border(
          left: BorderSide(color: palette.borderPrimary, width: 3),
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(LaSpacing.sm),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (widget.type != LaVariant.neutral)
              Padding(
                padding: const EdgeInsets.only(right: LaSpacing.xs, top: 1),
                child:
                    widget.icon ??
                    Icon(
                      _iconFor(widget.type),
                      size: 19,
                      color: palette.borderPrimary,
                    ),
              ),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    widget.message,
                    style: TextStyle(
                      color: context.laColors.contentNeutralPrimary,
                      fontSize: LaFontSize.xs,
                      fontWeight: LaFontWeight.medium,
                      fontFamily: LaFontFamily.base,
                      fontFamilyFallback: LaFontFamily.fallback,
                    ),
                  ),
                  if (widget.additional != null &&
                      (!widget.expandable || _expanded)) ...[
                    const SizedBox(height: LaSpacing.x3xs),
                    Text(
                      widget.additional!,
                      style: TextStyle(
                        color: context.laColors.contentNeutralPrimary,
                        fontSize: 11,
                        fontWeight: LaFontWeight.regular,
                        fontFamily: LaFontFamily.base,
                        fontFamilyFallback: LaFontFamily.fallback,
                      ),
                    ),
                  ],
                  if (widget.additional != null && widget.expandable)
                    Padding(
                      padding: const EdgeInsets.only(top: LaSpacing.x3xs),
                      child: GestureDetector(
                        onTap: () => setState(() => _expanded = !_expanded),
                        child: Text(
                          _expanded ? 'Show less' : 'Show more',
                          style: TextStyle(
                            color: palette.contentTertiary,
                            fontSize: LaFontSize.xs,
                            fontWeight: LaFontWeight.medium,
                            decoration: TextDecoration.underline,
                          ),
                        ),
                      ),
                    ),
                  if (hasActions)
                    Padding(
                      padding: const EdgeInsets.only(top: LaSpacing.xs),
                      child: Wrap(
                        spacing: LaSpacing.x2xs,
                        children: [
                          if (widget.primaryLabel != null)
                            LaButton(
                              label: widget.primaryLabel,
                              size: LaButtonSize.nano,
                              color: _categoryFor(widget.type),
                              onPressed: widget.onPrimaryPressed,
                            ),
                          if (widget.secondaryLabel != null)
                            LaButton(
                              label: widget.secondaryLabel,
                              size: LaButtonSize.nano,
                              variant: LaButtonVariant.tertiary,
                              color: _categoryFor(widget.type),
                              onPressed: widget.onSecondaryPressed,
                            ),
                        ],
                      ),
                    ),
                ],
              ),
            ),
            if (widget.canClose)
              InkResponse(
                onTap: () {
                  setState(() => _closed = true);
                  widget.onClose?.call();
                },
                radius: 16,
                child: Icon(
                  Icons.close,
                  size: 16,
                  color: context.laColors.contentNeutralSecondary,
                ),
              ),
          ],
        ),
      ),
    );
  }
}
