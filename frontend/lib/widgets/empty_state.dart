import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

/// Centered icon + title + message, used by any screen whose collection
/// (proposals, history, ...) is currently empty.
class EmptyState extends StatelessWidget {
  final IconData icon;
  final String title;
  final String message;

  const EmptyState({
    super.key,
    required this.icon,
    required this.title,
    required this.message,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(LaSpacing.lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 48, color: colors.contentNeutralTertiary),
            const SizedBox(height: LaSpacing.sm),
            Text(
              title,
              style: LaTextStyles.labelLg.copyWith(
                color: colors.contentNeutralPrimary,
              ),
            ),
            const SizedBox(height: LaSpacing.x2xs),
            Text(
              message,
              textAlign: TextAlign.center,
              style: LaTextStyles.bodySm.copyWith(
                color: colors.contentNeutralTertiary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
