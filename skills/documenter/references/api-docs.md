# API / Reference Documentation

## Structure per function/method

```markdown
### `function_name(param1, param2, **kwargs)`

One-line description of what it does.

**Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `param1` | `str` | required | What it's for. Valid values or constraints. |
| `param2` | `int \| None` | `None` | What it controls. What happens when None. |

**Returns:** `ReturnType` — what the return value represents.

**Raises:**
- `ValueError` — when param1 is empty
- `ConnectionError` — when the service is unreachable

**Example:**
```python
# Basic usage
result = function_name("hello", param2=42)
print(result)  # → ExpectedOutput

# Edge case: None param
result = function_name("hello")
print(result)  # → FallbackBehavior

# Error case
function_name("")  # → raises ValueError
```

**Notes:**
- This function is idempotent — calling it twice with same args produces same result.
- Thread-safe. Can be called from multiple coroutines concurrently.
```

## Structure per class

```markdown
## `ClassName`

One-line: what this class represents and when to use it.

**Constructor:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `config` | `Config` | required | Configuration object |

**Example:**
```python
obj = ClassName(config=Config(timeout=30))
result = obj.process(data)
```

### Methods

#### `process(data) → Result`
{same structure as function above}

#### `validate(item) → bool`
{same structure}

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `is_ready` | `bool` | Whether the instance is initialized |
| `count` | `int` | Number of items processed |
```

## Structure per API endpoint

```markdown
### `POST /api/v1/resource`

Create a new resource.

**Request:**
```json
{
  "name": "string (required, 1-255 chars)",
  "type": "enum: standard | premium",
  "metadata": "object (optional)"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "name": "string",
  "type": "string",
  "created_at": "ISO 8601"
}
```

**Errors:**
| Status | Body | When |
|--------|------|------|
| 400 | `{"error": "name_required"}` | Name is empty or missing |
| 409 | `{"error": "duplicate_name"}` | Resource with that name exists |
| 500 | `{"error": "internal"}` | Unexpected server error |

**Example:**
```bash
curl -X POST https://api.example.com/v1/resource \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-resource", "type": "standard"}'
```

**Notes:**
- Idempotent with same name — returns existing resource if duplicate.
- Rate limited: 100 req/min per API key.
```

## Structure per CLI command

```markdown
### `mycli add <name> [--priority <level>] [--file <path>]`

Add a new item.

**Arguments:**
| Arg | Type | Required | Description |
|-----|------|----------|-------------|
| `name` | string | yes | The item name. Quoted if contains spaces. |

**Options:**
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--priority` | enum | `medium` | `high`, `medium`, or `low` |
| `--file` | path | `~/.tasks/tasks.json` | Custom storage file |

**Example:**
```bash
# Basic
mycli add "Buy groceries"

# With priority
mycli add "Deploy hotfix" --priority high

# Custom file
mycli add "Team task" --file ./team-tasks.json
```

**Exit codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Invalid input |
| 2 | File not found or not writable |
```

## Extracting examples from tests

When the codebase has tests, prefer extracting examples from them:

1. Find test files for the module being documented
2. Identify the simplest test that shows basic usage → becomes the "basic" example
3. Find tests with edge cases → become "edge case" examples
4. Find tests that assert errors → become "error case" examples

Adapt test code to documentation style (remove test framework boilerplate,
add comments explaining what's happening), but preserve the actual values
and assertions.
