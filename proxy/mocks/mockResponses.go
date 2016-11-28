package mocks

const BindingResponse = "{ " +
	"\"kind\": \"Status\"," +
	"\"apiVersion\": \"v1\"," +
	"\"metadata\": {}," +
	"\"status\": \"Success\"," +
	"\"code\": 201}"

// this response should be based on scheduler request and it should just add
// selfLink, uid, resourceVersion and creationTimestamp fields
const EventResponse = "{" +
	"\"kind\": \"Event\"," +
	"\"apiVersion\": \"v1\"," +
	"\"metadata\": {" +
	"\"name\": \"nginx-without-nodename.148a9fcaa3a27080\"," +
	"\"namespace\": \"default\"," +
	"\"selfLink\": \"/api/v1/namespaces/default/events/nginx-without-nodename.148a9fcaa3a27080\"," +
	"\"uid\": \"851920f3-b3e6-11e6-9514-000c2999b232\"," +
	"\"resourceVersion\": \"114\"," +
	"\"creationTimestamp\": \"2016-11-26T14:42:13Z\"}," +
	"\"involvedObject\": {\"kind\": \"Pod\"," +
	"\"namespace\": \"default\"," +
	"\"name\": \"nginx-without-nodename\"," +
	"\"uid\": \"b0220242-b346-11e6-a633-000c2999b232\"," +
	"\"apiVersion\": \"v1\"," +
	"\"resourceVersion\": \"614\"}," +
	"\"reason\": \"Scheduled\"," +
	"\"message\": \"Successfully assigned nginx-without-nodename to ubuntu\"," +
	"\"source\": {" +
	"\"component\": \"default-scheduler\"}," +
	"\"firstTimestamp\": \"2016-11-26T14:38:40Z\"," +
	"\"lastTimestamp\": \"2016-11-26T14:38:40Z\"," +
	"\"count\": 1," +
	"\"type\": \"Normal\"}"
