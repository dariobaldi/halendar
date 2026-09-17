import 'package:halendar_front/components/my_textformfield.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

import '../widgets/halendar_logo.dart';
import '../widgets/theme_toggle_button.dart';

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
    String errorMessage = 'Invalid credentials';

    try {
      responseCode = await AuthService.instance
          .authenticate(username, password)
          .timeout(const Duration(seconds: 3));
    } catch (_) {
      errorMessage = 'The server is not responding';
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

  Form loginForm(BuildContext context) {
    return Form(
      key: _formKey,
      child: AutofillGroup(
        child: Column(
          children: [
            MyTextFormField(
              controller: usernameController,
              lableText: 'Username',
              obscureText: false,
              onEnter: signUserIn,
              hints: const [AutofillHints.username, AutofillHints.email],
              width: null,
            ),

            const SizedBox(height: 20),
            // password input
            MyTextFormField(
              controller: passwordController,
              lableText: 'Password',
              obscureText: !showPassword,
              onEnter: signUserIn,
              hints: const [AutofillHints.password],
              width: null,
              suffixIcon: IconButton(
                icon: Icon(
                  showPassword ? Icons.visibility_off : Icons.visibility,
                ),
                onPressed: () {
                  setState(() {
                    showPassword = !showPassword;
                  });
                },
              ),
            ),

            const SizedBox(height: 20),
            // login button
            // Matches MyTextFormField's own horizontal padding (20 each
            // side) so the button lines up with the actual input boxes
            // instead of the wider row that contains them.
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20.0),
              child: LaButton(
                label: 'Sign In',
                fullWidth: true,
                onPressed: () => signUserIn(context),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        actions: const [ThemeToggleButton()],
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(LaSpacing.md),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 420),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Center(
                    child: HalendarLogo(
                      strokeColor: colors.contentNeutralPrimary,
                      // backgroundErrorPrimary (not contentErrorPrimary) --
                      // the latter is a dark-mode *text-on-error* token and
                      // resolves to a near-white color, whereas this one is
                      // the same saturated red in both themes.
                      dotColor: colors.backgroundErrorPrimary,
                    ),
                  ),
                  const SizedBox(height: LaSpacing.sm),
                  Text(
                    'Halendar',
                    textAlign: TextAlign.center,
                    style: LaTextStyles.h4.copyWith(
                      color: colors.contentNeutralPrimary,
                    ),
                  ),
                  const SizedBox(height: LaSpacing.x2xs),
                  Text(
                    'Sign in to your account',
                    textAlign: TextAlign.center,
                    style: LaTextStyles.bodySm.copyWith(
                      color: colors.contentNeutralTertiary,
                    ),
                  ),
                  const SizedBox(height: LaSpacing.md),
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(LaSpacing.base),
                      child: loginForm(context),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
