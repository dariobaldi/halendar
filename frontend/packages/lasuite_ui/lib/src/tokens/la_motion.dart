import 'package:flutter/animation.dart';

/// Motion tokens (design token `globals.transitions`): shared duration and
/// easing curves used across interactive components.
abstract final class LaMotion {
  static const Duration duration = Duration(milliseconds: 250);
  static const Duration toastSlideIn = Duration(milliseconds: 1000);
  static const Duration toastSlideOut = Duration(milliseconds: 300);
  static const Duration tooltip = Duration(milliseconds: 200);

  /// `cubic-bezier(0.32, 0, 0.67, 0)`.
  static const Curve easeIn = Cubic(0.32, 0, 0.67, 0);

  /// `cubic-bezier(0.33, 1, 0.68, 1)`.
  static const Curve easeOut = Cubic(0.33, 1, 0.68, 1);

  /// `cubic-bezier(0.65, 0, 0.35, 1)`.
  static const Curve easeInOut = Cubic(0.65, 0, 0.35, 1);
}
