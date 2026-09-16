import 'package:flutter/material.dart';

import '../../theme/lasuite_theme.dart';
import '../../tokens/tokens.dart';
import '../../utils/la_semantic_palette.dart';
import '../loader/la_loader.dart';

/// Button size, mirroring the upstream `nano` / `small` / `medium` scale.
enum LaButtonSize { nano, small, medium }

/// Button color, mirroring the upstream semantic `color` prop.
typedef LaButtonColor = LaSemanticCategory;

/// Button visual treatment, mirroring the upstream `variant` prop:
/// - [primary]: filled, high emphasis.
/// - [secondary]: tinted background, medium emphasis.
/// - [tertiary]: text-only, low emphasis.
/// - [bordered]: outlined, text-only with a neutral border.
enum LaButtonVariant { primary, secondary, tertiary, bordered }

/// Where [LaButton.icon] is placed relative to [LaButton.label].
enum LaIconPosition { left, right }

/// The La Suite numérique button.
///
/// Reproduces the upstream Button's size/color/variant matrix, its
/// icon-only and full-width modes, and its hover/focus-visible/pressed/
/// disabled states, using Flutter-native gesture and focus handling instead
/// of CSS pseudo-classes.
class LaButton extends StatefulWidget {
  const LaButton({
    super.key,
    this.label,
    this.icon,
    this.iconPosition = LaIconPosition.left,
    required this.onPressed,
    this.size = LaButtonSize.medium,
    this.color = LaButtonColor.brand,
    this.variant = LaButtonVariant.primary,
    this.fullWidth = false,
    this.loading = false,
    this.autofocus = false,
    this.semanticLabel,
  }) : assert(
         label != null || icon != null,
         'LaButton requires a label and/or an icon.',
       );

  final String? label;
  final Widget? icon;
  final LaIconPosition iconPosition;

  /// Called when the button is activated. The button renders as disabled
  /// when this is null.
  final VoidCallback? onPressed;

  final LaButtonSize size;
  final LaButtonColor color;
  final LaButtonVariant variant;
  final bool fullWidth;

  /// Shows a spinner in place of the label/icon and disables interaction,
  /// without changing the button's layout size.
  final bool loading;

  final bool autofocus;

  /// Overrides the accessibility label (defaults to [label]); required when
  /// the button is icon-only.
  final String? semanticLabel;

  bool get _iconOnly => icon != null && label == null;

  double get _height => switch (size) {
    LaButtonSize.nano => 24,
    LaButtonSize.small => 32,
    LaButtonSize.medium => 40,
  };

  double get _fontSize => switch (size) {
    LaButtonSize.nano => LaFontSize.xs,
    LaButtonSize.small => LaFontSize.sm,
    LaButtonSize.medium => LaFontSize.md,
  };

  double get _horizontalPadding => switch (size) {
    LaButtonSize.nano => LaSpacing.xs,
    LaButtonSize.small => LaSpacing.sm,
    LaButtonSize.medium => LaSpacing.base,
  };

  double get _iconSize => switch (size) {
    LaButtonSize.nano => 16,
    LaButtonSize.small => 18,
    LaButtonSize.medium => 20,
  };

  @override
  State<LaButton> createState() => _LaButtonState();
}

class _LaButtonState extends State<LaButton> {
  bool _hovered = false;
  bool _pressed = false;

  /// Whether the button reacts to input (hover/press/tap).
  bool get _enabled => widget.onPressed != null && !widget.loading;

  /// Whether the button is painted in its disabled palette. A button that
  /// has a handler but is [LaButton.loading] keeps its normal colors (with
  /// the spinner swapped in for the content) rather than looking disabled.
  bool get _looksDisabled => widget.onPressed == null;

  @override
  Widget build(BuildContext context) {
    final palette = LaSemanticPalette.of(context.laColors, widget.color);
    final disabledBg = context.laColors.backgroundDisabledPrimary;
    final disabledFg = context.laColors.contentDisabledPrimary;

    final Color background;
    final Color foreground;
    Border? border;

    if (_looksDisabled) {
      background = switch (widget.variant) {
        LaButtonVariant.primary => disabledBg,
        LaButtonVariant.secondary =>
          context.laColors.backgroundDisabledSecondary,
        LaButtonVariant.tertiary ||
        LaButtonVariant.bordered => Colors.transparent,
      };
      foreground = disabledFg;
      if (widget.variant == LaButtonVariant.bordered) {
        border = Border.all(color: context.laColors.borderDisabledPrimary);
      }
    } else {
      switch (widget.variant) {
        case LaButtonVariant.primary:
          background = _hovered || _pressed
              ? palette.backgroundPrimaryHover
              : palette.backgroundPrimary;
          foreground = palette.contentOn;
        case LaButtonVariant.secondary:
          background = _hovered || _pressed
              ? palette.backgroundSecondaryHover
              : palette.backgroundSecondary;
          foreground = palette.contentSecondary;
        case LaButtonVariant.tertiary:
          background = _hovered || _pressed
              ? context.laColors.backgroundNeutralTertiary
              : Colors.transparent;
          foreground = palette.contentTertiary;
        case LaButtonVariant.bordered:
          background = _hovered || _pressed
              ? context.laColors.backgroundNeutralTertiary
              : Colors.transparent;
          foreground = palette.contentTertiary;
          border = Border.all(color: context.laColors.borderNeutralTertiary);
      }
    }

    final Widget content = widget.loading
        ? Row(
            mainAxisSize: MainAxisSize.min,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              SizedBox(
                height: widget._iconSize,
                width: widget._iconSize,
                child: LaLoader(size: LaLoaderSize.small, color: foreground),
              ),
            ],
          )
        : _ButtonContent(
            label: widget.label,
            icon: widget.icon,
            iconPosition: widget.iconPosition,
            foreground: foreground,
            fontSize: widget._fontSize,
            iconSize: widget._iconSize,
          );

