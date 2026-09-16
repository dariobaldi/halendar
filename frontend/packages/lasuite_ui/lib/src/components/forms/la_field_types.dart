/// Validation state shared by every form control (text field, text area,
/// checkbox, radio, switch, select).
enum LaFieldState { normal, success, error }

/// How a form control's label is displayed, mirroring the upstream
/// `FieldVariant`:
/// - [floating]: the label serves as the placeholder when empty and floats
///   above the value once focused/filled.
/// - [classic]: the label always sits above the control.
/// - [inline]: the label sits in a left column, the control in a right
///   column, on the same row.
enum LaFieldVariant { floating, classic, inline }
