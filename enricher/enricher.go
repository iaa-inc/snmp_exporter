// Package enricher is responsible for processing the collected data, and enriching it with IAA specific labels.
package enricher

import (
	"regexp"
	"strings"

	"github.com/iaa-inc/gosdk"
	"github.com/iaa-inc/gosdk/admin"
)

// port_a1b2c3d
var portParser = regexp.MustCompile(`port_([a-zA-Z0-9]{7})`)
var extremePortRegex = regexp.MustCompile(`:`)

type Enricher struct {
	api   *gosdk.AdminClient
	cache *Cache
}

func NewEnricher(
	api *gosdk.AdminClient,
	cache *Cache,
) *Enricher {
	return &Enricher{
		api:   api,
		cache: cache,
	}
}

func (e *Enricher) Enrich(target string, labels map[string]string) map[string]string {
	// strip "ix.asn.au" from target if it's there
	target = strings.Replace(target, ".ix.asn.au", "", -1)

	ifAlias := labels["ifAlias"]
	ifName := labels["ifName"]
	ifDescr := labels["ifDescr"]

	portId := ""
	if ifAlias != "" {
		matches := portParser.FindStringSubmatch(ifAlias)
		if len(matches) > 1 {
			portId = matches[1]
		}
	}

	var port *admin.Port

	if portId != "" {
		port = e.cache.GetPort(portId)
	}

	if port == nil {
		port = e.cache.GetPortByIfDescr(ifName, target)
	}

	if port == nil {
		port = e.cache.GetPortByIfDescr(ifDescr, target)
	}

	if port == nil {
		// if ifName has a ':' in it, it may be a dumb extreme, so let's split that and only use the second part
		parts := extremePortRegex.Split(ifName, 2)
		if len(parts) > 1 {
			port = e.cache.GetPortByIfDescr(parts[1], target)
		}
	}

	if port != nil {
		labels["port_id"] = port.Service_id
		labels["member"] = port.Account.Name
		labels["exchange"] = port.Exchange.Name
		labels["facility"] = port.Facility.Name
	}

	return labels
}
