import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

/// The email body, collapsed to two lines by default. Tapping the "..."
/// reveals the rest -- keeps the card short without hiding the source
/// message away behind a separate screen.
class EmailPreview extends StatefulWidget {
  final String excerpt;

  const EmailPreview({super.key, required this.excerpt});

  @override
  State<EmailPreview> createState() => _EmailPreviewState();
}

class _EmailPreviewState extends State<EmailPreview> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    // Collapsed view reads as a flowing snippet; expanded view keeps the
    // sender's original line breaks.
    final collapsedText = widget.excerpt
        .replaceAll(RegExp(r'\s*\n+\s*'), ' ')
        .trim();
    return InkWell(
      borderRadius: BorderRadius.circular(LaRadius.md),
      hoverColor: colors.backgroundBrandTertiaryHover,
      splashColor: colors.backgroundBrandTertiaryHover,
      onTap: () => setState(() => _expanded = !_expanded),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(LaSpacing.sm),
        decoration: BoxDecoration(
          color: colors.backgroundNeutralTertiary,
          borderRadius: BorderRadius.circular(LaRadius.md),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Text(
                _expanded ? widget.excerpt : collapsedText,
                maxLines: _expanded ? null : 2,
                overflow: _expanded
                    ? TextOverflow.visible
                    : TextOverflow.ellipsis,
                style: LaTextStyles.bodySm.copyWith(
                  color: colors.contentNeutralSecondary,
                ),
              ),
            ),
            const SizedBox(width: LaSpacing.x2xs),
            Icon(
              _expanded ? Icons.expand_less : Icons.more_horiz,
              size: 18,
              color: colors.contentNeutralTertiary,
            ),
          ],
        ),
      ),
    );
  }
}
