import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';

/// Separator stroke weight, mirroring the upstream `thin` / `double` widths.
enum LaSeparatorWidth { thin, double }

/// A horizontal rule, typically used between stacked sections.
class LaHorizontalSeparator extends StatelessWidget {
  const LaHorizontalSeparator({
    super.key,
    this.width = LaSeparatorWidth.thin,
    this.withPadding = true,
    this.color,
  });

  final LaSeparatorWidth width;
  final bool withPadding;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final line = Container(
      height: width == LaSeparatorWidth.thin ? 1 : 2,
      color: color ?? context.laColors.borderSurfacePrimary,
    );
    return withPadding
        ? Padding(padding: const EdgeInsets.symmetric(vertical: 8), child: line)
        : line;
  }
}

/// A vertical rule, typically used between inline items (e.g. toolbar
/// groups).
class LaVerticalSeparator extends StatelessWidget {
  const LaVerticalSeparator({
    super.key,
    this.width = LaSeparatorWidth.thin,
    this.withPadding = true,
    this.color,
  });

  final LaSeparatorWidth width;
  final bool withPadding;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final line = Container(
      width: width == LaSeparatorWidth.thin ? 1 : 2,
      color: color ?? context.laColors.borderSurfacePrimary,
    );
    return withPadding
        ? Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            child: line,
          )
        : line;
  }
}
