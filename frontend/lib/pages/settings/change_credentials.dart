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
          title: "Password changed",
          content: "",
          type: "success",
          duration: 3,
        );
        return;
      } else {
        addNotification(
          title: "Error",
          content: "Error: ${response.body}",
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
                "Change credentials",
                style: TextStyle(fontSize: 20),
              ),
              const SizedBox(height: 15),
              Stack(
                alignment: Alignment.centerRight,
                children: [
                  MyTextFormField(
                    controller: _oldPassword,
                    lableText: 'Current',
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
                    lableText: 'New',
                    obscureText: !_showNew,
                    onEnter: submitForm,
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return 'This value cannot be empty';
                      }
                      if (value.length < 8 || value.length > 72) {
                        return 'The password must be between 8 and 72 characters';
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
                    lableText: 'Confirm new',
                    obscureText: !_showNewDuplcate,
                    onEnter: submitForm,
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return 'This value cannot be empty';
                      }
                      if (_newPassword.text != _newPasswordDuplicate.text) {
                        return 'The passwords do not match';
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
                child: const Text('Update'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
