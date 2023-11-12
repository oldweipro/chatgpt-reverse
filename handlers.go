package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	arkose "github.com/xqdoo00o/funcaptcha"
)

func optionsHandler(c *gin.Context) {
	// Set headers for CORS
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "*")
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
func chatCompletions(c *gin.Context) {
	proxyUrl := ProxyUrl
	var originalRequest APIRequest
	err := c.BindJSON(&originalRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{
			"message": "Request must be proper JSON",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(400, gin.H{"error": gin.H{
			"message": "missing header parameter Authorization",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
		return
	}
	puid := c.GetHeader("PUid")
	if puid == "" {
		c.JSON(400, gin.H{"error": gin.H{
			"message": "missing header parameter PUid",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
		return
	}
	accessToken := ""
	if authHeader != "" {
		customAccessToken := strings.Replace(authHeader, "Bearer ", "", 1)
		// Check if customAccessToken starts with sk-
		if strings.HasPrefix(customAccessToken, "eyJhbGciOiJSUzI1NiI") {
			accessToken = customAccessToken
		}
	}
	if accessToken == "" {
		c.JSON(400, gin.H{"error": gin.H{
			"message": "wrong header parameter Authorization",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
		return
	}

	// Convert the chat request to a ChatGPT request
	translatedRequest := ConvertAPIRequest(originalRequest, puid, proxyUrl)

	response, err := POSTConversation(translatedRequest, accessToken, puid, proxyUrl)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "error sending request",
		})
		return
	}
	defer response.Body.Close()
	if HandleRequestError(c, response) {
		return
	}
	var fullResponse string
	defer response.Body.Close()
	for i := 3; i > 0; i-- {
		var continueInfo *ContinueInfo
		var responsePart string
		responsePart, continueInfo = Handler(c, response, originalRequest.Stream)
		fullResponse += responsePart
		if continueInfo == nil {
			break
		}
		println("Continuing conversation")
		translatedRequest.Messages = nil
		translatedRequest.Action = "continue"
		translatedRequest.ConversationID = continueInfo.ConversationID
		translatedRequest.ParentMessageID = continueInfo.ParentID
		response, err = POSTConversation(translatedRequest, accessToken, puid, proxyUrl)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "error sending request",
			})
			return
		}
		if HandleRequestError(c, response) {
			return
		}
	}
	if !originalRequest.Stream {
		completion := NewChatCompletion(fullResponse)
		completion.Model = originalRequest.Model
		c.JSON(200, completion)
	} else {
		c.String(200, "data: [DONE]\n\n")
	}

}

func ConvertAPIRequest(apiRequest APIRequest, puid string, proxyUrl string) ChatGPTRequest {
	chatgptRequest := NewChatGPTRequest()
	token, err := arkose.GetOpenAIAuthToken(puid, proxyUrl)
	if err == nil {
		chatgptRequest.ArkoseToken = token
	} else {
		fmt.Println("Error getting Arkose token: ", err)
		return chatgptRequest
	}
	if strings.HasPrefix(apiRequest.Model, "gpt-3.5") {
		chatgptRequest.Model = "text-davinci-002-render-sha"
	}
	if strings.HasPrefix(apiRequest.Model, "gpt-4") {
		chatgptRequest.Model = apiRequest.Model
		// Cover some models like gpt-4-32k
		if len(apiRequest.Model) >= 7 && apiRequest.Model[6] >= 48 && apiRequest.Model[6] <= 57 {
			chatgptRequest.Model = "gpt-4"
		}
	}
	if apiRequest.PluginIDs != nil {
		chatgptRequest.PluginIDs = apiRequest.PluginIDs
		chatgptRequest.Model = "gpt-4-plugins"
	}
	for _, message := range apiRequest.Messages {
		if message.Role == "system" {
			message.Role = "critic"
		}
		chatgptRequest.AddMessage(message.Role, message.Content)
	}
	return chatgptRequest
}

var (
	jar     = tlsclient.NewCookieJar()
	options = []tlsclient.HttpClientOption{
		tlsclient.WithTimeoutSeconds(360),
		tlsclient.WithClientProfile(profiles.Okhttp4Android13),
		tlsclient.WithNotFollowRedirects(),
		tlsclient.WithCookieJar(jar), // create cookieJar instance and pass it as argument
		// Disable SSL verification
		tlsclient.WithInsecureSkipVerify(),
	}
	client, _ = tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
)

func init() {
	arkose.SetTLSClient(&client)
}
func POSTConversation(message ChatGPTRequest, accessToken string, puid string, proxy string) (*http.Response, error) {
	if proxy != "" {
		client.SetProxy(proxy)
	}

	apiUrl := "https://chat.openai.com/backend-api/conversation"

	// JSONify the body and add it to the request
	bodyJson, err := json.Marshal(message)
	if err != nil {
		return &http.Response{}, err
	}

	request, err := http.NewRequest(http.MethodPost, apiUrl, bytes.NewBuffer(bodyJson))
	if err != nil {
		return &http.Response{}, err
	}
	// Clear cookies
	if puid != "" {
		request.Header.Set("Cookie", "_puid="+puid+";")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
	request.Header.Set("Accept", "*/*")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	if err != nil {
		return &http.Response{}, err
	}
	response, err := client.Do(request)
	return response, err
}

func HandleRequestError(c *gin.Context, response *http.Response) bool {
	if response.StatusCode != 200 {
		// Try read response body as JSON
		var errorResponse map[string]interface{}
		err := json.NewDecoder(response.Body).Decode(&errorResponse)
		if err != nil {
			// Read response body
			body, _ := io.ReadAll(response.Body)
			c.JSON(500, gin.H{"error": gin.H{
				"message": "Unknown error",
				"type":    "internal_server_error",
				"param":   nil,
				"code":    "500",
				"details": string(body),
			}})
			return true
		}
		c.JSON(response.StatusCode, gin.H{"error": gin.H{
			"message": errorResponse["detail"],
			"type":    response.Status,
			"param":   nil,
			"code":    "error",
		}})
		return true
	}
	return false
}

type ContinueInfo struct {
	ConversationID string `json:"conversation_id"`
	ParentID       string `json:"parent_id"`
}

func Handler(c *gin.Context, response *http.Response, stream bool) (string, *ContinueInfo) {
	maxTokens := false

	// Create a bufio.Reader from the response body
	reader := bufio.NewReader(response.Body)

	// Read the response byte by byte until a newline character is encountered
	if stream {
		// Response content type is text/event-stream
		c.Header("Content-Type", "text/event-stream")
	} else {
		// Response content type is application/json
		c.Header("Content-Type", "application/json")
	}
	var finishReason string
	var previousText StringStruct
	var originalResponse ChatGPTResponse
	var isRole = true
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", nil
		}
		if len(line) < 6 {
			continue
		}
		// Remove "data: " from the beginning of the line
		line = line[6:]
		// Check if line starts with [DONE]
		if !strings.HasPrefix(line, "[DONE]") {
			// Parse the line as JSON

			err = json.Unmarshal([]byte(line), &originalResponse)
			if err != nil {
				continue
			}
			if originalResponse.Error != nil {
				c.JSON(500, gin.H{"error": originalResponse.Error})
				return "", nil
			}
			if originalResponse.Message.Author.Role != "assistant" || originalResponse.Message.Content.Parts == nil {
				continue
			}
			if originalResponse.Message.Metadata.MessageType != "next" && originalResponse.Message.Metadata.MessageType != "continue" || originalResponse.Message.EndTurn != nil {
				continue
			}
			responseString := ConvertToString(&originalResponse, &previousText, isRole)
			isRole = false
			if stream {
				_, err = c.Writer.WriteString(responseString)
				if err != nil {
					return "", nil
				}
			}
			// Flush the response writer buffer to ensure that the client receives each line as it's written
			c.Writer.Flush()

			if originalResponse.Message.Metadata.FinishDetails != nil {
				if originalResponse.Message.Metadata.FinishDetails.Type == "max_tokens" {
					maxTokens = true
				}
				finishReason = originalResponse.Message.Metadata.FinishDetails.Type
			}

		} else {
			if stream {
				finalLine := StopChunk(finishReason)
				c.Writer.WriteString("data: " + finalLine.String() + "\n\n")
			}
		}
	}
	if !maxTokens {
		return previousText.Text, nil
	}
	return previousText.Text, &ContinueInfo{
		ConversationID: originalResponse.ConversationID,
		ParentID:       originalResponse.Message.ID,
	}
}

func ConvertToString(chatgptResponse *ChatGPTResponse, previousText *StringStruct, role bool) string {
	translatedResponse := NewChatCompletionChunk(strings.ReplaceAll(chatgptResponse.Message.Content.Parts[0], *&previousText.Text, ""))
	if role {
		translatedResponse.Choices[0].Delta.Role = chatgptResponse.Message.Author.Role
	}
	previousText.Text = chatgptResponse.Message.Content.Parts[0]
	return "data: " + translatedResponse.String() + "\n\n"

}
