// Temporary script-as-test used to rasterize HalendarLogo into master PNGs
// for app icon generation. Not a real test; deleted after use.
import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:halendar_front/widgets/halendar_logo.dart';

const strokeColor = Color(0xFF25252F);
const dotColor = Color(0xFFD7010E);
const bgColor = Color(0xFFFFFFFF);

const outDir = '/private/tmp/claude-501/-Users-dario-42-halendar/81a4eeab-cf8a-4d7f-b76e-f74ad168abad/scratchpad';

Future<void> captureIcon(
  WidgetTester tester, {
  required String fileName,
  required double canvasSize,
  required double logoSize,
}) async {
  final key = GlobalKey();
  await tester.pumpWidget(
    Directionality(
      textDirection: TextDirection.ltr,
      child: RepaintBoundary(
        key: key,
        child: Container(
          width: canvasSize,
          height: canvasSize,
          color: bgColor,
          child: Center(
            child: HalendarLogo(
              size: logoSize,
              strokeColor: strokeColor,
              dotColor: dotColor,
            ),
          ),
        ),
      ),
    ),
  );
  await tester.pump();

  final boundary =
      key.currentContext!.findRenderObject() as RenderRepaintBoundary;
  final image = await boundary.toImage(pixelRatio: 1.0);
  final byteData = await image.toByteData(format: ui.ImageByteFormat.png);
  final file = File('$outDir/$fileName');
  await file.writeAsBytes(byteData!.buffer.asUint8List());
}

void main() {
  testWidgets('generate icon masters', (tester) async {
    const canvas = 1024.0;
    await tester.binding.setSurfaceSize(const Size(canvas, canvas));
    tester.view.physicalSize = const Size(canvas, canvas);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    // Standard icon: logo fills most of the canvas with a small margin.
    await captureIcon(
      tester,
      fileName: 'master_icon.png',
      canvasSize: canvas,
      logoSize: canvas * 0.86,
    );

    // Maskable icon: logo kept within the ~80% safe-zone circle so it
    // survives circular/squircle masking on Android launchers.
    await captureIcon(
      tester,
      fileName: 'master_maskable.png',
      canvasSize: canvas,
      logoSize: canvas * 0.60,
    );
  });
}
