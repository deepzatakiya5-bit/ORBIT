import dotenv from "dotenv";
import { z } from "zod";
dotenv.config();
const envSchema = z.object({
    PORT: z.coerce.number().default(8090),
    DATABASE_URL: z.string().min(1),
    REDIS_URL: z.string().min(1),
    LLM_PROVIDER: z.enum(["ollama", "gemini"]).default("ollama"),
    OLLAMA_BASE_URL: z.string().default("http://localhost:11434"),
    OLLAMA_CHAT_MODEL: z.string().default("qwen2.5:7b-instruct"),
    OLLAMA_EMBED_MODEL: z.string().default("nomic-embed-text"),
    GEMINI_API_KEY: z.string().optional(),
    GEMINI_MODEL: z.string().default("gemini-2.5-flash"),
    MEMORY_CONTEXT_BUDGET_TOKENS: z.coerce.number().default(900),
});
export const env = envSchema.parse(process.env);
