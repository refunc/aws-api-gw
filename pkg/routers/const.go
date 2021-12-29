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

var (
	allowSignHeaders = map[string]struct{}{
		"Cache-Control":                         {},
		"Content-Disposition":                   {},
		"Content-Encoding":                      {},
		"Content-Language":                      {},
		"Content-Md5":                           {},
		"Content-Type":                          {},
		"Expires":                               {},
		"If-Match":                              {},
		"If-Modified-Since":                     {},
		"If-None-Match":                         {},
		"If-Unmodified-Since":                   {},
		"Range":                                 {},
		"X-Amz-Acl":                             {},
		"X-Amz-Copy-Source":                     {},
		"X-Amz-Copy-Source-If-Match":            {},
		"X-Amz-Copy-Source-If-Modified-Since":   {},
		"X-Amz-Copy-Source-If-None-Match":       {},
		"X-Amz-Copy-Source-If-Unmodified-Since": {},
		"X-Amz-Copy-Source-Range":               {},
		"X-Amz-Copy-Source-Server-Side-Encryption-Customer-Algorithm": {},
		"X-Amz-Copy-Source-Server-Side-Encryption-Customer-Key":       {},
		"X-Amz-Copy-Source-Server-Side-Encryption-Customer-Key-Md5":   {},
		"X-Amz-Grant-Full-control":                                    {},
		"X-Amz-Grant-Read":                                            {},
		"X-Amz-Grant-Read-Acp":                                        {},
		"X-Amz-Grant-Write":                                           {},
		"X-Amz-Grant-Write-Acp":                                       {},
		"X-Amz-Metadata-Directive":                                    {},
		"X-Amz-Mfa":                                                   {},
		"X-Amz-Request-Payer":                                         {},
		"X-Amz-Server-Side-Encryption":                                {},
		"X-Amz-Server-Side-Encryption-Aws-Kms-Key-Id":                 {},
		"X-Amz-Server-Side-Encryption-Customer-Algorithm":             {},
		"X-Amz-Server-Side-Encryption-Customer-Key":                   {},
		"X-Amz-Server-Side-Encryption-Customer-Key-Md5":               {},
		"X-Amz-Storage-Class":                                         {},
		"X-Amz-Tagging":                                               {},
		"X-Amz-Website-Redirect-Location":                             {},
		"X-Amz-Content-Sha256":                                        {},
	}
)

type Config struct {
	Rbac bool
}
