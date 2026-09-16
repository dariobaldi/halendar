import 'package:halendar_front/pages/login.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:flex_color_scheme/flex_color_scheme.dart';
import 'package:flutter/material.dart';
import 'package:hive_flutter/hive_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  // Using Hive for local storage
  await Hive.initFlutter();
  var _ = await Hive.openBox('localDB');
  await AuthService.instance.init();

  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<AuthUser?>(
      stream: AuthService.instance.isLoggedInStream,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasData) {
          return MaterialApp(
            theme: FlexThemeData.light(scheme: FlexScheme.deepPurple),
            darkTheme: FlexThemeData.dark(scheme: FlexScheme.material),
            home: Scaffold(body: Text("Halendar")),
          );
        }
        return MaterialApp(
          theme: FlexThemeData.light(scheme: FlexScheme.deepPurple),
          darkTheme: FlexThemeData.dark(scheme: FlexScheme.deepPurple),
          home: const LoginPage(),
        );
      },
    );
  }
}
