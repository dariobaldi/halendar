import 'package:flutter/material.dart';

/// The calendar-and-dot mark. Drawn rather than shipped as an image so a
/// single widget can re-color itself for light/dark mode (stroke color is
/// caller-supplied) instead of needing two separate asset files.
class HalendarLogo extends StatelessWidget {
  final double size;
  final Color strokeColor;
  final Color dotColor;

  const HalendarLogo({
    super.key,
    this.size = 96,
    required this.strokeColor,
    required this.dotColor,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: size,
      height: size,
      child: CustomPaint(
        painter: _HalendarLogoPainter(
          strokeColor: strokeColor,
          dotColor: dotColor,
        ),
      ),
    );
  }
}

class _HalendarLogoPainter extends CustomPainter {
  final Color strokeColor;
  final Color dotColor;

  _HalendarLogoPainter({required this.strokeColor, required this.dotColor});

  @override
  void paint(Canvas canvas, Size size) {
    final w = size.width;
    final h = size.height;

    final strokePaint = Paint()
      ..color = strokeColor
      ..style = PaintingStyle.stroke
      ..strokeWidth = w * 0.055
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;

    // Card body.
    final bodyRect = Rect.fromLTWH(w * 0.16, h * 0.32, w * 0.68, h * 0.52);
    canvas.drawRRect(
      RRect.fromRectAndRadius(bodyRect, Radius.circular(w * 0.12)),
      strokePaint,
    );

    // Header divider, separating the top strip from the body.
    final dividerY = bodyRect.top + bodyRect.height * 0.24;
    canvas.drawLine(
      Offset(bodyRect.left, dividerY),
      Offset(bodyRect.right, dividerY),
      strokePaint,
    );

    // Binder rings hanging over the top edge.
    const ringCount = 4;
    final ringWidth = w * 0.05;
    final ringHeight = h * 0.16;
    final ringCenterY = bodyRect.top - ringHeight * 0.1;
    final firstRingX = bodyRect.left + bodyRect.width * 0.14;
    final lastRingX = bodyRect.right - bodyRect.width * 0.14;
    final step = (lastRingX - firstRingX) / (ringCount - 1);
    for (var i = 0; i < ringCount; i++) {
      final cx = firstRingX + step * i;
      final ringRect = Rect.fromCenter(
        center: Offset(cx, ringCenterY),
        width: ringWidth,
        height: ringHeight,
      );
      canvas.drawRRect(
        RRect.fromRectAndRadius(ringRect, Radius.circular(ringWidth / 2)),
        strokePaint,
      );
    }

    // Red dot with a soft glow, centered a little below the divider.
    final dotCenter = bodyRect.center + Offset(0, bodyRect.height * 0.08);
    final dotRadius = w * 0.065;

    final glowPaint = Paint()
      ..color = dotColor.withValues(alpha: 0.35)
      ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 14);
    canvas.drawCircle(dotCenter, dotRadius * 1.8, glowPaint);

    canvas.drawCircle(dotCenter, dotRadius, Paint()..color = dotColor);
  }

  @override
  bool shouldRepaint(covariant _HalendarLogoPainter oldDelegate) {
    return oldDelegate.strokeColor != strokeColor ||
        oldDelegate.dotColor != dotColor;
  }
}
