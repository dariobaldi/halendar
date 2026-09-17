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
        final currentMode = _resolveViewMode(MediaQuery.sizeOf(context).width);
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
          body: proposals.isEmpty
              ? widget.emptyState
              : LayoutBuilder(
                  builder: (context, constraints) {
                    final mode = _resolveViewMode(constraints.maxWidth);
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
                              )
                            : _ProposalList(
                                proposals: proposals,
                                store: widget.store,
                                readOnly: widget.readOnly,
                              ),
                      ),
                    );
                  },
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

  const _ProposalList({
    required this.proposals,
    required this.store,
    required this.readOnly,
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
          key: ValueKey(proposal.id),
          proposal: proposal,
          store: store,
          readOnly: readOnly,
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

  const _ProposalGrid({
    required this.proposals,
    required this.store,
    required this.readOnly,
    required this.minCardWidth,
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
              children: proposals.map((proposal) {
                return SizedBox(
                  width: cardWidth,
                  child: ProposalCard(
                    key: ValueKey(proposal.id),
                    proposal: proposal,
                    store: store,
                    readOnly: readOnly,
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
