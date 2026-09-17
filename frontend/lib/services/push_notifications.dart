import 'dart:convert';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';

import 'api.dart';
import 'auth.dart';

const _androidNotificationChannel = AndroidNotificationChannel(
  'proposals',
  'Meeting proposals',
  description: 'New meeting requests detected in your inbox.',
  importance: Importance.high,
);

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
  final _localNotifications = FlutterLocalNotificationsPlugin();

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
        await _initLocalNotifications();

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
    _setPendingProposalId(message.data['proposal_id']);
  }

  void _setPendingProposalId(Object? proposalId) {
    if (proposalId is String && proposalId.isNotEmpty) {
      pendingProposalId.value = proposalId;
    }
  }

  // Registers the native Android notification channel used for proposal
  // pushes and wires up tap handling for notifications shown by
  // _onForegroundMessage below (tapping a system notification for a
  // backgrounded/closed app is handled separately, via onMessageOpenedApp /
  // getInitialMessage).
  Future<void> _initLocalNotifications() async {
    await _localNotifications
        .resolvePlatformSpecificImplementation<
          AndroidFlutterLocalNotificationsPlugin
        >()
        ?.createNotificationChannel(_androidNotificationChannel);

    await _localNotifications.initialize(
      settings: const InitializationSettings(
        android: AndroidInitializationSettings('@mipmap/ic_launcher'),
      ),
      onDidReceiveNotificationResponse: (response) {
        _setPendingProposalId(response.payload);
      },
    );
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
  // foreground (by design, so the app can decide how to present it) -- show it
  // as a native Android notification ourselves via flutter_local_notifications
  // instead, so the experience matches a backgrounded/closed app exactly.
  //
  // Tapping it goes through onDidReceiveNotificationResponse (see
  // _initLocalNotifications), which needs to behave the same as tapping a
  // real system notification would (see _handleNotificationTap) -- otherwise
  // a proposal notification that happens to arrive while the app is already
  // open would be the one case where tapping it doesn't take you to the
  // message.
  Future<void> _onForegroundMessage(RemoteMessage message) async {
    final notification = message.notification;
    if (notification == null) return;

    await _localNotifications.show(
      // Android notification ids are 32-bit; a plain object hashCode isn't
      // guaranteed to fit, so derive one from the clock instead.
      id: DateTime.now().millisecondsSinceEpoch.remainder(1 << 31),
      title: notification.title ?? 'Halendar',
      body: notification.body ?? '',
      notificationDetails: NotificationDetails(
        android: AndroidNotificationDetails(
          _androidNotificationChannel.id,
          _androidNotificationChannel.name,
          channelDescription: _androidNotificationChannel.description,
          importance: _androidNotificationChannel.importance,
          priority: Priority.high,
        ),
      ),
      payload: message.data['type'] == 'proposal'
          ? message.data['proposal_id']
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
