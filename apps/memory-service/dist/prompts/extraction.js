export function buildExtractionPrompt(input) {
    return `You are a high-precision memory extraction engine for a long-term AI companion.
Return JSON only. No markdown. No prose.

Rules:
1) Extract ONLY long-term meaningful information.
2) Do not store one-off trivia unless it affects life continuity.
3) Every memory must be grounded in explicit user evidence from the transcript.
4) If confidence < 0.45, omit the memory.
5) Detect contradictions against prior facts if clear (same contradiction_group).
6) Event dates should be ISO strings. If uncertain, mark inferred=true and lower confidence.

Output JSON schema:
{
  "memories": [{
    "type": "FACT|PREFERENCE|GOAL|RELATIONSHIP|EVENT|EMOTIONAL_PATTERN|IDENTITY|ROUTINE",
    "content": "short grounded statement",
    "confidence": 0.0,
    "emotionalWeight": 0.0,
    "explicitImportance": 0.0,
    "goalRelevance": 0.0,
    "userEmphasis": 0.0,
    "contradictionGroup": "optional normalized key",
    "validity": "ACTIVE|SUPERSEDED|INACTIVE|CONFLICTED",
    "metadata": {}
  }],
  "timelineEvents": [{
    "eventType": "TRAVEL|HOMETOWN_VISIT|CAREER_CHANGE|RELATIONSHIP_MILESTONE|ACHIEVEMENT|HEALTH|ROUTINE_CHANGE|CUSTOM",
    "title": "event title",
    "description": "optional details",
    "startDate": "ISO datetime",
    "endDate": "ISO datetime optional",
    "entities": ["people/org/place"],
    "location": "optional",
    "inferred": false,
    "confidence": 0.0,
    "metadata": {}
  }],
  "emotionalPatterns": [],
  "shouldPersist": true,
  "qualityScore": 0.0
}

User profile:
${JSON.stringify(input.userProfile)}

Conversation ID:
${input.conversationId}

Transcript:
${JSON.stringify(input.messages)}
`;
}
export function buildSummaryPrompt(input) {
    return `Create a ${input.period.toLowerCase()} life-continuity summary for an AI companion.
Avoid generic recap. Focus on:
- life events and transitions
- emotional shifts
- recurring goals and progress/regression
- relationship changes
- routines and identity evolution

Return JSON only:
{
  "summary": "compact narrative paragraph",
  "highlights": ["bullet-level structured highlights"]
}

Source slices:
${JSON.stringify(input.slices)}
`;
}
