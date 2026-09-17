import 'dart:async';
import 'dart:convert';

import 'package:halendar_front/services/api.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:halendar_front/services/websocket.dart';
import 'package:flutter/foundation.dart';

/// One messaging account the user has connected (Gmail today, more providers
/// later). Credentials never appear here -- the backend keeps those encrypted
/// and only ever exposes this summary.
class EmailAccount {
  final String id;
  final String provider;
  final String emailAddress;
  final String status; // active | error | revoked
  final String? lastError;
  final DateTime? lastSyncedAt;
  final DateTime createdAt;

  EmailAccount({
    required this.id,
    required this.provider,
    required this.emailAddress,
    required this.status,
    this.lastError,
    this.lastSyncedAt,
    required this.createdAt,
  });

  bool get isActive => status == 'active';

  factory EmailAccount.fromJson(Map<String, dynamic> json) {
    return EmailAccount(
      id: json['id'] as String,
      provider: json['provider'] as String,
      emailAddress: json['email_address'] as String,
      status: json['status'] as String,
      lastError: json['last_error'] as String?,
      lastSyncedAt: json['last_synced_at'] == null
          ? null
          : DateTime.parse(json['last_synced_at']).toLocal(),
      createdAt: DateTime.parse(json['created_at']).toLocal(),
    );
  }
}

class EmailAccountsModel extends ChangeNotifier {
  List<EmailAccount> _accounts = [];
  bool _fetching = true;

  StreamSubscription? _ws;

  List<EmailAccount> get accounts => _accounts;
  bool get fetching => _fetching;

  void init() {
    fetch();
    _ws = AuthService.instance.ws.stream.listen((message) {
      final Map<String, dynamic> data = jsonDecode(message);
      if (data.containsKey("reconnected") ||
          (validTypeMessage("email_accounts", data) && data['refresh'] == true)) {
        fetch();
      }
    });
  }

  void end() {
    _ws?.cancel();
  }

  Future<void> fetch() async {
    try {
      final response = await apiRequest('GET', 'v1/email-accounts', true, null, {});
      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(utf8.decode(response.bodyBytes));
        final List<dynamic> raw = data['accounts'];
        _accounts = raw.map((j) => EmailAccount.fromJson(j)).toList();
      }
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "EmailAccountsModel.fetch()");
    }
    _fetching = false;
    notifyListeners();
  }

  /// Asks the backend to start a provider's OAuth2 flow, returning the URL to open in
  /// a webview. Null on failure (a notification is already shown).
  Future<String?> beginConnect(String provider) async {
    try {
      final response = await apiRequest(
        'GET',
        'v1/email-accounts/$provider/connect',
        true,
        null,
        {},
      );
      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(utf8.decode(response.bodyBytes));
        return data['auth_url'] as String?;
      }
      addNotification(
        title: "Couldn't start connection",
        content: response.body,
        type: "error",
      );
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "EmailAccountsModel.beginConnect()");
    }
    return null;
  }

  Future<void> disconnect(EmailAccount account) async {
    try {
      final response = await apiRequest(
        'DELETE',
        'v1/email-accounts/${account.id}',
        true,
        null,
        {},
      );
      if (response.statusCode == 200) {
        _accounts.removeWhere((a) => a.id == account.id);
        notifyListeners();
      } else {
        addNotification(
          title: "Couldn't disconnect account",
          content: response.body,
          type: "error",
        );
      }
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "EmailAccountsModel.disconnect()");
    }
  }
}
