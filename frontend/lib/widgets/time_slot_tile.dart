import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

import '../models/time_slot.dart';
import '../utils/date_format.dart';

/// A single row inside the flat slot list. Free, selectable slots respond
/// to [onTap]; busy slots (or any slot when [onTap] is null) render as a
/// plain, dimmed, non-interactive row.
///
/// Built from a plain `Material` + `InkWell` instead of `RadioListTile`:
/// that widget only forwards `hoverColor` to its tiny radio control, not
/// to the row itself, which left the whole-row hover as a flat, sharp
/// edged, default-grey rectangle. Here the hover/selected fill is an
/// explicit rounded, softly tinted background -- consistent with the
/// email and reply boxes elsewhere on the card.
class TimeSlotRow extends StatelessWidget {
  final TimeSlot slot;
  final bool isSelected;
  final VoidCallback? onTap;

  const TimeSlotRow({
    super.key,
    required this.slot,
    this.isSelected = false,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final interactive = onTap != null && slot.isFree;
    final borderRadius = BorderRadius.circular(LaRadius.md);

    final icon = isSelected
        ? Icons.check_circle
        : slot.isFree
        ? Icons.radio_button_unchecked
        : Icons.remove_circle_outline;
    final iconColor = isSelected
        ? colors.contentSuccessTertiary
        : colors.contentNeutralTertiary;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: LaSpacing.x4xs),
      child: Semantics(
        button: interactive,
        selected: isSelected,
        child: Material(
          color: isSelected
              ? colors.backgroundBrandTertiary
              : Colors.transparent,
          borderRadius: borderRadius,
          clipBehavior: Clip.antiAlias,
          child: InkWell(
            onTap: interactive ? onTap : null,
            hoverColor: colors.backgroundBrandTertiaryHover,
            splashColor: colors.backgroundBrandSecondary,
            child: Opacity(
              opacity: slot.isFree ? 1 : 0.5,
              child: Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: LaSpacing.sm,
                  vertical: LaSpacing.sm,
                ),
                child: Row(
                  children: [
                    Icon(icon, color: iconColor, size: 22),
                    const SizedBox(width: LaSpacing.sm),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          formatDay(slot.start),
                          style: LaTextStyles.labelMd.copyWith(
                            color: colors.contentNeutralPrimary,
                          ),
                        ),
                        const SizedBox(height: LaSpacing.x4xs),
                        Text(
                          slot.isFree
                              ? formatTimeRange(slot.start, slot.end)
                              : '${formatTimeRange(slot.start, slot.end)} · Busy',
                          style: LaTextStyles.bodySm.copyWith(
                            color: colors.contentNeutralTertiary,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
