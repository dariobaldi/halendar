import 'package:flutter/foundation.dart';

import '../data/mock_data.dart';
import '../models/proposal.dart';
import '../models/time_slot.dart';
import '../utils/draft_builder.dart';

/// Single source of truth for proposals during the demo. Every mutation
/// here corresponds to an explicit user action -- nothing changes status
/// on its own.
class ProposalsStore extends ChangeNotifier {
  final List<Proposal> _proposals = buildMockProposals();

  List<Proposal> get needsAction =>
      _proposals.where((p) => p.status == ProposalStatus.pending).toList();

  List<Proposal> get history => _proposals
      .where(
        (p) => p.status == ProposalStatus.confirmed ||
            p.status == ProposalStatus.rejected,
      )
      .toList();

  Proposal byId(String id) => _proposals.firstWhere((p) => p.id == id);

  /// Selects a different free slot. Regenerates the draft to reference
  /// the new time, unless the user has already edited it by hand.
  void selectSlot(String id, TimeSlot slot) {
    final proposal = byId(id);
    proposal.selectedSlot = slot;
    if (!proposal.draftEditedByUser) {
      proposal.responseDraft = buildResponseDraft(
        senderFirstName: proposal.senderName.split(' ').first,
        slot: slot,
      );
    }
    notifyListeners();
  }

  void updateDraft(String id, String text) {
    final proposal = byId(id);
    proposal.responseDraft = text;
    proposal.draftEditedByUser = true;
    notifyListeners();
  }

  void confirm(String id) {
    byId(id).status = ProposalStatus.confirmed;
    notifyListeners();
  }

  void reject(String id) {
    byId(id).status = ProposalStatus.rejected;
    notifyListeners();
  }
}
