package apis

type EventSourceMappingConfiguration struct {
	EventSourceArn         string                 `json:"EventSourceArn"` //arn:<trigger-type>:<trigger-name>
	FunctionArn            string                 `json:"FunctionName"`
	SelfManagedEventSource SelfManagedEventSource `json:"SelfManagedEventSource"`
}

type SelfManagedEventSource struct {
	Endpoints map[string][]string `json:"Endpoints"`
}