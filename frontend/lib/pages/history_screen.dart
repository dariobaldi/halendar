import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

import '../state/proposals_store.dart';
import '../widgets/proposal_card.dart';
import '../widgets/theme_toggle_button.dart';

class HistoryScreen extends StatelessWidget {
  final ProposalsStore store;

  const HistoryScreen({super.key, required this.store});

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: store,
      builder: (context, _) {
        final proposals = store.history;
        final colors = context.laColors;
        return Scaffold(
          appBar: AppBar(
            title: const Text('History'),
            actions: const [ThemeToggleButton()],
          ),
          body: store.fetching && proposals.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : RefreshIndicator(
                  onRefresh: store.fetch,
                  child: proposals.isEmpty
                      ? ListView(
                          // Plain Center isn't scrollable, so the pull gesture
                          // above would never register with nothing to show yet.
                          physics: const AlwaysScrollableScrollPhysics(),
                          children: [
                            Padding(
                              padding: const EdgeInsets.symmetric(
                                horizontal: LaSpacing.lg,
                                vertical: LaSpacing.xl,
                              ),
                              child: Column(
                                children: [
                                  Icon(
                                    Icons.history,
                                    size: 48,
                                    color: colors.contentNeutralTertiary,
                                  ),
                                  const SizedBox(height: LaSpacing.sm),
                                  Text(
                                    'No history yet',
                                    style: LaTextStyles.labelLg.copyWith(
                                      color: colors.contentNeutralPrimary,
                                    ),
                                  ),
                                  const SizedBox(height: LaSpacing.x2xs),
                                  Text(
                                    'Confirmed or deleted proposals appear here.',
                                    textAlign: TextAlign.center,
                                    style: LaTextStyles.bodySm.copyWith(
                                      color: colors.contentNeutralTertiary,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        )
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
                              readOnly: true,
                              initiallyExpanded: index == 0,
                            );
                          },
                        ),
                ),
        );
      },
    );
  }
}
