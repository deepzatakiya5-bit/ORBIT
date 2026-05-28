import { decayQueue, summarizationQueue } from "./queues.js";
import { prisma } from "../db/prisma.js";
import { logger } from "../config/logger.js";

async function schedule() {
  await decayQueue.add(
    "daily-decay",
    {},
    { repeat: { pattern: "0 2 * * *" }, removeOnComplete: 100, removeOnFail: 100 },
  );

  const users = await prisma.user.findMany({ select: { id: true } });
  for (const user of users) {
    await summarizationQueue.add(
      "daily-summary",
      { userId: user.id, day: new Date().toISOString() },
      { repeat: { pattern: "15 2 * * *" }, removeOnComplete: 100, removeOnFail: 100 },
    );
  }
  logger.info({ users: users.length }, "Scheduled recurring memory jobs");
}

schedule().catch((err) => {
  logger.error({ err }, "Failed to schedule jobs");
  process.exit(1);
});
