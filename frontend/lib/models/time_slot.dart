/// "unknown" is a real, distinct outcome from the backend: the calendar check
/// itself failed (e.g. no calendar connected, or a CalDAV/Google request errored)
/// rather than confirming the slot is actually busy. Treated as non-selectable,
/// same as busy, but shown with its own label so it doesn't read as a false "Busy".
enum SlotAvailability { free, busy, unknown }

SlotAvailability _availabilityFromJson(String value) {
  switch (value) {
    case 'free':
      return SlotAvailability.free;
    case 'busy':
      return SlotAvailability.busy;
    default:
      return SlotAvailability.unknown;
  }
}

class TimeSlot {
  /// The backend's email_event_slots.id -- sent back when the user picks this slot.
  final String id;
  final DateTime start;
  final DateTime end;
  final SlotAvailability availability;

  const TimeSlot({
    required this.id,
    required this.start,
    required this.end,
    required this.availability,
  });

  bool get isFree => availability == SlotAvailability.free;

  factory TimeSlot.fromJson(Map<String, dynamic> json) {
    return TimeSlot(
      id: json['id'] as String,
      start: DateTime.parse(json['start_at']).toLocal(),
      end: DateTime.parse(json['end_at']).toLocal(),
      availability: _availabilityFromJson(json['availability'] as String),
    );
  }
}
