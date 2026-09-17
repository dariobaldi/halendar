import 'package:flutter/material.dart';

import '../state/proposals_store.dart';
import '../widgets/empty_state.dart';
import '../widgets/proposals_view.dart';

class HistoryScreen extends StatelessWidget {
  final ProposalsStore store;

  const HistoryScreen({super.key, required this.store});

  @override
  Widget build(BuildContext context) {
    return ProposalsView(
      title: 'History',
      store: store,
      selector: (s) => s.history,
      readOnly: true,
      emptyState: const EmptyState(
        icon: Icons.history,
        title: 'No history yet',
        message: 'Confirmed or deleted proposals appear here.',
      ),
    );
  }
}
