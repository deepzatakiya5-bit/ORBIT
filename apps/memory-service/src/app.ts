import express from "express";
import { logger } from "./config/logger.js";
import { routes } from "./interfaces/http/routes.js";

export function createApp() {
  const app = express();
  app.use(express.json({ limit: "2mb" }));
  app.use((req, _res, next) => {
    logger.info({ method: req.method, path: req.path }, "memory-service request");
    next();
  });
  app.use(routes);

  app.use(
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    (err: unknown, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
      logger.error({ err }, "Unhandled memory-service error");
      res.status(500).json({ error: "internal_error" });
    },
  );
  return app;
}
