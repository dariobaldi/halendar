import 'package:halendar_front/components/my_button.dart';
import 'package:halendar_front/components/my_container.dart';
import 'package:halendar_front/components/my_textformfield.dart';
import 'package:halendar_front/components/responsive.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginMenuSPage();
}

class _LoginMenuSPage extends State<LoginPage> {
  // Login text cotrollers
  final usernameController = TextEditingController();
  final passwordController = TextEditingController();
  bool showPassword = false;
  final _formKey = GlobalKey<FormState>();

  // Sign in method
  void signUserIn(BuildContext context) async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    showDialog(
      context: context,
      builder: (context) {
        return const Center(child: CircularProgressIndicator());
      },
    );

    final username = usernameController.text;
    final password = passwordController.text;

    int responseCode;
    String errorMessage = 'Identifiants non valides';

    try {
      responseCode = await AuthService.instance
          .authenticate(username, password)
          .timeout(const Duration(seconds: 3));
    } catch (_) {
      errorMessage = 'Le serveur ne répond pas';
      responseCode = 408;
    }

    if (context.mounted) {
      Navigator.pop(context);
    }

    if (responseCode != 201) {
      TextInput.finishAutofillContext();
      if (!context.mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Row(
            children: [
              const Icon(Icons.error),
              const SizedBox(width: 5),
              Expanded(
                child: Text(
                  errorMessage,
                  maxLines: 10,
                  overflow: TextOverflow.clip,
                  style: TextStyle(
                    color: Theme.of(context).colorScheme.onError,
                    fontSize: 18,
                  ),
                ),
              ),
            ],
          ),
          backgroundColor: Theme.of(context).colorScheme.error,
        ),
      );
    }
  }

  Widget loginMenu(BuildContext context) {
    return SingleChildScrollView(
      scrollDirection: Axis.vertical,
      child: MyContainer(
        width: 350,
        height: 470,
        color: Theme.of(context).colorScheme.surface,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // Logo
            Image.asset(
              "./lib/images/logo/ic_launcher.png",
              height: 200,
              width: 290,
            ),

            // username input
            loginForm(context),
          ],
        ),
      ),
    );
  }

  Form loginForm(BuildContext context) {
    return Form(
      key: _formKey,
      child: AutofillGroup(
        child: Column(
          children: [
            MyTextFormField(
              controller: usernameController,
              lableText: 'Pseudo',
              obscureText: false,
              onEnter: signUserIn,
              hints: const [AutofillHints.username, AutofillHints.email],
            ),

            const SizedBox(height: 20),
            // password input
            Stack(
              alignment: Alignment.centerRight,
              children: [
                MyTextFormField(
                  controller: passwordController,
                  lableText: 'Mot de passe',
                  obscureText: !showPassword,
                  onEnter: signUserIn,
                  hints: const [AutofillHints.password],
                ),
                Positioned(
                  right: 25,
                  child: Checkbox(
                    checkColor: Theme.of(context).colorScheme.onPrimary,
                    activeColor: Theme.of(context).colorScheme.primary,
                    value: showPassword,
                    onChanged: (bool? value) {
                      setState(() {
                        showPassword = value ?? false;
                      });
                    },
                  ),
                ),
              ],
            ),

            const SizedBox(height: 20),
            // login button
            MyButton(text: 'Se Connecter', onTap: signUserIn),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Theme.of(context).colorScheme.surface,
      body: Responsive(
        mobile: LoginPageMobile(loginMenu: loginMenu(context)),
        tablet: LoginPageDesktop(loginMenu: loginMenu(context)),
        desktop: LoginPageDesktop(loginMenu: loginMenu(context)),
      ),
    );
  }
}

class LoginPageMobile extends StatelessWidget {
  final Widget loginMenu;
  const LoginPageMobile({super.key, required this.loginMenu});

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Stack(
        alignment: Alignment.center,
        children: [
          Container(),
          Positioned(
            top: 0,
            left: 0,
            child: Image.asset("lib/images/graphics/bg_top_left.png", scale: 6),
          ),
          Positioned(
            bottom: 0,
            right: 0,
            child: Image.asset(
              "lib/images/graphics/bg_bottom_right.png",
              scale: 6,
            ),
          ),
          loginMenu,
        ],
      ),
    );
  }
}

class LoginPageDesktop extends StatelessWidget {
  final Widget loginMenu;
  const LoginPageDesktop({super.key, required this.loginMenu});

  @override
  Widget build(BuildContext context) {
    return Stack(
      alignment: Alignment.center,
      children: [
        Container(),
        Positioned(
          top: 0,
          left: 0,
          child: Image.asset("lib/images/graphics/bg_top_left.png", scale: 3),
        ),
        Positioned(
          bottom: 0,
          right: 0,
          child: Image.asset(
            "lib/images/graphics/bg_bottom_right.png",
            scale: 3,
          ),
        ),
        loginMenu,
      ],
    );
  }
}
