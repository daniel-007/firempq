package sqsproto

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/vburenin/firempq/pqueue"
	"github.com/vburenin/firempq/qmgr"
	"github.com/vburenin/firempq/server/sqsproto/change_message_visibility"
	"github.com/vburenin/firempq/server/sqsproto/change_message_visibility_batch"
	"github.com/vburenin/firempq/server/sqsproto/create_queue"
	"github.com/vburenin/firempq/server/sqsproto/delete_message"
	"github.com/vburenin/firempq/server/sqsproto/delete_message_batch"
	"github.com/vburenin/firempq/server/sqsproto/delete_queue"
	"github.com/vburenin/firempq/server/sqsproto/get_queue_attributes"
	"github.com/vburenin/firempq/server/sqsproto/get_queue_url"
	"github.com/vburenin/firempq/server/sqsproto/list_queues"
	"github.com/vburenin/firempq/server/sqsproto/purge_queue"
	"github.com/vburenin/firempq/server/sqsproto/receive_message"
	"github.com/vburenin/firempq/server/sqsproto/send_message"
	"github.com/vburenin/firempq/server/sqsproto/send_message_batch"
	"github.com/vburenin/firempq/server/sqsproto/set_queue_attributes"
	"github.com/vburenin/firempq/server/sqsproto/sqs_response"
	"github.com/vburenin/firempq/server/sqsproto/sqserr"
	"github.com/vburenin/firempq/server/sqsproto/urlutils"
)

type SQSRequestHandler struct {
	ServiceManager *qmgr.ServiceManager
}

func ParseQueueName(urlPath string) (string, error) {
	f := strings.SplitN(urlPath, "/", 2)
	if len(f) == 2 && f[0] == "pqueue" {
		return f[0], nil
	}
	return "", sqserr.MalformedInputError("Invalid URL Format")
}

func (self *SQSRequestHandler) handleManageActions(sqsQuery *urlutils.SQSQuery) sqs_response.SQSResponse {
	switch sqsQuery.Action {
	case "CreateQueue":
		return create_queue.CreateQueue(self.ServiceManager, sqsQuery)
	case "GetQueueUrl":
		return get_queue_url.GetQueueUrl(self.ServiceManager, sqsQuery)
	case "ListQueues":
		return list_queues.ListQueues(self.ServiceManager, sqsQuery)
	case "ListDeadLetterSourceQueues":
	}
	return sqserr.InvalidActionError(sqsQuery.Action)
}

func (self *SQSRequestHandler) handleQueueActions(pq *pqueue.PQueue, sqsQuery *urlutils.SQSQuery) sqs_response.SQSResponse {
	switch sqsQuery.Action {
	case "SendMessage":
		return send_message.SendMessage(pq, sqsQuery)
	case "SendMessageBatch":
		return send_message_batch.SendMessageBatch(pq, sqsQuery)
	case "DeleteMessage":
		return delete_message.DeleteMessage(pq, sqsQuery)
	case "DeleteMessageBatch":
		return delete_message_batch.DeleteMessageBatch(pq, sqsQuery)
	case "ReceiveMessage":
		return receive_message.ReceiveMessage(pq, sqsQuery)
	case "ChangeMessageVisibility":
		return change_message_visibility.ChangeMessageVisibility(pq, sqsQuery)
	case "ChangeMessageVisibilityBatch":
		return change_message_visibility_batch.ChangeMessageVisibilityBatch(pq, sqsQuery)
	case "DeleteQueue":
		return delete_queue.DeleteQueue(self.ServiceManager, sqsQuery)
	case "PurgeQueue":
		return purge_queue.PurgeQueue(pq, sqsQuery)
	case "GetQueueAttributes":
		return get_queue_attributes.GetQueueAttributes(pq, sqsQuery)
	case "SetQueueAttributes":
		return set_queue_attributes.SetQueueAttributes(pq, sqsQuery)
	case "AddPermission":
	case "RemovePermission":
	}
	return sqserr.InvalidActionError(sqsQuery.Action)
}

func (self *SQSRequestHandler) dispatchSQSQuery(r *http.Request) sqs_response.SQSResponse {
	var queuePath string

	sqsQuery, err := urlutils.ParseSQSQuery(r)
	if err != nil {
		return sqserr.ServiceDeniedError()
	}
	if sqsQuery.QueueUrl != "" {
		queueUrl, err := url.ParseRequestURI(sqsQuery.QueueUrl)
		if err != nil {
			return sqserr.ServiceDeniedError()
		}
		queuePath = queueUrl.Path
	} else {
		queuePath = r.URL.Path
	}
	if strings.HasPrefix(queuePath, "/queue/") {
		sqsQuery.QueueName = strings.SplitN(queuePath, "/queue/", 2)[1]
		svc, ok := self.ServiceManager.GetService(sqsQuery.QueueName)
		if !ok {
			return sqserr.QueueDoesNotExist()
		}
		pq, _ := svc.(*pqueue.PQueue)
		return self.handleQueueActions(pq, sqsQuery)

	} else if r.URL.Path == "/" {
		return self.handleManageActions(sqsQuery)
	}
	return sqserr.ServiceDeniedError()
}

func (self *SQSRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := self.dispatchSQSQuery(r)
	if resp == nil {
		return
	}

	//log.Info(resp.XmlDocument())
	w.WriteHeader(resp.HttpCode())
	io.WriteString(w, resp.XmlDocument())
	io.WriteString(w, "\n")
}
