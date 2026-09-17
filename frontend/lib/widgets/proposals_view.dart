import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';

import '../models/proposal.dart';
import '../state/proposals_store.dart';
import '../utils/layout.dart';
import 'proposal_card.dart';
import 'theme_toggle_button.dart';

enum ProposalViewMode { list, grid }

/// Renders a selection of proposals as either a single stacked column or a
/// responsive grid, with a toggle in the app bar to switch between them.
/// Shared by the Proposals and History screens so they stay visually and
/// behaviorally identical.
class ProposalsView extends StatefulWidget {
  final String title;
  final ProposalsStore store;
  final List<Proposal> Function(ProposalsStore store) selector;
  final Widget emptyState;
  final bool readOnly;

  const ProposalsView({
    super.key,
    required this.title,
    required this.store,
    required this.selector,
    required this.emptyState,
    this.readOnly = false,
  });

  @override
  State<ProposalsView> createState() => _ProposalsViewState();
}

class _ProposalsViewState extends State<ProposalsView> {
  // Null until the user explicitly picks a view -- until then the layout
  // just follows the window width, so a desktop-sized window opens in grid
  // view without permanently locking narrower windows into it.
  ProposalViewMode? _viewMode;

  static const double _minGridCardWidth = 340;

  bool _skippingAll = false;

  Future<void> _skipAllSuggested() async {
    setState(() => _skippingAll = true);
    await widget.store.skipAllSuggested();
    if (mounted) setState(() => _skippingAll = false);
  }

  // GlobalKeys (rather than plain ValueKeys) so a specific card's rendered
  // BuildContext can be found later, to scroll it into view on a notification
  // tap. Only meaningful for the actionable (non-readOnly) view -- a pending
  // proposal a notification points to never appears in History anyway.
  final Map<String, GlobalKey> _cardKeys = {};

  GlobalKey _keyFor(String id) => _cardKeys.putIfAbsent(id, () => GlobalKey());

  @override
  void initState() {
    super.initState();
    if (!widget.readOnly) {
      widget.store.addListener(_maybeScrollToFocused);
      _maybeScrollToFocused();
    }
  }

  @override
  void dispose() {
    if (!widget.readOnly) {
      widget.store.removeListener(_maybeScrollToFocused);
    }
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

  ProposalViewMode _resolveViewMode(double width) {
    return _viewMode ??
        (LaBreakpoints.isAtLeastTablet(width)
            ? ProposalViewMode.grid
            : ProposalViewMode.list);
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: widget.store,
      builder: (context, _) {
        final proposals = widget.selector(widget.store);
        final focusId = widget.readOnly ? null : widget.store.focusProposalId;
        final currentMode = _resolveViewMode(MediaQuery.sizeOf(context).width);
        // History's proposals are already resolved (confirmed/rejected) -- bulk-skip
        // only ever makes sense for the actionable (non-readOnly) list.
        final suggestedSkipCount = widget.readOnly
            ? 0
            : proposals.where((p) => p.suggestedSkip).length;
        return Scaffold(
          appBar: AppBar(
            title: Text(widget.title),
            actions: [
              if (proposals.isNotEmpty)
                _ViewModeToggle(
                  mode: currentMode,
                  onChanged: (mode) => setState(() => _viewMode = mode),
                ),
              const ThemeToggleButton(),
            ],
          ),
          body: widget.store.fetching && proposals.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : Column(
                  children: [
                    if (suggestedSkipCount > 0)
                      _SuggestedSkipBanner(
                        count: suggestedSkipCount,
                        busy: _skippingAll,
                        onSkipAll: _skipAllSuggested,
                      ),
                    Expanded(
                      child: RefreshIndicator(
                        onRefresh: widget.store.fetch,
                        child: proposals.isEmpty
                            // Plain Center isn't scrollable, so the pull gesture
                            // above would never register with nothing to show yet.
                            ? ListView(
                                physics: const AlwaysScrollableScrollPhysics(),
                                children: [widget.emptyState],
                              )
                            : LayoutBuilder(
                                builder: (context, constraints) {
                                  final mode = _resolveViewMode(
                                    constraints.maxWidth,
                                  );
                                  return Align(
                                    alignment: Alignment.topCenter,
                                    child: ConstrainedBox(
                                      constraints: const BoxConstraints(
                                        maxWidth: kPageContentMaxWidth,
                                      ),
                                      child: mode == ProposalViewMode.grid
                                          ? _ProposalGrid(
                                              proposals: proposals,
                                              store: widget.store,
                                              readOnly: widget.readOnly,
                                              minCardWidth: _minGridCardWidth,
                                              focusId: focusId,
                                              keyFor: _keyFor,
                                            )
                                          : _ProposalList(
                                              proposals: proposals,
                                              store: widget.store,
                                              readOnly: widget.readOnly,
                                              focusId: focusId,
                                              keyFor: _keyFor,
                                            ),
                                    ),
                                  );
                                },
                              ),
                      ),
                    ),
                  ],
                ),
        );
      },
    );
  }
}

class _ProposalList extends StatelessWidget {
  final List<Proposal> proposals;
  final ProposalsStore store;
  final bool readOnly;
  final String? focusId;
  final GlobalKey Function(String id) keyFor;

