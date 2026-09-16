import 'dart:convert';

import 'package:halendar_front/components/my_textformfield.dart';
import 'package:halendar_front/services/api.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:flutter/material.dart';

class ChangeCredentialsPage extends StatefulWidget {
  const ChangeCredentialsPage({super.key});

  @override
  State<ChangeCredentialsPage> createState() => _ChangeCredentialsPageState();
}

class _ChangeCredentialsPageState extends State<ChangeCredentialsPage> {
  final _formKey = GlobalKey<FormState>();
  final _oldPassword = TextEditingController();
  bool _showOld = false;
  final _newPassword = TextEditingController();
  bool _showNew = false;
  final _newPasswordDuplicate = TextEditingController();
  bool _showNewDuplcate = false;

  Future<void> submitForm(BuildContext context) async {
    if (!_formKey.currentState!.validate()) {
      return;
    }
    try {
      final response = await apiRequest(
        'POST',
        'v1/users/change_password',
        true,
        json.encode({
          'username': AuthService.instance.user!.username,
          'password': _oldPassword.text,
          'new_password': _newPassword.text,
        }),
        {},
      );
      if (response.statusCode == 201) {
        AuthService.instance.logOut();
        addNotification(
          title: "Le mot de passe a été changé",
          content: "",
          type: "success",
          duration: 3,
        );
        return;
      } else {
        addNotification(
          title: "Erreur",
          content: "Erreur: ${response.body}",
          type: "error",
        );
      }
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "submitForm()");
    }
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        padding: const EdgeInsets.all(8.0),
        width: 350,
        child: Form(
          key: _formKey,
          child: Column(
            children: [
              Row(
                children: [
                  ActionChip(
                    onPressed: () {
                      Navigator.of(context).pop();
                    },
                    label: const SizedBox(
                      height: 40,
                      child: Icon(Icons.arrow_back),
                    ),
                  ),
                ],
              ),
              const Text(
                "Changer les identifiants",
                style: TextStyle(fontSize: 20),
              ),
              const SizedBox(height: 15),
              Stack(
                alignment: Alignment.centerRight,
                children: [
                  MyTextFormField(
                    controller: _oldPassword,
                    lableText: 'Ancien',
                    obscureText: !_showOld,
                    onEnter: submitForm,
                    hints: const [AutofillHints.password],
                  ),
                  Positioned(
                    right: 25,
                    child: Checkbox(
                      checkColor: Theme.of(context).colorScheme.onPrimary,
                      activeColor: Theme.of(context).colorScheme.primary,
                      value: _showOld,
                      onChanged: (bool? value) {
                        setState(() {
                          _showOld = value ?? false;
                        });
                      },
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              Stack(
                alignment: Alignment.centerRight,
                children: [
                  MyTextFormField(
                    controller: _newPassword,
                    lableText: 'Nouveau',
                    obscureText: !_showNew,
                    onEnter: submitForm,
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return 'La valeur ne peux pas être vide';
                      }
                      if (value.length < 8 || value.length > 72) {
                        return 'Le mot de passe doit avoir entre 8 et 72 charactères';
                      }
                      return null;
                    },
                  ),
                  Positioned(
                    right: 25,
                    child: Checkbox(
                      checkColor: Theme.of(context).colorScheme.onPrimary,
                      activeColor: Theme.of(context).colorScheme.primary,
                      value: _showNew,
                      onChanged: (bool? value) {
                        setState(() {
                          _showNew = value ?? false;
                        });
                      },
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              Stack(
                alignment: Alignment.centerRight,
                children: [
                  MyTextFormField(
                    controller: _newPasswordDuplicate,
                    lableText: 'Vérif nouveau',
                    obscureText: !_showNewDuplcate,
                    onEnter: submitForm,
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return 'La valeur ne peux pas être vide';
                      }
                      if (_newPassword.text != _newPasswordDuplicate.text) {
                        return 'Le mot de passe n\'est pas le même';
                      }
                      return null;
                    },
                  ),
                  Positioned(
                    right: 25,
                    child: Checkbox(
                      checkColor: Theme.of(context).colorScheme.onPrimary,
                      activeColor: Theme.of(context).colorScheme.primary,
                      value: _showNewDuplcate,
                      onChanged: (bool? value) {
                        setState(() {
                          _showNewDuplcate = value ?? false;
                        });
                      },
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              FilledButton(
                onPressed: () {
                  submitForm(context);
                },
                child: const Text('Mettre à jour'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
