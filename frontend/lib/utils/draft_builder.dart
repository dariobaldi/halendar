import '../models/time_slot.dart';
import 'date_format.dart';

/// Builds the AI-drafted reply when the user picks a different slot than the one
/// the backend auto-selected. Mirrors the backend's own phrasing/signature
/// (draftEventReply in cmd/api/email_import.go) closely enough to be a reasonable
/// stand-in without a round trip -- the user can always edit it before sending.
String buildResponseDraft({
  required String senderFirstName,
  required String userName,
  required TimeSlot? slot,
}) {
  final body = slot == null
      ? 'Sorry, none of the proposed times work for me. Could you '
          'suggest some other options?'
      : '${formatDay(slot.start)} at '
          '${formatTimeRange(slot.start, slot.end).split('–').first} '
          'works great for me.';

  return 'Hi $senderFirstName,\n\n$body\n\n—\n'
      'Sent by the Halendar Assistant on behalf of $userName.';
}
