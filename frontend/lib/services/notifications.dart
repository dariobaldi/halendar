import 'package:halendar_front/services/auth.dart';
import 'package:flutter/material.dart';

class HalendarNotification {
  final String id;
  final String title;
  final String content;
  final String imageUrl;
  final String type;
  final int duration;

  HalendarNotification({
    required this.title,
    required this.content,
    required this.imageUrl,
    this.type = "",
    this.duration = 15,
  }) : id = DateTime.now().microsecondsSinceEpoch.toString();
}

Color notificationColor(BuildContext context, String type) {
  switch (type) {
    case "success":
      return Theme.of(context).colorScheme.primary;
    case "on_success":
      return Theme.of(context).colorScheme.onPrimary;
    case "error":
      return Theme.of(context).colorScheme.error;
    case "on_error":
      return Theme.of(context).colorScheme.onError;
  }
  if (type.startsWith("on_")) {
    return Theme.of(context).colorScheme.onSecondary;
  } else {
    return Theme.of(context).colorScheme.secondary;
  }
}

Widget buildNotificationCard(
  HalendarNotification notification,
  BuildContext context,
  Animation<double> animation,
) {
  return SizeTransition(
    sizeFactor: animation,
    child: FadeTransition(
      opacity: animation,
      child: Padding(
        padding: const EdgeInsets.only(bottom: 8.0),
        child: Material(
          elevation: 4,
          borderRadius: BorderRadius.circular(12),
          color: notificationColor(context, notification.type),
          child: SelectionArea(
            child: Container(
              padding: const EdgeInsets.all(8),
              child: Row(
                spacing: 3,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (notification.imageUrl != "")
                    Padding(
                      padding: const EdgeInsets.only(right: 10),
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.network(
                          "${notification.imageUrl}&width=70",
                          width: 60,
                          height: 60,
                          fit: BoxFit.cover,
                          errorBuilder:
                              (
                                BuildContext context,
                                Object exception,
                                StackTrace? stackTrace,
                              ) {
                                return const SizedBox(
                                  width: 60,
                                  height: 60,
                                  child: Icon(Icons.broken_image),
                                );
                              },
                        ),
                      ),
                    ),

                  Expanded(
                    child: Column(
                      spacing: 3,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          spacing: 3,
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            IconButton(
                              onPressed: () {
                                AuthService.instance.removeNotification(
                                  notification.id,
                                );
                              },
                              icon: const Icon(Icons.close),
                              padding: EdgeInsets.zero,
                              constraints: const BoxConstraints(),
                              color: notificationColor(
                                context,
                                "on_${notification.type}",
                              ),
                            ),

                            const SizedBox(width: 5),

                            Expanded(
                              child: Text(
                                notification.title,
                                style: TextStyle(
                                  color: notificationColor(
                                    context,
                                    "on_${notification.type}",
                                  ),
                                  fontSize: 18,
                                ),
                              ),
                            ),
                          ],
                        ),

                        if (notification.content != "")
                          const SizedBox(height: 4),

                        if (notification.content != "")
                          Text(
                            notification.content,
                            softWrap: true,
                            style: TextStyle(
                              color: notificationColor(
                                context,
                                "on_${notification.type}",
                              ),
                              fontSize: 22,
                            ),
                          ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    ),
  );
}
