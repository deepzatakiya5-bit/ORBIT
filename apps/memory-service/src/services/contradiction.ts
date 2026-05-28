import { PrismaClient } from "@prisma/client";

export async function resolveContradictions(args: {
  prisma: PrismaClient;
  userId: string;
  contradictionGroup?: string | null;
  memoryId: string;
  confidence: number;
  createdAt: Date;
}) {
  if (!args.contradictionGroup) return;

  const peers = await args.prisma.memory.findMany({
    where: {
      userId: args.userId,
      contradictionGroup: args.contradictionGroup,
      id: { not: args.memoryId },
      validity: "ACTIVE",
    },
    orderBy: { createdAt: "desc" },
    take: 5,
  });

  for (const peer of peers) {
    const shouldSupersede =
      args.confidence >= peer.confidence ||
      args.createdAt.getTime() > peer.createdAt.getTime();
    if (shouldSupersede) {
      await args.prisma.memory.update({
        where: { id: peer.id },
        data: {
          validity: "SUPERSEDED",
          validUntil: args.createdAt,
          metadata: {
            ...(peer.metadata as object),
            supersededBy: args.memoryId,
          },
        },
      });
    }
  }
}
