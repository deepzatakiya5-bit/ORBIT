import { TimelineEventType } from "@prisma/client";
export async function getTimeline(prisma, userId, limit = 100) {
    return prisma.timelineEvent.findMany({
        where: { userId },
        orderBy: { startDate: "desc" },
        take: limit,
    });
}
export async function getLastHometownVisit(prisma, userId) {
    return prisma.timelineEvent.findFirst({
        where: { userId, eventType: TimelineEventType.HOMETOWN_VISIT },
        orderBy: { startDate: "desc" },
    });
}
export async function getTravelFrequency(prisma, userId) {
    const events = await prisma.timelineEvent.findMany({
        where: {
            userId,
            eventType: { in: [TimelineEventType.TRAVEL, TimelineEventType.HOMETOWN_VISIT] },
        },
        select: { startDate: true },
    });
    const byMonth = new Map();
    for (const event of events) {
        const key = `${event.startDate.getUTCFullYear()}-${String(event.startDate.getUTCMonth() + 1).padStart(2, "0")}`;
        byMonth.set(key, (byMonth.get(key) ?? 0) + 1);
    }
    return Array.from(byMonth, ([month, count]) => ({ month, count })).sort((a, b) => a.month.localeCompare(b.month));
}
