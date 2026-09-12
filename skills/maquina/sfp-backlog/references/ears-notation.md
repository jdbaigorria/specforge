# EARS Notation Reference

Easy Approach to Requirements Syntax — structured natural language for unambiguous requirements.

## Patterns

### Ubiquitous (always active)
```
THE SYSTEM SHALL [behavior]
```
Example: THE SYSTEM SHALL encrypt all data at rest.

### Event-Driven (triggered by event)
```
WHEN [event]
THE SYSTEM SHALL [behavior]
```
Example: WHEN a user submits the login form, THE SYSTEM SHALL validate credentials.

### State-Driven (while condition holds)
```
WHILE [state]
THE SYSTEM SHALL [behavior]
```
Example: WHILE the system is in maintenance mode, THE SYSTEM SHALL reject new connections.

### Unwanted (negative scenarios)
```
IF [unwanted condition]
THEN THE SYSTEM SHALL [mitigation]
```
Example: IF the database is unreachable, THEN THE SYSTEM SHALL return cached data.

### Optional (user-configurable)
```
WHERE [feature is enabled]
THE SYSTEM SHALL [behavior]
```
Example: WHERE dark mode is enabled, THE SYSTEM SHALL render the UI with dark theme.

### Complex (combination)
```
WHILE [state], WHEN [event]
THE SYSTEM SHALL [behavior]
```
Example: WHILE the user is authenticated, WHEN the session expires, THE SYSTEM SHALL prompt re-authentication.

## Guidelines

- One requirement per EARS statement (no "and" combining two behaviors)
- Use active voice: "THE SYSTEM SHALL create" not "a record shall be created"
- Be specific: "within 200ms" not "quickly"
- Each requirement must be independently testable
- Avoid implementation details in requirements (say WHAT, not HOW)
