import 'package:flutter/material.dart';
import 'package:lasuite_ui/lasuite_ui.dart';
import 'package:url_launcher/url_launcher.dart';

import '../models/proposal.dart';
import '../state/proposals_store.dart';
import '../utils/date_format.dart';
import 'email_preview.dart';
import 'status_badge.dart';
import 'time_slot_tile.dart';

/// Everything a proposal needs is right here: the email, the proposed
/// slots, the draft reply, and the three actions (Send / Edit / Delete).
/// No separate screen to open first -- this card *is* the interaction,
/// the way an inline calendar invite or a push notification lets you
/// act without a detour through a detail view.
class ProposalCard extends StatefulWidget {
  final Proposal proposal;
  final ProposalsStore store;
  final bool readOnly;
  final bool initiallyExpanded;

  const ProposalCard({
    super.key,
    required this.proposal,
    required this.store,
    this.readOnly = false,
    this.initiallyExpanded = false,
  });

  @override
  State<ProposalCard> createState() => _ProposalCardState();
}

class _ProposalCardState extends State<ProposalCard> {
  late TextEditingController _draftController;
  final FocusNode _draftFocus = FocusNode();
  bool _editingDraft = false;
  late bool _expanded;

  @override
  void initState() {
    super.initState();
    _draftController = TextEditingController(
      text: widget.proposal.responseDraft,
    );
    _expanded = widget.initiallyExpanded;
  }

  @override
  void dispose() {
    _draftController.dispose();
    _draftFocus.dispose();
    super.dispose();
  }

  void _saveDraft() {
    widget.store.updateDraft(widget.proposal.id, _draftController.text);
    setState(() => _editingDraft = false);
  }

