import 'dart:convert';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';

import 'api.dart';
import 'auth.dart';

/// Wires this device up to receive push notifications: initializes Firebase,
/// requests notification permission, and keeps the backend's record of this
/// device's FCM token up to date (POST /v1/devices) — including whenever the
/// platform issues a refreshed token, which happens periodically for any device.
///
/// Android-only for now: Firebase.initializeApp() reads android/app/
/// google-services.json natively, with no Dart-side Firebase config needed.
class PushNotificationsService {
  PushNotificationsService._();
  static final instance = PushNotificationsService._();

  bool _initialized = false;

  /// Set whenever a notification is tapped and carries a proposal to jump to --
  /// tapping works whether the app was backgrounded or fully closed, so this needs
  /// to survive until whatever UI is ready (HomeShell may not exist yet at the
  /// moment a cold-start tap is detected) reads and clears it.
  static final ValueNotifier<String?> pendingProposalId = ValueNotifier(null);

  /// Call once a user is signed in (the backend scopes devices to a user, so
  /// there's nothing to register before that). Safe to call again on every
  /// login — Firebase/listener setup only happens once per app run.
  Future<void> registerForUser() async {
    if (kIsWeb) return;

    try {
      if (!_initialized) {
        await Firebase.initializeApp();

        final settings = await FirebaseMessaging.instance.requestPermission();
        if (settings.authorizationStatus == AuthorizationStatus.denied) {
          return;
        }

        FirebaseMessaging.instance.onTokenRefresh.listen(_sendTokenToBackend);
        FirebaseMessaging.onMessage.listen(_onForegroundMessage);
        // The app was backgrounded (not closed) and got brought back to the
        // foreground by a notification tap.
        FirebaseMessaging.onMessageOpenedApp.listen(_handleNotificationTap);
        _initialized = true;

        // The app was fully closed and this tap is what launched it -- checked
        // once, right after the listeners above are in place.
        final initialMessage = await FirebaseMessaging.instance
            .getInitialMessage();
        if (initialMessage != null) {
          _handleNotificationTap(initialMessage);
        }
      }

      final token = await FirebaseMessaging.instance.getToken();
      if (token != null) {
        await _sendTokenToBackend(token);
      }
    } catch (err, stackTrace) {
      devNotification(
        err: err,
        stackTrace: stackTrace,
        title: "PushNotificationsService.registerForUser()",
        showInScreen: false,
      );
    }
  }

  // The backend attaches {"type": "proposal", "proposal_id": "..."} to a new
  // meeting request's notification (see notifyNewProposal in the Go backend) --
  // HomeShell picks this up to switch to the Proposals tab and focus that card.
  void _handleNotificationTap(RemoteMessage message) {
    final proposalId = message.data['proposal_id'];
    if (proposalId is String && proposalId.isNotEmpty) {
      pendingProposalId.value = proposalId;
    }
  }

  Future<void> _sendTokenToBackend(String token) async {
    if (AuthService.instance.token.isEmpty) return;
    final response = await apiRequest(
      'POST',
      'v1/devices',
      true,
      json.encode({'push_token': token, 'platform': 'android'}),
      {},
    );
    if (response.statusCode != 200) {
      debugPrint('Failed to register device for push: ${response.body}');
    }
  }

  // FCM does not display a system notification while the app is in the
  // foreground (by design, so the app can decide how to present it) — show it
  // as an in-app banner via the existing notification system instead. For a
  // native system-style banner even in the foreground, add
  // flutter_local_notifications and show it from here.
  //
  // Tapping this banner needs to behave the same as tapping a real system
  // notification would (see _handleNotificationTap) -- otherwise a proposal
  // notification that happens to arrive while the app is already open would be the
  // one case where tapping it doesn't take you to the message.
  void _onForegroundMessage(RemoteMessage message) {
    addNotification(
      title: message.notification?.title ?? 'Halendar',
      content: message.notification?.body ?? '',
      type: 'info',
      onTap: message.data['type'] == 'proposal'
          ? () => _handleNotificationTap(message)
          : null,
    );
  }

  /// Call before AuthService.logOut() clears the session, so this device stops
  /// receiving notifications meant for the now-signed-out account.
  Future<void> unregister() async {
    if (kIsWeb || !_initialized) return;
    try {
      final token = await FirebaseMessaging.instance.getToken();
      if (token == null || AuthService.instance.token.isEmpty) return;
      await apiRequest(
        'DELETE',
        'v1/devices',
        true,
        json.encode({'push_token': token}),
        {},
      );
    } catch (err, stackTrace) {
      devNotification(
        err: err,
        stackTrace: stackTrace,
        title: "PushNotificationsService.unregister()",
        showInScreen: false,
      );
    }
  }
}
