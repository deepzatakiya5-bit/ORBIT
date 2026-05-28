import dayjs from "dayjs";
import { MemoryType } from "@prisma/client";
export function recencyScore(lastSignalAt, now = new Date()) {
    const days = Math.max(0, dayjs(now).diff(lastSignalAt, "day"));
    return Math.exp(-days / 45);
}
export function decayHalfLifeDays(type) {
    switch (type) {
        case MemoryType.IDENTITY:
            return 540;
        case MemoryType.GOAL:
        case MemoryType.RELATIONSHIP:
            return 240;
        case MemoryType.PREFERENCE:
        case MemoryType.ROUTINE:
            return 120;
        case MemoryType.FACT:
            return 90;
        default:
            return 150;
    }
}
export function updateDecay(args) {
    const now = args.now ?? new Date();
    const anchor = args.lastReferencedAt ?? args.createdAt;
    const days = Math.max(0, dayjs(now).diff(anchor, "day"));
    const halfLife = decayHalfLifeDays(args.type);
    return Math.pow(0.5, days / halfLife);
}
export function calculateSalience(args) {
    const repetition = Math.min(1, Math.log1p(args.frequency) / Math.log(10));
    const weighted = args.recencyScore * 0.2 +
        args.emotionalWeight * 0.2 +
        repetition * 0.15 +
        args.explicitImportance * 0.15 +
        args.userEmphasis * 0.1 +
        args.goalRelevance * 0.2;
    return Math.max(0, Math.min(1, weighted * args.decayScore));
}
export function reinforceMemory(currentSalience, strength = 0.08) {
    return Math.min(1, currentSalience + strength);
}
