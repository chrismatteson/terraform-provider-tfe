package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceTFETeam() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamCreate,
		Read:   resourceTFETeamRead,
		Update: resourceTFETeamUpdate,
		Delete: resourceTFETeamDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFETeamImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"organization_access": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"manage_policies": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"manage_workspaces": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"manage_vcs_settings": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceTFETeamCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get team attributes.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)
	access := getTeamOrganizationAccess(d)

	// Create a new options struct.
	options := tfe.TeamCreateOptions{
		Name:               tfe.String(name),
		OrganizationAccess: access,
	}

	log.Printf("[DEBUG] Create team %s for organization: %s", name, organization)
	team, err := tfeClient.Teams.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating team %s for organization %s: %v", name, organization, err)
	}

	d.SetId(team.ID)

	return resourceTFETeamRead(d, meta)
}

func resourceTFETeamRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of team: %s", d.Id())
	team, err := tfeClient.Teams.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Team %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of team %s: %v", d.Id(), err)
	}

	// Update the config.
	d.Set("name", team.Name)
	d.Set("organization_access.0.manage_policies", team.OrganizationAccess.ManagePolicies)
	d.Set("organization_access.0.manage_workspaces", team.OrganizationAccess.ManageWorkspaces)
	d.Set("organization_access.0.manage_vcs_settings", team.OrganizationAccess.ManageVCSSettings)

	return nil
}

func resourceTFETeamUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	access := getTeamOrganizationAccess(d)

	// create an options struct
	options := tfe.TeamUpdateOptions{
		Name:               tfe.String(name),
		OrganizationAccess: access,
	}

	log.Printf("[DEBUG] Update team: %s", d.Id())
	_, err := tfeClient.Teams.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf(
			"Error updating team %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFETeamDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete team: %s", d.Id())
	err := tfeClient.Teams.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting team %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFETeamImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid team import format: %s (expected <ORGANIZATION>/<TEAM ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("organization", s[0])
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}

func getTeamOrganizationAccess(d *schema.ResourceData) *tfe.OrganizationAccessOptions {
	options := &tfe.OrganizationAccessOptions{}

	if attr, ok := d.GetOk("organization_access.0.manage_policies"); ok {
		options.ManagePolicies = tfe.Bool(attr.(bool))
	}

	if attr, ok := d.GetOk("organization_access.0.manage_workspaces"); ok {
		options.ManageWorkspaces = tfe.Bool(attr.(bool))
	}

	if attr, ok := d.GetOk("organization_access.0.manage_vcs_settings"); ok {
		options.ManageVCSSettings = tfe.Bool(attr.(bool))
	}

	return options
}
