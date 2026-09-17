# App Icon Package (Flutter)

Generated from your calendar logo. Contains ready-to-use icon files for
Android, iOS, and Web, plus the master source images.

## What's inside

```
assets/icon/
  icon.png                 1024x1024 flat icon (opaque bg) — master source
  icon_foreground.png      1024x1024 transparent glyph — Android adaptive foreground
  icon_background.png      1024x1024 solid background — Android adaptive background

android/app/src/main/res/
  mipmap-mdpi|hdpi|xhdpi|xxhdpi|xxxhdpi/ic_launcher.png        legacy launcher icons
  mipmap-mdpi|hdpi|xhdpi|xxhdpi|xxxhdpi/ic_launcher_foreground.png
  mipmap-mdpi|hdpi|xhdpi|xxhdpi|xxxhdpi/ic_launcher_background.png
  mipmap-anydpi-v26/ic_launcher.xml     adaptive icon definition

ios/Runner/Assets.xcassets/AppIcon.appiconset/
  Contents.json + all required @1x/@2x/@3x PNGs (20–1024pt)

web/
  favicon.ico, favicon-16x16.png, favicon-32x32.png
  apple-touch-icon.png
  icons/Icon-192.png, Icon-512.png, Icon-maskable-192.png, Icon-maskable-512.png
  manifest.json (PWA manifest referencing the icons above)

flutter_launcher_icons.yaml    optional config if you'd rather regenerate via the package
```

## Option A — Drop the files straight in (fastest)

1. Copy the `android/`, `ios/`, and `web/` folders into your Flutter project root,
   merging with your existing folders (these paths match Flutter's default
   project structure, so files will land in the right place).
2. Also copy `assets/icon/` into your project (handy to keep the source art around).
3. Rebuild the app — `flutter clean && flutter run` — so Xcode/Gradle picks up
   the new icons.
4. For web, make sure your `web/index.html` `<head>` references match:

```html
<link rel="icon" type="image/png" href="favicon.ico"/>
<link rel="apple-touch-icon" href="apple-touch-icon.png">
<link rel="manifest" href="manifest.json">
```

   (Flutter's default `web/index.html` already includes most of these — just
   confirm the filenames match, since this package replaces the default
   `favicon.png` with a full favicon set.)

## Option B — Regenerate with flutter_launcher_icons (recommended for future updates)

This is the standard Flutter tool and will keep every platform in sync any
time you change the source art.

1. Copy `assets/icon/icon.png` (and `icon_foreground.png`/`icon_background.png`
   if you want adaptive icons) into your project's `assets/icon/` folder.
2. Add the dev dependency and config below to your `pubspec.yaml` (or use the
   included `flutter_launcher_icons.yaml`):

```yaml
dev_dependencies:
  flutter_launcher_icons: ^0.14.1

flutter_launcher_icons:
  android: true
  ios: true
  image_path: "assets/icon/icon.png"
  adaptive_icon_background: "assets/icon/icon_background.png"
  adaptive_icon_foreground: "assets/icon/icon_foreground.png"
  web:
    generate: true
    image_path: "assets/icon/icon.png"
    background_color: "#F8F8F9"
    theme_color: "#1E1C26"
```

3. Run:

```bash
flutter pub get
flutter pub run flutter_launcher_icons
```

This regenerates every Android/iOS/Web icon file automatically from the one
source image.

## Colors used

- Calendar glyph / hangers: `#1E1C26` (near-black navy)
- Red dot: `#C42D22`
- Card interior gradient: white → `#FCE0DE`
- Background: `#F8F8F9`

## Note

The icon was newly redrawn (not a pixel copy) to match the style of the
image you shared, so it's safe to use as your own app's branding.