    final button = MouseRegion(
      cursor: _enabled ? SystemMouseCursors.click : SystemMouseCursors.basic,
      onEnter: (_) => setState(() => _hovered = true),
      onExit: (_) => setState(() {
        _hovered = false;
        _pressed = false;
      }),
      child: GestureDetector(
        onTapDown: _enabled ? (_) => setState(() => _pressed = true) : null,
        onTapUp: _enabled ? (_) => setState(() => _pressed = false) : null,
        onTapCancel: _enabled ? () => setState(() => _pressed = false) : null,
        onTap: widget.loading ? null : widget.onPressed,
        child: FocusableActionDetector(
          autofocus: widget.autofocus,
          enabled: _enabled,
          mouseCursor: SystemMouseCursors.click,
          child: Builder(
            builder: (context) {
              final focused = Focus.of(context).hasPrimaryFocus;
              return AnimatedContainer(
                duration: LaMotion.duration,
                curve: LaMotion.easeOut,
                height: widget._height,
                width: widget._iconOnly ? widget._height : null,
                padding: widget._iconOnly
                    ? EdgeInsets.zero
                    : EdgeInsets.symmetric(
                        horizontal: widget._horizontalPadding,
                      ),
                decoration: BoxDecoration(
                  color: background,
                  borderRadius: BorderRadius.circular(LaRadius.sm),
                  border: border,
                  boxShadow: focused && _enabled
                      ? [
                          BoxShadow(
                            color: palette.borderPrimary,
                            spreadRadius: 2,
                          ),
                        ]
                      : null,
                ),
                // No `alignment` here: Container/AnimatedContainer expand to
                // fill bounded parent constraints whenever alignment is set,
                // which would stretch non-fullWidth buttons inside a Wrap or
                // Row. Height is fixed above, so Row's default centered
                // crossAxisAlignment already vertically centers the content;
                // width stays shrink-wrapped to it.
                child: content,
              );
            },
          ),
        ),
      ),
    );

    final result = Semantics(
      button: true,
      enabled: _enabled,
      label: widget.semanticLabel ?? widget.label,
      child: ExcludeSemantics(child: button),
    );

    return widget.fullWidth
        ? SizedBox(width: double.infinity, child: result)
        : result;
  }
}

class _ButtonContent extends StatelessWidget {
  const _ButtonContent({
    required this.label,
    required this.icon,
    required this.iconPosition,
    required this.foreground,
    required this.fontSize,
    required this.iconSize,
  });

  final String? label;
  final Widget? icon;
  final LaIconPosition iconPosition;
  final Color foreground;
  final double fontSize;
  final double iconSize;

  @override
  Widget build(BuildContext context) {
    final iconWidget = icon == null
        ? null
        : IconTheme.merge(
            data: IconThemeData(color: foreground, size: iconSize),
            child: icon!,
          );

    final text = label == null
        ? null
        : Text(
            label!,
            style: TextStyle(
              color: foreground,
              fontSize: fontSize,
              fontWeight: LaFontWeight.medium,
              fontFamily: LaFontFamily.base,
              fontFamilyFallback: LaFontFamily.fallback,
            ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          );

    // Always wrap in a Row, even for a single child: the AnimatedContainer
    // above gives its child a *tight* height (matching the button size) but
    // a loose width, and only Row's `crossAxisAlignment.center` (its
    // default) actually centers a shorter child within that tight height —
    // a bare Text/Icon would just sit at the top of the box instead.
    final List<Widget> children;
    if (iconWidget == null) {
      children = [text!];
    } else if (text == null) {
      children = [iconWidget];
    } else if (iconPosition == LaIconPosition.left) {
      children = [
        iconWidget,
        const SizedBox(width: LaSpacing.x3xs),
        Flexible(child: text),
      ];
    } else {
      children = [
        Flexible(child: text),
        const SizedBox(width: LaSpacing.x3xs),
        iconWidget,
      ];
    }

    // `mainAxisAlignment: center` matters for the icon-only case: the
    // AnimatedContainer gives that Row a *tight* square (width == height ==
    // button size), so MainAxisSize.min can no longer shrink it to the
    // icon's own width — without centering, the icon would hug the row's
    // start edge instead of sitting in the middle of the square button.
    return Row(
      mainAxisSize: MainAxisSize.min,
      mainAxisAlignment: MainAxisAlignment.center,
      children: children,
    );
  }
}
