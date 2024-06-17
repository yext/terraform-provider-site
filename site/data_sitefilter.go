package site

import (
	"fmt"

	"github.com/gobwas/glob"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceSitefilter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSitefilterRead,
		Schema: map[string]*schema.Schema{
			"filter": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateGlob,
			},
			"separator": &schema.Schema{
				Type:         schema.TypeString,
				Default:      ".",
				ValidateFunc: validateSeparator,
			},
			"site_configs": &schema.Schema{
				Type:     schema.TypeMap,
				Required: true,
			},
			"sites": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func validateGlob(val interface{}, key string) ([]string, []error) {
	if _, err := glob.Compile(val.(string)); err != nil {
		return nil, []error{fmt.Errorf("%q must be a valid glob: %w", key, err)}
	}
	return nil, nil
}

func validateSeparator(val interface{}, key string) ([]string, []error) {
	if len([]rune(val.(string))) != 1 {
		return nil, []error{fmt.Errorf("%q must be a single character", key)}
	}
	return nil, nil
}

func dataSourceSitefilterRead(d *schema.ResourceData, _ any) error {
	var (
		filter         = d.Get("filter").(string)
		separator      = d.Get("separator").(string)
		rawSiteConfigs = d.Get("site_configs").(map[string]any)
	)

	filterGlob, err := glob.Compile(filter, []rune(separator)[0])
	if err != nil {
		return fmt.Errorf("invalid filter: %w", err)
	}

	siteMetadata, err := unmarshalSiteMetadata(rawSiteConfigs)
	if err != nil {
		return fmt.Errorf("invalid site_configs: %w", err)
	}

	filteredSites := map[string]any{}

	for key, meta := range siteMetadata {
		if filterGlob.Match(meta.FQN()) {
			filteredSites[key] = rawSiteConfigs[key]
		}
	}

	d.SetId(filter)
	d.Set("sites", filteredSites)
	return nil
}

func unmarshalSiteMetadata(rawSiteConfigs map[string]any) (map[string]SiteMetadata, error) {
	metas := map[string]SiteMetadata{}

	for key, raw := range rawSiteConfigs {
		asMap, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid site_configs[%d]: must be a map", key)
		}
		meta, err := NewSiteMetadata(asMap)
		if err != nil {
			return nil, fmt.Errorf("invalid site_configs[%d]: %w", key, err)
		}
		metas[key] = meta
	}

	return metas, nil
}

// SiteMetadata contains the metadata of a site required to construct its
// fully qualified name (FQN).
type SiteMetadata map[string]string

// NewSiteMetadata returns a SiteMetadata constructed from the specified raw config map.
func NewSiteMetadata(rawConfig map[string]any) (SiteMetadata, error) {
	m := map[string]string{}

	for _, k := range siteConfigRequiredKeys {
		v, err := configAsString(rawConfig, k)
		if err != nil {
			return SiteMetadata(m), fmt.Errorf("invalid key %q: %w", k, err)
		}
		m[k] = v
	}
	return SiteMetadata(m), nil
}

var siteConfigRequiredKeys = []string{
	"env",
	"partition",
	"servingRole",
	"dataCenter",
}

// FQN returns the fully qualified name of the site described by its metadata.
func (m SiteMetadata) FQN() string {
	return fmt.Sprintf(
		"%s.%s.%s.%s",
		m["env"],
		m["partition"],
		m["servingRole"],
		m["dataCenter"],
	)
}

func configAsString(rawSiteConfig map[string]any, key string) (string, error) {
	if v, ok := rawSiteConfig[key]; !ok {
		return "", fmt.Errorf("%q is missing", key)
	} else if vStr, ok := v.(string); !ok {
		return "", fmt.Errorf("%q is not a string", key)
	} else {
		return vStr, nil
	}
}
