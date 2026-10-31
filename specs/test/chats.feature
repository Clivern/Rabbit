Feature: Assistant chats

  Each member can chat with the dashboard assistant in a workspace. A chat
  belongs to the user who started it: nobody else lists or reads it, and
  workspace access keys cannot use chats at all. Sending a message stores the
  question and streams the answer as server-sent events; the answer is stored
  once it is written, even if the client leaves mid-stream.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/chats
    POST   /api/v1/workspaces/{workspaceId}/chats
    GET    /api/v1/workspaces/{workspaceId}/chats/{chatId}
    DELETE /api/v1/workspaces/{workspaceId}/chats/{chatId}
    POST   /api/v1/workspaces/{workspaceId}/chats/{chatId}/messages

  Scenarios tagged @llm need the server's AI provider configured (app.llm.api_key)
  and make real model calls.

  Background:
    Given the run's workspace exists

  # ---------------------------------------------------------------------------
  # Chats
  # ---------------------------------------------------------------------------

  @readonly @smoke
  Scenario: The user can list their chats
    When the user GETs chats
    Then the response is 200
    And chats is an array

  @mutates @smoke
  Scenario: Starting a chat returns an empty idle chat
    When the user POSTs chats
    Then the response is 201
    And it has an id, title is "" and status is "idle"
    And it has no messages field

  @mutates
  Scenario: A new chat shows in the list
    Given the user started a chat
    When the user GETs chats
    Then the list holds that chat with status "idle" and no preview

  @mutates
  Scenario: A chat reads back by id
    Given the user started a chat
    When the user GETs chats/<that chat's id>
    Then the response is 200
    And id is that chat's id and status is "idle"

  @readonly
  Scenario: Reading an unknown chat returns not found
    When the user GETs chats/<a random UUID>
    Then the response is 404
    And the errorMessage is "Chat not found."

  @mutates
  Scenario: Deleting a chat removes it
    Given the user started a chat
    When the user DELETEs chats/<that chat's id>
    Then the response is 204
    And reading chats/<that chat's id> returns 404
    And the list no longer holds it

  @readonly
  Scenario: Deleting an unknown chat returns not found
    When the user DELETEs chats/<a random UUID>
    Then the response is 404

  @mutates
  Scenario: Access keys cannot use chats
    Given an access key with every permission in AccessKeyPermissions
    When that key GETs chats
    Then the response is 403
    When that key POSTs chats
    Then the response is 403

  # ---------------------------------------------------------------------------
  # Sending messages
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A message needs content
    Given the user started a chat
    When the user POSTs chats/<that chat's id>/messages with body {}
    Then the response is 400
    And the body is JSON with an errorMessage

  @mutates
  Scenario: A message has at most 4000 characters
    Given the user started a chat
    When the user POSTs chats/<that chat's id>/messages with content of 4001 characters
    Then the response is 400

  @readonly
  Scenario: Sending to an unknown chat returns not found
    When the user POSTs chats/<a random UUID>/messages with content "hello"
    Then the response is 404

  @mutates @llm @smoke
  Scenario: Sending a message streams the answer and stores it
    Given the user started a chat
    When the user POSTs chats/<that chat's id>/messages with content "  What is this workspace called?  "
    Then the response is 200 with Content-Type "text/event-stream"
    And the first event is "chat" with status "running", title "What is this workspace called?"
      and one message of kind "user" with content "What is this workspace called?"
    And one or more "delta" events follow, each with a non-empty content
    And the last event is "done" with status "idle" and two messages: the user's, then one of kind "assistant"
    And the assistant message's content equals the delta contents joined in order, trimmed
    And reading the chat returns the same two messages with status "idle"

  @mutates @llm
  Scenario: The chat title is the first 60 characters of the first question
    Given the user started a chat
    When the user sends a 100 character question and reads the stream to the end
    Then the chat's title is the question's first 60 characters
    When the user sends "and the second one?" and reads the stream to the end
    Then the title has not changed

  @mutates @llm
  Scenario: A second message waits for the first answer
    Given the user started a chat
    And the user sent "Which agents are registered?" and has not read the stream yet
    When the user POSTs chats/<that chat's id>/messages with content "hello" while the first is answering
    Then the response is 409
    And the errorMessage is "The assistant is still answering the last message."

  @mutates @llm @slow
  Scenario: The answer is stored when the client leaves mid-stream
    Given the user started a chat
    When the user POSTs chats/<that chat's id>/messages with content "Tell me about this workspace"
    And closes the connection right after the "chat" event
    Then within 60 seconds reading the chat returns status "idle" and an "assistant" message

  @mutates @llm
  Scenario: The newest message is the chat's preview
    Given the user started a chat
    And the user sent "What is this workspace called?" and read the stream to the end
    When the user GETs chats
    Then that chat's preview is the assistant's answer