  void _startEditing() {
    // The store mutates Proposal in place, so the controller can only be
    // trusted to match the live draft at the moment editing starts (e.g.
    // after picking a different slot regenerated it) -- resync here.
    _draftController.text = widget.proposal.responseDraft;
    setState(() => _editingDraft = true);
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => _draftFocus.requestFocus(),
    );
  }

  void _cancelEditing() {
    _draftController.text = widget.proposal.responseDraft;
    setState(() => _editingDraft = false);
  }

  Future<void> _confirmSend() async {
    final proposal = widget.proposal;
    final slot = proposal.selectedSlot;
    final colors = context.laColors;
    final confirmed = await showLaModal<bool>(
      context: context,
      title: 'Send this reply?',
      size: LaModalSize.small,
      builder: (context) => Padding(
        padding: const EdgeInsets.only(bottom: LaSpacing.base),
        child: Text(
          slot == null
              ? 'This reply will be sent to ${proposal.senderName}. No '
                    'event will be created.'
              : 'An event will be created: ${formatSlot(slot)}. This '
                    'reply will be sent to ${proposal.senderName}.',
          style: LaTextStyles.bodySm.copyWith(
            color: colors.contentNeutralSecondary,
          ),
        ),
      ),
      rightActions: [
        LaButton(
          label: 'Cancel',
          variant: LaButtonVariant.tertiary,
          color: LaButtonColor.neutral,
          onPressed: () => Navigator.pop(context, false),
        ),
        LaButton(
          label: 'Confirm',
          onPressed: () => Navigator.pop(context, true),
        ),
      ],
    );
    if (confirmed == true && mounted) {
      widget.store.confirm(proposal.id);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            slot == null
                ? 'Reply sent.'
                : 'Reply sent. Event added to your calendar.',
          ),
        ),
      );
    }
  }

  Future<void> _confirmDelete() async {
    final colors = context.laColors;
    final confirmed = await showLaModal<bool>(
      context: context,
      title: 'Delete this proposal?',
      size: LaModalSize.small,
      builder: (context) => Padding(
        padding: const EdgeInsets.only(bottom: LaSpacing.base),
        child: Text(
          'No email will be sent and no event will be created. The '
          'proposal will be archived.',
          style: LaTextStyles.bodySm.copyWith(
            color: colors.contentNeutralSecondary,
          ),
        ),
      ),
      rightActions: [
        LaButton(
          label: 'Back',
          variant: LaButtonVariant.tertiary,
          color: LaButtonColor.neutral,
          onPressed: () => Navigator.pop(context, false),
        ),
        LaButton(
          label: 'Delete',
          color: LaButtonColor.error,
          onPressed: () => Navigator.pop(context, true),
        ),
      ],
    );
    if (confirmed == true && mounted) {
      widget.store.reject(widget.proposal.id);
    }
  }

  Future<void> _openInMailApp() async {
    final proposal = widget.proposal;
    final uri = Uri(
      scheme: 'mailto',
      path: proposal.senderEmail,
      query: 'subject=${Uri.encodeComponent('Re: ${proposal.subject}')}',
    );
    var launched = false;
    try {
      launched = await launchUrl(uri);
    } catch (_) {
      launched = false;
    }
    if (!launched && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Could not open your mail app.')),
      );
    }
  }

  String _collapsedSummary() {
    final proposal = widget.proposal;
    if (proposal.status == ProposalStatus.confirmed) {
      return proposal.selectedSlot == null
          ? 'Reply sent'
          : 'Confirmed for ${formatSlot(proposal.selectedSlot!)}';
    }
    if (proposal.status == ProposalStatus.rejected) {
      return 'Deleted';
    }
    if (proposal.needsManualReview) {
      return 'Needs manual review';
    }
    final slot = proposal.selectedSlot;
    return slot == null
        ? 'None of the proposed times are available'
        : formatSlot(slot);
  }

  Widget _header(BuildContext context) {
    final colors = context.laColors;
    final proposal = widget.proposal;
    // No hover wash on the row itself -- for a collapsed card this row
    // *is* almost the whole card, so a full-width tint just looks like
    // the entire card changing color. Only the chevron gets its own
    // small, contained hover circle; the rest of the row stays tappable
    // (a generous touch target) but visually quiet.
    return InkWell(
      borderRadius: BorderRadius.circular(LaRadius.md),
      hoverColor: Colors.transparent,
      highlightColor: Colors.transparent,
      splashColor: colors.backgroundBrandTertiaryHover,
      onTap: () => setState(() => _expanded = !_expanded),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: LaSpacing.x4xs),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            LaAvatar(name: proposal.senderName),
            const SizedBox(width: LaSpacing.xs),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    proposal.senderName,
                    style: LaTextStyles.labelLg.copyWith(
                      color: colors.contentNeutralPrimary,
                    ),
                  ),
                  Text(
                    proposal.subject,
                    style: LaTextStyles.bodySm.copyWith(
                      color: colors.contentNeutralSecondary,
                    ),
                  ),
                  const SizedBox(height: LaSpacing.x4xs),
                  Text(
                    formatReceivedAt(proposal.receivedAt),
                    style: LaTextStyles.caption.copyWith(
                      color: colors.contentNeutralTertiary,
                    ),
                  ),
                  if (!_expanded) ...[
                    const SizedBox(height: LaSpacing.x2xs),
                    Text(
                      _collapsedSummary(),
                      style: LaTextStyles.labelMd.copyWith(
                        color: colors.contentNeutralPrimary,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(width: LaSpacing.xs),
            Padding(
              // Nudges the trailing cluster down to sit level with the
              // sender name instead of the card's very top edge.
              padding: const EdgeInsets.only(top: 1),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  StatusBadge(proposal: proposal),
                  const SizedBox(width: LaSpacing.x2xs),
                  ClipOval(
                    child: Material(
                      color: Colors.transparent,
                      child: InkWell(
                        onTap: () => setState(() => _expanded = !_expanded),
                        hoverColor: colors.backgroundBrandTertiaryHover,
                        splashColor: colors.backgroundBrandSecondary,
                        child: Padding(
                          padding: const EdgeInsets.all(LaSpacing.x4xs),
                          child: Icon(
                            _expanded ? Icons.expand_less : Icons.expand_more,
                            size: 22,
                            color: colors.contentNeutralTertiary,
                          ),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _slotsSection() {
    final proposal = widget.proposal;
    final selected = proposal.selectedSlot;

    return Column(
      children: proposal.slots.map((slot) {
        return TimeSlotRow(
          slot: slot,
          isSelected: selected != null && slot.start == selected.start,
          onTap: () => widget.store.selectSlot(proposal.id, slot),
        );
      }).toList(),
    );
  }

  Widget _draftSection() {
    final colors = context.laColors;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Reply',
          style: LaTextStyles.labelMd.copyWith(
            color: colors.contentNeutralPrimary,
          ),
        ),
        const SizedBox(height: LaSpacing.x2xs),
        if (_editingDraft)
          LaTextArea(
            controller: _draftController,
            focusNode: _draftFocus,
            minLines: 5,
            maxLines: 10,
          )
        else
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(LaSpacing.sm),
            decoration: BoxDecoration(
              color: colors.backgroundNeutralTertiary,
              borderRadius: BorderRadius.circular(LaRadius.md),
            ),
            child: Text(
              widget.proposal.responseDraft,
              style: LaTextStyles.bodySm.copyWith(
                color: colors.contentNeutralSecondary,
              ),
            ),
          ),
      ],
    );
  }

  Widget _actions() {
    if (_editingDraft) {
      return Row(
        children: [
          Expanded(
            child: LaButton(
              label: 'Cancel',
              variant: LaButtonVariant.bordered,
              color: LaButtonColor.neutral,
              fullWidth: true,
              onPressed: _cancelEditing,
            ),
          ),
          const SizedBox(width: LaSpacing.xs),
          Expanded(
            child: LaButton(
              label: 'Save',
              variant: LaButtonVariant.secondary,
              fullWidth: true,
              onPressed: _saveDraft,
            ),
          ),
        ],
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // A fullWidth button stretches to match the card, which is fine in
        // the narrower grid view but reads as an oversized bar in the
        // wider single-column list view -- cap it instead of letting it
        // track the card's width.
        Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 400),
            child: LaButton(
              label: 'Send',
              icon: const Icon(Icons.send),
              fullWidth: true,
              onPressed: _confirmSend,
            ),
          ),
        ),
        const SizedBox(height: LaSpacing.x3xs),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            LaButton(
              label: 'Edit',
              variant: LaButtonVariant.tertiary,
              color: LaButtonColor.neutral,
              size: LaButtonSize.small,
              onPressed: _startEditing,
            ),
            const SizedBox(width: LaSpacing.x2xs),
            LaButton(
              label: 'Delete',
              variant: LaButtonVariant.tertiary,
              color: LaButtonColor.error,
              size: LaButtonSize.small,
              onPressed: _confirmDelete,
            ),
          ],
        ),
      ],
    );
  }

  Widget _readOnlyNote() {
    final proposal = widget.proposal;
    if (proposal.status == ProposalStatus.confirmed) {
      return LaAlert(
        type: LaVariant.success,
        message: proposal.selectedSlot == null
            ? 'Reply sent.'
            : 'Confirmed for ${formatSlot(proposal.selectedSlot!)}.',
      );
    }
    return const LaAlert(
      type: LaVariant.neutral,
      message: 'Deleted: no email was sent.',
    );
  }

  Widget _expandedContent() {
    final proposal = widget.proposal;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: LaSpacing.xs),
        EmailPreview(excerpt: proposal.emailExcerpt),
        if (widget.readOnly) ...[
          const SizedBox(height: LaSpacing.xs),
          _readOnlyNote(),
        ] else ...[
          if (proposal.needsManualReview) ...[
            const SizedBox(height: LaSpacing.xs),
            const LaAlert(
              type: LaVariant.neutral,
              message:
                  'The AI could not identify a clear time in this email. '
                  'Open it in your mail app to read the full message, or '
                  'send a quick reply asking when works.',
            ),
            const SizedBox(height: LaSpacing.xs),
            LaButton(
              label: 'Open in Mail',
              icon: const Icon(Icons.open_in_new),
              variant: LaButtonVariant.bordered,
              color: LaButtonColor.neutral,
              onPressed: _openInMailApp,
            ),
          ] else ...[
            SizedBox(height: LaSpacing.md),
            Text(
              'Time slots',
              style: LaTextStyles.labelMd.copyWith(
                color: context.laColors.contentNeutralPrimary,
              ),
            ),
            const SizedBox(height: LaSpacing.x3xs),
            _slotsSection(),
          ],
          const SizedBox(height: LaSpacing.md),
          _draftSection(),
          const SizedBox(height: LaSpacing.sm),
          _actions(),
        ],
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(LaSpacing.sm),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _header(context),
            AnimatedSize(
              duration: LaMotion.duration,
              curve: LaMotion.easeOut,
              alignment: Alignment.topCenter,
              child: _expanded
                  ? _expandedContent()
                  : const SizedBox(width: double.infinity),
            ),
          ],
        ),
      ),
    );
  }
}
