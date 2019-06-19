package tm1

import (
	"encoding/json"
)

//TransactionLogsRespose is a response to GetTransactionLogs function
type TransactionLogsRespose struct {
	OdataContext   string                `json:"@odata.context"`
	OdataDeltaLink string                `json:"@odata.deltaLink"`
	Value          []TransactionLogEntry `json:"value"`
}

//MessageLogsRespose is a response to GetTransactionLogs function
type MessageLogsRespose struct {
	OdataContext   string                `json:"@odata.context"`
	OdataDeltaLink string                `json:"@odata.deltaLink"`
	Value          []TransactionLogEntry `json:"value"`
}

//MessageLogEntry is a single message log entry
type MessageLogEntry struct {
	ID        int    `json:"ID"`
	ThreadID  int    `json:"ThreadID"`
	SessionID int    `json:"SessionID"`
	Level     string `json:"Level"`
	TimeStamp string `json:"TimeStamp"`
	Logger    string `json:"Logger"`
	Message   string `json:"Message"`
}

//TransactionLogEntry is a single transaction log entry
type TransactionLogEntry struct {
	ID              int         `json:"ID"`
	ChangeSetID     string      `json:"ChangeSetID"`
	TimeStamp       string      `json:"TimeStamp"`
	ReplicationTime string      `json:"ReplicationTime"`
	User            string      `json:"User"`
	Cube            string      `json:"Cube"`
	Tuple           []string    `json:"Tuple"`
	OldValue        interface{} `json:"OldValue"`
	NewValue        interface{} `json:"NewValue"`
	StatusMessage   string      `json:"StatusMessage"`
}

//GetTransactionLogs method gets transaction logs from tm1
func (s Tm1Session) GetTransactionLogs(filter, deltaLink string) (TransactionLogsRespose, error) {
	transactionLogs := TransactionLogsRespose{}
	var path string
	if deltaLink != "" {
		path = "/" + deltaLink
	} else {
		if filter != "" {
			path = "/TransactionLogEntries?$orderby=TimeStamp&$filter=" + filter
		} else {
			path = "/TransactionLogEntries?$orderby=TimeStamp"
		}

	}

	res, err := s.Tm1SendHttpRequest("GET", path, nil)
	if err != nil {
		return transactionLogs, err
	}

	json.Unmarshal(res, &transactionLogs)

	return transactionLogs, nil
}

//GetMessageLogs method gets tm1 logs
func (s Tm1Session) GetMessageLogs(filter, deltaLink string) (MessageLogsRespose, error) {
	messageLogs := MessageLogsRespose{}
	var path string
	if deltaLink != "" {
		path = "/" + deltaLink
	} else {
		if filter != "" {
			path = "/MessageLogEntries?$orderby=TimeStamp&$filter=" + filter
		} else {
			path = "/MessageLogEntries?$orderby=TimeStamp"
		}

	}

	res, err := s.Tm1SendHttpRequest("GET", path, nil)
	if err != nil {
		return messageLogs, err
	}

	json.Unmarshal(res, &messageLogs)

	return messageLogs, nil
}
