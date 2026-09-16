import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';

/// One tab's identity, label and panel content.
class LaTabItem {
  const LaTabItem({
    required this.id,
    required this.label,
    this.icon,
    required this.content,
  });

  final String id;
  final String label;
  final IconData? icon;
  final Widget content;
}

/// A simple, self-contained tab strip + panel, selecting between
/// [LaTabItem.content] widgets.
class LaTabs extends StatefulWidget {
  const LaTabs({
    super.key,
    required this.tabs,
    this.initialTabId,
    this.fullWidth = false,
    this.onChanged,
  }) : assert(tabs.length > 0, 'LaTabs requires at least one tab.');

  final List<LaTabItem> tabs;
  final String? initialTabId;
  final bool fullWidth;
  final ValueChanged<String>? onChanged;

  @override
  State<LaTabs> createState() => _LaTabsState();
}

class _LaTabsState extends State<LaTabs> {
  late String _selectedId = widget.initialTabId ?? widget.tabs.first.id;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final selected = widget.tabs.firstWhere(
      (t) => t.id == _selectedId,
      orElse: () => widget.tabs.first,
    );

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        DecoratedBox(
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(color: colors.borderSurfacePrimary),
            ),
          ),
          child: Row(
            mainAxisSize: widget.fullWidth
                ? MainAxisSize.max
                : MainAxisSize.min,
            children: [
              for (final tab in widget.tabs)
                if (widget.fullWidth)
                  Expanded(
                    child: _TabButton(
                      tab: tab,
                      selected: tab.id == _selectedId,
                      onTap: () => _select(tab.id),
                    ),
                  )
                else
                  _TabButton(
                    tab: tab,
                    selected: tab.id == _selectedId,
                    onTap: () => _select(tab.id),
                  ),
            ],
          ),
        ),
        const SizedBox(height: LaSpacing.sm),
        selected.content,
      ],
    );
  }

  void _select(String id) {
    if (id == _selectedId) return;
    setState(() => _selectedId = id);
    widget.onChanged?.call(id);
  }
}

class _TabButton extends StatefulWidget {
  const _TabButton({
    required this.tab,
    required this.selected,
    required this.onTap,
  });

  final LaTabItem tab;
  final bool selected;
  final VoidCallback onTap;

  @override
  State<_TabButton> createState() => _TabButtonState();
}

class _TabButtonState extends State<_TabButton> {
  bool _hovered = false;

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final color = widget.selected
        ? colors.contentBrandPrimary
        : colors.contentNeutralSecondary;

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hovered = true),
      onExit: (_) => setState(() => _hovered = false),
      child: GestureDetector(
        onTap: widget.onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(
            horizontal: LaSpacing.sm,
            vertical: LaSpacing.x2xs,
          ),
          decoration: BoxDecoration(
            color: _hovered && !widget.selected
                ? colors.backgroundNeutralTertiary
                : null,
            border: Border(
              bottom: BorderSide(
                color: widget.selected
                    ? colors.borderBrandPrimary
                    : Colors.transparent,
                width: 2,
              ),
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (widget.tab.icon != null) ...[
                Icon(widget.tab.icon, size: 16, color: color),
                const SizedBox(width: LaSpacing.x3xs),
              ],
              Text(
                widget.tab.label,
                style: TextStyle(
                  color: color,
                  fontSize: LaFontSize.sm,
                  fontWeight: widget.selected
                      ? LaFontWeight.bold
                      : LaFontWeight.medium,
                  fontFamily: LaFontFamily.base,
                  fontFamilyFallback: LaFontFamily.fallback,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
