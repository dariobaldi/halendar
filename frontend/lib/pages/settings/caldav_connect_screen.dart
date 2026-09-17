import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

import '../../models/calendar_account.dart';

/// Connects a CalDAV calendar (Apple iCloud, a groupware's calendar such as La Suite
/// numérique if it speaks CalDAV, ...) directly from a URL/username/app password --
/// no redirect flow, unlike the OAuth providers.
///
/// [model] is passed explicitly rather than looked up via Provider: this screen is
/// pushed onto the app's Navigator, which sits outside the ChangeNotifierProvider
/// scope CalendarAccountsScreen wraps its own subtree in.
class CaldavConnectScreen extends StatefulWidget {
  final CalendarAccountsModel model;

  const CaldavConnectScreen({super.key, required this.model});

  @override
  State<CaldavConnectScreen> createState() => _CaldavConnectScreenState();
}

class _CaldavConnectScreenState extends State<CaldavConnectScreen> {
  final _formKey = GlobalKey<FormState>();
  final _urlController = TextEditingController(
    text: 'https://caldav.icloud.com/',
  );
  final _userController = TextEditingController();
  final _passController = TextEditingController();
  final _calendarController = TextEditingController();
  final _timezoneController = TextEditingController();

  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _urlController.dispose();
    _userController.dispose();
    _passController.dispose();
    _calendarController.dispose();
    _timezoneController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() {
      _submitting = true;
      _error = null;
    });

    final error = await widget.model.connectCaldav(
      url: _urlController.text.trim(),
      user: _userController.text.trim(),
      pass: _passController.text,
      calendar: _calendarController.text.trim(),
      timezone: _timezoneController.text.trim(),
    );

    if (!mounted) return;
    if (error == null) {
      Navigator.of(context).pop(true);
      return;
    }
    setState(() {
      _submitting = false;
      _error = error;
    });
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    return Scaffold(
      appBar: AppBar(title: const Text('Connect a CalDAV calendar')),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(LaSpacing.base),
          children: [
            Text(
              'Works with Apple iCloud and most groupware/CalDAV servers. Use an '
              'app-specific password, not your account password.',
              style: LaTextStyles.bodySm.copyWith(
                color: colors.contentNeutralSecondary,
              ),
            ),
            const SizedBox(height: LaSpacing.base),
            TextFormField(
              controller: _urlController,
              decoration: const InputDecoration(
                labelText: 'Server URL',
                border: OutlineInputBorder(),
              ),
              keyboardType: TextInputType.url,
              validator: (v) =>
                  (v == null || v.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: LaSpacing.sm),
            TextFormField(
              controller: _userController,
              decoration: const InputDecoration(
                labelText: 'Username / email',
                border: OutlineInputBorder(),
              ),
              validator: (v) =>
                  (v == null || v.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: LaSpacing.sm),
            TextFormField(
              controller: _passController,
              decoration: const InputDecoration(
                labelText: 'App password',
                border: OutlineInputBorder(),
              ),
              obscureText: true,
              validator: (v) => (v == null || v.isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: LaSpacing.sm),
            TextFormField(
              controller: _calendarController,
              decoration: const InputDecoration(
                labelText: 'Calendar name (optional)',
                helperText: 'Leave blank to use the first available calendar',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: LaSpacing.sm),
            TextFormField(
              controller: _timezoneController,
              decoration: const InputDecoration(
                labelText: 'Timezone (optional)',
                helperText: 'e.g. Europe/Paris -- defaults to Europe/Paris',
                border: OutlineInputBorder(),
              ),
            ),
            if (_error != null) ...[
              const SizedBox(height: LaSpacing.base),
              Text(
                _error!,
                style: LaTextStyles.bodySm.copyWith(
                  color: colors.contentErrorPrimary,
                ),
              ),
            ],
            const SizedBox(height: LaSpacing.base),
            FilledButton(
              onPressed: _submitting ? null : _submit,
              child: _submitting
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Text('Connect'),
            ),
          ],
        ),
      ),
    );
  }
}
