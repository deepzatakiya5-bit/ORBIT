import { Worker } from "bullmq";
import { connection } from "../queues.js";
import { prisma } from "../../db/prisma.js";
import { processConversationMemories } from "../../services/extractionService.js";
import { logger } from "../../config/logger.js";

new Worker(
  "memory-extraction",
  async (job) => {
    await processConversationMemories({
      prisma,
      userId: job.data.userId,
      conversationId: job.data.conversationId,
      messageIds: job.data.messageIds,
    });
  },
  { connection, concurrency: 5 },
).on("failed", (job, err) => {
  logger.error({ jobId: job?.id, err }, "memory-extraction job failed");
});

logger.info("Extraction worker started");
