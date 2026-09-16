import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';

/// Page navigation control with numbered pages (collapsed with an ellipsis
/// for large page counts) and an optional "go to page" field.
class LaPagination extends StatefulWidget {
  const LaPagination({
    super.key,
    required this.page,
    required this.pagesCount,
    required this.onPageChange,
    this.displayGoto = false,
  });

  /// Current page, 1-indexed.
  final int page;
  final int pagesCount;
  final ValueChanged<int> onPageChange;
  final bool displayGoto;

  @override
  State<LaPagination> createState() => _LaPaginationState();
}

class _LaPaginationState extends State<LaPagination> {
  late final TextEditingController _gotoController = TextEditingController(
    text: '${widget.page}',
  );

  @override
  void didUpdateWidget(covariant LaPagination oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.page != widget.page) {
      _gotoController.text = '${widget.page}';
    }
  }

  @override
  void dispose() {
    _gotoController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (widget.pagesCount <= 1) return const SizedBox.shrink();

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        _NavButton(
          icon: Icons.chevron_left,
          onTap: widget.page > 1
              ? () => widget.onPageChange(widget.page - 1)
              : null,
        ),
        for (final entry in _pagesToDisplay(widget.page, widget.pagesCount))
          entry == null
              ? const Padding(
                  padding: EdgeInsets.symmetric(horizontal: 4),
                  child: Text('…'),
                )
              : _PageButton(
                  page: entry,
                  selected: entry == widget.page,
                  onTap: () => widget.onPageChange(entry),
                ),
        _NavButton(
          icon: Icons.chevron_right,
          onTap: widget.page < widget.pagesCount
              ? () => widget.onPageChange(widget.page + 1)
              : null,
        ),
        if (widget.displayGoto) ...[
          const SizedBox(width: LaSpacing.sm),
          SizedBox(
            width: 48,
            height: 32,
            child: TextField(
              controller: _gotoController,
              keyboardType: TextInputType.number,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: LaFontSize.sm),
              decoration: InputDecoration(
                isDense: true,
                contentPadding: const EdgeInsets.symmetric(vertical: 6),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(LaRadius.sm),
                  borderSide: BorderSide(
                    color: context.laColors.borderNeutralTertiary,
                  ),
                ),
              ),
              onSubmitted: (value) {
                final target = int.tryParse(value);
                if (target != null &&
                    target >= 1 &&
                    target <= widget.pagesCount) {
                  widget.onPageChange(target);
                } else {
                  _gotoController.text = '${widget.page}';
                }
              },
            ),
          ),
        ],
      ],
    );
  }

  /// Builds a windowed page list around [page] out of [count] pages, using
  /// `null` entries as ellipsis markers.
  static List<int?> _pagesToDisplay(int page, int count) {
    const siblings = 1;
    final pages = <int?>[];
    final start = (page - siblings).clamp(1, count);
    final end = (page + siblings).clamp(1, count);

    pages.add(1);
    if (start > 2) pages.add(null);
    for (var i = start; i <= end; i++) {
      if (i != 1 && i != count) pages.add(i);
    }
    if (end < count - 1) pages.add(null);
    if (count > 1) pages.add(count);
    return pages;
  }
}

class _NavButton extends StatelessWidget {
  const _NavButton({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final enabled = onTap != null;
    return IconButton(
      onPressed: onTap,
      icon: Icon(icon, size: 18),
      color: enabled
          ? context.laColors.contentNeutralPrimary
          : context.laColors.contentDisabledPrimary,
      splashRadius: 18,
    );
  }
}

class _PageButton extends StatelessWidget {
  const _PageButton({
    required this.page,
    required this.selected,
    required this.onTap,
  });

  final int page;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    return InkWell(
      onTap: onTap,
      customBorder: const CircleBorder(),
      child: Container(
        height: 32,
        width: 32,
        alignment: Alignment.center,
        margin: const EdgeInsets.symmetric(horizontal: 2),
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: selected ? colors.backgroundBrandPrimary : Colors.transparent,
        ),
        child: Text(
          '$page',
          style: TextStyle(
            color: selected
                ? colors.contentOnBrand
                : colors.contentNeutralPrimary,
            fontSize: LaFontSize.sm,
            fontWeight: selected ? LaFontWeight.bold : LaFontWeight.regular,
          ),
        ),
      ),
    );
  }
}
