import { PrismaClient } from "@prisma/client";
import dayjs from "dayjs";
import { embedText } from "../llm/geminiClient.js";
import { RetrievalInput } from "../types/memory.js";

export type RetrievedMemory = {
  id: string;
  content: string;
  type: string;
  salience: number;
  confidence: number;
  decayScore: number;
  recencyBoost: number;
  semanticScore: number;
  finalScore: number;
};

export async function retrieveMemories(
  prisma: PrismaClient,
  input: RetrievalInput,
): Promise<RetrievedMemory[]> {
  const limit = input.limit ?? 20;
  const now = input.now ?? new Date();
  const vector = await embedText(input.query);
  const vectorLiteral = `[${vector.join(",")}]`;

  const rows = await prisma.$queryRawUnsafe<Array<{
    id: string;
    content: string;
    type: string;
    salience: number;
    confidence: number;
    decay_score: number;
    updated_at: Date;
    similarity: number;
  }>>(
    `SELECT m.id, m.content, m.type::text, m.salience, m.confidence, m.decay_score, m.updated_at,
            (1 - (me.embedding <=> $2::vector)) AS similarity
       FROM memory m
       JOIN memory_embedding me ON me.memory_id = m.id
      WHERE m.user_id = $1
        AND m.validity = 'ACTIVE'
      ORDER BY similarity DESC
      LIMIT $3`,
    input.userId,
    vectorLiteral,
    limit * 3,
  );

  return rows
    .map((row) => {
      const ageDays = Math.max(0, dayjs(now).diff(row.updated_at, "day"));
      const recencyBoost = Math.exp(-ageDays / 60);
      const finalScore =
        row.similarity * 0.4 +
        row.salience * 0.25 +
        row.confidence * 0.15 +
        row.decay_score * 0.1 +
        recencyBoost * 0.1;
      return {
        id: row.id,
        content: row.content,
        type: row.type,
        salience: row.salience,
        confidence: row.confidence,
        decayScore: row.decay_score,
        recencyBoost,
        semanticScore: row.similarity,
        finalScore,
      };
    })
    .sort((a, b) => b.finalScore - a.finalScore)
    .slice(0, limit);
}
