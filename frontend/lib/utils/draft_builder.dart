import '../models/time_slot.dart';
import 'date_format.dart';

/// Builds the AI-drafted reply. When [slot] is null, none of the sender's
/// proposed times were free, so the draft simply declines and asks for
/// other options -- kept generic on purpose, no calendar lookups involved.
String buildResponseDraft({
  required String senderFirstName,
  required TimeSlot? slot,
}) {
  final body = slot == null
      ? "Sorry, none of the proposed times work for me. Could you "
          'suggest some other options?'
      : '${formatDay(slot.start)} at '
          '${formatTimeRange(slot.start, slot.end).split('–').first} '
          'works great for me.';

  return 'Hi $senderFirstName,\n\n$body\n\nBest regards,\n'
      "Sent by Ariane's Assistant";
}

/// Builds a generic reply for when the sender wants to meet but never
/// proposed any time at all -- there is nothing to accept or decline,
/// just a request for them to suggest some options.
String buildOpenEndedDraft({required String senderFirstName}) {
  return 'Hi $senderFirstName,\n\nSounds good -- could you let me know '
      "what times work for you?\n\nBest regards,\nSent by Ariane's Assistant";
}
