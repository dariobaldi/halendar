import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../models/email_account.dart';

/// Lets the user connect messaging accounts (Gmail today) that get imported and
/// analyzed for meeting requests in the background. Adding another provider later
/// is a matter of another entry in _providerLabels/_connect, not a new screen.
class EmailAccountsScreen extends StatelessWidget {
  const EmailAccountsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => EmailAccountsModel()..init(),
      child: const _EmailAccountsView(),
    );
  }
}

const _providerLabels = {'gmail': 'Gmail'};

class _EmailAccountsView extends StatefulWidget {
  const _EmailAccountsView();

  @override
  State<_EmailAccountsView> createState() => _EmailAccountsViewState();
}

class _EmailAccountsViewState extends State<_EmailAccountsView>
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
    context.read<EmailAccountsModel>().end();
    super.dispose();
  }

  // Google's OAuth policy disallows signing in from an embedded webview, so consent
  // happens in the system browser; there's no in-app callback to await. The backend
  // pushes a websocket refresh the moment it completes, and resuming the app (the
  // user switching back after finishing in the browser) is the fallback for when
  // that socket didn't survive being backgrounded.
  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      context.read<EmailAccountsModel>().fetch();
    }
  }

  Future<void> _connect(String provider) async {
    setState(() => _connecting = true);
    final model = context.read<EmailAccountsModel>();
    final authUrl = await model.beginConnect(provider);
    if (authUrl != null) {
      await launchUrl(Uri.parse(authUrl), mode: LaunchMode.externalApplication);
    }
    if (mounted) setState(() => _connecting = false);
  }

  Future<void> _confirmDisconnect(EmailAccount account) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Disconnect ${account.emailAddress}?'),
        content: const Text(
          'Halendar will stop importing mail from this account. Its imported message history is removed too.',
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
      await context.read<EmailAccountsModel>().disconnect(account);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final model = context.watch<EmailAccountsModel>();

    return Scaffold(
      appBar: AppBar(title: const Text('Connected accounts')),
      body: Column(
        children: [
          if (model.fetching) const LinearProgressIndicator(),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(LaSpacing.base),
              children: [
                if (!model.fetching && model.accounts.isEmpty)
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: LaSpacing.lg),
                    child: Column(
                      children: [
                        Icon(
                          Icons.alternate_email,
                          size: 48,
                          color: colors.contentNeutralTertiary,
                        ),
                        const SizedBox(height: LaSpacing.sm),
                        Text(
                          'No accounts connected yet',
                          style: LaTextStyles.labelLg.copyWith(
                            color: colors.contentNeutralPrimary,
                          ),
                        ),
                        const SizedBox(height: LaSpacing.x2xs),
                        Text(
                          'Connect a mailbox so Halendar can watch it for '
                          'meeting requests.',
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
                        Icons.mail_outline,
                        color: account.isActive
                            ? colors.contentNeutralSecondary
                            : colors.contentErrorPrimary,
                      ),
                      title: Text(account.emailAddress),
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
                  onPressed: _connecting ? null : () => _connect('gmail'),
                  icon: _connecting
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(Icons.add),
                  label: const Text('Connect Gmail'),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
