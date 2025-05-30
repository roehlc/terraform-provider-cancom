package ipam

import (
	"context"

	"github.com/cancom/terraform-provider-cancom/client"
	client_ipam "github.com/cancom/terraform-provider-cancom/client/services/ipam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceHost() *schema.Resource {
	return &schema.Resource{
		Description: `IP Management --- IPAM host allows you to assign host ip-addresses from a network that must be enabled for host assignment. The network is identified by it's id (crn).  
You manage networks and enable them for host-assignments by using the portal or the network resource of this provider.`,
		CreateContext: resourceHostCreate,
		ReadContext:   resourceHostRead,
		UpdateContext: resourceHostUpdate,
		DeleteContext: resourceHostDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_crn": {
				Type:     schema.TypeString,
				Computed: false,
				Required: true,
				ForceNew: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
				ForceNew: true,
			},
			"name_tag": {
				Type:     schema.TypeString,
				Computed: false,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := m.(*client.CcpClient).GetService("ip-management")
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	id := d.Id()

	resp, err := (*client_ipam.Client)(c).GetHost(id)

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name_tag", resp.NameTag)
	d.Set("address", resp.Address)
	d.Set("description", resp.Description)
	d.Set("id", resp.ID)

	return diags
}

func resourceHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := m.(*client.CcpClient).GetService("ip-management")
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	host := &client_ipam.HostCreateRequest{
		NetworkCrn:  d.Get("network_crn").(string),
		NameTag:     d.Get("name_tag").(string),
		Operation:   "assign_address",
		Description: d.Get("description").(string),
		Qualifier:   d.Get("qualifier").(string),
	}

	resp, err := (*client_ipam.Client)(c).CreateHost(host)

	if err != nil {
		return diag.FromErr(err)
	}

	id := resp.ID

	d.Set("address", resp.Address)
	d.SetId(id)

	resourceHostRead(ctx, d, m)

	return diags

}

func resourceHostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := m.(*client.CcpClient).GetService("ip-management")
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	id := d.Id()

	host := &client_ipam.HostUpdateRequest{
		NetworkCrn:  d.Get("network_crn").(string),
		NameTag:     d.Get("name_tag").(string),
		Description: d.Get("description").(string),
	}

	_, err = (*client_ipam.Client)(c).UpdateHost(id, host)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	resourceHostRead(ctx, d, m)

	return diags
}

func resourceHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := m.(*client.CcpClient).GetService("ip-management")
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	id := d.Id()

	err = (*client_ipam.Client)(c).DeleteHost(id)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
