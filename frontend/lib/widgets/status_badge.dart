import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

import '../models/proposal.dart';

(String, LaBadgeType) _styleFor(Proposal proposal) {
  if (proposal.status == ProposalStatus.rejected) {
    return ('Deleted', LaBadgeType.neutral);
  }
  if (proposal.status == ProposalStatus.confirmed) {
    return ('Confirmed', LaBadgeType.success);
  }
  if (proposal.needsManualReview) {
    return ('Needs review', LaBadgeType.neutral);
  }
  return ('Pending', LaBadgeType.info);
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
