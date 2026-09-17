import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:halendar_front/services/api.dart';
import 'package:halendar_front/services/auth.dart';

/// Which model email analysis uses for the current user: the shared local Ollama
/// instance by default, or their own connected Claude API key once activated.
class AISettings {
  final String provider; // 'ollama' | 'claude'
  final bool hasApiKey;

  const AISettings({required this.provider, required this.hasApiKey});

  static const fallback = AISettings(provider: 'ollama', hasApiKey: false);

  bool get isClaudeActive => provider == 'claude';

  factory AISettings.fromJson(Map<String, dynamic> json) {
    return AISettings(
      provider: json['provider'] as String,
      hasApiKey: json['has_api_key'] as bool? ?? false,
    );
  }
}

class AISettingsModel extends ChangeNotifier {
  AISettings _settings = AISettings.fallback;
  bool _fetching = true;
  bool _busy = false;

  AISettings get settings => _settings;
  bool get fetching => _fetching;
  bool get busy => _busy;

  Future<void> fetch() async {
    try {
      final response = await apiRequest('GET', 'v1/ai-settings', true, null, {});
      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(
          utf8.decode(response.bodyBytes),
        );
        _settings = AISettings.fromJson(data['ai_settings']);
      }
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "AISettingsModel.fetch()");
    }
    _fetching = false;
    notifyListeners();
  }

  /// Saves (or replaces) the user's Claude API key. The backend verifies it with a
  /// live request before storing it, so a typo surfaces here rather than on the next
  /// background analysis pass. Doesn't itself activate Claude -- call [setProvider]
  /// separately. Returns null on success, an error message otherwise.
  Future<String?> connectClaude(String apiKey) =>
      _mutate('PUT', 'v1/ai-settings/claude-key', {'api_key': apiKey});

  /// Removes the stored key, which also forces the provider back to Ollama
  /// server-side.
  Future<String?> disconnectClaude() =>
      _mutate('DELETE', 'v1/ai-settings/claude-key', null);

  /// Switches which model future analysis uses. The backend rejects switching to
  /// 'claude' if no key is on file.
  Future<String?> setProvider(String provider) =>
      _mutate('PUT', 'v1/ai-settings/provider', {'provider': provider});

  Future<String?> _mutate(
    String method,
    String path,
    Map<String, dynamic>? body,
  ) async {
    _busy = true;
    notifyListeners();
    try {
      final response = await apiRequest(
        method,
        path,
        true,
        body == null ? null : json.encode(body),
        {},
      );
      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(
          utf8.decode(response.bodyBytes),
        );
        _settings = AISettings.fromJson(data['ai_settings']);
        return null;
      }
      return _extractError(response.bodyBytes);
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "AISettingsModel.$method $path");
      return 'Something went wrong. Please try again.';
    } finally {
      _busy = false;
      notifyListeners();
    }
  }

  String _extractError(List<int> bodyBytes) {
    final body = utf8.decode(bodyBytes);
    try {
      final decoded = json.decode(body);
      final error = decoded['error'];
      if (error is String) return error;
      if (error is Map) return error.values.join(', ');
    } catch (_) {}
    return body;
  }
}
