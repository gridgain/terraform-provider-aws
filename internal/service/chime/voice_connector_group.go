// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_chime_voice_connector_group")
func ResourceVoiceConnectorGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVoiceConnectorGroupCreate,
		ReadWithoutTimeout:   resourceVoiceConnectorGroupRead,
		UpdateWithoutTimeout: resourceVoiceConnectorGroupUpdate,
		DeleteWithoutTimeout: resourceVoiceConnectorGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"connector": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"voice_connector_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						"priority": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 99),
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
		},
	}
}

func resourceVoiceConnectorGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.CreateVoiceConnectorGroupInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("connector"); ok && v.(*schema.Set).Len() > 0 {
		input.VoiceConnectorItems = expandVoiceConnectorItems(v.(*schema.Set).List())
	}

	resp, err := conn.CreateVoiceConnectorGroup(ctx, input)
	if err != nil || resp.VoiceConnectorGroup == nil {
		return diag.Errorf("creating Chime Voice Connector group: %s", err)
	}

	d.SetId(aws.ToString(resp.VoiceConnectorGroup.VoiceConnectorGroupId))

	return resourceVoiceConnectorGroupRead(ctx, d, meta)
}

func resourceVoiceConnectorGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	resp, err := FindVoiceConnectorResourceWithRetry(ctx, d.IsNewResource(), func() (*awstypes.VoiceConnectorGroup, error) {
		return findVoiceConnectorGroupByID(ctx, conn, d.Id())
	})

	if tfresource.TimedOut(err) {
		resp, err = findVoiceConnectorGroupByID(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Voice conector group %s not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Voice Connector Group (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.Name)

	if err := d.Set("connector", flattenVoiceConnectorItems(resp.VoiceConnectorItems)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Chime Voice Connector group items (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceVoiceConnectorGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.UpdateVoiceConnectorGroupInput{
		Name:                  aws.String(d.Get("name").(string)),
		VoiceConnectorGroupId: aws.String(d.Id()),
	}

	if d.HasChange("connector") {
		if v, ok := d.GetOk("connector"); ok {
			input.VoiceConnectorItems = expandVoiceConnectorItems(v.(*schema.Set).List())
		}
	} else if !d.IsNewResource() {
		input.VoiceConnectorItems = make([]awstypes.VoiceConnectorItem, 0)
	}

	if _, err := conn.UpdateVoiceConnectorGroup(ctx, input); err != nil {
		return diag.Errorf("updating Chime Voice Connector group (%s): %s", d.Id(), err)
	}

	return resourceVoiceConnectorGroupRead(ctx, d, meta)
}

func resourceVoiceConnectorGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	if v, ok := d.GetOk("connector"); ok && v.(*schema.Set).Len() > 0 {
		if err := resourceVoiceConnectorGroupUpdate(ctx, d, meta); err != nil {
			return err
		}
	}

	input := &chimesdkvoice.DeleteVoiceConnectorGroupInput{
		VoiceConnectorGroupId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteVoiceConnectorGroup(ctx, input); err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			log.Printf("[WARN] Chime Voice conector group %s not found", d.Id())
			return nil
		}
		return diag.Errorf("deleting Chime Voice Connector group (%s): %s", d.Id(), err)
	}

	return nil
}

func expandVoiceConnectorItems(data []interface{}) []awstypes.VoiceConnectorItem {
	var connectorsItems []awstypes.VoiceConnectorItem

	for _, rItem := range data {
		item := rItem.(map[string]interface{})
		connectorsItems = append(connectorsItems, awstypes.VoiceConnectorItem{
			VoiceConnectorId: aws.String(item["voice_connector_id"].(string)),
			Priority:         aws.Int32(int32(item["priority"].(int))),
		})
	}

	return connectorsItems
}

func flattenVoiceConnectorItems(connectors []awstypes.VoiceConnectorItem) []interface{} {
	var rawConnectors []interface{}

	for _, c := range connectors {
		rawC := map[string]interface{}{
			"priority":           c.Priority,
			"voice_connector_id": aws.ToString(c.VoiceConnectorId),
		}
		rawConnectors = append(rawConnectors, rawC)
	}
	return rawConnectors
}

func findVoiceConnectorGroupByID(ctx context.Context, conn *chimesdkvoice.Client, id string) (*awstypes.VoiceConnectorGroup, error) {
	in := &chimesdkvoice.GetVoiceConnectorGroupInput{
		VoiceConnectorGroupId: aws.String(id),
	}

	resp, err := conn.GetVoiceConnectorGroup(ctx, in)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if resp == nil || resp.VoiceConnectorGroup == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if err != nil {
		return nil, err
	}

	return resp.VoiceConnectorGroup, nil
}