  const _ProposalList({
    required this.proposals,
    required this.store,
    required this.readOnly,
    required this.focusId,
    required this.keyFor,
  });

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      padding: const EdgeInsets.all(LaSpacing.base),
      itemCount: proposals.length,
      separatorBuilder: (context, index) =>
          const SizedBox(height: LaSpacing.sm),
      itemBuilder: (context, index) {
        final proposal = proposals[index];
        return ProposalCard(
          key: keyFor(proposal.id),
          proposal: proposal,
          store: store,
          readOnly: readOnly,
          initiallyExpanded: focusId != null
              ? proposal.id == focusId
              : index == 0,
        );
      },
    );
  }
}

class _ProposalGrid extends StatelessWidget {
  final List<Proposal> proposals;
  final ProposalsStore store;
  final bool readOnly;
  final double minCardWidth;
  final String? focusId;
  final GlobalKey Function(String id) keyFor;

  const _ProposalGrid({
    required this.proposals,
    required this.store,
    required this.readOnly,
    required this.minCardWidth,
    required this.focusId,
    required this.keyFor,
  });

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(LaSpacing.base),
      child: LayoutBuilder(
        builder: (context, constraints) {
          const spacing = LaSpacing.sm;
          // Fits as many columns as the width allows, up to 3 -- narrower
          // "desktop" windows (e.g. a tablet, or a split screen) fall back
          // to 2 or 1 instead of squeezing 3 cramped columns in.
          final columns = (constraints.maxWidth / minCardWidth).floor().clamp(
            1,
            3,
          );
          final cardWidth =
              (constraints.maxWidth - spacing * (columns - 1)) / columns;
          // Wrap shrink-wraps to the width of its widest run, so with fewer
          // cards than columns (a single, short run) it would end up
          // narrower than the container and get centered by the Align
          // above it. Forcing it to the full available width keeps its
          // own (left-aligned) content honest instead.
          return SizedBox(
            width: constraints.maxWidth,
            child: Wrap(
              alignment: WrapAlignment.start,
              spacing: spacing,
              runSpacing: spacing,
              children: proposals.asMap().entries.map((entry) {
                final index = entry.key;
                final proposal = entry.value;
                return SizedBox(
                  width: cardWidth,
                  child: ProposalCard(
                    key: keyFor(proposal.id),
                    proposal: proposal,
                    store: store,
                    readOnly: readOnly,
                    initiallyExpanded: focusId != null
                        ? proposal.id == focusId
                        : index == 0,
                  ),
                );
              }).toList(),
            ),
          );
        },
      ),
    );
  }
}

/// A dismissal-free call to action above the list: rather than making the user tap
/// Skip on every promotional/automated message one at a time, offer to clear all of
/// them in one tap. Stays visible (not part of the scrolling list) since it reflects
/// the whole list's state, not one item's.
class _SuggestedSkipBanner extends StatelessWidget {
  final int count;
  final bool busy;
  final VoidCallback onSkipAll;

  const _SuggestedSkipBanner({
    required this.count,
    required this.busy,
    required this.onSkipAll,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(
        horizontal: LaSpacing.base,
        vertical: LaSpacing.sm,
      ),
      color: colors.backgroundNeutralTertiary,
      child: Row(
        children: [
          Icon(
            Icons.filter_alt_outlined,
            size: 18,
            color: colors.contentNeutralSecondary,
          ),
          const SizedBox(width: LaSpacing.x2xs),
          Expanded(
            child: Text(
              count == 1
                  ? "1 message looks like it doesn't need a reply"
                  : "$count messages look like they don't need a reply",
              style: LaTextStyles.bodySm.copyWith(
                color: colors.contentNeutralSecondary,
              ),
            ),
          ),
          LaButton(
            label: 'Skip all',
            size: LaButtonSize.small,
            variant: LaButtonVariant.bordered,
            color: LaButtonColor.neutral,
            loading: busy,
            onPressed: onSkipAll,
          ),
        ],
      ),
    );
  }
}

class _ViewModeToggle extends StatelessWidget {
  final ProposalViewMode mode;
  final ValueChanged<ProposalViewMode> onChanged;

  const _ViewModeToggle({required this.mode, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: LaSpacing.x2xs),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          _ViewModeButton(
            icon: Icons.view_agenda_outlined,
            tooltip: 'List view',
            selected: mode == ProposalViewMode.list,
            onPressed: () => onChanged(ProposalViewMode.list),
          ),
          _ViewModeButton(
            icon: Icons.grid_view_outlined,
            tooltip: 'Grid view',
            selected: mode == ProposalViewMode.grid,
            onPressed: () => onChanged(ProposalViewMode.grid),
          ),
        ],
      ),
    );
  }
}

class _ViewModeButton extends StatelessWidget {
  final IconData icon;
  final String tooltip;
  final bool selected;
  final VoidCallback onPressed;

  const _ViewModeButton({
    required this.icon,
    required this.tooltip,
    required this.selected,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.laColors;
    return ClipRRect(
      borderRadius: BorderRadius.circular(LaRadius.md),
      child: Material(
        color: selected ? colors.backgroundBrandTertiary : Colors.transparent,
        child: IconButton(
          tooltip: tooltip,
          icon: Icon(icon),
          color: selected
              ? colors.contentBrandPrimary
              : colors.contentNeutralTertiary,
          hoverColor: colors.backgroundBrandTertiaryHover,
          onPressed: onPressed,
        ),
      ),
    );
  }
}
