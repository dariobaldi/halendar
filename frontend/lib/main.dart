import 'package:halendar_front/pages/home_shell.dart';
import 'package:halendar_front/pages/login.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:halendar_front/state/theme_controller.dart';
import 'package:flutter/material.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  // Using Hive for local storage
  await Hive.initFlutter();
  var _ = await Hive.openBox('localDB');
  await AuthService.instance.init();
  await initializeDateFormatting('en_US');

  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ValueListenableBuilder<ThemeMode>(
      valueListenable: ThemeController.instance,
      builder: (context, mode, _) {
        return StreamBuilder<AuthUser?>(
          stream: AuthService.instance.isLoggedInStream,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return MaterialApp(
                theme: _extendTheme(LaSuiteTheme.light(), LaColors.light),
                darkTheme: _extendTheme(LaSuiteTheme.dark(), LaColors.dark),
                themeMode: mode,
                home: const Scaffold(
                  body: Center(child: CircularProgressIndicator()),
                ),
              );
            }
            return MaterialApp(
              title: 'Halendar',
              theme: _extendTheme(LaSuiteTheme.light(), LaColors.light),
              darkTheme: _extendTheme(LaSuiteTheme.dark(), LaColors.dark),
              themeMode: mode,
              home: snapshot.hasData ? const HomeShell() : const LoginPage(),
            );
          },
        );
      },
    );
  }
}

/// The kit ships tokens, a theme and components, but no Card or
/// NavigationBar -- those are app-level layout choices, not design-system
/// atoms. This styles both from the same surface/border/brand/radius
/// tokens as everything else, so they read as part of the same system
/// rather than Flutter's Material defaults (whose bottom-nav indicator in
/// particular would otherwise fall back to a generic tonal purple instead
/// of the kit's own brand color).
ThemeData _extendTheme(ThemeData theme, LaColors colors) {
  return theme.copyWith(
    cardTheme: CardThemeData(
      elevation: 0,
      color: colors.surfacePrimary,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(LaRadius.lg),
        side: BorderSide(color: colors.borderSurfacePrimary),
      ),
    ),
    navigationBarTheme: NavigationBarThemeData(
      backgroundColor: colors.surfacePrimary,
      indicatorColor: colors.backgroundBrandSecondary,
      iconTheme: WidgetStateProperty.resolveWith(
        (states) => IconThemeData(
          color: states.contains(WidgetState.selected)
              ? colors.contentBrandPrimary
              : colors.contentNeutralTertiary,
        ),
      ),
      labelTextStyle: WidgetStateProperty.resolveWith(
        (states) => LaTextStyles.labelSm.copyWith(
          color: states.contains(WidgetState.selected)
              ? colors.contentBrandPrimary
              : colors.contentNeutralTertiary,
        ),
      ),
    ),
  );
}
