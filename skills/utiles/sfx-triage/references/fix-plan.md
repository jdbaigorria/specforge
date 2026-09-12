# Fix Plan Strategy

## TDD Approach: Test First, Fix Second

When root cause is found, always write the failing test before the fix:

1. **Write a test that reproduces the bug.** This test MUST fail with the current code.
   It proves the bug exists and defines what "fixed" means.

2. **Describe the fix.** What code changes make the test pass.

3. **Write a regression prevention test.** If the bug is a boundary condition or edge case,
   add a test that covers the general pattern — not just this specific instance.

### Example

```python
# Step 1: Failing test (proves the bug)
def test_email_with_special_characters():
    """Bug: /users endpoint returns 500 when email contains ñ"""
    response = client.post("/users", json={"email": "josé@example.com"})
    assert response.status_code == 201  # Currently returns 500

# Step 3: Regression prevention (covers the pattern)
@pytest.mark.parametrize("email", [
    "josé@example.com",      # ñ
    "naïve@example.com",     # ï
    "café@example.com",      # é
    "user+tag@example.com",  # +
])
def test_email_with_unicode_characters(email):
    response = client.post("/users", json={"email": email})
    assert response.status_code == 201
```

## Scope Assessment

Classify the fix to recommend the right next step:

### Trivial (<10 lines, single file, no architectural change)
- Off-by-one error
- Missing null check
- Wrong string comparison
- Typo in constant

→ Recommend: "Apply directly. Here's the fix."

### Small (10-50 lines, 1-3 files, no architectural change)
- Missing validation
- Error handling gap
- Wrong business logic branch
- Missing edge case handling

→ Recommend: "Run `sf new \"<the bug>\"` — a `tipo: bug` skips planning and review."

### Medium (50+ lines, multiple files, may change interfaces)
- Missing feature (assumed but never implemented)
- Incorrect data model
- Wrong integration pattern
- Performance issue requiring restructure

→ Recommend: "This is not a bug fix any more. Run `sf new \"<the capability>\"` as a normal story."

### Large (systemic, cross-cutting, architectural)
- Fundamental design flaw
- Security vulnerability in shared component
- Data corruption requiring migration
- Concurrency issue in core architecture

→ Recommend: "This may need a design change. Review the fix plan and consider
`sf new` as a normal story, so it gets planned."

## Why It Wasn't Caught

Document why the bug made it to production (or current state). This informs
whether it's a one-off or a systemic gap:

- **No test coverage:** The happy path was tested but not this edge case
- **Assumption mismatch:** Developer assumed X, reality is Y
- **Environment difference:** Works locally, fails with production data/config
- **Regression:** Unrelated change broke this path, no integration test caught it
- **Missing requirement:** The spec never mentioned this scenario

If the reason is systemic (keeps happening across features), note it. sf-check's
backprop mechanism will promote it to invariant after 3 occurrences.

## Fix pattern: condition-based waiting

When the bug is a race — flaky test, "works when I step through it", passes
locally and fails in CI — the reflex is a sleep. A sleep is not a fix: it trades
a race for a slower race, and the duration is a guess that rots as the machine
changes.

Wait on the **condition**, not on the clock:

```python
# ✗ arbitrary — too short and it flakes, too long and every run pays for it
time.sleep(2)
assert widget.is_loaded()

# ✓ polls the actual condition, with a ceiling so a real hang still fails
wait_until(lambda: widget.is_loaded(), timeout=5, interval=0.05)
```

Three requirements for the replacement:

1. **A ceiling.** Unbounded polling turns a failing test into a hanging one,
   which is strictly worse — a hang tells you nothing and blocks the suite.
2. **A useful failure message.** On timeout, say what condition never became
   true. `TimeoutError` alone sends the next reader back to square one.
3. **The condition is the real one.** Polling a proxy ("the spinner is gone")
   for the thing you care about ("the data is rendered") reintroduces the race
   one layer down.

Most test frameworks ship this: `WebDriverWait` (Selenium), `waitFor`
(Testing Library), `Eventually` (Gomega), `pytest-timeout` plus a poll helper.
Reach for the one already in the project before writing your own.
