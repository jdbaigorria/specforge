# Design-First Workflow

## When to Use

- You know the technical stack and constraints before knowing exact requirements
- The project has strict non-functional requirements (latency, throughput, compliance)
- You're porting an existing architecture or design document
- You want to explore what's technically feasible before committing to requirements

## Flow (inverts Step 2 and Step 3 of Requirements-First)

### Step 2 becomes: Generate design.md FIRST

Ask the user about:
- **Technical constraints:** stack, services, infrastructure
- **Non-functional requirements:** performance, scalability, security, compliance
- **Integration points:** existing systems, APIs, databases
- **Architecture preferences:** patterns, opinions, prior art

Generate `design.md` with the same structure as Requirements-First.

→ 🔴 **GATE**: Present `design.md`. Wait for approval.

### Step 3 becomes: Derive requirements.md FROM design

Key difference: requirements are scoped to what the architecture can deliver.
Don't generate requirements that the design can't support.

For each component in the design, ask:
- What user-facing behavior does this enable?
- What system behaviors are implied?
- What edge cases does this architecture create?

Generate `requirements.md` with the same structure as Requirements-First.
Add a note at the top:

```markdown
> ℹ️ These requirements were derived from the technical design.
> They represent what is feasible given the chosen architecture.
```

→ 🔴 **GATE**: Present `requirements.md`. Wait for approval.

### Step 4: Same — generate tasks.md

Identical to Requirements-First. Tasks trace to requirements regardless of
which was generated first.

## Detail Levels

### High Level Design
Full architecture: components, interactions, data flow, deployment.
Best for: complex systems, team collaboration, thorough documentation.

### Low Level Design
Implementation-focused: pseudocode, interfaces, data structures.
Best for: rapid prototyping, solo development, feasibility checks.

Ask the user which level they want, or infer:
- Mentions "architecture", "diagram", "components" → High Level
- Mentions "algorithm", "interface", "prototype" → Low Level
- Ambiguous → default to High Level

## Uploading Existing Designs

If the user provides an existing design document (architecture diagram, 
RFC, design doc, whiteboard photo):
1. Parse the provided design
2. Formalize into `design.md` format
3. Present for approval
4. Derive requirements from the formalized design
