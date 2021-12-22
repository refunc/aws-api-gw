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

type FunctionConfiguration struct {
	Architectures              []string                    `json:"Architectures"`
	CodeSha256                 string                      `json:"CodeSha256"`
	CodeSize                   int64                       `json:"CodeSize"`
	DeadLetterConfig           map[string]string           `json:"DeadLetterConfig"`
	Description                string                      `json:"Description"`
	Environment                FunctionEnvironment         `json:"Environment"`
	FileSystemConfigs          []map[string]string         `json:"FileSystemConfigs"`
	FunctionArn                string                      `json:"FunctionArn"`
	FunctionName               string                      `json:"FunctionName"`
	Handler                    string                      `json:"Handler"`
	ImageConfigResponse        FunctionImageConfigResponse `json:"ImageConfigResponse"`
	KMSKeyArn                  string                      `json:"KMSKeyArn"`
	LastModified               string                      `json:"LastModified"`
	LastUpdateStatus           string                      `json:"LastUpdateStatus"`
	LastUpdateStatusReason     string                      `json:"LastUpdateStatusReason"`
	LastUpdateStatusReasonCode string                      `json:"LastUpdateStatusReasonCode"`
	Layers                     []FunctionLayers            `json:"Layers"`
	MasterArn                  string                      `json:"MasterArn"`
	MemorySize                 int64                       `json:"MemorySize"`
	PackageType                string                      `json:"PackageType"`
	RevisionId                 string                      `json:"RevisionId"`
	Role                       string                      `json:"Role"`
	Runtime                    string                      `json:"Runtime"`
	SigningJobArn              string                      `json:"SigningJobArn"`
	SigningProfileVersionArn   string                      `json:"SigningProfileVersionArn"`
	State                      string                      `json:"State"`
	StateReason                string                      `json:"StateReason"`
	StateReasonCode            string                      `json:"StateReasonCode"`
	Timeout                    int64                       `json:"Timeout"`
	TracingConfig              map[string]string           `json:"TracingConfig"`
	Version                    string                      `json:"Version"`
	VpcConfig                  FunctionVpcConfig           `json:"VpcConfig"`
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
