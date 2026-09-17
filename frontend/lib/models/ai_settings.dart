import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:halendar_front/services/api.dart';
import 'package:halendar_front/services/auth.dart';

/// Which model email analysis uses for the current user: the shared local Ollama
/// instance by default, or their own connected Claude or Gemini API key once
/// activated. A user can have both a Claude and a Gemini key on file at once and
/// switch between them freely -- only one is ever active at a time.
class AISettings {
  final String provider; // 'ollama' | 'claude' | 'gemini'
  final bool hasClaudeKey;
  final bool hasGeminiKey;

  const AISettings({
    required this.provider,
    required this.hasClaudeKey,
    required this.hasGeminiKey,
  });

  static const fallback = AISettings(
    provider: 'ollama',
    hasClaudeKey: false,
    hasGeminiKey: false,
  );

  bool hasKeyFor(String provider) => switch (provider) {
    'claude' => hasClaudeKey,
    'gemini' => hasGeminiKey,
    _ => true, // ollama needs no key
  };

  factory AISettings.fromJson(Map<String, dynamic> json) {
    return AISettings(
      provider: json['provider'] as String,
      hasClaudeKey: json['has_claude_key'] as bool? ?? false,
      hasGeminiKey: json['has_gemini_key'] as bool? ?? false,
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

  /// Saves (or replaces) the user's API key for provider ('claude' or 'gemini'). The
  /// backend verifies it with a live request before storing it, so a typo surfaces
  /// here rather than on the next background analysis pass. Doesn't itself activate
  /// the provider -- call [setProvider] separately. Returns null on success, an
  /// error message otherwise.
  Future<String?> connectKey(String provider, String apiKey) =>
      _mutate('PUT', 'v1/ai-settings/$provider/key', {'api_key': apiKey});

  /// Removes the stored key for provider, which also forces the active provider
  /// back to Ollama server-side if it was the one just disconnected.
  Future<String?> disconnectKey(String provider) =>
      _mutate('DELETE', 'v1/ai-settings/$provider/key', null);

  /// Switches which model future analysis uses. The backend rejects switching to
  /// 'claude'/'gemini' if no key is on file for it.
  Future<String?> setProvider(String provider) =>
      _mutate('PATCH', 'v1/ai-settings/provider', {'provider': provider});

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
