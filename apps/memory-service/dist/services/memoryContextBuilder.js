import { retrieveMemories } from "./retrievalService.js";
export async function buildMemoryContext(args) {
    const [topMemories, daily, weekly] = await Promise.all([
        retrieveMemories(args.prisma, {
            userId: args.userId,
            query: args.query,
            limit: 20,
        }),
        args.prisma.dailySummary.findMany({
            where: { userId: args.userId },
            orderBy: { dayDate: "desc" },
            take: 3,
        }),
        args.prisma.weeklySummary.findMany({
            where: { userId: args.userId },
            orderBy: { weekStart: "desc" },
            take: 2,
        }),
    ]);
    const identity = topMemories.filter((m) => m.type === "IDENTITY").slice(0, 6);
    const unresolvedGoals = topMemories.filter((m) => m.type === "GOAL").slice(0, 6);
    const emotionalState = topMemories.filter((m) => m.type === "EMOTIONAL_PATTERN").slice(0, 5);
    const lines = [];
    lines.push("## Long-term Memory Context");
    lines.push("Identity anchors:");
    identity.forEach((m) => lines.push(`- ${m.content}`));
    lines.push("Active goals:");
    unresolvedGoals.forEach((m) => lines.push(`- ${m.content}`));
    lines.push("Recent emotional patterns:");
    emotionalState.forEach((m) => lines.push(`- ${m.content}`));
    lines.push("Recent summaries:");
    daily.forEach((s) => lines.push(`- [Daily ${s.dayDate.toISOString().slice(0, 10)}] ${s.summary}`));
    weekly.forEach((s) => lines.push(`- [Weekly ${s.weekStart.toISOString().slice(0, 10)}] ${s.summary}`));
    let context = lines.join("\n");
    const approxTokens = Math.ceil(context.length / 4);
    if (approxTokens > args.budgetTokens) {
        context = context.slice(0, args.budgetTokens * 4);
    }
    await args.prisma.memoryAccessLog.createMany({
        data: topMemories.slice(0, 12).map((m) => ({
            userId: args.userId,
            memoryId: m.id,
            queryText: args.query,
            retrievalScore: m.finalScore,
            usedInPrompt: true,
        })),
    });
    return { context, memories: topMemories };
}
