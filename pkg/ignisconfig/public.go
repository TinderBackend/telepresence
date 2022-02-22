package ignisconfig

import "github.com/TinderBackend/telepresence/v2/pkg/iputil"

// DefaultAlsoProxy is a set of default CIDRs that will be auto-added to the AlsoProxy block
// supported here -> https://www.telepresence.io/docs/latest/reference/config/#Values
var DefaultAlsoProxy []*iputil.Subnet

// ServerURL is a URL pointing to the traffic manager server. The address format is
// "[objkind/]objname[.objnamespace]:port" e.g. svc/recsv2.edward-owens:8080.
var ServerURL = "svc/traffic-manager."
