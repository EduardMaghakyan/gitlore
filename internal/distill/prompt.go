package distill

const promptTemplate = `You are summarizing a coding session for a developer's commit history.

Here is the conversation between the developer and their coding agent:
<conversation>
%s
</conversation>

Here is the resulting diff:
<diff>
%s
</diff>

Write exactly 3 sentences:
1. What the developer was trying to achieve (intent, not implementation)
2. What actually changed at a behavioral level (not line-by-line)
3. Anything left incomplete, deferred, or worth flagging for future reference

Be concise. Write in plain past tense. Do not mention file names or line numbers.
Output only the 3 sentences, nothing else.`
