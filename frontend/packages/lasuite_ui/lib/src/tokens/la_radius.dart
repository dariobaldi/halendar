/// Border radius scale, collected from the per-component tokens of the
/// La Suite numérique UI Kit (buttons, inputs, badges, modals, tooltips...).
abstract final class LaRadius {
  static const double none = 0;
  static const double xs = 2; // checkbox, active/pressed button
  static const double sm = 4; // button, input hover/focus
  static const double md = 8; // input default, modal, tooltip
  static const double lg = 12; // badge
  static const double full = 999; // switch rail/handle, pill shapes
}
