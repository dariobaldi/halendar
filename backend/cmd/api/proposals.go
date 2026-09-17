package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/google/uuid"
)

// listProposalsHandler returns every proposal (pending, confirmed, or rejected)
// across all of the user's connected accounts. The frontend splits pending
// ("needs action") from confirmed/rejected ("history") itself, the same way it
// already does with the mock data it's replacing.
func (app *app) listProposalsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	proposals, err := app.models.Proposals.GetForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"proposals": proposals}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// selectProposalSlotHandler records which slot the user wants to go with, along with
// the draft text for it -- the frontend regenerates the draft locally whenever the
// user hasn't hand-edited it and sends both together in one call.
func (app *app) selectProposalSlotHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		SlotID        string `json:"slot_id"`
		ResponseDraft string `json:"response_draft"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	slotID, err := uuid.Parse(input.SlotID)
	if err != nil {
		app.badRequestResponse(w, r, errors.New(`"slot_id" must be a valid id`))
		return
	}

	if err := app.models.Proposals.SetSelectedSlot(id, user.ID, slotID, input.ResponseDraft); err != nil {
		app.respondToProposalUpdateError(w, r, err)
		return
	}
	app.respondUpdatedProposal(w, r, id, user.ID)
}

// updateProposalDraftHandler records the user's own edited reply text.
func (app *app) updateProposalDraftHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		ResponseDraft string `json:"response_draft"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.models.Proposals.SetDraft(id, user.ID, input.ResponseDraft); err != nil {
		app.respondToProposalUpdateError(w, r, err)
		return
	}
	app.respondUpdatedProposal(w, r, id, user.ID)
}

// confirmProposalHandler sends the drafted reply for real, through the connected
// account's SMTP, in the original message's thread. If a slot is selected, it also
// books that time on the connected calendar (best-effort: a booking failure is
// logged but doesn't stop the proposal from being confirmed -- the reply having
// actually been sent is what matters most). The status flip happens first, atomically
// gated on the proposal still being pending, so a duplicate request can't send twice.
func (app *app) confirmProposalHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	proposal, err := app.models.Proposals.GetByID(id, user.ID)
	if err != nil {
		app.respondToProposalUpdateError(w, r, err)
		return
	}

	if err := app.models.Proposals.ClaimPending(id, user.ID, data.ProposalStatusConfirmed); err != nil {
		app.respondToProposalUpdateError(w, r, err)
		return
	}

	if err := app.sendProposalReply(r.Context(), user.ID, *proposal); err != nil {
		// Sending failed -- give the proposal back so the user can retry, rather than
		// leaving it stuck "confirmed" with nothing actually sent.
		if revertErr := app.models.Proposals.SetStatus(id, user.ID, data.ProposalStatusPending); revertErr != nil {
			app.logger.Error("confirm proposal: reverting after failed send: " + revertErr.Error())
		}
		app.badRequestResponse(w, r, fmt.Errorf("could not send the reply: %w", err))
		return
	}

	app.respondUpdatedProposal(w, r, id, user.ID)
}

// rejectProposalHandler marks a proposal rejected (the frontend's "Delete"): archived,
// nothing sent, no event booked.
func (app *app) rejectProposalHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.models.Proposals.ClaimPending(id, user.ID, data.ProposalStatusRejected); err != nil {
		app.respondToProposalUpdateError(w, r, err)
		return
	}
	app.respondUpdatedProposal(w, r, id, user.ID)
}

// sendProposalReply sends proposal.ResponseDraft as a reply to the original message,
// in its thread, through the account it was imported from. When the proposal has an
// effective slot selected, it also books that time on the user's connected calendar.
func (app *app) sendProposalReply(ctx context.Context, userID uuid.UUID, proposal data.Proposal) error {
	account, err := app.models.EmailAccounts.Get(proposal.EmailAccountID, userID)
	if err != nil {
		return fmt.Errorf("loading email account: %w", err)
	}
	mailbox, err := app.connectMailbox(ctx, *account)
	if err != nil {
		return fmt.Errorf("connecting mailbox: %w", err)
	}
	original, err := mailbox.Read(ctx, proposal.IMAPUID)
	if err != nil {
		return fmt.Errorf("reading original message: %w", err)
	}
	if _, err := mailbox.Reply(ctx, *original, proposal.ResponseDraft); err != nil {
		return fmt.Errorf("sending reply: %w", err)
	}

	if slot := proposal.EffectiveSlot(); slot != nil && slot.Availability == data.SlotAvailabilityFree {
		calSource, _ := app.userCalendarSource(ctx, userID)
		if calSource == nil {
			app.logger.Error("confirm proposal: no calendar connected, skipping event booking")
		} else {
			description := fmt.Sprintf("Booked by the Halendar Assistant from an email from %s <%s>.", proposal.SenderName, proposal.SenderEmail)
			title := firstNonEmpty(proposal.Title, proposal.Subject)
			if err := calSource.AddEvent(ctx, title, proposal.Location, description, slot.StartAt, slot.EndAt); err != nil {
				app.logger.Error("confirm proposal: booking calendar event: " + err.Error())
			}
		}
	}
	return nil
}

func (app *app) respondToProposalUpdateError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, data.ErrRecordNotFound):
		app.notFoundResponse(w, r)
	default:
		app.serverErrorResponse(w, r, err)
	}
}

// respondUpdatedProposal re-fetches and returns the proposal so the frontend can
// reconcile its local state with whatever the backend actually stored, rather than
// trusting its own optimistic update was applied verbatim.
func (app *app) respondUpdatedProposal(w http.ResponseWriter, r *http.Request, id, userID uuid.UUID) {
	proposal, err := app.models.Proposals.GetByID(id, userID)
	if err != nil {
		app.respondToProposalUpdateError(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"proposal": proposal}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
