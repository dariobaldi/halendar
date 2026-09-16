import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';
import '../button/la_button.dart';

/// Modal width preset, mirroring the upstream `ModalSize` enum.
enum LaModalSize { small, medium, large, extraLarge, full }

double? _widthFor(LaModalSize size, double screenWidth) => switch (size) {
  LaModalSize.small => 342,
  LaModalSize.medium => 600,
  LaModalSize.large => 800,
  LaModalSize.extraLarge => screenWidth * 0.75,
  LaModalSize.full => null,
};

/// Shows a [LaModal] as a dialog route and returns the value passed to
/// `Navigator.pop`.
Future<T?> showLaModal<T>({
  required BuildContext context,
  String? title,
  required WidgetBuilder builder,
  LaModalSize size = LaModalSize.medium,
  List<Widget>? rightActions,
  List<Widget>? leftActions,
  bool barrierDismissible = true,
}) {
  return showDialog<T>(
    context: context,
    barrierDismissible: barrierDismissible,
    barrierColor: const Color(0x99000000),
    builder: (dialogContext) => LaModal(
      title: title,
      size: size,
      rightActions: rightActions,
      leftActions: leftActions,
      onClose: () => Navigator.of(dialogContext).pop(),
      child: builder(dialogContext),
    ),
  );
}

/// A dialog surface with a title bar, scrollable body and an optional
/// left/right-aligned action row — matching the upstream Modal's default
/// layout. Prefer [showLaModal] to present it as a route.
class LaModal extends StatelessWidget {
  const LaModal({
    super.key,
    this.title,
    required this.child,
    this.size = LaModalSize.medium,
    this.leftActions,
    this.rightActions,
    this.onClose,
  });

  final String? title;
  final Widget child;
  final LaModalSize size;
  final List<Widget>? leftActions;
  final List<Widget>? rightActions;
  final VoidCallback? onClose;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final screenSize = MediaQuery.sizeOf(context);
    final width = _widthFor(size, screenSize.width);
    final isFull = size == LaModalSize.full;

    final hasActions =
        (leftActions?.isNotEmpty ?? false) ||
        (rightActions?.isNotEmpty ?? false);

    return Dialog(
      insetPadding: isFull ? EdgeInsets.zero : const EdgeInsets.all(24),
      backgroundColor: colors.surfaceSecondary,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(isFull ? 0 : LaRadius.md),
      ),
      child: SizedBox(
        width: width,
        height: isFull ? screenSize.height : null,
        child: Column(
          mainAxisSize: isFull ? MainAxisSize.max : MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (title != null || onClose != null)
              Padding(
                padding: const EdgeInsets.fromLTRB(
                  LaSpacing.md,
                  LaSpacing.sm,
                  LaSpacing.sm,
                  LaSpacing.sm,
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        title ?? '',
                        style: TextStyle(
                          color: colors.contentNeutralPrimary,
                          fontSize: LaFontSize.lg,
                          fontWeight: LaFontWeight.bold,
                          fontFamily: LaFontFamily.base,
                          fontFamilyFallback: LaFontFamily.fallback,
                        ),
                      ),
                    ),
                    if (onClose != null)
                      LaButton(
                        icon: const Icon(Icons.close),
                        variant: LaButtonVariant.tertiary,
                        color: LaButtonColor.neutral,
                        size: LaButtonSize.small,
                        semanticLabel: 'Close',
                        onPressed: onClose,
                      ),
                  ],
                ),
              ),
            Flexible(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: LaSpacing.md),
                child: isFull ? SingleChildScrollView(child: child) : child,
              ),
            ),
            const SizedBox(height: LaSpacing.md),
            if (hasActions)
              Padding(
                padding: const EdgeInsets.fromLTRB(
                  LaSpacing.md,
                  LaSpacing.sm,
                  LaSpacing.md,
                  LaSpacing.md,
                ),
                child: Row(
                  children: [
                    ...?leftActions,
                    const Spacer(),
                    for (final action in rightActions ?? [])
                      Padding(
                        padding: const EdgeInsets.only(left: LaSpacing.x2xs),
                        child: action,
                      ),
                  ],
                ),
              ),
          ],
        ),
      ),
    );
  }
}
