# Bug Report

## Missing sender role validation in signaling

### Severity
Medium

### Summary
The signaling server validates the target client type for `offer` and `answer`, but it does not validate that the sender has the expected role.

### Impact
A registered client can send protocol-invalid SDP messages:

- any registered client can send an `offer` to a broadcaster
- any registered client can send an `answer` to a viewer

This breaks the expected signaling contract and allows a buggy or malicious client to inject invalid session descriptions into another peer's WebRTC flow.

### Expected behavior

- only viewers should be allowed to send `offer` messages to broadcasters
- only broadcasters should be allowed to send `answer` messages to viewers
- invalid role/message combinations should be rejected and logged

### Current behavior
In `main.go`, the `offer` case only checks that the target is a broadcaster, and the `answer` case only checks that the target is a viewer.

### Suggested fix
Add sender-role validation in the signaling switch:

- reject `offer` unless `client.Type == "viewer"`
- reject `answer` unless `client.Type == "broadcaster"`
- keep the existing target validation

### Affected area
- `main.go`
