import { env } from "../config/env.js";
export async function generateJson(prompt) {
    const response = await fetch(`${env.OLLAMA_BASE_URL}/api/chat`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            model: env.OLLAMA_CHAT_MODEL,
            stream: false,
            format: "json",
            options: { temperature: 0.2 },
            messages: [{ role: "user", content: prompt }],
        }),
    });
    if (!response.ok) {
        throw new Error(`Ollama chat failed: ${response.status} ${await response.text()}`);
    }
    const payload = (await response.json());
    if (payload.error)
        throw new Error(`Ollama chat error: ${payload.error}`);
    const text = payload.message?.content?.trim();
    if (!text)
        throw new Error("Ollama returned empty response");
    return safeJsonParse(text);
}
export async function embedText(text) {
    const response = await fetch(`${env.OLLAMA_BASE_URL}/api/embeddings`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            model: env.OLLAMA_EMBED_MODEL,
            prompt: text,
        }),
    });
    if (!response.ok) {
        throw new Error(`Ollama embeddings failed: ${response.status} ${await response.text()}`);
    }
    const data = (await response.json());
    if (data.error)
        throw new Error(`Ollama embedding error: ${data.error}`);
    const values = data.embedding ?? data.embeddings?.[0];
    if (!values?.length) {
        throw new Error("Failed to produce embedding");
    }
    return normalizeEmbedding(values, 1536);
}
function safeJsonParse(text) {
    try {
        return JSON.parse(text);
    }
    catch {
        const match = text.match(/\{[\s\S]*\}/);
        if (!match)
            throw new Error("Model output is not valid JSON");
        return JSON.parse(match[0]);
    }
}
function normalizeEmbedding(values, targetDims) {
    if (values.length === targetDims)
        return values;
    if (values.length > targetDims)
        return values.slice(0, targetDims);
    const out = values.slice();
    while (out.length < targetDims)
        out.push(0);
    return out;
}
