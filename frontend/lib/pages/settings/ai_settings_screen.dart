import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';
import 'package:provider/provider.dart';

import '../../models/ai_settings.dart';

/// Lets the user connect their own Claude API key and choose whether email analysis
/// uses it instead of the shared local Ollama instance. Connecting a key doesn't by
/// itself switch anything over -- the switch only takes effect (and can only be
/// turned on) once a key is on file.
class AISettingsScreen extends StatelessWidget {
  const AISettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => AISettingsModel()..fetch(),
      child: const _AISettingsView(),
    );
  }
}

class _AISettingsView extends StatefulWidget {
  const _AISettingsView();

  @override
  State<_AISettingsView> createState() => _AISettingsViewState();
}

class _AISettingsViewState extends State<_AISettingsView> {
  final _apiKeyController = TextEditingController();
  bool _editingKey = false;
  String? _error;

  @override
  void dispose() {
    _apiKeyController.dispose();
    super.dispose();
  }

  Future<void> _connect() async {
    final apiKey = _apiKeyController.text.trim();
    if (apiKey.isEmpty) return;
    setState(() => _error = null);
    final error = await context.read<AISettingsModel>().connectClaude(apiKey);
    if (!mounted) return;
    if (error == null) {
      _apiKeyController.clear();
      setState(() => _editingKey = false);
    } else {
      setState(() => _error = error);
    }
  }

  Future<void> _confirmDisconnect() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Disconnect Claude?'),
        content: const Text(
          'Email analysis will go back to using the local model. You can '
          'reconnect a key at any time.',
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
      await context.read<AISettingsModel>().disconnectClaude();
    }
  }

  Future<void> _toggleActive(bool useClaude) async {
    final model = context.read<AISettingsModel>();
    final error = await model.setProvider(useClaude ? 'claude' : 'ollama');
    if (error != null && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(error)));
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final model = context.watch<AISettingsModel>();
    final settings = model.settings;

    return Scaffold(
      appBar: AppBar(title: const Text('AI model')),
      body: model.fetching
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(LaSpacing.base),
              children: [
                Text(
                  'By default, Halendar analyzes your email with a local model '
                  'running on our server. Connect your own Claude API key to use '
                  'it instead -- useful if the local model\'s results aren\'t '
                  'reliable enough for you.',
                  style: LaTextStyles.bodySm.copyWith(
                    color: colors.contentNeutralSecondary,
                  ),
                ),
                const SizedBox(height: LaSpacing.base),
                Card(
                  child: Column(
                    children: [
                      ListTile(
                        leading: Icon(
                          Icons.auto_awesome,
                          color: settings.hasApiKey
                              ? colors.contentNeutralSecondary
                              : colors.contentNeutralTertiary,
                        ),
                        title: const Text('Claude'),
                        subtitle: Text(
                          settings.hasApiKey
                              ? 'API key connected'
                              : 'No API key connected',
                        ),
                        trailing: settings.hasApiKey
                            ? IconButton(
                                icon: const Icon(Icons.link_off),
                                tooltip: 'Disconnect',
                                onPressed: model.busy
                                    ? null
                                    : _confirmDisconnect,
                              )
                            : null,
                      ),
                      if (settings.hasApiKey) ...[
                        Divider(height: 1, color: colors.borderSurfacePrimary),
                        ListTile(
                          leading: Icon(
                            Icons.smart_toy_outlined,
                            color: colors.contentNeutralSecondary,
                          ),
                          title: const Text('Use Claude for analysis'),
                          subtitle: Text(
                            settings.isClaudeActive
                                ? 'Active -- new mail is analyzed with Claude'
                                : 'Off -- new mail is analyzed with the local model',
                          ),
                          trailing: LaSwitch(
                            value: settings.isClaudeActive,
                            onChanged: model.busy ? null : _toggleActive,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                if (!settings.hasApiKey) ...[
                  const SizedBox(height: LaSpacing.base),
                  if (!_editingKey)
                    FilledButton.icon(
                      onPressed: () => setState(() => _editingKey = true),
                      icon: const Icon(Icons.add),
                      label: const Text('Connect Claude API key'),
                    )
                  else
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(LaSpacing.base),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            TextFormField(
                              controller: _apiKeyController,
                              autofocus: true,
                              obscureText: true,
                              decoration: const InputDecoration(
                                labelText: 'Anthropic API key',
                                hintText: 'sk-ant-...',
                                border: OutlineInputBorder(),
                              ),
                              onFieldSubmitted: (_) => _connect(),
                            ),
                            if (_error != null) ...[
                              const SizedBox(height: LaSpacing.sm),
                              Text(
                                _error!,
                                style: LaTextStyles.bodySm.copyWith(
                                  color: colors.contentErrorPrimary,
                                ),
                              ),
                            ],
                            const SizedBox(height: LaSpacing.sm),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.end,
                              children: [
                                TextButton(
                                  onPressed: model.busy
                                      ? null
                                      : () => setState(() {
                                          _editingKey = false;
                                          _error = null;
                                          _apiKeyController.clear();
                                        }),
                                  child: const Text('Cancel'),
                                ),
                                const SizedBox(width: LaSpacing.x2xs),
                                FilledButton(
                                  onPressed: model.busy ? null : _connect,
                                  child: model.busy
                                      ? const SizedBox(
                                          width: 16,
                                          height: 16,
                                          child: CircularProgressIndicator(
                                            strokeWidth: 2,
                                          ),
                                        )
                                      : const Text('Connect'),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ),
                ],
              ],
            ),
    );
  }
}
