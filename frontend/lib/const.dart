import 'package:halendar_front/models/users.dart';
import 'package:flutter/material.dart';

const String uuidNil = "00000000-0000-0000-0000-000000000000";

User unkownUser = User(
  id: "",
  createdAt: DateTime.now(),
  name: "?",
  email: "?",
  username: "?",
  activated: false,
  accessLevel: 0,
  version: 0,
);

String getUserName(String id, List<User> users) {
  return users
      .firstWhere((user) => user.id == id, orElse: () => unkownUser)
      .name;
}

Widget accessLevelIcon(int level) {
  late IconData icon;
  if (level < 5) {
    icon = Icons.person;
  } else if (level <= 5) {
    icon = Icons.person_add;
  } else if (level <= 10) {
    icon = Icons.engineering;
  } else if (level <= 100) {
    icon = Icons.computer;
  } else {
    icon = Icons.question_answer;
  }
  return Icon(icon);
}

String accessLevelName(int level) {
  if (level < 5) {
    return "Employee";
  } else if (level <= 5) {
    return "Manager";
  } else if (level <= 10) {
    return "Admin";
  } else if (level <= 100) {
    return "Dev";
  } else {
    return "Unknown";
  }
}
