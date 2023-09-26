package apis

type ListURLResponse struct {
	FunctionUrlConfigs []FunctionURLConfig `json:"FunctionUrlConfigs"`
	NextMarker         string              `json:"NextMarker,omitempty"`
}

type FunctionURLConfig struct {
	AuthType         string  `json:"AuthType"`
	Cors             URLCors `json:"Cors,omitempty"`
	CreationTime     string  `json:"CreationTime"`
	FunctionArn      string  `json:"FunctionArn"`
	FunctionUrl      string  `json:"FunctionUrl"`
	LastModifiedTime string  `json:"LastModifiedTime"`
}

type URLCors struct {
	AllowCredentials bool     `json:"AllowCredentials,omitempty"`
	AllowHeaders     []string `json:"AllowHeaders,omitempty"`
	AllowMethods     []string `json:"AllowMethods,omitempty"`
	AllowOrigins     []string `json:"AllowOrigins,omitempty"`
	ExposeHeaders    []string `json:"ExposeHeaders,omitempty"`
	MaxAge           int      `json:"MaxAge,omitempty"`
}
