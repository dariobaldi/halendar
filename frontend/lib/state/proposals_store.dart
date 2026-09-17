import 'dart:async';
import 'dart:convert';

import 'package:halendar_front/services/api.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:halendar_front/services/websocket.dart';
import 'package:flutter/foundation.dart';

import '../models/proposal.dart';
import '../models/time_slot.dart';
import '../utils/draft_builder.dart';

/// Single source of truth for proposals, backed by the /v1/proposals API. Every
/// mutation here corresponds to an explicit user action, applied locally right away
/// (so the UI feels instant) and then persisted to the backend in the background --
/// a failure to persist shows a notification but doesn't roll back the local change,
/// consistent with how the rest of the app's stores handle it.
class ProposalsStore extends ChangeNotifier {
  List<Proposal> _proposals = [];
  bool _fetching = true;

  StreamSubscription? _ws;

  List<Proposal> get needsAction =>
      _proposals.where((p) => p.status == ProposalStatus.pending).toList();

  List<Proposal> get history => _proposals
      .where(
        (p) => p.status == ProposalStatus.confirmed ||
            p.status == ProposalStatus.rejected,
      )
      .toList();

  bool get fetching => _fetching;

  Proposal byId(String id) => _proposals.firstWhere((p) => p.id == id);

  /// Set once, right after a notification tap, to tell ProposalsListScreen which
  /// card to expand and scroll to. Read once by the next build, then cleared
  /// quietly (see [clearFocusQuietly]) rather than left to keep forcing that one
  /// card open on every future rebuild.
  String? focusProposalId;

  void focusOn(String id) {
    focusProposalId = id;
    notifyListeners();
  }

  /// Clears the focus without notifying listeners, so the card that was just
  /// expanded because of it doesn't immediately collapse again -- only the next
  /// independent rebuild (a refresh, a new proposal arriving, ...) will stop
  /// treating it as forced-open.
  void clearFocusQuietly() {
    focusProposalId = null;
  }

  void init() {
    fetch();
    _ws = AuthService.instance.ws.stream.listen((message) {
      final Map<String, dynamic> data = jsonDecode(message);
      if (data.containsKey("reconnected") ||
          (validTypeMessage("email_messages", data) &&
              data['refresh'] == true)) {
        fetch();
      }
    });
  }

  void end() {
    _ws?.cancel();
  }

  Future<void> fetch() async {
    try {
      final response = await apiRequest('GET', 'v1/proposals', true, null, {});
      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(
          utf8.decode(response.bodyBytes),
        );
        final List<dynamic> raw = data['proposals'];
        _proposals = raw.map((j) => Proposal.fromJson(j)).toList();
      }
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "ProposalsStore.fetch()");
    }
    _fetching = false;
    notifyListeners();
  }

  /// Selects a different free slot. Regenerates the draft to reference the new
  /// time, unless the user has already edited it by hand.
  Future<void> selectSlot(String id, TimeSlot slot) async {
    final proposal = byId(id);
    proposal.selectedSlot = slot;
    if (!proposal.draftEditedByUser) {
      final user = AuthService.instance.user;
      proposal.responseDraft = buildResponseDraft(
        senderFirstName: proposal.senderName.split(' ').first,
        userName: user?.name ?? user?.username ?? 'the user',
        slot: slot,
      );
    }
    notifyListeners();

    try {
      final response = await apiRequest(
        'PATCH',
        'v1/proposals/$id/slot',
        true,
        json.encode({'slot_id': slot.id, 'response_draft': proposal.responseDraft}),
        {},
      );
      if (response.statusCode != 200) {
        addNotification(
          title: "Couldn't save your selection",
          content: response.body,
          type: "error",
        );
      }
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "ProposalsStore.selectSlot()");
    }
  }

  Future<void> updateDraft(String id, String text) async {
    final proposal = byId(id);
    proposal.responseDraft = text;
    proposal.draftEditedByUser = true;
    notifyListeners();

    try {
      final response = await apiRequest(
        'PATCH',
        'v1/proposals/$id/draft',
        true,
        json.encode({'response_draft': text}),
        {},
      );
      if (response.statusCode != 200) {
        addNotification(
          title: "Couldn't save your edit",
          content: response.body,
          type: "error",
        );
      }
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "ProposalsStore.updateDraft()");
    }
  }

  /// Confirms the proposal: the backend sends the drafted reply for real (and books
  /// the slot, if any). Returns null on success, an error message otherwise -- on
  /// failure (e.g. the send itself failed) the optimistic "confirmed" status is
  /// rolled back, since nothing actually went out and the card still needs the
  /// user's attention.
  Future<String?> confirm(String id) => _setStatus(id, ProposalStatus.confirmed, 'confirm');

  /// Skips the proposal: archived, nothing sent. Same rollback-on-failure behavior
  /// as [confirm].
  Future<String?> reject(String id) => _setStatus(id, ProposalStatus.rejected, 'reject');

  /// Skips every currently-pending proposal the AI suggested skipping (see
  /// Proposal.suggestedSkip) in one go -- the bulk counterpart to tapping Skip on
  /// each one individually. Proposals that fail to skip stay in the list (same
  /// rollback as a single reject) rather than silently vanishing.
  Future<void> skipAllSuggested() async {
    final ids = needsAction.where((p) => p.suggestedSkip).map((p) => p.id).toList();
    await Future.wait(ids.map(reject));
  }

  Future<String?> _setStatus(String id, ProposalStatus newStatus, String action) async {
    final proposal = byId(id);
    final previousStatus = proposal.status;
    proposal.status = newStatus;
    notifyListeners();

    final error = await _postStatus(id, action);
    if (error != null) {
      proposal.status = previousStatus;
      notifyListeners();
    }
    return error;
  }

  Future<String?> _postStatus(String id, String action) async {
    try {
      final response = await apiRequest(
        'POST',
        'v1/proposals/$id/$action',
        true,
        null,
        {},
      );
      if (response.statusCode == 200) return null;
      addNotification(
        title: "Couldn't update proposal",
        content: response.body,
        type: "error",
      );
      return response.body;
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "ProposalsStore.$action()");
      return 'Something went wrong. Please try again.';
    }
  }

  /// Re-runs AI analysis on this proposal's source message on demand -- e.g. right
  /// after switching to Claude, or to retry a bad extraction. Runs on the backend's
  /// own time (an LLM call, not instant), so callers should show a loading state
  /// rather than expect this to resolve immediately.
  Future<void> reanalyze(String id) async {
    try {
      final response = await apiRequest(
        'POST',
        'v1/proposals/$id/reanalyze',
        true,
        null,
        {},
      );
      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(
          utf8.decode(response.bodyBytes),
        );
        final updated = Proposal.fromJson(data['proposal']);
        final index = _proposals.indexWhere((p) => p.id == id);
        if (index != -1) _proposals[index] = updated;
        notifyListeners();
        return;
      }
      if (response.statusCode == 404) {
        // Rare: the message (or its account) was removed elsewhere in the meantime.
        // A re-analysis that just decides it's not a meeting request after all comes
        // back as a normal 200 with is_meeting_request: false, handled above -- it
        // still shows up here, just without slots/draft to act on.
        _proposals.removeWhere((p) => p.id == id);
        notifyListeners();
        addNotification(
          title: "This message is no longer available",
          content: "It may have been removed elsewhere.",
          type: "info",
        );
        return;
      }
      addNotification(
        title: "Couldn't re-analyze",
        content: response.body,
        type: "error",
      );
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "ProposalsStore.reanalyze()");
    }
  }
}
