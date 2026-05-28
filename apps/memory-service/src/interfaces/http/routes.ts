import { Router } from "express";
import { z } from "zod";
import { prisma } from "../../db/prisma.js";
import { retrieveMemories } from "../../services/retrievalService.js";
import { buildMemoryContext } from "../../services/memoryContextBuilder.js";
import { reinforceMemory } from "../../services/salience.js";
import {
  getLastHometownVisit,
  getTimeline,
  getTravelFrequency,
} from "../../services/timelineService.js";
import { extractionQueue, summarizationQueue } from "../../jobs/queues.js";
import { env } from "../../config/env.js";

export const routes = Router();

const userParam = z.object({ userId: z.string().uuid() });

routes.get("/health", (_req, res) => res.json({ status: "ok" }));

// Internal trigger from chat backend after new messages.
routes.post("/internal/events/conversation-updated", async (req, res, next) => {
  try {
    const body = z
      .object({
        userId: z.string().uuid(),
        conversationId: z.string().uuid(),
        messageIds: z.array(z.string().uuid()).optional(),
      })
      .parse(req.body);
    await extractionQueue.add(
      "extract",
      body,
      { attempts: 5, backoff: { type: "exponential", delay: 2000 } },
    );
    await summarizationQueue.add("daily", { userId: body.userId });
    res.status(202).json({ queued: true });
  } catch (error) {
    next(error);
  }
});

routes.get("/v1/users/:userId/memory/context", async (req, res, next) => {
  try {
    const { userId } = userParam.parse(req.params);
    const query = z.string().min(1).parse(req.query.query);
    const result = await buildMemoryContext({
      prisma,
      userId,
      query,
      budgetTokens: env.MEMORY_CONTEXT_BUDGET_TOKENS,
    });
    res.json(result);
  } catch (error) {
    next(error);
  }
});

routes.get("/v1/users/:userId/memory/search", async (req, res, next) => {
  try {
    const { userId } = userParam.parse(req.params);
    const query = z.string().min(1).parse(req.query.query);
    const memories = await retrieveMemories(prisma, { userId, query, limit: 30 });
    res.json({ memories });
  } catch (error) {
    next(error);
  }
});

routes.get("/v1/users/:userId/memory", async (req, res, next) => {
  try {
    const { userId } = userParam.parse(req.params);
    const type = req.query.type ? String(req.query.type) : undefined;
    const memories = await prisma.memory.findMany({
      where: {
        userId,
        validity: "ACTIVE",
        ...(type ? { type: type as any } : {}),
      },
      orderBy: [{ salience: "desc" }, { updatedAt: "desc" }],
      take: 100,
    });
    res.json({ memories });
  } catch (error) {
    next(error);
  }
});

routes.post("/v1/users/:userId/memory/:memoryId/reinforce", async (req, res, next) => {
  try {
    const params = z.object({
      userId: z.string().uuid(),
      memoryId: z.string().uuid(),
    }).parse(req.params);
    const existing = await prisma.memory.findUniqueOrThrow({ where: { id: params.memoryId } });
    if (existing.userId !== params.userId) {
      return res.status(404).json({ error: "memory not found" });
    }
    const salience = reinforceMemory(existing.salience);
    const memory = await prisma.memory.update({
      where: { id: params.memoryId },
      data: {
        salience,
        frequency: { increment: 1 },
        lastReferencedAt: new Date(),
      },
    });
    res.json({ memory });
  } catch (error) {
    next(error);
  }
});

routes.post("/v1/users/:userId/memory/:memoryId/feedback", async (req, res, next) => {
  try {
    const params = z.object({
      userId: z.string().uuid(),
      memoryId: z.string().uuid(),
    }).parse(req.params);
    const feedback = z
      .object({
        helpful: z.boolean(),
        correction: z.string().optional(),
      })
      .parse(req.body);
    const memory = await prisma.memory.update({
      where: { id: params.memoryId },
      data: {
        confidence: feedback.helpful ? { increment: 0.05 } : { decrement: 0.1 },
        metadata: {
          feedback,
          feedbackAt: new Date().toISOString(),
        },
      },
    });
    res.json({ memory });
  } catch (error) {
    next(error);
  }
});

