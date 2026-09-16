import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';
import '../../utils/la_semantic_palette.dart';
import '../../utils/la_variant.dart';

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

/// A single toast notification's visual content. Normally shown through
/// [LaToastMessenger.show] rather than instantiated directly.
class LaToast extends StatelessWidget {
  const LaToast({
    super.key,
    required this.message,
    this.type = LaVariant.info,
    this.onDismiss,
    this.action,
  });

  final String message;
  final LaVariant type;
  final VoidCallback? onDismiss;
  final Widget? action;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final iconColor = LaSemanticPalette.of(
      colors,
      _categoryFor(type),
    ).borderPrimary;

    return Material(
      color: Colors.transparent,
      child: Container(
        constraints: const BoxConstraints(minWidth: 280, maxWidth: 420),
        padding: const EdgeInsets.symmetric(
          horizontal: LaSpacing.sm,
          vertical: LaSpacing.sm,
        ),
        decoration: BoxDecoration(
          color: colors.backgroundNeutralTertiary,
          borderRadius: BorderRadius.circular(LaRadius.md),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.15),
              blurRadius: 12,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (type != LaVariant.neutral) ...[
              Icon(_iconFor(type), size: 19, color: iconColor),
              const SizedBox(width: LaSpacing.xs),
            ],
            Flexible(
              child: Text(
                message,
                style: TextStyle(
                  color: colors.contentNeutralPrimary,
                  fontSize: LaFontSize.sm,
                  fontFamily: LaFontFamily.base,
                  fontFamilyFallback: LaFontFamily.fallback,
                ),
              ),
            ),
            if (action != null) ...[
              const SizedBox(width: LaSpacing.xs),
              action!,
            ],
            if (onDismiss != null) ...[
              const SizedBox(width: LaSpacing.xs),
              InkResponse(
                onTap: onDismiss,
                radius: 14,
                child: Icon(
                  Icons.close,
                  size: 16,
                  color: colors.contentNeutralSecondary,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

/// Shows [LaToast]s in a top-of-screen [Overlay] stack, auto-dismissing
/// after [duration].
abstract final class LaToastMessenger {
  static void show(
    BuildContext context, {
    required String message,
    LaVariant type = LaVariant.info,
    Duration duration = const Duration(seconds: 4),
    Widget? action,
  }) {
    final overlay = Overlay.of(context, rootOverlay: true);
    late OverlayEntry entry;
    entry = OverlayEntry(
      builder: (_) => _ToastSlot(
        message: message,
        type: type,
        duration: duration,
        action: action,
        onRemove: () => entry.remove(),
      ),
    );
    overlay.insert(entry);
  }
}

class _ToastSlot extends StatefulWidget {
  const _ToastSlot({
    required this.message,
    required this.type,
    required this.duration,
    required this.onRemove,
    this.action,
  });

  final String message;
  final LaVariant type;
  final Duration duration;
  final Widget? action;
  final VoidCallback onRemove;

  @override
  State<_ToastSlot> createState() => _ToastSlotState();
}

class _ToastSlotState extends State<_ToastSlot>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller = AnimationController(
    vsync: this,
    duration: LaMotion.toastSlideOut,
  );
  late final Animation<double> _fade = CurvedAnimation(
    parent: _controller,
    curve: LaMotion.easeOut,
  );

  @override
  void initState() {
    super.initState();
    _controller.forward();
    Future.delayed(widget.duration, _dismiss);
  }

  Future<void> _dismiss() async {
    if (!mounted) return;
    await _controller.reverse();
    widget.onRemove();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Positioned(
      top: MediaQuery.paddingOf(context).top + LaSpacing.sm,
      left: 0,
      right: 0,
      child: Center(
        child: FadeTransition(
          opacity: _fade,
          child: SlideTransition(
            position: Tween<Offset>(
              begin: const Offset(0, -0.3),
              end: Offset.zero,
            ).animate(_fade),
            child: LaToast(
              message: widget.message,
              type: widget.type,
              action: widget.action,
              onDismiss: _dismiss,
            ),
          ),
        ),
      ),
    );
  }
}
