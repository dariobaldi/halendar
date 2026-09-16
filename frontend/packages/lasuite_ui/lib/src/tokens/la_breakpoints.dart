/// Responsive breakpoint tokens (design token `globals.breakpoints`), in
/// logical pixels.
abstract final class LaBreakpoints {
  static const double xxs = 320;
  static const double xs = 480;
  static const double mobile = 768;
  static const double tablet = 1024;

  /// True when [width] is at or above the "mobile" breakpoint (768px).
  static bool isAtLeastMobile(double width) => width >= mobile;

  /// True when [width] is at or above the "tablet" breakpoint (1024px).
  static bool isAtLeastTablet(double width) => width >= tablet;
}
