package routers

const (
	LambdaLabelServiceAccount = "lambda.refunc.io/service-account"
	authorizationHeader       = "Authorization"
	authHeaderSignatureElem   = "Signature="
	signatureQueryKey         = "X-Amz-Signature"

	authHeaderPrefix = "AWS4-HMAC-SHA256"
	timeFormat       = "20060102T150405Z"
	shortTimeFormat  = "20060102"
	awsV4Request     = "aws4_request"
)

type Config struct {
	Rbac bool
}
