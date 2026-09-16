import 'package:intl/intl.dart';

import '../models/time_slot.dart';

/// Formats a reception date the way an inbox usually does: relative for
/// today/yesterday, absolute otherwise.
String formatReceivedAt(DateTime dateTime) {
  final now = DateTime.now();
  final today = DateTime(now.year, now.month, now.day);
  final day = DateTime(dateTime.year, dateTime.month, dateTime.day);
  final time = DateFormat.Hm('en_US').format(dateTime);

  if (day == today) {
    return 'Received today at $time';
  }
  if (day == today.subtract(const Duration(days: 1))) {
    return 'Received yesterday at $time';
  }
  final date = DateFormat('MMMM d', 'en_US').format(dateTime);
  return 'Received on $date at $time';
}

/// "Friday, September 18"
String formatDay(DateTime dateTime) {
  return DateFormat('EEEE, MMMM d', 'en_US').format(dateTime);
}

/// "10:00-11:00"
String formatTimeRange(DateTime start, DateTime end) {
  final startStr = DateFormat.Hm('en_US').format(start);
  final endStr = DateFormat.Hm('en_US').format(end);
  return '$startStr–$endStr';
}

/// "Friday, September 18, 10:00-11:00"
String formatSlot(TimeSlot slot) {
  return '${formatDay(slot.start)}, ${formatTimeRange(slot.start, slot.end)}';
}
