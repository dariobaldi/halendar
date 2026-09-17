import 'time_slot.dart';

/// Status of a meeting proposal. A proposal that needs manual review is
/// still [pending] (see [Proposal.needsManualReview]); there is no
/// separate status for it, nor for the case where none of the sender's
/// times were free -- that just produces a decline draft instead of an
/// accept draft, still within the normal pending flow.
enum ProposalStatus { pending, confirmed, rejected }

/// A meeting proposal extracted from an incoming email, enriched with
/// calendar availability and an AI-drafted reply. Nothing here has been
/// sent or written to the calendar yet -- it is only a suggestion until
/// the user confirms it.
class Proposal {
  final String id;
  final String senderName;
  final String senderEmail;
  final String subject;
  final DateTime receivedAt;
  final String emailExcerpt;

  /// The sender's proposed times, free or busy.
  final List<TimeSlot> slots;

  /// The slot the user intends to go with. Null when none of [slots] are
  /// free -- there is then nothing to confirm, only a decline to send.
  TimeSlot? selectedSlot;

  /// True when the AI could not extract any usable time reference from
  /// the email. A normal, low-frequency edge case, not a system failure.
  final bool needsManualReview;

  String responseDraft;

  /// Once the user edits the draft by hand, picking a different slot no
  /// longer silently overwrites their wording.
  bool draftEditedByUser;

  ProposalStatus status;

  Proposal({
    required this.id,
    required this.senderName,
    required this.senderEmail,
    required this.subject,
    required this.receivedAt,
    required this.emailExcerpt,
    this.slots = const [],
    this.selectedSlot,
    this.needsManualReview = false,
    this.responseDraft = '',
    this.draftEditedByUser = false,
    this.status = ProposalStatus.pending,
  });

  bool get hasFreeSlot => selectedSlot != null;

  factory Proposal.fromJson(Map<String, dynamic> json) {
    final slots = (json['slots'] as List<dynamic>? ?? [])
        .map((s) => TimeSlot.fromJson(s as Map<String, dynamic>))
        .toList();

    // The backend leaves selected_slot_id unset until the user actively picks a
    // slot different from the one it auto-selected -- the same default (first free
    // slot) applies here so the UI shows a sensible pick from the start.
    final selectedSlotId = json['selected_slot_id'] as String?;
    TimeSlot? selectedSlot;
    if (selectedSlotId != null) {
      for (final s in slots) {
        if (s.id == selectedSlotId) {
          selectedSlot = s;
          break;
        }
      }
    } else {
      for (final s in slots) {
        if (s.isFree) {
          selectedSlot = s;
          break;
        }
      }
    }

    return Proposal(
      id: json['id'] as String,
      senderName: (json['sender_name'] as String?)?.trim().isNotEmpty == true
          ? json['sender_name'] as String
          : (json['sender_email'] as String),
      senderEmail: json['sender_email'] as String,
      subject: (json['subject'] as String?)?.isNotEmpty == true
          ? json['subject'] as String
          : '(no subject)',
      receivedAt: DateTime.parse(json['received_at']).toLocal(),
      emailExcerpt: json['email_excerpt'] as String? ?? '',
      slots: slots,
      selectedSlot: selectedSlot,
      needsManualReview: json['needs_manual_review'] as bool? ?? false,
      responseDraft: json['response_draft'] as String? ?? '',
      draftEditedByUser: json['draft_edited_by_user'] as bool? ?? false,
      status: switch (json['status']) {
        'confirmed' => ProposalStatus.confirmed,
        'rejected' => ProposalStatus.rejected,
        _ => ProposalStatus.pending,
      },
    );
  }
}
