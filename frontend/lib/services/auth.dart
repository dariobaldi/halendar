import 'dart:async';
import 'package:halendar_front/const.dart';
import 'package:halendar_front/services/notifications.dart';
import 'package:halendar_front/services/push_notifications.dart';
import 'package:halendar_front/services/websocket.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:hive_flutter/hive_flutter.dart';
import 'dart:convert';

import '../env.dart';

class AuthService {
  final _box = Hive.box('localDB');
  final ws = WebSocketMain();

  // Notifications Stuff
  final GlobalKey<AnimatedListState> _listKey = GlobalKey<AnimatedListState>();
  final List<HalendarNotification> _notifications = [];
  final StreamController<List<HalendarNotification>> _notificationsController =
      StreamController<List<HalendarNotification>>.broadcast();

  // Make Auth a Singleton
  AuthService._();
  static final instance = AuthService._();

  Future<void> init() async {
    await readUserFromBox();
  }

  AuthUser? _user;
  final StreamController<AuthUser?> _controller = StreamController<AuthUser?>();

  AuthUser? get user => _user;
  String get token => _user?.token ?? '';
  DateTime get expiry =>
      _user?.expiry ?? DateTime.now().add(const Duration(days: -1));
  int get accessLevel => _user?.accessLevel ?? 0;

  GlobalKey<AnimatedListState> get listKey => _listKey;
  List<HalendarNotification> get notifications => _notifications;
  Stream<List<HalendarNotification>> get notificationsStream =>
      _notificationsController.stream;

  // import user from Hive box
  Future<void> readUserFromBox() async {
    if (_box.get('token') == null) {
      return;
    }
    _user = AuthUser(
      id: _box.get('id') ?? uuidNil,
      name: _box.get('name'),
      username: _box.get('username') ?? "?",
      token: _box.get('token'),
      expiry: _box.get('expiry'),
      accessLevel: _box.get('access_level'),
    );
    _user = checkToken();
    if (ws.channel == null && _user != null) await ws.connect();
  }

  void saveUserToBox() {
    _box.put('id', _user?.id);
    _box.put('name', _user?.name);
    _box.put('username', _user?.username);
    _box.put('token', _user?.token);
    _box.put('expiry', _user?.expiry);
    _box.put('access_level', _user?.accessLevel);
  }

  AuthUser? checkToken() {
    bool isValid = expiry.isAfter(DateTime.now());
    bool hasToken = (token != '');
    return (isValid && hasToken) ? _user : null;
  }

  // Call to the Back-end for authentication token
  Future<int> authenticate(String username, String password) async {
    final logginType = username.contains('@') ? 'email' : 'username';
    Map<String, String> credentials = {
      logginType: username,
      'password': password,
    };

    // Convert the credentials map to JSON
    String jsonBody = json.encode(credentials);

    late Uri uri;
    // Make the HTTP request to the API endpoint
    if (kDebugMode) {
      uri = Uri.http(backendURL, 'v1/users/authentication');
    } else {
      uri = Uri.https(backendURL, 'v1/users/authentication');
    }
    var response = await http.post(
      uri,
      headers: {'Content-Type': 'application/json'},
      body: jsonBody,
    );

    // Check if the request was successful
    if (response.statusCode == 201) {
      // Parse the response body to get the token
      var data = json.decode(response.body);
      _user = AuthUser.fromJson(data);
      saveUserToBox();
      ws.connect();
    } else {
      _user = null;
    }
    _controller.add(_user);
    return response.statusCode;
  }

  void logOut() {
    PushNotificationsService.instance.unregister(); // uses the still-valid token
    _user = null;
    saveUserToBox();
    _controller.add(_user);
    ws.close();
  }

  Stream<AuthUser?> get isLoggedInStream {
    _controller.add(user);

    checkTokenStream.listen((_) {
      _controller.add(user);
    });

    return _controller.stream;
  }

  Stream<void> get checkTokenStream async* {
    while (true) {
      await Future.delayed(const Duration(hours: 1));
      yield checkToken();
    }
  }

  // Notifications Handling
  void addNotification(HalendarNotification newNotification) {
    _notifications.insert(0, newNotification);
    _notificationsController.add(_notifications);
    _listKey.currentState?.insertItem(
      0,
      duration: const Duration(milliseconds: 300),
    );
    Timer(Duration(seconds: newNotification.duration), () {
      removeNotification(newNotification.id);
    });
  }

  void removeNotification(String id) {
    final index = _notifications.indexWhere((element) => element.id == id);

    if (index >= 0) {
      final removedItem = _notifications[index];

      _notifications.removeAt(index);
      _notificationsController.add(_notifications);

      _listKey.currentState?.removeItem(
        index,
        (context, animation) =>
            buildNotificationCard(removedItem, context, animation),
        duration: const Duration(milliseconds: 300),
      );
    }
  }
}

void addNotification({
  required String title,
  required String content,
  String imageUrl = "",
  String type = "",
  int duration = 15,
}) {
  AuthService.instance.addNotification(
    HalendarNotification(
      title: title,
      content: content,
      imageUrl: imageUrl,
      type: type,
      duration: duration,
    ),
  );
}

void devNotification({
  required Object err,
  required StackTrace stackTrace,
  String title = "",
  bool showInScreen = true,
}) {
  debugPrint("----------------------------");
  debugPrint("Title: $title");
  debugPrint("ERROR: $err");
  debugPrint("STACK TRACE:");
  debugPrint(stackTrace.toString());
  debugPrint("----------------------------");

  if (showInScreen) {
    addNotification(
      title: title != "" ? "Internal error: $title" : "Internal error",
      content: "Error: $err",
      type: "error",
    );
  }
}

class AuthUser {
  String id;
  String name;
  String username;
  String token;
  DateTime expiry;
  int accessLevel;

  AuthUser({
    required this.id,
    required this.name,
    required this.username,
    required this.token,
    required this.expiry,
    required this.accessLevel,
  });

  factory AuthUser.fromJson(Map<String, dynamic> json) {
    return AuthUser(
      id: json['user']['id'],
      name: json['user']['name'],
      username: json['user']['username'],
      token: json['user']['token'],
      expiry: DateTime.parse(json['user']['expiry']),
      accessLevel: json['user']['access_level'],
    );
  }
}

var anonymousUser = AuthUser(
  id: uuidNil,
  name: "Anonime",
  username: "?",
  token: "X",
  expiry: DateTime.now(),
  accessLevel: 0,
);

bool isWorkingHours() {
  final now = DateTime.now();

  if (now.weekday == DateTime.saturday || now.weekday == DateTime.sunday) {
    return false;
  }

  return now.hour >= 7 && now.hour < 18;
}
