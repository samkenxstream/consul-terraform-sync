package tftmpl

import (
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

var (
	_ Monitor = (*CatalogServicesMonitor)(nil)
)

// CatalogServicesMonitor handles appending templating for the catalog-service
// run monitor
type CatalogServicesMonitor struct {
	Regexp     string
	Datacenter string
	Namespace  string
	NodeMeta   map[string]string

	SourceIncludesVar bool
}

// isServicesVar returns false because the tmplfunc returns a catalog_services
// variable, not a services variable
func (m CatalogServicesMonitor) isServicesVar() bool {
	return false
}

func (m CatalogServicesMonitor) SourceIncludesVariable() bool {
	return m.SourceIncludesVar
}

func (m CatalogServicesMonitor) appendModuleAttribute(body *hclwrite.Body) {
	body.SetAttributeTraversal("catalog_services", hcl.Traversal{
		hcl.TraverseRoot{Name: "var"},
		hcl.TraverseAttr{Name: "catalog_services"},
	})
}

func (m CatalogServicesMonitor) appendTemplate(w io.Writer) error {
	q := m.hcatQuery()

	if m.SourceIncludesVar {
		_, err := fmt.Fprintf(w, catalogServicesIncludesVarTmpl, q)
		if err != nil {
			err = fmt.Errorf("unable to write catalog-service template to include variable, error: %v", err)
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, catalogServicesEmptyTmpl, q); err != nil {
		err = fmt.Errorf("unable to write catalog-service empty template, error %v", err)
		return err
	}
	return nil
}

func (m CatalogServicesMonitor) appendVariable(w io.Writer) error {
	_, err := w.Write(variableCatalogServices)
	return err
}

func (m CatalogServicesMonitor) hcatQuery() string {
	var opts []string

	if m.Regexp != "" {
		opts = append(opts, fmt.Sprintf("regexp=%s", m.Regexp))
	}

	if m.Datacenter != "" {
		opts = append(opts, fmt.Sprintf("dc=%s", m.Datacenter))
	}

	if m.Namespace != "" {
		opts = append(opts, fmt.Sprintf("ns=%s", m.Namespace))
	}

	for k, v := range m.NodeMeta {
		opts = append(opts, fmt.Sprintf("node-meta=%s:%s", k, v))
	}

	if len(opts) > 0 {
		return `"` + strings.Join(opts, `" "`) + `" ` // deliberate space at end
	}
	return ""
}

var catalogServicesIncludesVarTmpl = fmt.Sprintf(`
catalog_services = {%s}
`, catalogServicesBaseTmpl)

const catalogServicesBaseTmpl = `
{{- with $catalogServices := catalogServicesRegistration %s}}
  {{- range $cs := $catalogServices }}
  "{{ $cs.Name }}" = {{ HCLServiceTags $cs.Tags }}
{{- end}}{{- end}}
`

const catalogServicesEmptyTmpl = `
{{- with $catalogServices := catalogServicesRegistration %s}}
  {{- range $cs := $catalogServices }}
    {{- /* Empty template. Detects changes in catalog-services */ -}}
{{- end}}{{- end}}
`

// variableCatalogServices is required for modules that include catalog-services
// information. It is versioned to track compatibility between the generated
// root module and modules that include catalog-services.
var variableCatalogServices = []byte(`
# Catalog Services definition protocol v0
variable "catalog_services" {
  description = "Consul catalog service names and list of all known tags for a given service"
  type        = map(list(string))
}
`)
