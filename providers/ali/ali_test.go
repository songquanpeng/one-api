package ali_test

import (
	"net/http"
	"one-api/common/config"
	"one-api/common/test"
	"one-api/model"
)

func setupAliTestServer() (baseUrl string, server *test.ServerTest, teardown func()) {
	server = test.NewTestServer()
	ts := server.TestServer(func(w http.ResponseWriter, r *http.Request) bool {
		return test.OpenAICheck(w, r)
	})
	ts.Start()
	teardown = ts.Close

	baseUrl = ts.URL
	return
}

func getAliChannel(baseUrl string) model.Channel {
	return test.GetChannel(config.ChannelTypeAli, baseUrl, "", "", "")
}
