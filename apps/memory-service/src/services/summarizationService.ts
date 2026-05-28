import dayjs from "dayjs";
import { PrismaClient, SummaryPeriod } from "@prisma/client";
import { buildSummaryPrompt } from "../prompts/extraction.js";
import { generateJson } from "../llm/geminiClient.js";

type SummaryPayload = { summary: string; highlights: string[] };

export async function generateDailySummary(prisma: PrismaClient, userId: string, day: Date) {
  const start = dayjs(day).startOf("day").toDate();
  const end = dayjs(day).endOf("day").toDate();

  const messages = await prisma.message.findMany({
    where: {
      conversation: { userId },
      createdAt: { gte: start, lte: end },
    },
    orderBy: { createdAt: "asc" },
  });

  if (!messages.length) return;
  const slices = messages.map((m) => `${m.role}: ${m.content}`);
  const payload = await generateJson<SummaryPayload>(
    buildSummaryPrompt({ period: "DAILY", slices }),
  );

  await prisma.dailySummary.upsert({
    where: { userId_dayDate: { userId, dayDate: start } },
    create: { userId, dayDate: start, summary: payload.summary, highlights: payload.highlights },
    update: { summary: payload.summary, highlights: payload.highlights },
  });
}

export async function rollupSummary(prisma: PrismaClient, userId: string, period: SummaryPeriod, anchor: Date) {
  if (period === "WEEKLY") {
    const start = dayjs(anchor).startOf("week").toDate();
    const end = dayjs(anchor).endOf("week").toDate();
    const days = await prisma.dailySummary.findMany({
      where: { userId, dayDate: { gte: start, lte: end } },
      orderBy: { dayDate: "asc" },
    });
    if (!days.length) return;
    const payload = await generateJson<SummaryPayload>(
      buildSummaryPrompt({ period: "WEEKLY", slices: days.map((d) => d.summary) }),
    );
    await prisma.weeklySummary.upsert({
      where: { userId_weekStart: { userId, weekStart: start } },
      create: { userId, weekStart: start, weekEnd: end, summary: payload.summary, highlights: payload.highlights },
      update: { summary: payload.summary, highlights: payload.highlights },
    });
    return;
  }

  if (period === "MONTHLY") {
    const monthStart = dayjs(anchor).startOf("month").toDate();
    const monthEnd = dayjs(anchor).endOf("month").toDate();
    const weeks = await prisma.weeklySummary.findMany({
      where: { userId, weekStart: { gte: monthStart, lte: monthEnd } },
      orderBy: { weekStart: "asc" },
    });
    if (!weeks.length) return;
    const payload = await generateJson<SummaryPayload>(
      buildSummaryPrompt({ period: "MONTHLY", slices: weeks.map((w) => w.summary) }),
    );
    await prisma.monthlySummary.upsert({
      where: { userId_monthDate: { userId, monthDate: monthStart } },
      create: { userId, monthDate: monthStart, summary: payload.summary, highlights: payload.highlights },
      update: { summary: payload.summary, highlights: payload.highlights },
    });
    return;
  }

  if (period === "YEARLY") {
    const yearStart = dayjs(anchor).startOf("year").toDate();
    const yearEnd = dayjs(anchor).endOf("year").toDate();
    const months = await prisma.monthlySummary.findMany({
      where: { userId, monthDate: { gte: yearStart, lte: yearEnd } },
      orderBy: { monthDate: "asc" },
    });
    if (!months.length) return;
    const payload = await generateJson<SummaryPayload>(
      buildSummaryPrompt({ period: "YEARLY", slices: months.map((m) => m.summary) }),
    );
    await prisma.yearlySummary.upsert({
      where: { userId_yearDate: { userId, yearDate: yearStart } },
      create: { userId, yearDate: yearStart, summary: payload.summary, highlights: payload.highlights },
      update: { summary: payload.summary, highlights: payload.highlights },
    });
  }
}
