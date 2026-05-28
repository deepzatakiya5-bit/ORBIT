import dayjs from "dayjs";
import { Worker } from "bullmq";
import { connection } from "../queues.js";
import { prisma } from "../../db/prisma.js";
import { generateDailySummary, rollupSummary } from "../../services/summarizationService.js";
import { logger } from "../../config/logger.js";

new Worker(
  "memory-summarization",
  async (job) => {
    const userId: string = job.data.userId;
    const day = job.data.day ? new Date(job.data.day) : new Date();
    await generateDailySummary(prisma, userId, day);
    await rollupSummary(prisma, userId, "WEEKLY", day);
    if (dayjs(day).date() === dayjs(day).daysInMonth()) {
      await rollupSummary(prisma, userId, "MONTHLY", day);
    }
    if (dayjs(day).month() === 11 && dayjs(day).date() === 31) {
      await rollupSummary(prisma, userId, "YEARLY", day);
    }
  },
  { connection, concurrency: 3 },
).on("failed", (job, err) => {
  logger.error({ jobId: job?.id, err }, "memory-summarization job failed");
});

logger.info("Summarization worker started");
