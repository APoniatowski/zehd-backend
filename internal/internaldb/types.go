package internaldb

type CollectionData struct {
	FrontendName string `json:"frontendName"`
	TimeDate     int64  `json:"timeDate"`
	IP           string `json:"ip"`
	Port         int    `json:"port"`
	Path         string `json:"path"`
	Method       string `json:"method"`
	XForwardFor  string `json:"XForwardFor"`
	XRealIP      string `json:"XRealIP"`
	UserAgent    string `json:"useragent"`
	Via          string `json:"via"`
	Age          string `json:"age"`
	CFIPCountry  string `json:"CF-IPCountry"`
}

type BannedData struct {
	FrontendName    string `json:"frontendName"`
	TimeDateBanned  int64  `json:"timeDateBanned"`
	TimeDateChecked int64  `json:"timeDateChecked"`
	IP              string `json:"ip"`
	DomainName      string `json:"domainName"`
	Banned          bool   `json:"banned"`
}
