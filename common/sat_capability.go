package common

const (
	SATV2AccessReadOnly  = "readonly"
	SATV2AccessReadWrite = "readwrite"
)

type SATV2Domain string

const (
	SATV2DomainCore    SATV2Domain = "core"
	SATV2DomainTagging SATV2Domain = "tagging"
	SATV2DomainSystem  SATV2Domain = "system"
	SATV2DomainMetrics SATV2Domain = "metrics"
)

type SATV2RouteDomainMapping struct {
	Prefix string
	Domain SATV2Domain
}

var SATV2RouteMappings = []SATV2RouteDomainMapping{
	{Prefix: "/taggingservice", Domain: SATV2DomainTagging},
	{Prefix: "/metrics", Domain: SATV2DomainMetrics},
	{Prefix: "/queries/filters/downloadlocation", Domain: SATV2DomainSystem},
	{Prefix: "/updates/filters/downloadlocation", Domain: SATV2DomainSystem},
	{Prefix: "/roundrobinfilter", Domain: SATV2DomainSystem},
	{Prefix: "/rfc/recooking", Domain: SATV2DomainSystem},
	{Prefix: "/rfc/preprocess", Domain: SATV2DomainSystem},
	{Prefix: "/appsettings", Domain: SATV2DomainSystem},
	{Prefix: "/canarysettings", Domain: SATV2DomainSystem},
	{Prefix: "/lockdownsettings", Domain: SATV2DomainSystem},
	{Prefix: "/wakeuppool", Domain: SATV2DomainSystem},
	{Prefix: "/dataservice", Domain: SATV2DomainCore},
	{Prefix: "/estbfirmware", Domain: SATV2DomainCore},
	{Prefix: "/queries", Domain: SATV2DomainCore},
	{Prefix: "/updates", Domain: SATV2DomainCore},
	{Prefix: "/delete", Domain: SATV2DomainCore},
	{Prefix: "/model", Domain: SATV2DomainCore},
	{Prefix: "/environment", Domain: SATV2DomainCore},
	{Prefix: "/genericnamespacedlist", Domain: SATV2DomainCore},
	{Prefix: "/firmwarerule", Domain: SATV2DomainCore},
	{Prefix: "/firmwareruletemplate", Domain: SATV2DomainCore},
	{Prefix: "/firmwareconfig", Domain: SATV2DomainCore},
	{Prefix: "/percentfilter", Domain: SATV2DomainCore},
	{Prefix: "/amv", Domain: SATV2DomainCore},
	{Prefix: "/activationminimumversion", Domain: SATV2DomainCore},
	{Prefix: "/settings", Domain: SATV2DomainCore},
	{Prefix: "/setting", Domain: SATV2DomainCore},
	{Prefix: "/featurerule", Domain: SATV2DomainCore},
	{Prefix: "/feature", Domain: SATV2DomainCore},
	{Prefix: "/rfc", Domain: SATV2DomainCore},
	{Prefix: "/changelog", Domain: SATV2DomainCore},
	{Prefix: "/log", Domain: SATV2DomainCore},
	{Prefix: "/reportpage", Domain: SATV2DomainCore},
	{Prefix: "/stats", Domain: SATV2DomainCore},
	{Prefix: "/migration", Domain: SATV2DomainCore},
	{Prefix: "/dcm", Domain: SATV2DomainCore},
	{Prefix: "/telemetry", Domain: SATV2DomainCore},
	{Prefix: "/change", Domain: SATV2DomainCore},
	{Prefix: "/penetrationdata", Domain: SATV2DomainCore},
}

func BuildSATV2Capability(domain string, access string) string {
	return "xconf:" + domain + ":" + access
}

func HasSATV2Capability(capabilities []string, domain string, access string) bool {
	required := BuildSATV2Capability(domain, access)
	for _, capability := range capabilities {
		if capability == required {
			return true
		}
	}
	return false
}
