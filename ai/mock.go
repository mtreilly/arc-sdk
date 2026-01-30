// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package ai

import (
	"context"
)

// MockClient is a mock AI client for testing.
type MockClient struct {
	// Response to return from Ask
	Response Response

	// Error to return from Ask
	Err error

	// RecordedRequests stores all requests made
	RecordedRequests []Request

	// AvailableModels is the list of models to return
	AvailableModels []string
}

// NewMockClient creates a new mock client.
func NewMockClient() *MockClient {
	return &MockClient{
		RecordedRequests: make([]Request, 0),
		AvailableModels:  []string{"mock-model"},
	}
}

// Ask implements AIClient.
func (m *MockClient) Ask(ctx context.Context, req Request) (Response, error) {
	m.RecordedRequests = append(m.RecordedRequests, req)
	return m.Response, m.Err
}

// Models implements AIClient.
func (m *MockClient) Models() []string {
	return m.AvailableModels
}

// WithResponse sets the response to return.
func (m *MockClient) WithResponse(text string) *MockClient {
	m.Response = Response{Text: text, Model: "mock-model"}
	return m
}

// WithError sets the error to return.
func (m *MockClient) WithError(err error) *MockClient {
	m.Err = err
	return m
}

// LastRequest returns the most recent request.
func (m *MockClient) LastRequest() Request {
	if len(m.RecordedRequests) == 0 {
		return Request{}
	}
	return m.RecordedRequests[len(m.RecordedRequests)-1]
}

// SmartMockClient is a mock that returns different responses based on prompts.
type SmartMockClient struct {
	Responses map[string]Response
	Default   Response
}

// NewSmartMockClient creates a new smart mock client.
func NewSmartMockClient() *SmartMockClient {
	return &SmartMockClient{
		Responses: make(map[string]Response),
		Default:   Response{Text: "mock response", Model: "mock-model"},
	}
}

// Ask implements AIClient.
func (m *SmartMockClient) Ask(ctx context.Context, req Request) (Response, error) {
	if resp, ok := m.Responses[req.Prompt]; ok {
		return resp, nil
	}
	return m.Default, nil
}

// Models implements AIClient.
func (m *SmartMockClient) Models() []string {
	return []string{"mock-model"}
}

// OnPrompt sets a response for a specific prompt.
func (m *SmartMockClient) OnPrompt(prompt string, text string) *SmartMockClient {
	m.Responses[prompt] = Response{Text: text, Model: "mock-model"}
	return m
}
