# lasuite_ui

A Flutter-native implementation of the [La Suite numérique UI Kit](https://github.com/suitenumerique/ui-kit) design system: design tokens, an app theme, and a set of reusable, idiomatic Flutter widgets.

This is **not** a port of the upstream React/TypeScript code. The design tokens (colors, spacing, typography, radii, motion) and the component inventory (variants, states, sizes) were read from the upstream kit and re-implemented from scratch as Dart/Flutter widgets built on Flutter's own theming, gesture and accessibility APIs. See [ATTRIBUTION.md](ATTRIBUTION.md) for details.

## Screenshots

Captured from the showcase app in `example/`.

| | |
|---|---|
| ![Showcase home](doc/screenshots/home-light.png) | ![Showcase home, dark theme](doc/screenshots/home-dark.png) |
| ![Design tokens page](doc/screenshots/tokens.png) | ![Button variants, colors, sizes, states](doc/screenshots/button.png) |
| ![Alert types and expandable alert](doc/screenshots/alert.png) | ![Modal dialog](doc/screenshots/modal.png) |
| ![Form controls: text field variants, icons, validation states](doc/screenshots/forms.png) | |

## Features

**Design tokens** (`lib/src/tokens/`)
- `LaPalette` — the full raw color ramps (brand, gray, info, success, warning, error, red, orange, brown, yellow, green, blue-1, blue-2, purple, pink, black, white), 50–950 steps.
- `LaColors` — a `ThemeExtension` with ~100 semantic colors (backgroundXPrimary/Secondary/Tertiary, contentX…, borderX… for brand/neutral/info/success/warning/error), for both light and dark.
- `LaAccentColors` — 11 named accent hues (for avatars, tags, data-viz).
- `LaSpacing`, `LaFontSize`, `LaFontWeight`, `LaFontFamily`, `LaTextStyles`, `LaRadius`, `LaMotion`, `LaBreakpoints`.

**Theme** (`lib/src/theme/`)
- `LaSuiteTheme.light()` / `LaSuiteTheme.dark()` — ready-to-use `ThemeData` for `MaterialApp.theme`/`darkTheme`.
- `context.laColors` — reads the current `LaColors` off `Theme.of(context)`.

**Components** (`lib/src/components/`)
- `LaButton` — sizes (nano/small/medium) × colors (brand/neutral/info/success/warning/error) × variants (primary/secondary/tertiary/bordered), icons, loading, full-width.
- `LaBadge` — 6 semantic types, uppercased option.
- `LaAlert` — info/success/warning/error/neutral, expandable additional content, actions, dismiss.
- `LaToast` / `LaToastMessenger` — overlay-based toast notifications.
- `LaModal` / `showLaModal` — dialogs with 5 size presets and a left/right action row.
- `LaTabs` — a self-contained tab strip + panel.
- `LaPagination` — numbered pages with ellipsis collapsing + optional "go to page".
- `LaTooltip` — hover/long-press hint styled to the kit.
- `LaLoader` — small/medium spinner.
- `LaHorizontalSeparator` / `LaVerticalSeparator`.
- `LaAvatar` — initials avatar colored from the accent palette.
- Forms: `LaTextField`, `LaTextArea`, `LaCheckbox`, `LaRadio`/`LaRadioGroup`, `LaSwitch`, `LaSelect` — all sharing `LaFieldState` (normal/success/error) and, for text inputs, `LaFieldVariant` (floating/classic/inline label placement).

## Getting started

Add the package (from this repo, until/unless it's published):

```yaml
dependencies:
  lasuite_ui:
    path: ../path/to/ui-flutter-components
```

Install the theme once at the root of your app:

```dart
import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

void main() => runApp(const MyApp());

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      theme: LaSuiteTheme.light(),
      darkTheme: LaSuiteTheme.dark(),
      home: const HomePage(),
    );
  }
}
```

## Usage

```dart
LaButton(
  label: 'Save',
  color: LaButtonColor.brand,
  variant: LaButtonVariant.primary,
  icon: const Icon(Icons.check),
  onPressed: () {},
)

LaBadge(label: 'New', type: LaBadgeType.success)

LaAlert(
  message: 'Your export is ready.',
  type: LaVariant.success,
  canClose: true,
)

LaTextField(
  label: 'Email',
  state: LaFieldState.error,
  helperText: 'Enter a valid email address.',
)

// Reading a design token directly:
Container(color: context.laColors.backgroundBrandTertiary)
Container(color: LaPalette.brand[500])
```

If you only want the tokens/theme (e.g. to theme your own widgets), just import `lasuite_ui` and use `LaColors` / `LaPalette` / `LaSpacing` / etc. — no component usage required.

## Showcase app

`example/` is a runnable Flutter app listing every component and its variants/states:

```bash
cd example
flutter pub get
flutter run -d chrome   # or: -d macos / -d linux / -d windows / a connected device
```

## Additional information

- Design tokens and component behavior are derived from https://github.com/suitenumerique/ui-kit (MIT licensed) — see [ATTRIBUTION.md](ATTRIBUTION.md).
- Contributions: keep new components consistent with the existing token usage (`LaColors`, `LaSpacing`, etc.) rather than hard-coded values, and add both a showcase page and widget tests.
