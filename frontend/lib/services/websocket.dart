import 'dart:async';
import 'dart:convert';
import 'package:halendar_front/services/api.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:flutter/material.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/status.dart' as status;

class WebSocketService {
  final String name;
  WebSocketChannel? _channel;
  Uri? uri;
  Timer? _reconnectTimer;
  bool _isConnected = true;

  WebSocketService(this.name, {this.uri});

  // bool get isConnected => _isConnected;

  void connect(void Function(dynamic)? onData, Function notify) async {
    // Connect to the WebSocket
    if (uri == null) {
      _channel = await apiWebSocket(name);
    } else {
      _channel = WebSocketChannel.connect(uri!);
    }

    if (_channel != null) {
      _isConnected = true;
    } else {
      _isConnected = false;
    }
    notify();

    // Listen for incoming messages
    _channel?.stream.listen(
      onData,
      onDone: () {
        debugPrint('Connection closed');
        _isConnected = false; // Update status to disconnected
        notify();
        _scheduleReconnect(onData, notify);
      },
      onError: (error) {
        debugPrint('Error: $error');
        _isConnected = false; // Update status to disconnected
        notify();
        _scheduleReconnect(onData, notify);
      },
    );
  }

  bool isConnected() {
    return _isConnected;
  }

  void _scheduleReconnect(void Function(dynamic)? onData, Function notify) {
    // Cancel any existing reconnect attempts
    _reconnectTimer?.cancel();

    // Schedule a reconnect attempt after a delay
    _reconnectTimer = Timer(const Duration(seconds: 5), () {
      debugPrint('Attempting to reconnect...');
      connect(onData, notify); // Reconnect to the WebSocket
    });
  }

  void sendMessage(String message) {
    if (_isConnected && _channel != null) {
      _channel?.sink.add(message);
    } else {
      debugPrint('Cannot send message, not connected');
    }
  }

  void close() {
    _reconnectTimer?.cancel();
    _channel?.sink.close(status.normalClosure);
  }
}

class WebSocketMain {
  WebSocketChannel? _channel;
  Uri? _uri;
  final _controller = StreamController.broadcast();
  Timer? _reconnectTimer;
  bool _isConnected = true;
  bool _manuallyDisconnected = false;
  int _reconnectAttempts = 0;

  WebSocketMain();

  WebSocketChannel? get channel => _channel;
  bool get isConnected => _isConnected;
  Stream get stream => _controller.stream;

  Future<void> connect() async {
    try {
      _uri = await getWsTokenUri("halendar");
      if (_uri == null) {
        _isConnected = false;
        _controller.add('{"disconnected": true, "type": "internal"}');
        return;
      }
      _channel = WebSocketChannel.connect(_uri!);

      if (_channel != null) {
        _isConnected = true;
        _controller.add('{"reconnected": true, "type": "internal"}');
      } else {
        _isConnected = false;
        _controller.add('{"disconnected": true, "type": "internal"}');
      }

      // Listen for incoming messages
      _channel?.stream.listen(
        (data) {
          _reconnectAttempts = 0;
          _controller.add(data);
        },
        onDone: () {
          _controller.add('Connection closed : ${DateTime.now().toString()}');
          debugPrint('Connection closed');
          _controller.add('{"disconnected": true, "type": "internal"}');
          _isConnected = false;
          _scheduleReconnect();
        },
        onError: (error) {
          _controller.add('Error: $error');
          debugPrint('Error: $error');
          _controller.add('{"disconnected": true, "type": "internal"}');
          _isConnected = false;
          _scheduleReconnect();
        },
      );
    } catch (err, stackTrace) {
      devNotification(
        err: err,
        stackTrace: stackTrace,
        title: "WebSocketMain::connect()",
      );
      _isConnected = false;
      _controller.add('{"disconnected": true, "type": "internal"}');
      _scheduleReconnect();
    }
  }

  void _scheduleReconnect() {
    _reconnectAttempts++;
    final delay = Duration(seconds: _reconnectAttempts.clamp(1, 2));

    // Cancel any existing reconnect attempts
    _reconnectTimer?.cancel();
    // Schedule a reconnect attempt after a delay
    _reconnectTimer = Timer(delay, () {
      debugPrint('Attempting to reconnect...');
      if (!_manuallyDisconnected) {
        connect(); // Reconnect to the WebSocket
      }
    });
  }

  void reconnect() {
    _reconnectAttempts = 0;
    connect(); // Reconnect to the WebSocket
  }

  void send(dynamic data) {
    if (_channel != null) {
      _channel!.sink.add(jsonEncode(data));
    }
  }

  void close() {
    _manuallyDisconnected = true;
    _reconnectAttempts = 0;
    _reconnectTimer?.cancel();
    _channel?.sink.close(status.normalClosure);
  }
}

bool validTypeMessage(String currentType, Map<String, dynamic> messageData) {
  if (messageData.containsKey("type")) {
    String messageType = messageData["type"];
    if (currentType == messageType) return true;
  }
  return false;
}
