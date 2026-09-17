import 'package:flutter/material.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:lasuite_ui/lasuite_ui.dart';
import 'package:url_launcher/url_launcher.dart';

import 'settings/calendar_accounts_screen.dart';
import 'settings/email_accounts_screen.dart';
import '../widgets/theme_toggle_button.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  bool _notificationsEnabled = true;

  Future<void> _openMailApp() async {
    var launched = false;
    try {
      launched = await launchUrl(Uri(scheme: 'mailto'));
    } catch (_) {
      launched = false;
    }
    if (!launched && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Could not open your mail app.')),
      );
    }
  }

  void _signOut() {
    AuthService.instance.logOut();
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final user = AuthService.instance.user;
    final displayName = user?.name ?? user?.username ?? 'Signed in';
    return Scaffold(
      appBar: AppBar(
        title: const Text('Settings'),
        actions: const [ThemeToggleButton()],
      ),
      body: ListView(
        padding: const EdgeInsets.all(LaSpacing.base),
        children: [
          Card(
            child: ListTile(
              leading: LaAvatar(name: displayName),
              title: Text(
                displayName,
                style: LaTextStyles.labelLg.copyWith(
                  color: colors.contentNeutralPrimary,
                ),
              ),
              subtitle: Text(
                user?.username ?? '',
                style: LaTextStyles.bodySm.copyWith(
                  color: colors.contentNeutralSecondary,
                ),
              ),
            ),
          ),
          const SizedBox(height: LaSpacing.base),
          Card(
            child: Column(
              children: [
                ListTile(
                  leading: Icon(
                    Icons.alternate_email,
                    color: colors.contentNeutralSecondary,
                  ),
                  title: const Text('Connected email accounts'),
                  trailing: Icon(
                    Icons.chevron_right,
                    size: 18,
                    color: colors.contentNeutralTertiary,
                  ),
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (_) => const EmailAccountsScreen(),
                    ),
                  ),
                ),
                Divider(height: 1, color: colors.borderSurfacePrimary),
                ListTile(
                  leading: Icon(
                    Icons.calendar_month_outlined,
                    color: colors.contentNeutralSecondary,
                  ),
                  title: const Text('Connected calendars'),
                  trailing: Icon(
                    Icons.chevron_right,
                    size: 18,
                    color: colors.contentNeutralTertiary,
                  ),
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (_) => const CalendarAccountsScreen(),
                    ),
                  ),
                ),
                Divider(height: 1, color: colors.borderSurfacePrimary),
                ListTile(
                  leading: Icon(
                    Icons.mail_outline,
                    color: colors.contentNeutralSecondary,
                  ),
                  title: const Text('Open Mail'),
                  trailing: Icon(
                    Icons.open_in_new,
                    size: 18,
                    color: colors.contentNeutralTertiary,
                  ),
                  onTap: _openMailApp,
                ),
                Divider(height: 1, color: colors.borderSurfacePrimary),
                ListTile(
                  leading: Icon(
                    Icons.notifications_outlined,
                    color: colors.contentNeutralSecondary,
                  ),
                  title: const Text('Notifications'),
                  trailing: LaSwitch(
                    value: _notificationsEnabled,
                    onChanged: (value) =>
                        setState(() => _notificationsEnabled = value),
                  ),
                  onTap: () => setState(
                    () => _notificationsEnabled = !_notificationsEnabled,
                  ),
                ),
                Divider(height: 1, color: colors.borderSurfacePrimary),
                ListTile(
                  leading: Icon(
                    Icons.logout,
                    color: colors.contentNeutralSecondary,
                  ),
                  title: const Text('Sign out'),
                  onTap: _signOut,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
