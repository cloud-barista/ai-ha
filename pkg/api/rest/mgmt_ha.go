package rest

import (
	// swag resolves the model names in the annotations through this file's imports.
	_ "github.com/cloud-barista/ai-ha/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

// The redundancy API (TM-AI-HA-02, plan/ha-roadmap.md). Nothing behind it
// exists yet: every route answers 501 and documents the contract it will keep.

// RestGetHAGroups godoc
// @ID GetHAGroups
// @Summary List HA groups
// @Description Not implemented yet: answers 501. Will list HA groups with their state, verdict and reasons.
// @Tags [Redundancy] HA Group Management
// @Produce json
// @Success 200 {object} model.HAGroupListResp
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups [get]
func RestGetHAGroups(c echo.Context) error {
	return notImplemented(c)
}

// RestPostHAGroup godoc
// @ID PostHAGroup
// @Summary Register an HA group
// @Description Not implemented yet: answers 501. Will register two systems, the volume replicated between them and the service running on it. Registering configures nothing: see setup.
// @Tags [Redundancy] HA Group Management
// @Accept json
// @Produce json
// @Param haGroupReq body model.HAGroupReq true "HA group definition"
// @Success 201 {object} model.HAGroupInfo
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups [post]
func RestPostHAGroup(c echo.Context) error {
	return notImplemented(c)
}

// RestGetHAGroup godoc
// @ID GetHAGroup
// @Summary Get an HA group
// @Description Not implemented yet: answers 501. Will return the group's state, the verdict with its reasons, and each node's health and replication status.
// @Tags [Redundancy] HA Group Management
// @Produce json
// @Param groupId path string true "HA group ID"
// @Success 200 {object} model.HAGroupInfo
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups/{groupId} [get]
func RestGetHAGroup(c echo.Context) error {
	return notImplemented(c)
}

// RestPostHAGroupSetup godoc
// @ID PostHAGroupSetup
// @Summary Configure an HA group's replication
// @Description Not implemented yet: answers 501. Will queue replication setup on the standby and then the primary, ending in the initial synchronization. Follow it by the group's state.
// @Tags [Redundancy] HA Group Management
// @Produce json
// @Param groupId path string true "HA group ID"
// @Success 202 {object} model.HAGroupInfo
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups/{groupId}/setup [post]
func RestPostHAGroupSetup(c echo.Context) error {
	return notImplemented(c)
}

// RestPostHAGroupTeardown godoc
// @ID PostHAGroupTeardown
// @Summary Remove an HA group's replication
// @Description Not implemented yet: answers 501. Will stop and clean up replication on both nodes.
// @Tags [Redundancy] HA Group Management
// @Produce json
// @Param groupId path string true "HA group ID"
// @Success 202 {object} model.HAGroupInfo
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups/{groupId}/teardown [post]
func RestPostHAGroupTeardown(c echo.Context) error {
	return notImplemented(c)
}

// RestPostHAGroupSwitchover godoc
// @ID PostHAGroupSwitchover
// @Summary Switch an HA group to its standby
// @Description Not implemented yet: answers 501. Will demote the primary, and only once that is confirmed promote the standby. For a planned move while both nodes are healthy. Will be refused while another job runs on either node.
// @Tags [Redundancy] HA Group Management
// @Accept json
// @Produce json
// @Param groupId path string true "HA group ID"
// @Param haApproval body model.HAApproval true "Approval"
// @Success 202 {object} model.HATransitionInfo
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups/{groupId}/switchover [post]
func RestPostHAGroupSwitchover(c echo.Context) error {
	return notImplemented(c)
}

// RestPostHAGroupFailover godoc
// @ID PostHAGroupFailover
// @Summary Approve an HA group's failover
// @Description Not implemented yet: answers 501. Will promote the standby of a group whose primary stopped answering. The framework never does this on its own judgement: the caller must confirm the primary is fenced, and accept any replication gap.
// @Tags [Redundancy] HA Group Management
// @Accept json
// @Produce json
// @Param groupId path string true "HA group ID"
// @Param haFailoverReq body model.HAFailoverReq true "Approval and fencing confirmation"
// @Success 202 {object} model.HATransitionInfo
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups/{groupId}/failover [post]
func RestPostHAGroupFailover(c echo.Context) error {
	return notImplemented(c)
}

// RestPostHAGroupResume godoc
// @ID PostHAGroupResume
// @Summary Resume an HA group's primary after a reboot
// @Description Not implemented yet: answers 501. Nothing promotes automatically, so a rebooted primary comes back secondary with its service stopped. This will promote it again without changing roles.
// @Tags [Redundancy] HA Group Management
// @Accept json
// @Produce json
// @Param groupId path string true "HA group ID"
// @Param haApproval body model.HAApproval true "Approval"
// @Success 202 {object} model.HATransitionInfo
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups/{groupId}/resume [post]
func RestPostHAGroupResume(c echo.Context) error {
	return notImplemented(c)
}

// RestPostHAGroupRejoin godoc
// @ID PostHAGroupRejoin
// @Summary Rejoin a returned node to its HA group
// @Description Not implemented yet: answers 501. Will discard what the old primary wrote after the switch and resynchronize it. This cannot be undone, so it only ever runs on explicit request.
// @Tags [Redundancy] HA Group Management
// @Accept json
// @Produce json
// @Param groupId path string true "HA group ID"
// @Param haApproval body model.HAApproval true "Approval"
// @Success 202 {object} model.HATransitionInfo
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups/{groupId}/rejoin [post]
func RestPostHAGroupRejoin(c echo.Context) error {
	return notImplemented(c)
}

// RestPostHAGroupAcknowledge godoc
// @ID PostHAGroupAcknowledge
// @Summary Clear an HA group's error state
// @Description Not implemented yet: answers 501. A group in error stays there until an operator has dealt with the cause and says so here.
// @Tags [Redundancy] HA Group Management
// @Accept json
// @Produce json
// @Param groupId path string true "HA group ID"
// @Param haApproval body model.HAApproval true "Approval"
// @Success 200 {object} model.HAGroupInfo
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups/{groupId}/acknowledge [post]
func RestPostHAGroupAcknowledge(c echo.Context) error {
	return notImplemented(c)
}

// RestGetHAGroupTransitions godoc
// @ID GetHAGroupTransitions
// @Summary List an HA group's transitions
// @Description Not implemented yet: answers 501. Will list the group's switches and recoveries, with who approved each and whether fencing was confirmed.
// @Tags [Redundancy] HA Group Management
// @Produce json
// @Param groupId path string true "HA group ID"
// @Param limit query int false "How many to return, 1-200" default(50)
// @Param offset query int false "How many to skip" default(0)
// @Success 200 {object} model.HATransitionListResp
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/groups/{groupId}/transitions [get]
func RestGetHAGroupTransitions(c echo.Context) error {
	return notImplemented(c)
}

// RestGetHAEvents godoc
// @ID GetHAEvents
// @Summary List HA state changes
// @Description Not implemented yet: answers 501. Will list state changes in sequence order, so that a reader can resume from the last one it saw.
// @Tags [Redundancy] HA Group Management
// @Produce json
// @Param groupId query string false "Only this group's events"
// @Param after query int false "Only events after this sequence number" default(0)
// @Param limit query int false "How many to return, 1-200" default(50)
// @Success 200 {object} model.HAEventListResp
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /ha/events [get]
func RestGetHAEvents(c echo.Context) error {
	return notImplemented(c)
}
