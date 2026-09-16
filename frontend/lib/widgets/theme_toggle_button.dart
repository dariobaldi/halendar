import 'package:flutter/material.dart';

import '../state/theme_controller.dart';

/// Light/dark mode toggle, meant to sit in the top-right corner of every
/// screen's app bar.
class ThemeToggleButton extends StatelessWidget {
  const ThemeToggleButton({super.key});

  @override
  Widget build(BuildContext context) {
    return ValueListenableBuilder<ThemeMode>(
      valueListenable: ThemeController.instance,
      builder: (context, mode, _) {
        final isDark = mode == ThemeMode.dark;
        // AppBar's default title sits 16px from the left edge
        // (titleSpacing). IconButton's own 48x48 tap target already
        // centers its 24px icon with a 12px inherent gap from the trailing
        // edge, so 4px more here mirrors that 16px exactly without
        // shrinking the tap target.
        return Padding(
          padding: const EdgeInsets.only(right: 4),
          child: IconButton(
            tooltip: isDark ? 'Switch to light mode' : 'Switch to dark mode',
            icon: Icon(
              isDark ? Icons.light_mode_outlined : Icons.dark_mode_outlined,
            ),
            onPressed: ThemeController.instance.toggle,
          ),
        );
      },
    );
  }
}
