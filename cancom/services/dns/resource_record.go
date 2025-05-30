package dns

import (
	"context"
	"strings"

	"github.com/cancom/terraform-provider-cancom/cancom/util"
	"github.com/cancom/terraform-provider-cancom/client"
	client_dns "github.com/cancom/terraform-provider-cancom/client/services/dns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// This is a bit of a hack to prevent multiple API requests for the same zone. The zone api locks after each individual request to wait until
// the record is available on all resolvers. Creating many records might therefore run into timeout and 429 Too Many Requests issues.
// This can still happen if the API is triggered externally, hence we also use retries for those cases.
var resourceRecordApiLock = util.NewLock()

func resourceRecord() *schema.Resource {
	return &schema.Resource{
		Description:   "DNS --- Defines a DNS record.",
		CreateContext: resourceRecordCreate,
		ReadContext:   resourceRecordRead,
		UpdateContext: resourceRecordUpdate,
		DeleteContext: resourceRecordDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the record",
				Required:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of the record (i.e. A, CNAME, ...)",
				Required:    true,
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Target of the record",
			},
			"ttl": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "TTL, if not set, defaults to the zones TTL",
			},
			"zone_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Zone that this record belongs to",
			},
			"zone_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The uuid of the zone",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The uuid of the record",
			},
			"comments": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "You can optionally add comments to records to describe their intended usage",
			},
			"last_change_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date at which the record was last updated",
			},
			"tenant": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The tenant that the record belongs to.",
			},
		},
	}
}

func resourceRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := m.(*client.CcpClient).GetService("domdns")
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	record := &client_dns.RecordCreateRequest{
		Name:     d.Get("name").(string),
		Type:     d.Get("type").(string),
		Content:  d.Get("content").(string),
		TTL:      d.Get("ttl").(int),
		ZoneName: d.Get("zone_name").(string),
		Tenant:   d.Get("tenant").(string),
	}

	if err := resourceRecordApiLock.Lock(ctx); err != nil {
		return diag.FromErr(err)
	}
	defer resourceRecordApiLock.Unlock()

	resp, err := (*client_dns.Client)(c).CreateRecord(record)

	if err != nil {
		return diag.FromErr(err)
	}

	id := resp.ID

	d.Set("last_change_date", resp.LastChangeDate)
	d.Set("zone_id", resp.ZoneID)
	d.Set("comments", resp.Comments)

	d.SetId(id)

	resourceRecordRead(ctx, d, m)

	return diags

}

func resourceRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := m.(*client.CcpClient).GetService("domdns")
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	id := d.Id()
	zoneName := d.Get("zone_name").(string)

	resp, err := (*client_dns.Client)(c).GetRecord(id, zoneName)

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", strings.Replace(resp.Name, "."+resp.ZoneName, "", 1))
	d.Set("type", resp.Type)
	d.Set("content", resp.Content)
	d.Set("ttl", resp.TTL)
	d.Set("zone_name", resp.ZoneName)
	d.Set("zone_id", resp.ZoneID)
	d.Set("id", resp.ID)
	d.Set("comments", resp.Comments)
	d.Set("last_change_date", resp.LastChangeDate)
	d.Set("tenant", resp.Tenant)

	return diags
}

func resourceRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := m.(*client.CcpClient).GetService("domdns")
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	id := d.Id()

	record := &client_dns.RecordUpdateRequest{
		Name:     d.Get("name").(string),
		Type:     d.Get("type").(string),
		Content:  d.Get("content").(string),
		TTL:      d.Get("ttl").(int),
		ZoneName: d.Get("zone_name").(string),
		ZoneID:   d.Get("zone_id").(string),
		Tenant:   d.Get("tenant").(string),
	}

	if err := resourceRecordApiLock.Lock(ctx); err != nil {
		return diag.FromErr(err)
	}
	defer resourceRecordApiLock.Unlock()

	_, err = (*client_dns.Client)(c).UpdateRecord(id, record)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	resourceRecordRead(ctx, d, m)

	return diags
}

func resourceRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := m.(*client.CcpClient).GetService("domdns")
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	id := d.Id()

	if err := resourceRecordApiLock.Lock(ctx); err != nil {
		return diag.FromErr(err)
	}
	defer resourceRecordApiLock.Unlock()

	err = (*client_dns.Client)(c).DeleteRecord(id)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
