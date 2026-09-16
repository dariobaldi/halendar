import '../models/proposal.dart';
import '../models/time_slot.dart';
import '../utils/draft_builder.dart';

DateTime _at(DateTime day, int hour, [int minute = 0]) =>
    DateTime(day.year, day.month, day.day, hour, minute);

TimeSlot? _firstFree(List<TimeSlot> slots) {
  for (final slot in slots) {
    if (slot.isFree) return slot;
  }
  return null;
}

/// Builds a realistic set of proposals covering every case the UI needs
/// to communicate: a straightforward pending request, one where the
/// sender's own times were all busy (so the draft simply declines and
/// asks for other options), one the AI could not parse at all, one
/// already confirmed, and one already cancelled.
List<Proposal> buildMockProposals() {
  final now = DateTime.now();
  final today = DateTime(now.year, now.month, now.day);
  final in2Days = today.add(const Duration(days: 2));
  final in3Days = today.add(const Duration(days: 3));

  final claireSlots = [
    TimeSlot(
      start: _at(in2Days, 14),
      end: _at(in2Days, 15),
      availability: SlotAvailability.busy,
    ),
    TimeSlot(
      start: _at(in3Days, 10),
      end: _at(in3Days, 11),
      availability: SlotAvailability.free,
    ),
    TimeSlot(
      start: _at(in3Days, 15),
      end: _at(in3Days, 16),
      availability: SlotAvailability.free,
    ),
  ];

  final julienSlots = [
    TimeSlot(
      start: _at(in2Days, 9),
      end: _at(in2Days, 9, 30),
      availability: SlotAvailability.busy,
    ),
    TimeSlot(
      start: _at(in2Days, 11),
      end: _at(in2Days, 11, 30),
      availability: SlotAvailability.free,
    ),
  ];

  final budgetSlots = [
    TimeSlot(
      start: _at(in2Days, 9),
      end: _at(in2Days, 10),
      availability: SlotAvailability.busy,
    ),
    TimeSlot(
      start: _at(in2Days, 16),
      end: _at(in2Days, 17),
      availability: SlotAvailability.busy,
    ),
    TimeSlot(
      start: _at(in3Days, 8),
      end: _at(in3Days, 9),
      availability: SlotAvailability.busy,
    ),
  ];

  final sprintSlot = TimeSlot(
    start: _at(today.subtract(const Duration(days: 3)), 9),
    end: _at(today.subtract(const Duration(days: 3)), 9, 30),
    availability: SlotAvailability.free,
  );

  final retroSlot = TimeSlot(
    start: _at(today.subtract(const Duration(days: 5)), 17),
    end: _at(today.subtract(const Duration(days: 5)), 17, 30),
    availability: SlotAvailability.free,
  );

  final claireSelected = _firstFree(claireSlots);
  final julienSelected = _firstFree(julienSlots);
  final budgetSelected = _firstFree(budgetSlots); // null: none are free

  return [
    Proposal(
      id: 'claire-martin-project-sync',
      senderName: 'Claire Martin',
      senderEmail: 'claire.martin@example.com',
      subject: 'Project sync',
      receivedAt: _at(today, 9, 12),
      emailExcerpt:
          'Hi Ariane,\n\nWould you be available Thursday 2-3pm, or Friday '
          'either in the morning (10-11am) or late afternoon (3-4pm), for '
          "a quick sync on the project's progress?\n\nBest,\nClaire",
      slots: claireSlots,
      selectedSlot: claireSelected,
      responseDraft: buildResponseDraft(
        senderFirstName: 'Claire',
        slot: claireSelected,
      ),
    ),
    Proposal(
      id: 'julien-dupont-candidate-interview',
      senderName: 'Julien Dupont',
      senderEmail: 'julien.dupont@example.com',
      subject: 'Candidate interview',
      receivedAt: _at(today.subtract(const Duration(days: 1)), 16, 40),
      emailExcerpt:
          "Hi,\n\nFor the candidate's interview, do you have a slot at "
          '9-9:30am or 11-11:30am early next week?\n\nThanks,\nJulien',
      slots: julienSlots,
      selectedSlot: julienSelected,
      responseDraft: buildResponseDraft(
        senderFirstName: 'Julien',
        slot: julienSelected,
      ),
    ),
    Proposal(
      id: 'marc-bellamy-budget-review',
      senderName: 'Marc Bellamy',
      senderEmail: 'marc.bellamy@example.com',
      subject: 'Budget review',
      receivedAt: _at(today, 8, 5),
      emailExcerpt:
          'Hi Ariane,\n\nCan we lock in the budget review Tuesday '
          '9-10am, Tuesday 4-5pm, or Wednesday 8-9am?\n\nMarc',
      slots: budgetSlots,
      selectedSlot: budgetSelected,
      responseDraft: buildResponseDraft(
        senderFirstName: 'Marc',
        slot: budgetSelected,
      ),
    ),
    Proposal(
      id: 'sophie-lambert-client-follow-up',
      senderName: 'Sophie Lambert',
      senderEmail: 'sophie.lambert@example.com',
      subject: 'Client follow-up',
      receivedAt: _at(today, 7, 20),
      emailExcerpt:
          "Hi Ariane,\n\nWe should catch up soon on the client follow-up, "
          'let me know what works for you.\n\nSophie',
      needsManualReview: true,
      responseDraft: buildOpenEndedDraft(senderFirstName: 'Sophie'),
    ),
    Proposal(
      id: 'nadia-belkacem-sprint-kickoff',
      senderName: 'Nadia Belkacem',
      senderEmail: 'nadia.belkacem@example.com',
      subject: 'Sprint kickoff',
      receivedAt: _at(today.subtract(const Duration(days: 3)), 8, 50),
      emailExcerpt:
          "Hi Ariane,\n\nShall we lock in the sprint kickoff Thursday "
          'morning?\n\nNadia',
      slots: [sprintSlot],
      selectedSlot: sprintSlot,
      responseDraft: buildResponseDraft(
        senderFirstName: 'Nadia',
        slot: sprintSlot,
      ),
      status: ProposalStatus.confirmed,
    ),
    Proposal(
      id: 'thomas-girard-team-retro',
      senderName: 'Thomas Girard',
      senderEmail: 'thomas.girard@example.com',
      subject: 'Team retro',
      receivedAt: _at(today.subtract(const Duration(days: 5)), 11, 5),
      emailExcerpt:
          'Hi Ariane,\n\nAvailable for the retro late afternoon?\n\n'
          'Thomas',
      slots: [retroSlot],
      selectedSlot: retroSlot,
      responseDraft: buildResponseDraft(
        senderFirstName: 'Thomas',
        slot: retroSlot,
      ),
      status: ProposalStatus.rejected,
    ),
  ];
}
