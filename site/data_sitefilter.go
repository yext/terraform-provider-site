package site

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gobwas/glob"
	"github.com/hashicorp/terraform/helper/schema"

	"gopkg.in/yaml.v2"
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
				Optional:     true,
				Default:      ".",
				ValidateFunc: validateSeparator,
			},
			"site_yamls": &schema.Schema{
				Type:         schema.TypeMap,
				Required:     true,
				ValidateFunc: validateSiteConfigs,
			},
			"sites": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func validateGlob(val interface{}, key string) ([]string, []error) {
	// Whitespace is ignored for these filters.
	filter := removeSpaces(val.(string))

	if _, err := glob.Compile(filter); err != nil {
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

func validateSiteConfigs(val interface{}, key string) ([]string, []error) {
	var errs []error
	for id, siteYAML := range val.(map[string]any) {
		if _, err := NewSiteMetadata(siteYAML.(string)); err != nil {
			errs = append(errs, fmt.Errorf("invalid %q[%d]: %w", key, id, err))
		}
	}
	return nil, errs
}

func dataSourceSitefilterRead(d *schema.ResourceData, _ any) error {
	var (
		filter         = d.Get("filter").(string)
		separator      = d.Get("separator").(string)
		rawSiteConfigs = d.Get("site_yamls").(map[string]any)
	)

	// Whitespace is ignored for these filters.
	filter = removeSpaces(filter)

	filterGlob, err := glob.Compile(filter, []rune(separator)[0])
	if err != nil {
		return fmt.Errorf("invalid filter: %w", err)
	}

	siteMetadata, err := unmarshalSiteMetadata(rawSiteConfigs)
	if err != nil {
		return fmt.Errorf("invalid site_yamls: %w", err)
	}

	var matchedSites []string

	for id, meta := range siteMetadata {
		if filterGlob.Match(meta.FQN()) {
			matchedSites = append(matchedSites, id)
		}
	}

	d.SetId(filter)
	d.Set("sites", matchedSites)
	return nil
}

func unmarshalSiteMetadata(rawSiteConfigs map[string]any) (map[string]SiteMetadata, error) {
	metas := map[string]SiteMetadata{}

	for id, siteYAML := range rawSiteConfigs {
		meta, err := NewSiteMetadata(siteYAML.(string))
		if err != nil {
			return nil, fmt.Errorf("invalid site_yamls[%d]: %w", id, err)
		}
		metas[id] = meta
	}

	return metas, nil
}

// SiteMetadata contains the metadata of a site required to construct its
// fully qualified name (FQN).
type SiteMetadata struct {
	Env         string `yaml:"env"`
	Partition   string `yaml:"partition"`
	ServingRole string `yaml:"servingRole"`
	DataCenter  string `yaml:"dataCenter"`
}

// NewSiteMetadata returns a SiteMetadata constructed from the specified raw config map.
func NewSiteMetadata(configYAML string) (SiteMetadata, error) {
	var m SiteMetadata
	err := yaml.Unmarshal([]byte(configYAML), &m)
	if err != nil {
		return m, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	switch {
	case m.Env == "":
		return m, fmt.Errorf("missing key: env")
	case m.Partition == "":
		return m, fmt.Errorf("missing key: partition")
	case m.ServingRole == "":
		return m, fmt.Errorf("missing key: servingRole")
	case m.DataCenter == "":
		return m, fmt.Errorf("missing key: dataCenter")
	}

	return m, nil
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
		m.Env,
		m.Partition,
		m.ServingRole,
		m.DataCenter,
	)
}

// removeSpaces returns the whitespace-less version of s.
func removeSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
