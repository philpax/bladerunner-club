package main

import (
	comatproto "github.com/bluesky-social/indigo/api/atproto"
	toolsozone "github.com/bluesky-social/indigo/api/ozone"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/automod"
)

func accountHasLabel(am *automod.AccountMeta, label string) bool {
	// check if label is already applied
	for _, l := range am.AccountLabels {
		if l == label {
			return true
		}
	}
	return false
}

func addAccountLabel(c *automod.RecordContext, did syntax.DID, label string) error {
	am := c.GetAccountMeta(did)
	eng := c.InternalEngine()
	if eng.OzoneClient == nil {
		c.Logger.Warn("skipping label addition (no ozone client)", "did", did, "label", label)
		return nil
	}

	// check if label is already applied
	for _, l := range am.AccountLabels {
		if l == label {
			return nil
		}
	}

	// send label via engine
	c.Logger.Warn("adding label", "did", did, "label", label)
	comment := "auto-adding label"
	_, err := toolsozone.ModerationEmitEvent(c.Ctx, eng.OzoneClient, &toolsozone.ModerationEmitEvent_Input{
		CreatedBy: eng.OzoneClient.Auth.Did,
		Event: &toolsozone.ModerationEmitEvent_Input_Event{
			ModerationDefs_ModEventLabel: &toolsozone.ModerationDefs_ModEventLabel{
				CreateLabelVals: []string{label},
				NegateLabelVals: []string{},
				Comment:         &comment,
			},
		},
		Subject: &toolsozone.ModerationEmitEvent_Input_Subject{
			AdminDefs_RepoRef: &comatproto.AdminDefs_RepoRef{
				Did: did.String(),
			},
		},
	})
	return err
}

func removeAccountLabel(c *automod.RecordContext, did syntax.DID, label string) error {
	am := c.GetAccountMeta(did)
	eng := c.InternalEngine()
	if eng.OzoneClient == nil {
		c.Logger.Warn("skipping label removal (no ozone client)", "did", did, "label", label)
		return nil
	}

	// check if label is already applied
	exists := false
	for _, l := range am.AccountLabels {
		if l == label {
			exists = true
			break
		}
	}
	if !exists {
		return nil
	}

	// send label via engine
	c.Logger.Warn("removing label", "did", did, "label", label)
	comment := "auto-removing label"
	_, err := toolsozone.ModerationEmitEvent(c.Ctx, eng.OzoneClient, &toolsozone.ModerationEmitEvent_Input{
		CreatedBy: eng.OzoneClient.Auth.Did,
		Event: &toolsozone.ModerationEmitEvent_Input_Event{
			ModerationDefs_ModEventLabel: &toolsozone.ModerationDefs_ModEventLabel{
				CreateLabelVals: []string{},
				NegateLabelVals: []string{label},
				Comment:         &comment,
			},
		},
		Subject: &toolsozone.ModerationEmitEvent_Input_Subject{
			AdminDefs_RepoRef: &comatproto.AdminDefs_RepoRef{
				Did: did.String(),
			},
		},
	})
	return err
}
