import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:halendar_front/services/api.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:lasuite_ui/lasuite_ui.dart';
import 'package:provider/provider.dart';

import '../../models/ai_settings.dart';

/// One key-bearing provider's connect/disconnect UI can register with this so the
/// screen knows its display name, key-field hint, and API console link -- adding a
/// third provider later is a matter of one more entry here, not a new screen.
class _ProviderInfo {
  final String id;
  final String label;
  final String keyLabel;
  final String keyHint;

  const _ProviderInfo({
    required this.id,
    required this.label,
    required this.keyLabel,
    required this.keyHint,
  });
}

const _providers = [
  _ProviderInfo(
    id: 'claude',
    label: 'Claude',
    keyLabel: 'Anthropic API key',
    keyHint: 'sk-ant-...',
  ),
  _ProviderInfo(
    id: 'gemini',
    label: 'Gemini',
    keyLabel: 'Google AI API key',
    keyHint: 'AIza...',
  ),
];

/// Lets the user connect their own Claude and/or Gemini API key and choose which
/// model (those, or the shared local Ollama instance) email analysis uses. Connecting
/// a key doesn't by itself switch anything over -- a provider can only be selected
/// once a key is on file for it, and a user can hold keys for more than one at a time
/// and switch freely between them.
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

/// A single, tappable option row for choosing the active provider -- a plain
/// `Material` + `InkWell` rather than `RadioListTile`, matching TimeSlotRow's own
/// reasoning (see its doc comment): full-row hover/selected fill, not just around a
/// tiny control, and no dependency on RadioListTile's now-deprecated
/// groupValue/onChanged API.
class _SelectableProviderTile extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final bool selected;
  final VoidCallback? onTap;

  const _SelectableProviderTile({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.selected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final enabled = onTap != null;

    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        hoverColor: colors.backgroundBrandTertiaryHover,
        splashColor: colors.backgroundBrandSecondary,
        child: Padding(
          padding: const EdgeInsets.symmetric(
            horizontal: LaSpacing.sm,
            vertical: LaSpacing.sm,
          ),
          child: Row(
            children: [
              Icon(
                selected
                    ? Icons.radio_button_checked
                    : Icons.radio_button_unchecked,
                color: selected
                    ? colors.contentBrandPrimary
                    : colors.contentNeutralTertiary,
                size: 22,
              ),
              const SizedBox(width: LaSpacing.sm),
              Icon(
                icon,
                color: enabled
                    ? colors.contentNeutralSecondary
                    : colors.contentNeutralTertiary,
              ),
              const SizedBox(width: LaSpacing.sm),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: LaTextStyles.labelMd.copyWith(
                        color: enabled
                            ? colors.contentNeutralPrimary
                            : colors.contentNeutralTertiary,
                      ),
                    ),
                    Text(
                      subtitle,
                      style: LaTextStyles.bodySm.copyWith(
                        color: colors.contentNeutralTertiary,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _AISettingsView extends StatefulWidget {
  const _AISettingsView();

  @override
  State<_AISettingsView> createState() => _AISettingsViewState();
}

class _AISettingsViewState extends State<_AISettingsView> {
  bool _reanalyzingAll = false;

  Future<void> _selectProvider(BuildContext context, String provider) async {
    final model = context.read<AISettingsModel>();
    final error = await model.setProvider(provider);
    if (error != null && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(error)));
    }
  }

  Future<void> _confirmReanalyzeAll() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Re-analyze all messages?'),
        content: const Text(
          'Every imported message is re-analyzed with the model currently '
          'active above. Useful after switching models, or after connecting '
          'a new API key, so past mail gets a fresh look instead of only new '
          'mail going forward. This can take a while for a large mailbox.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Re-analyze'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;

    setState(() => _reanalyzingAll = true);
    try {
      final response = await apiRequest(
        'POST',
        'v1/email-messages/reanalyze',
        true,
        null,
        {},
      );
      if (!mounted) return;
      if (response.statusCode == 202) {
        final Map<String, dynamic> data = json.decode(
          utf8.decode(response.bodyBytes),
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Re-analyzing ${data['count']} messages -- the Messages page '
              'will update as each one finishes.',
            ),
          ),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Could not start re-analysis: ${response.body}')),
        );
      }
    } catch (err, stackTrace) {
      devNotification(
        err: err,
        stackTrace: stackTrace,
        title: "AISettingsScreen._confirmReanalyzeAll()",
      );
    } finally {
      if (mounted) setState(() => _reanalyzingAll = false);
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
                  'running on our server. Connect your own Claude or Gemini API '
                  'key to use it instead -- useful if the local model\'s results '
                  'aren\'t reliable enough for you.',
                  style: LaTextStyles.bodySm.copyWith(
                    color: colors.contentNeutralSecondary,
                  ),
                ),
                const SizedBox(height: LaSpacing.base),
                Card(
                  child: _SelectableProviderTile(
                    icon: Icons.dns_outlined,
                    title: 'Local model',
                    subtitle: 'Runs on our server, no setup needed',
                    selected: settings.provider == 'ollama',
                    onTap: model.busy
                        ? null
                        : () => _selectProvider(context, 'ollama'),
                  ),
                ),
                for (final provider in _providers) ...[
                  const SizedBox(height: LaSpacing.sm),
                  _ProviderCard(provider: provider),
                ],
                const SizedBox(height: LaSpacing.lg),
                Text(
                  'Maintenance',
                  style: LaTextStyles.labelMd.copyWith(
                    color: colors.contentNeutralPrimary,
                  ),
                ),
                const SizedBox(height: LaSpacing.x2xs),
                Card(
                  child: ListTile(
                    leading: _reanalyzingAll
                        ? const SizedBox(
                            width: 24,
                            height: 24,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Icon(
                            Icons.refresh,
                            color: colors.contentNeutralSecondary,
                          ),
                    title: const Text('Re-analyze all messages'),
                    subtitle: const Text(
                      'Re-run analysis on every imported message with the '
                      'active model above',
                    ),
                    onTap: _reanalyzingAll ? null : _confirmReanalyzeAll,
                  ),
                ),
              ],
            ),
    );
  }
}

