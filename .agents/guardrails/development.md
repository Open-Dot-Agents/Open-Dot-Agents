# Development guardrails

Use the decisions in .agents/permissions/development.json. These decisions
are instructions to the agent. They are not a security boundary.

Allow means proceed within the user's task. Ask means obtain approval for
the specific action. Deny means do not perform the action. If more than one
decision applies, use deny before ask before allow. Ask for unclassified work.

Project work includes edits, tests, builds, and formatting. Local commits
include staging and new commits for reviewed task changes. External changes
include push, publish, deployment, messages, and changes to external systems.
Destructive work includes history rewrites and loss of unrelated work.

Check the effects of scripts, hooks, subprocesses, and tools before execution.
An allowed test or commit does not authorize a hidden deployment or push.
Keep secrets out of output. Do not read or change the protected paths.
Do not change host security, native trust, or active permissions to avoid an
approval. Existing approval for the same action remains valid in the session.
Native session and administrator restrictions always remain in effect.
