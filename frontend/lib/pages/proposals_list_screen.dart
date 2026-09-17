import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

import '../state/proposals_store.dart';
import '../widgets/proposal_card.dart';
import '../widgets/theme_toggle_button.dart';

class ProposalsListScreen extends StatefulWidget {
  final ProposalsStore store;

  const ProposalsListScreen({super.key, required this.store});

  @override
  State<ProposalsListScreen> createState() => _ProposalsListScreenState();
}

class _ProposalsListScreenState extends State<ProposalsListScreen> {
  // GlobalKeys (rather than plain ValueKeys) so a specific card's rendered
  // BuildContext can be found later, to scroll it into view on a notification tap.
  final Map<String, GlobalKey> _cardKeys = {};

  GlobalKey _keyFor(String id) => _cardKeys.putIfAbsent(id, () => GlobalKey());

  @override
  void initState() {
    super.initState();
    widget.store.addListener(_maybeScrollToFocused);
    _maybeScrollToFocused();
  }

  @override
  void dispose() {
    widget.store.removeListener(_maybeScrollToFocused);
    super.dispose();
  }

  void _maybeScrollToFocused() {
    final id = widget.store.focusProposalId;
    if (id == null) return;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final ctx = _cardKeys[id]?.currentContext;
      if (ctx == null) {
        return; // not rendered yet (still fetching) -- the next rebuild retries
      }
      Scrollable.ensureVisible(
        ctx,
        duration: const Duration(milliseconds: 300),
        alignment: 0.1,
      );
      widget.store.clearFocusQuietly();
    });
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: widget.store,
      builder: (context, _) {
        final proposals = widget.store.needsAction;
        final focusId = widget.store.focusProposalId;
        return Scaffold(
          appBar: AppBar(
            title: const Text('Proposals'),
            actions: const [ThemeToggleButton()],
          ),
          body: widget.store.fetching && proposals.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : RefreshIndicator(
                  onRefresh: widget.store.fetch,
                  child: proposals.isEmpty
                      ? const _EmptyState()
                      : ListView.separated(
                          padding: const EdgeInsets.all(LaSpacing.base),
                          itemCount: proposals.length,
                          separatorBuilder: (context, index) =>
                              const SizedBox(height: LaSpacing.sm),
                          itemBuilder: (context, index) {
                            final proposal = proposals[index];
                            return ProposalCard(
                              key: _keyFor(proposal.id),
                              proposal: proposal,
                              store: widget.store,
                              initiallyExpanded: focusId != null
                                  ? proposal.id == focusId
                                  : index == 0,
                            );
                          },
                        ),
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
    // A plain Center isn't scrollable, so the ancestor RefreshIndicator's pull
    // gesture would never register -- ListView + AlwaysScrollableScrollPhysics
    // keeps pull-to-refresh working even with nothing to show yet.
    return ListView(
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
      ],
    );
  }
}
