import { calculateSalience, updateDecay } from "./salience.js";
export async function applyMemoryDecay(prisma) {
    const memories = await prisma.memory.findMany({
        where: { validity: "ACTIVE" },
        take: 5000,
        orderBy: { updatedAt: "desc" },
    });
    for (const memory of memories) {
        const decayScore = updateDecay({
            type: memory.type,
            createdAt: memory.createdAt,
            lastReferencedAt: memory.lastReferencedAt,
        });
        const salience = calculateSalience({
            recencyScore: memory.recencyScore,
            emotionalWeight: memory.emotionalWeight,
            frequency: memory.frequency,
            explicitImportance: memory.explicitImportance,
            userEmphasis: memory.userEmphasis,
            goalRelevance: memory.goalRelevance,
            decayScore,
        });
        await prisma.memory.update({
            where: { id: memory.id },
            data: { decayScore, salience },
        });
    }
}
