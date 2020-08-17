package awx

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awx "github.com/mrcrilly/goawx/client"
	"strconv"
)

func resourceCredentialAzure() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCredentialAzureCreate,
		ReadContext:   resourceCredentialAzureRead,
		UpdateContext: resourceCredentialAzureUpdate,
		DeleteContext: awxAPIDeleteByID,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"organisation_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"url": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"client": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"secret": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				//DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				//	// Because the AWX API always returns "$encrypted$" for secret strings, we can never
				//	// compare the value in the HCL to the API to determine if it has changed or not. This
				//	// results in Terraform always wanting to update the value. We can either leave it to
				//	// always update the value or use this function to essentially completely ignore changes
				//	// to this field.
				//	return true
				//},
			},
			"tenant": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCredentialAzureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	newCredential := map[string]interface{}{
		"name":            d.Get("name").(string),
		"description":     d.Get("description").(string),
		"organization":    d.Get("organisation_id").(int),
		"credential_type": 19, // Azure Key Vault
		"inputs": map[string]interface{}{
			"url":    d.Get("url").(string),
			"client": d.Get("client").(string),
			"secret": d.Get("secret").(string),
			"tenant": d.Get("tenant").(string),
		},
	}

	client := m.(*awx.AWX)
	cred, err := client.CredentialsService.CreateCredentials(newCredential, map[string]string{})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create new credentials",
			Detail:   fmt.Sprintf("Unable to create new credentials: %s", err.Error()),
		})
		return diags
	}

	d.SetId(strconv.Itoa(cred.ID))
	resourceCredentialAzureRead(ctx, d, m)

	return diags
}

func resourceCredentialAzureRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*awx.AWX)
	id, _ := strconv.Atoi(d.Id())
	cred, err := client.CredentialsService.GetCredentialsByID(id, map[string]string{})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to fetch credentials",
			Detail:   fmt.Sprintf("Unable to credentials with id %d: %s", id, err.Error()),
		})
		return diags
	}

	d.Set("name", cred.Name)
	d.Set("description", cred.Description)
	d.Set("organisation_id", cred.OrganizationID)
	d.Set("url", cred.Inputs["url"])
	d.Set("client", cred.Inputs["client"])
	//d.Set("secret", cred.Inputs["secret"])
	d.Set("tenant", cred.Inputs["tenant"])

	return diags
}

func resourceCredentialAzureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	keys := []string{
		"name",
		"description",
		"url",
		"client",
		//"secret",
		"tenant",
	}

	if d.HasChanges(keys...) {
		var err error

		id, _ := strconv.Atoi(d.Id())
		updatedCredential := map[string]interface{}{
			"name":            d.Get("name").(string),
			"description":     d.Get("description").(string),
			"organization":    d.Get("organisation_id").(int),
			"credential_type": 19, // Azure Key Vault
			"inputs": map[string]interface{}{
				"url":    d.Get("url").(string),
				"client": d.Get("client").(string),
				"secret": d.Get("secret").(string),
				"tenant": d.Get("tenant").(string),
			},
		}

		client := m.(*awx.AWX)
		_, err = client.CredentialsService.UpdateCredentialsByID(id, updatedCredential, map[string]string{})
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update existing credentials",
				Detail:   fmt.Sprintf("Unable to update existing credentials with id %d: %s", id, err.Error()),
			})
			return diags
		}
	}

	return resourceCredentialAzureRead(ctx, d, m)
}
