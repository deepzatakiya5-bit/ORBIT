import { z } from "zod";
import { MemoryType, TimelineEventType } from "@prisma/client";

export { MemoryType, TimelineEventType };

export const extractedMemorySchema = z.object({
  type: z.nativeEnum(MemoryType),
  content: z.string().min(8),
  confidence: z.number().min(0).max(1),
  emotionalWeight: z.number().min(0).max(1).default(0),
  explicitImportance: z.number().min(0).max(1).default(0),
  goalRelevance: z.number().min(0).max(1).default(0),
  userEmphasis: z.number().min(0).max(1).default(0),
  contradictionGroup: z.string().optional(),
  validity: z.enum(["ACTIVE", "SUPERSEDED", "INACTIVE", "CONFLICTED"]).default("ACTIVE"),
  metadata: z.record(z.any()).optional(),
});

export const extractedTimelineEventSchema = z.object({
  eventType: z.nativeEnum(TimelineEventType),
  title: z.string().min(3),
  description: z.string().optional(),
  startDate: z.string().datetime(),
  endDate: z.string().datetime().optional(),
  entities: z.array(z.string()).default([]),
  location: z.string().optional(),
  inferred: z.boolean().default(false),
  confidence: z.number().min(0).max(1).default(0.5),
  metadata: z.record(z.any()).optional(),
});

export const extractionOutputSchema = z.object({
  memories: z.array(extractedMemorySchema),
  timelineEvents: z.array(extractedTimelineEventSchema),
  emotionalPatterns: z.array(extractedMemorySchema).default([]),
  shouldPersist: z.boolean().default(true),
  qualityScore: z.number().min(0).max(1).default(0.5),
});

export type ExtractionOutput = z.infer<typeof extractionOutputSchema>;

export type RetrievalInput = {
  userId: string;
  query: string;
  now?: Date;
  limit?: number;
};
