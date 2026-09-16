enum SlotAvailability { free, busy }

class TimeSlot {
  final DateTime start;
  final DateTime end;
  final SlotAvailability availability;

  const TimeSlot({
    required this.start,
    required this.end,
    required this.availability,
  });

  bool get isFree => availability == SlotAvailability.free;
}
