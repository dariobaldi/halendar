import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../models/calendar_account.dart';
import 'caldav_connect_screen.dart';

/// Lets the user connect calendars to check availability against (Google Calendar
/// and CalDAV -- Apple iCloud, and any groupware that speaks it -- today, more
/// providers later). Connecting Gmail (see EmailAccountsScreen) links a Google
/// Calendar account automatically from the same grant; this screen is mainly for
/// connecting a calendar on its own, or a non-Google one.
class CalendarAccountsScreen extends StatelessWidget {
  const CalendarAccountsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => CalendarAccountsModel()..init(),
      child: const _CalendarAccountsView(),
    );
  }
}

const _providerLabels = {'google': 'Google Calendar', 'caldav': 'CalDAV'};

class _CalendarAccountsView extends StatefulWidget {
  const _CalendarAccountsView();

  @override
  State<_CalendarAccountsView> createState() => _CalendarAccountsViewState();
}

class _CalendarAccountsViewState extends State<_CalendarAccountsView>
    with WidgetsBindingObserver {
  bool _connecting = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    context.read<CalendarAccountsModel>().end();
    super.dispose();
  }

  // Same reasoning as EmailAccountsScreen: Google's OAuth consent happens in the
  // system browser, so there's no in-app callback to await -- the backend pushes a
  // websocket refresh on completion, and resuming the app is the fallback.
  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      context.read<CalendarAccountsModel>().fetch();
    }
  }

  Future<void> _connectGoogle() async {
    setState(() => _connecting = true);
    final model = context.read<CalendarAccountsModel>();
    final authUrl = await model.beginConnect('google');
    if (authUrl != null) {
      await launchUrl(
        Uri.parse(authUrl),
        mode: LaunchMode.externalApplication,
      );
    }
    if (mounted) setState(() => _connecting = false);
  }

  Future<void> _connectCaldav() async {
    final model = context.read<CalendarAccountsModel>();
    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => CaldavConnectScreen(model: model)),
    );
  }

  Future<void> _confirmDisconnect(CalendarAccount account) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Disconnect ${account.displayName}?'),
        content: const Text(
          'Halendar will no longer check this calendar for availability when '
          'a meeting request comes in.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Disconnect'),
          ),
        ],
      ),
    );
    if (confirmed == true && mounted) {
      await context.read<CalendarAccountsModel>().disconnect(account);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final model = context.watch<CalendarAccountsModel>();

    return Scaffold(
      appBar: AppBar(title: const Text('Connected calendars')),
      body: Column(
        children: [
          if (model.fetching) const LinearProgressIndicator(),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(LaSpacing.base),
              children: [
                if (!model.fetching && model.accounts.isEmpty)
                  Padding(
                    padding: const EdgeInsets.symmetric(
                      vertical: LaSpacing.lg,
                    ),
                    child: Column(
                      children: [
                        Icon(
                          Icons.calendar_month_outlined,
                          size: 48,
                          color: colors.contentNeutralTertiary,
                        ),
                        const SizedBox(height: LaSpacing.sm),
                        Text(
                          'No calendars connected yet',
                          style: LaTextStyles.labelLg.copyWith(
                            color: colors.contentNeutralPrimary,
                          ),
                        ),
                        const SizedBox(height: LaSpacing.x2xs),
                        Text(
                          'Connect a calendar so Halendar can check your '
                          'availability before proposing a reply.',
                          textAlign: TextAlign.center,
                          style: LaTextStyles.bodySm.copyWith(
                            color: colors.contentNeutralTertiary,
                          ),
                        ),
                      ],
                    ),
                  ),
                for (final account in model.accounts)
                  Card(
                    child: ListTile(
                      leading: Icon(
                        Icons.calendar_month_outlined,
                        color: account.isActive
                            ? colors.contentNeutralSecondary
                            : colors.contentErrorPrimary,
                      ),
                      title: Text(account.displayName),
                      subtitle: Text(
                        account.isActive
                            ? _providerLabels[account.provider] ??
                                  account.provider
                            : (account.lastError ?? 'Connection error'),
                        style: account.isActive
                            ? null
                            : TextStyle(color: colors.contentErrorPrimary),
                      ),
                      trailing: IconButton(
                        icon: const Icon(Icons.link_off),
                        tooltip: 'Disconnect',
                        onPressed: () => _confirmDisconnect(account),
                      ),
                    ),
                  ),
                const SizedBox(height: LaSpacing.base),
                FilledButton.icon(
                  onPressed: _connecting ? null : _connectGoogle,
                  icon: _connecting
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(Icons.add),
                  label: const Text('Connect Google Calendar'),
                ),
                const SizedBox(height: LaSpacing.sm),
                OutlinedButton.icon(
                  onPressed: _connectCaldav,
                  icon: const Icon(Icons.add),
                  label: const Text('Connect a CalDAV calendar'),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
