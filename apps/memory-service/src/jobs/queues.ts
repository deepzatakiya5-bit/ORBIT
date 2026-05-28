import { Queue } from "bullmq";
import { env } from "../config/env.js";

const connection = {
  url: env.REDIS_URL,
  maxRetriesPerRequest: null,
};

export const extractionQueue = new Queue("memory-extraction", { connection });
export const summarizationQueue = new Queue("memory-summarization", { connection });
export const decayQueue = new Queue("memory-decay", { connection });

export { connection };
