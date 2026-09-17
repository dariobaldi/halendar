import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

import '../models/proposal.dart';

(String, LaBadgeType) _styleFor(Proposal proposal) {
  if (proposal.status == ProposalStatus.rejected) {
    return ('Skipped', LaBadgeType.neutral);
  }
  if (proposal.status == ProposalStatus.confirmed) {
    return ('Confirmed', LaBadgeType.success);
  }
  if (proposal.suggestedSkip) {
    return ('Suggested: skip', LaBadgeType.neutral);
  }
  if (!proposal.isMeetingRequest) {
    return ('Not a meeting', LaBadgeType.neutral);
  }
  if (proposal.needsManualReview) {
    return ('Needs review', LaBadgeType.neutral);
  }
  // "Pending" read, in real user testing, as "a reply already went out" rather than
  // "there's a draft waiting for you to send" -- naming the actual state removes the
  // ambiguity.
  return ('Draft ready', LaBadgeType.info);
}

/// A proposal's status, shown with the kit's own pill badge.
class StatusBadge extends StatelessWidget {
  final Proposal proposal;

  const StatusBadge({super.key, required this.proposal});

  @override
  Widget build(BuildContext context) {
    final (label, type) = _styleFor(proposal);
    return LaBadge(label: label, type: type);
  }
}
