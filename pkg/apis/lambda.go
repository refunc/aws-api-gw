package apis

type CreateFunctionRequest struct {
	Architectures        []string            `json:"Architectures"`
	Code                 map[string]string   `json:"Code"`
	CodeSigningConfigArn string              `json:"CodeSigningConfigArn"`
	DeadLetterConfig     map[string]string   `json:"DeadLetterConfig"`
	Description          string              `json:"Description"`
	Environment          FunctionEnvironment `json:"Environment"`
	FileSystemConfigs    []map[string]string `json:"FileSystemConfigs"`
	FunctionName         string              `json:"FunctionName"`
	Handler              string              `json:"Handler"`
	ImageConfig          FunctionImageConfig `json:"ImageConfig"`
	KMSKeyArn            string              `json:"KMSKeyArn"`
	Layers               []string            `json:"Layers"`
	MemorySize           int64               `json:"MemorySize"`
	PackageType          string              `json:"PackageType"`
	Publish              bool                `json:"Publish"`
	Role                 string              `json:"Role"`
	Runtime              string              `json:"Runtime"`
	Tags                 map[string]string   `json:"Tags"`
	Timeout              int64               `json:"Timeout"`
	TracingConfig        map[string]string   `json:"TracingConfig"`
	VpcConfig            FunctionVpcConfig   `json:"VpcConfig"`
}

type CreateFunctionResponse struct {
	FunctionConfiguration `json:",inline"`
}

type GetFunctionResponse struct {
	Code          map[string]string     `json:"Code"`
	Concurrency   FunctionConcurrency   `json:"Concurrency"`
	Configuration FunctionConfiguration `json:"Configuration"`
	Tags          map[string]string     `json:"Tags"`
}

type ListFunctionResponse struct {
	Functions  []FunctionConfiguration `json:"Functions"`
	NextMarker string                  `json:"NextMarker,omitempty"`
}

type FunctionConfiguration struct {
	Architectures              []string                     `json:"Architectures,omitempty"`
	CodeSha256                 string                       `json:"CodeSha256,omitempty"`
	CodeSize                   int64                        `json:"CodeSize,omitempty"`
	DeadLetterConfig           map[string]string            `json:"DeadLetterConfig,omitempty"`
	Description                string                       `json:"Description,omitempty"`
	Environment                *FunctionEnvironment         `json:"Environment,omitempty"`
	FileSystemConfigs          []map[string]string          `json:"FileSystemConfigs,omitempty"`
	FunctionArn                string                       `json:"FunctionArn,omitempty"`
	FunctionName               string                       `json:"FunctionName,omitempty"`
	Handler                    string                       `json:"Handler,omitempty"`
	ImageConfigResponse        *FunctionImageConfigResponse `json:"ImageConfigResponse,omitempty"`
	KMSKeyArn                  string                       `json:"KMSKeyArn,omitempty"`
	LastModified               string                       `json:"LastModified,omitempty"`
	LastUpdateStatus           string                       `json:"LastUpdateStatus,omitempty"`
	LastUpdateStatusReason     string                       `json:"LastUpdateStatusReason,omitempty"`
	LastUpdateStatusReasonCode string                       `json:"LastUpdateStatusReasonCode,omitempty"`
	Layers                     []FunctionLayers             `json:"Layers,omitempty"`
	MasterArn                  string                       `json:"MasterArn,omitempty"`
	MemorySize                 int64                        `json:"MemorySize,omitempty"`
	PackageType                string                       `json:"PackageType,omitempty"`
	RevisionId                 string                       `json:"RevisionId,omitempty"`
	Role                       string                       `json:"Role,omitempty"`
	Runtime                    string                       `json:"Runtime,omitempty"`
	SigningJobArn              string                       `json:"SigningJobArn,omitempty"`
	SigningProfileVersionArn   string                       `json:"SigningProfileVersionArn,omitempty"`
	State                      string                       `json:"State,omitempty"`
	StateReason                string                       `json:"StateReason,omitempty"`
	StateReasonCode            string                       `json:"StateReasonCode,omitempty"`
	Timeout                    int64                        `json:"Timeout,omitempty"`
	TracingConfig              map[string]string            `json:"TracingConfig,omitempty"`
	Version                    string                       `json:"Version,omitempty"`
	VpcConfig                  *FunctionVpcConfig           `json:"VpcConfig,omitempty"`
}

type FunctionEnvironment struct {
	Error     map[string]string `json:"Error"`
	Variables map[string]string `json:"Variables"`
}

type FunctionImageConfig struct {
	Command          []string `json:"Command"`
	EntryPoint       []string `json:"EntryPoint"`
	WorkingDirectory string   `json:"WorkingDirectory"`
}

type FunctionImageConfigResponse struct {
	Error       map[string]string   `json:"Error"`
	ImageConfig FunctionImageConfig `json:"ImageConfig"`
}

type FunctionLayers struct {
	Arn                      string `json:"Arn"`
	CodeSize                 int64  `json:"CodeSize"`
	SigningJobArn            string `json:"SigningJobArn"`
	SigningProfileVersionArn string `json:"SigningProfileVersionArn"`
}

type FunctionVpcConfig struct {
	SecurityGroupIds []string `json:"SecurityGroupIds"`
	SubnetIds        []string `json:"SubnetIds"`
	VpcId            string   `json:"VpcId"`
}

type FunctionConcurrency struct {
	Concurrency int64 `json:"Concurrency"`
}