class _ProviderCard extends StatefulWidget {
  final _ProviderInfo provider;

  const _ProviderCard({required this.provider});

  @override
  State<_ProviderCard> createState() => _ProviderCardState();
}

class _ProviderCardState extends State<_ProviderCard> {
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
    final error = await context.read<AISettingsModel>().connectKey(
      widget.provider.id,
      apiKey,
    );
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
        title: Text('Disconnect ${widget.provider.label}?'),
        content: const Text(
          'Email analysis will go back to using the local model, if this was '
          'active. You can reconnect a key at any time.',
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
      await context.read<AISettingsModel>().disconnectKey(widget.provider.id);
    }
  }

  Future<void> _select() async {
    final model = context.read<AISettingsModel>();
    final error = await model.setProvider(widget.provider.id);
    if (error != null && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(error)));
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    final model = context.watch<AISettingsModel>();
    final settings = model.settings;
    final hasKey = settings.hasKeyFor(widget.provider.id);

    return Card(
      child: Column(
        children: [
          _SelectableProviderTile(
            icon: Icons.auto_awesome,
            title: widget.provider.label,
            subtitle: hasKey ? 'API key connected' : 'No API key connected',
            selected: settings.provider == widget.provider.id,
            // Can't select a provider with no key connected.
            onTap: (!hasKey || model.busy) ? null : _select,
          ),
          if (hasKey)
            Padding(
              padding: const EdgeInsets.fromLTRB(
                LaSpacing.base,
                0,
                LaSpacing.sm,
                LaSpacing.x2xs,
              ),
              child: Align(
                alignment: Alignment.centerRight,
                child: TextButton.icon(
                  onPressed: model.busy ? null : _confirmDisconnect,
                  icon: const Icon(Icons.link_off, size: 16),
                  label: const Text('Disconnect'),
                ),
              ),
            )
          else
            Padding(
              padding: const EdgeInsets.fromLTRB(
                LaSpacing.base,
                0,
                LaSpacing.base,
                LaSpacing.sm,
              ),
              child: _editingKey
                  ? Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        TextFormField(
                          controller: _apiKeyController,
                          autofocus: true,
                          obscureText: true,
                          decoration: InputDecoration(
                            labelText: widget.provider.keyLabel,
                            hintText: widget.provider.keyHint,
                            border: const OutlineInputBorder(),
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
                    )
                  : Align(
                      alignment: Alignment.centerLeft,
                      child: OutlinedButton.icon(
                        onPressed: () => setState(() => _editingKey = true),
                        icon: const Icon(Icons.add),
                        label: Text('Connect ${widget.provider.label} API key'),
                      ),
                    ),
            ),
        ],
      ),
    );
  }
}
