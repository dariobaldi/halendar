import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';

/// [LaLoader] size, mirroring the upstream `small` / `medium` scale.
enum LaLoaderSize { small, medium }

/// A spinning progress indicator, matching the La Suite numérique visual
/// language (a thin rotating arc) while delegating the animation to
/// Flutter's own [CircularProgressIndicator].
class LaLoader extends StatelessWidget {
  const LaLoader({
    super.key,
    this.size = LaLoaderSize.medium,
    this.color,
    this.semanticLabel,
  });

  final LaLoaderSize size;

  /// Defaults to the brand primary color.
  final Color? color;

  final String? semanticLabel;

  double get _dimension => switch (size) {
    LaLoaderSize.small => 16,
    LaLoaderSize.medium => 24,
  };

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: _dimension,
      width: _dimension,
      child: CircularProgressIndicator(
        strokeWidth: size == LaLoaderSize.small ? 2 : 2.5,
        valueColor: AlwaysStoppedAnimation<Color>(
          color ?? context.laColors.backgroundBrandPrimary,
        ),
        semanticsLabel: semanticLabel ?? 'Loading',
      ),
    );
  }
}
