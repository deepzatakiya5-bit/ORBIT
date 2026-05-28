import { PrismaClient } from "@prisma/client";
import { ExtractionOutput, extractionOutputSchema } from "../types/memory.js";
import { buildExtractionPrompt } from "../prompts/extraction.js";
import { generateJson, embedText } from "../llm/geminiClient.js";
import { calculateSalience, recencyScore, updateDecay } from "./salience.js";
import { resolveContradictions } from "./contradiction.js";
import { logger } from "../config/logger.js";

type ProcessArgs = {
  prisma: PrismaClient;
  userId: string;
  conversationId: string;
  messageIds?: string[];
};

export async function processConversationMemories(args: ProcessArgs) {
  const [user, messages] = await Promise.all([
    args.prisma.user.findUniqueOrThrow({ where: { id: args.userId } }),
    args.prisma.message.findMany({
      where: {
        conversationId: args.conversationId,
        ...(args.messageIds?.length ? { id: { in: args.messageIds } } : {}),
      },
      orderBy: { createdAt: "asc" },
      take: 120,
    }),
  ]);

  const prompt = buildExtractionPrompt({
    userProfile: {
      name: user.name,
      nickname: user.nickname,
      occupation: user.occupation,
      timezone: user.timezone,
      communicationStyle: user.communicationStyle,
      whyHere: user.whyHere,
    },
    conversationId: args.conversationId,
    messages: messages.map((m) => ({
      role: m.role,
      content: m.content,
      createdAt: m.createdAt.toISOString(),
    })),
  });

  const extraction: ExtractionOutput = extractionOutputSchema.parse(
    await generateJson<unknown>(prompt),
  );

  if (!extraction.shouldPersist || extraction.qualityScore < 0.45) {
    logger.info({ userId: args.userId }, "Skipping low quality extraction");
    return;
  }

  for (const mem of extraction.memories) {
    await upsertMemory(args, mem);
  }

  for (const event of extraction.timelineEvents) {
    await args.prisma.timelineEvent.create({
      data: {
        userId: args.userId,
        conversationId: args.conversationId,
        eventType: event.eventType,
        title: event.title,
        description: event.description,
        startDate: new Date(event.startDate),
        endDate: event.endDate ? new Date(event.endDate) : null,
        entities: event.entities,
        location: event.location,
        inferred: event.inferred,
        confidence: event.confidence,
        metadata: event.metadata ?? {},
      },
    });
  }
}

async function upsertMemory(
  args: ProcessArgs,
  mem: any,
) {
  if (!mem?.type || !mem?.content) return;

  const existing = await args.prisma.memory.findFirst({
    where: {
      userId: args.userId,
      content: mem.content,
      type: mem.type as any,
      validity: "ACTIVE",
    },
  });

  const now = new Date();
  if (existing) {
    const nextFrequency = existing.frequency + 1;
    const nextRecency = recencyScore(now, now);
    const nextDecay = updateDecay({
      type: existing.type,
      createdAt: existing.createdAt,
      lastReferencedAt: existing.lastReferencedAt,
      now,
    });
    const nextSalience = calculateSalience({
      recencyScore: nextRecency,
      emotionalWeight: Math.max(existing.emotionalWeight, mem.emotionalWeight ?? 0),
      frequency: nextFrequency,
      explicitImportance: Math.max(existing.explicitImportance, mem.explicitImportance ?? 0),
      userEmphasis: Math.max(existing.userEmphasis, mem.userEmphasis ?? 0),
      goalRelevance: Math.max(existing.goalRelevance, mem.goalRelevance ?? 0),
      decayScore: nextDecay,
    });

    await args.prisma.memory.update({
      where: { id: existing.id },
      data: {
        frequency: nextFrequency,
        confidence: Math.max(existing.confidence, mem.confidence),
        emotionalWeight: Math.max(existing.emotionalWeight, mem.emotionalWeight ?? 0),
        explicitImportance: Math.max(existing.explicitImportance, mem.explicitImportance ?? 0),
        goalRelevance: Math.max(existing.goalRelevance, mem.goalRelevance ?? 0),
        userEmphasis: Math.max(existing.userEmphasis, mem.userEmphasis ?? 0),
        recencyScore: nextRecency,
        decayScore: nextDecay,
        salience: nextSalience,
        lastReferencedAt: now,
      },
    });
    return;
  }

  const decay = updateDecay({
    type: mem.type as any,
    createdAt: now,
    now,
  });
  const salience = calculateSalience({
    recencyScore: 1,
    emotionalWeight: mem.emotionalWeight ?? 0,
    frequency: 1,
    explicitImportance: mem.explicitImportance ?? 0,
    userEmphasis: mem.userEmphasis ?? 0,
    goalRelevance: mem.goalRelevance ?? 0,
    decayScore: decay,
  });

  const created = await args.prisma.memory.create({
    data: {
      userId: args.userId,
      type: mem.type as any,
      content: mem.content,
      confidence: mem.confidence,
      emotionalWeight: mem.emotionalWeight ?? 0,
      explicitImportance: mem.explicitImportance ?? 0,
      userEmphasis: mem.userEmphasis ?? 0,
      goalRelevance: mem.goalRelevance ?? 0,
      recencyScore: 1,
      decayScore: decay,
      salience,
      validity: mem.validity as any,
      contradictionGroup: mem.contradictionGroup,
      sourceConversationId: args.conversationId,
      metadata: (mem.metadata ?? {}) as any,
    },
  });

  await resolveContradictions({
    prisma: args.prisma,
    userId: args.userId,
    contradictionGroup: mem.contradictionGroup,
    memoryId: created.id,
    confidence: created.confidence,
    createdAt: created.createdAt,
  });

  const embedding = await embedText(mem.content);
  await args.prisma.$executeRawUnsafe(
    `INSERT INTO memory_embedding (id, memory_id, provider, dimensions, embedding, created_at, updated_at)
     VALUES (gen_random_uuid(), $1, 'gemini', $2, $3::vector, now(), now())
     ON CONFLICT (memory_id) DO UPDATE SET embedding = excluded.embedding, updated_at = now()`,
    created.id,
    embedding.length,
    `[${embedding.join(",")}]`,
  );
}
