import { Worker } from "bullmq";
import { connection } from "../queues.js";
import { prisma } from "../../db/prisma.js";
import { applyMemoryDecay } from "../../services/decayService.js";
import { logger } from "../../config/logger.js";
new Worker("memory-decay", async () => {
    await applyMemoryDecay(prisma);
}, { connection, concurrency: 1 }).on("failed", (_job, err) => {
    logger.error({ err }, "memory-decay job failed");
});
logger.info("Decay worker started");
