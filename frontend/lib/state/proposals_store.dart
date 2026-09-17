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

  Future<void> confirm(String id) async {
    byId(id).status = ProposalStatus.confirmed;
    notifyListeners();
    await _postStatus(id, 'confirm');
  }

  Future<void> reject(String id) async {
    byId(id).status = ProposalStatus.rejected;
    notifyListeners();
    await _postStatus(id, 'reject');
  }

  Future<void> _postStatus(String id, String action) async {
    try {
      final response = await apiRequest(
        'POST',
        'v1/proposals/$id/$action',
        true,
        null,
        {},
      );
      if (response.statusCode != 200) {
        addNotification(
          title: "Couldn't update proposal",
          content: response.body,
          type: "error",
        );
      }
    } catch (err, stackTrace) {
      devNotification(err: err, stackTrace: stackTrace, title: "ProposalsStore.$action()");
    }
  }
}
