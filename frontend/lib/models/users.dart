import 'dart:async';
import 'dart:convert';

import 'package:halendar_front/services/api.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:halendar_front/services/websocket.dart';
import 'package:flutter/material.dart';

class User {
  String id; // UUID
  DateTime createdAt;
  String name;
  String email;
  String username;
  bool activated;
  int accessLevel;
  int version;

  User({
    required this.id,
    required this.createdAt,
    required this.name,
    required this.email,
    required this.username,
    required this.activated,
    required this.accessLevel,
    required this.version,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String,
      createdAt: DateTime.parse(json['created_at']).toLocal(),
      name: json['name'] as String,
      email: json['email'] as String,
      username: json['username'] as String,
      activated: json['activated'] as bool,
      accessLevel: json['access_level'] as int,
      version: json['version'] as int,
    );
  }
}

class UsersModel extends ChangeNotifier {
  List<User> _users = List.empty();
  bool _fetching = true;

  StreamSubscription? ws;

  UsersModel();

  List<User> get users => _users;
  bool get fetching => _fetching;

  void init() async {
    fetchUsers();
    wsListen();
  }

  void end() async {
    ws?.cancel();
  }

  void refresh() async {
    await fetchUsers();
    notifyListeners();
  }

  void wsListen() {
    ws = AuthService.instance.ws.stream.listen((message) {
      final Map<String, dynamic> messageData = jsonDecode(message);
      if (messageData.containsKey("reconnected")) {
        refresh();
        return;
      }
      if (!validTypeMessage("users", messageData)) return;
      if (messageData.containsKey("refresh")) {
        refresh();
        return;
      }
    });
  }

  Future<void> fetchUsers() async {
    await getAllUsers();
    _fetching = false;
    notifyListeners();
  }

  Future<void> getAllUsers() async {
    try {
      final response = await apiRequest('GET', 'v1/users', true, null, {});
      if (response.statusCode == 200) {
        final Map<String, dynamic> responseData = json.decode(
          utf8.decode(response.bodyBytes),
        );

        final List<dynamic> usersData = responseData['users'];
        _users = usersData.map((json) => User.fromJson(json)).toList();
        return;
      }
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "fetchUsers()");
    }
  }
}

Future<void> activateUser(User user) async {
  try {
    final response = await apiRequest(
      'PUT',
      'v1/users/activate',
      true,
      json.encode({
        'user_id': user.id,
        'activated': user.activated,
        'version': user.version,
      }),
      {},
    );
    if (response.statusCode != 200) {
      addNotification(
        title: "Error",
        content: "Error: ${response.body}",
        type: "error",
      );
    }
  } catch (err, stackTrace) {
    devNotification(err: err, stackTrace: stackTrace, title: "activateUser()");
  }
}

Future<void> changeUserLevel(User user, int level) async {
  try {
    final response = await apiRequest(
      'PATCH',
      'v1/user',
      true,
      json.encode({
        'id': user.id,
        'access_level': level,
        'version': user.version,
      }),
      {},
    );
    if (response.statusCode != 200) {
      addNotification(
        title: "Error",
        content: "Error: ${response.body}",
        type: "error",
      );
    }
  } catch (err, stackTrace) {
    devNotification(
      err: err,
      stackTrace: stackTrace,
      title: "changeUserLevel()",
    );
  }
}

Future<void> registerUser(User user, String password) async {
  try {
    final response = await apiRequest(
      'POST',
      'v1/users',
      true,
      json.encode({
        'name': user.name,
        'email': "${user.username}@jumajo.fr",
        'username': user.username,
        'password': password,
      }),
      {},
    );
    if (response.statusCode != 202) {
      addNotification(
        title: "Error",
        content: "Error: ${response.body}",
        type: "error",
      );
    }
  } catch (err, stackTrace) {
    devNotification(err: err, stackTrace: stackTrace, title: "registerUser()");
  }
}
