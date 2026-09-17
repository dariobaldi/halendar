import 'package:flutter/material.dart';

import '../state/proposals_store.dart';
import '../widgets/empty_state.dart';
import '../widgets/proposals_view.dart';

class ProposalsListScreen extends StatelessWidget {
  final ProposalsStore store;

  const ProposalsListScreen({super.key, required this.store});

  @override
  Widget build(BuildContext context) {
    return ProposalsView(
      title: 'Messages',
      store: store,
      selector: (s) => s.needsAction,
      emptyState: const EmptyState(
        icon: Icons.inbox_outlined,
        title: 'No messages pending',
        message:
            'New meeting requests detected in your emails will appear '
            'here.',
      ),
    );
  }
}
