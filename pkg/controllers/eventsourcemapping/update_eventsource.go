package eventsourcemapping

import (
	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
)

func UpdateEventSource(c *gin.Context) {
	//No suitable field to carry the payload, please fallback use delete&create to update.
	//https://docs.aws.amazon.com/lambda/latest/api/API_UpdateEventSourceMapping.html
	awsutils.AWSErrorResponse(c, 500, "ServiceException")
}
