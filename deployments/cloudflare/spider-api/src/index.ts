import { Container, getRandom } from "@cloudflare/containers";
import { env } from "cloudflare:workers";

const INSTANCE_COUNT = 3;

export class SpiderAPI extends Container {
  defaultPort = 8000;
  sleepAfter = "30m";
  envVars = {
    SPIDER_ENV: env.SPIDER_ENV,
    SPIDER_FAIL_MODE: env.SPIDER_FAIL_MODE,
    SPIDER_DEFAULT_DETECTOR: env.SPIDER_DEFAULT_DETECTOR,
    SPIDER_DEFAULT_THRESHOLD: env.SPIDER_DEFAULT_THRESHOLD,
    SPIDER_LOG_PROMPT_CONTENT: env.SPIDER_LOG_PROMPT_CONTENT,
    SPIDER_PERSIST_PROMPT_CONTENT: env.SPIDER_PERSIST_PROMPT_CONTENT,
    SPIDER_CORS_ORIGINS: env.SPIDER_CORS_ORIGINS,
    SPIDER_PROMPT_SHIELD_ENDPOINT: env.SPIDER_PROMPT_SHIELD_ENDPOINT,
    SPIDER_PROMPT_SHIELD_MODEL: env.SPIDER_PROMPT_SHIELD_MODEL,
    DATABASE_URL: env.DATABASE_URL,
    SPIDER_JWT_SECRET: env.SPIDER_JWT_SECRET,
    SPIDER_WORKER_TOKEN: env.SPIDER_WORKER_TOKEN,
  };
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const container = await getRandom(env.SPIDER_API, INSTANCE_COUNT);
    return container.fetch(request);
  },
};

interface Env {
  SPIDER_API: DurableObjectNamespace<SpiderAPI>;
  SPIDER_ENV: string;
  SPIDER_FAIL_MODE: string;
  SPIDER_DEFAULT_DETECTOR: string;
  SPIDER_DEFAULT_THRESHOLD: string;
  SPIDER_LOG_PROMPT_CONTENT: string;
  SPIDER_PERSIST_PROMPT_CONTENT: string;
  SPIDER_CORS_ORIGINS: string;
  SPIDER_PROMPT_SHIELD_ENDPOINT: string;
  SPIDER_PROMPT_SHIELD_MODEL: string;
  DATABASE_URL: string;
  SPIDER_JWT_SECRET: string;
  SPIDER_WORKER_TOKEN: string;
}
