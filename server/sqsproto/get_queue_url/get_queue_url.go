package get_queue_url

import (
	"encoding/xml"
	"net/http"

	"github.com/vburenin/firempq/apis"
	"github.com/vburenin/firempq/qmgr"
	"github.com/vburenin/firempq/server/sqsproto/sqs_response"
	"github.com/vburenin/firempq/server/sqsproto/sqserr"
	"github.com/vburenin/firempq/server/sqsproto/urlutils"
)

type GetQueueUrlResult struct {
	XMLName   xml.Name `xml:"http://queue.amazonaws.com/doc/2012-11-05/ GetQueueUrlResult"`
	QueueUrl  string   `xml:"GetQueueUrlResult>QueueUrl"`
	RequestId string   `xml:"ResponseMetadata>RequestId"`
}

func (r *GetQueueUrlResult) XmlDocument() string                  { return sqs_response.EncodeXml(r) }
func (r *GetQueueUrlResult) HttpCode() int                        { return http.StatusOK }
func (r *GetQueueUrlResult) BatchResult(docId string) interface{} { return nil }

func GetQueueUrl(svcMgr *qmgr.ServiceManager, sqsQuery *urlutils.SQSQuery) sqs_response.SQSResponse {
	svc, ok := svcMgr.GetService(sqsQuery.QueueName)
	if !ok {
		return sqserr.QueueDoesNotExist()
	}
	if svc.Info().Type != apis.ServiceTypePriorityQueue {
		return sqserr.QueueDoesNotExist()
	}

	return &GetQueueUrlResult{
		QueueUrl:  sqsQuery.Host + "/queue/" + sqsQuery.QueueName,
		RequestId: "1111-2222-3333",
	}
}
