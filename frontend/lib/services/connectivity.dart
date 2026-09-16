import 'dart:async';
import 'dart:convert';

import 'package:halendar_front/services/auth.dart';
import 'package:flutter/material.dart';

class NetworkAwareWidget extends StatefulWidget {
  final Widget child;

  const NetworkAwareWidget({super.key, required this.child});

  @override
  State<NetworkAwareWidget> createState() => _NetworkAwareWidgetState();
}

class _NetworkAwareWidgetState extends State<NetworkAwareWidget> {
  bool _serverConnected = true;
  StreamSubscription? ws;

  @override
  void initState() {
    super.initState();
    ws = AuthService.instance.ws.stream.listen((message) {
      final Map<String, dynamic> messageData = jsonDecode(message);
      if (messageData.containsKey("reconnected")) {
        setState(() {
          _serverConnected = true;
        });
        return;
      }
      if (messageData.containsKey("disconnected")) {
        setState(() {
          _serverConnected = false;
        });
        return;
      }
    });
  }

  @override
  void dispose() {
    ws?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Material(
      child: Stack(
        alignment: AlignmentDirectional.centerEnd,
        children: [
          widget.child,
          if (!_serverConnected && (AuthService.instance.accessLevel >= 5 || isWorkingHours()))
            Padding(
              padding: const EdgeInsets.all(8.0),
              child: Chip(
                label: Text(
                  'Déconnecté du serveur',
                  style: TextStyle(fontSize: 18),
                ),
                backgroundColor: Colors.red,
              ),
            ),
        ],
      ),
    );
  }
}
