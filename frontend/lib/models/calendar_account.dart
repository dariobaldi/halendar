import 'dart:async';
import 'dart:convert';

import 'package:halendar_front/services/api.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:halendar_front/services/websocket.dart';
import 'package:flutter/foundation.dart';

/// One calendar the user has connected (Google Calendar, a CalDAV server such as
/// Apple iCloud, ...), checked for availability when a proposal is generated.
class CalendarAccount {
  final String id;
  final String provider;
  final String displayName;
  final String status; // active | error | revoked
  final String? lastError;
  final DateTime createdAt;

  CalendarAccount({
    required this.id,
    required this.provider,
    required this.displayName,
    required this.status,
    this.lastError,
    required this.createdAt,
  });

  bool get isActive => status == 'active';

  factory CalendarAccount.fromJson(Map<String, dynamic> json) {
    return CalendarAccount(
      id: json['id'] as String,
      provider: json['provider'] as String,
      displayName: json['display_name'] as String,
      status: json['status'] as String,
      lastError: json['last_error'] as String?,
      createdAt: DateTime.parse(json['created_at']).toLocal(),
    );
  }
}

class CalendarAccountsModel extends ChangeNotifier {
  List<CalendarAccount> _accounts = [];
  bool _fetching = true;

  StreamSubscription? _ws;

  List<CalendarAccount> get accounts => _accounts;
  bool get fetching => _fetching;

  void init() {
    fetch();
    _ws = AuthService.instance.ws.stream.listen((message) {
      final Map<String, dynamic> data = jsonDecode(message);
      if (data.containsKey("reconnected") ||
          (validTypeMessage("calendar_accounts", data) &&
              data['refresh'] == true)) {
        fetch();
      }
    });
  }

  void end() {
    _ws?.cancel();
  }

  Future<void> fetch() async {
    try {
      final response = await apiRequest(
        'GET',
        'v1/calendar-accounts',
        true,
        null,
        {},
      );
      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(
          utf8.decode(response.bodyBytes),
        );
        final List<dynamic> raw = data['accounts'];
        _accounts = raw.map((j) => CalendarAccount.fromJson(j)).toList();
      }
    } catch (err, stackTrace) {
      devNotification(
        err: err,
        stackTrace: stackTrace,
        title: "CalendarAccountsModel.fetch()",
      );
    }
    _fetching = false;
    notifyListeners();
  }

  /// Asks the backend to start a provider's OAuth2 flow, returning the URL to open in
  /// the system browser. Null on failure (a notification is already shown).
  Future<String?> beginConnect(String provider) async {
    try {
      final response = await apiRequest(
        'GET',
        'v1/calendar-accounts/$provider/connect',
        true,
        null,
        {},
      );
      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(
          utf8.decode(response.bodyBytes),
        );
        return data['auth_url'] as String?;
      }
      addNotification(
        title: "Couldn't start connection",
        content: response.body,
        type: "error",
      );
    } catch (err, stackTrace) {
      devNotification(
        err: err,
        stackTrace: stackTrace,
        title: "CalendarAccountsModel.beginConnect()",
      );
    }
    return null;
  }

  /// Connects a CalDAV calendar directly (Apple iCloud, a groupware's calendar, ...):
  /// no redirect flow, the backend tests the credentials before storing anything.
  /// Returns an error message on failure, or null on success.
  Future<String?> connectCaldav({
    required String url,
    required String user,
    required String pass,
    String calendar = '',
    String timezone = '',
  }) async {
    try {
      final response = await apiRequest(
        'POST',
        'v1/calendar-accounts/caldav',
        true,
        json.encode({
          'url': url,
          'user': user,
          'pass': pass,
          'calendar': calendar,
          'timezone': timezone,
        }),
        {},
      );
      if (response.statusCode == 201) {
        await fetch();
        return null;
      }
      final body = utf8.decode(response.bodyBytes);
      try {
        final decoded = json.decode(body);
        final error = decoded['error'];
        if (error is String) return error;
        if (error is Map) return error.values.join(', ');
      } catch (_) {}
      return body;
    } catch (err, stackTrace) {
      devNotification(
        err: err,
        stackTrace: stackTrace,
        title: "CalendarAccountsModel.connectCaldav()",
      );
      return 'Something went wrong. Please try again.';
    }
  }

  Future<void> disconnect(CalendarAccount account) async {
    try {
      final response = await apiRequest(
        'DELETE',
        'v1/calendar-accounts/${account.id}',
        true,
        null,
        {},
      );
      if (response.statusCode == 200) {
        _accounts.removeWhere((a) => a.id == account.id);
        notifyListeners();
      } else {
        addNotification(
          title: "Couldn't disconnect calendar",
          content: response.body,
          type: "error",
        );
      }
    } catch (err, stackTrace) {
      devNotification(
        err: err,
        stackTrace: stackTrace,
        title: "CalendarAccountsModel.disconnect()",
      );
    }
  }
}