routes.delete("/v1/users/:userId/memory/:memoryId", async (req, res, next) => {
  try {
    const params = z.object({
      userId: z.string().uuid(),
      memoryId: z.string().uuid(),
    }).parse(req.params);
    await prisma.memory.update({
      where: { id: params.memoryId },
      data: { validity: "INACTIVE" },
    });
    res.status(204).send();
  } catch (error) {
    next(error);
  }
});

routes.get("/v1/users/:userId/timeline", async (req, res, next) => {
  try {
    const { userId } = userParam.parse(req.params);
    const events = await getTimeline(prisma, userId);
    res.json({ events });
  } catch (error) {
    next(error);
  }
});

routes.get("/v1/users/:userId/timeline/insights", async (req, res, next) => {
  try {
    const { userId } = userParam.parse(req.params);
    const [lastHometownVisit, travelFrequency] = await Promise.all([
      getLastHometownVisit(prisma, userId),
      getTravelFrequency(prisma, userId),
    ]);
    res.json({ lastHometownVisit, travelFrequency });
  } catch (error) {
    next(error);
  }
});

routes.get("/v1/users/:userId/summaries", async (req, res, next) => {
  try {
    const { userId } = userParam.parse(req.params);
    const [daily, weekly, monthly, yearly] = await Promise.all([
      prisma.dailySummary.findMany({ where: { userId }, orderBy: { dayDate: "desc" }, take: 14 }),
      prisma.weeklySummary.findMany({ where: { userId }, orderBy: { weekStart: "desc" }, take: 8 }),
      prisma.monthlySummary.findMany({ where: { userId }, orderBy: { monthDate: "desc" }, take: 12 }),
      prisma.yearlySummary.findMany({ where: { userId }, orderBy: { yearDate: "desc" }, take: 5 }),
    ]);
    res.json({ daily, weekly, monthly, yearly });
  } catch (error) {
    next(error);
  }
});

routes.get("/v1/users/:userId/export", async (req, res, next) => {
  try {
    const { userId } = userParam.parse(req.params);
    const [memories, timeline, summaries] = await Promise.all([
      prisma.memory.findMany({ where: { userId }, orderBy: { createdAt: "asc" } }),
      prisma.timelineEvent.findMany({ where: { userId }, orderBy: { startDate: "asc" } }),
      prisma.dailySummary.findMany({ where: { userId }, orderBy: { dayDate: "asc" } }),
    ]);
    res.json({ userId, exportedAt: new Date().toISOString(), memories, timeline, summaries });
  } catch (error) {
    next(error);
  }
});

// Admin / debugging endpoints
routes.get("/internal/debug/users/:userId/retrieval-viewer", async (req, res, next) => {
  try {
    const { userId } = userParam.parse(req.params);
    const query = z.string().min(1).parse(req.query.query);
    const memories = await retrieveMemories(prisma, { userId, query, limit: 50 });
    res.json({
      query,
      ranked: memories.map((m) => ({
        id: m.id,
        type: m.type,
        content: m.content,
        semanticScore: m.semanticScore,
        salience: m.salience,
        decayScore: m.decayScore,
        finalScore: m.finalScore,
      })),
    });
  } catch (error) {
    next(error);
  }
});

routes.get("/internal/debug/users/:userId/contradictions", async (req, res, next) => {
  try {
    const { userId } = userParam.parse(req.params);
    const groups = await prisma.memory.groupBy({
      by: ["contradictionGroup"],
      where: { userId, contradictionGroup: { not: null } },
      _count: true,
      having: { contradictionGroup: { _count: { gt: 1 } } },
    });
    res.json({ groups });
  } catch (error) {
    next(error);
  }
});
