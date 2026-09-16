import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

import '../state/proposals_store.dart';
import '../widgets/proposal_card.dart';
import '../widgets/theme_toggle_button.dart';

class ProposalsListScreen extends StatelessWidget {
  final ProposalsStore store;

  const ProposalsListScreen({super.key, required this.store});

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: store,
      builder: (context, _) {
        final proposals = store.needsAction;
        return Scaffold(
          appBar: AppBar(
            title: const Text('Proposals'),
            actions: const [ThemeToggleButton()],
          ),
          body: proposals.isEmpty
              ? const _EmptyState()
              : ListView.separated(
                  padding: const EdgeInsets.all(LaSpacing.base),
                  itemCount: proposals.length,
                  separatorBuilder: (context, index) =>
                      const SizedBox(height: LaSpacing.sm),
                  itemBuilder: (context, index) {
                    final proposal = proposals[index];
                    return ProposalCard(
                      key: ValueKey(proposal.id),
                      proposal: proposal,
                      store: store,
                      initiallyExpanded: index == 0,
                    );
                  },
                ),
        );
      },
    );
  }
}

class _EmptyState extends StatelessWidget {
  const _EmptyState();

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(LaSpacing.lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.inbox_outlined,
              size: 48,
              color: colors.contentNeutralTertiary,
            ),
            const SizedBox(height: LaSpacing.sm),
            Text(
              'No proposals pending',
              style: LaTextStyles.labelLg.copyWith(
                color: colors.contentNeutralPrimary,
              ),
            ),
            const SizedBox(height: LaSpacing.x2xs),
            Text(
              'New meeting requests detected in your emails will appear '
              'here.',
              textAlign: TextAlign.center,
              style: LaTextStyles.bodySm.copyWith(
                color: colors.contentNeutralTertiary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
