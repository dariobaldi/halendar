import 'dart:convert';

import 'package:halendar_front/env.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:web_socket_channel/web_socket_channel.dart';

Future<http.Response> apiRequest(String method, String endpoint, bool auth,
    Object? body, Map<String, dynamic> queryParameters) {
  final Uri uri = kDebugMode
      ? Uri.http(backendURL, endpoint, queryParameters)
      : Uri.https(backendURL, endpoint, queryParameters);

  late Map<String, String> headers;
  if (auth) {
    final token = AuthService.instance.token;
    headers = {'Authorization': 'Bearer $token'};
  } else {
    headers = {};
  }

  switch (method) {
    case "GET":
      return http.get(uri, headers: headers);
    case "PATCH":
      return http.patch(uri, headers: headers, body: body);
    case "POST":
      return http.post(uri, headers: headers, body: body);
    case "PUT":
      return http.put(uri, headers: headers, body: body);
    case "DELETE":
      return http.delete(uri, headers: headers, body: body);
    default:
      return http.get(uri, headers: headers);
  }
}

Future<WebSocketChannel?> apiWebSocket(String channel) async {
  final Uri? uri = await getWsTokenUri(channel);
  if (uri == null){
    return null;
  }

  return WebSocketChannel.connect(uri);
}

Future<Uri?> getWsTokenUri(String channel) async {
  var response = await apiRequest(
    "POST",
    "v1/websocket/token",
    true,
    null,
    {},
  );

  final Map<String, dynamic> responseData = json.decode(response.body);
  final String? token = responseData['token'];

  if (token == null){
    return null;
  }

  final Uri uri = kDebugMode
      ? Uri.parse('ws://$backendURL/v1/ws/$channel/$token')
      : Uri.parse('wss://$backendURL/v1/ws/$channel/$token');
  return uri;
}

Future<int> forceImportOrders() async {
  final response = await apiRequest('POST', 'v1/orders/import', true, null, {});
  return response.statusCode;
}
