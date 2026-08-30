import { Container, getContainer } from "@cloudflare/containers";
import { env } from "cloudflare:workers";

export class PromptShield extends Container {
  defaultPort = 8081;
  sleepAfter = "2h";
  envVars = {
    SPIDER_PROMPT_SHIELD_MODEL: env.SPIDER_PROMPT_SHIELD_MODEL,
    HF_TOKEN: env.HF_TOKEN,
  };
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const container = getContainer(env.PROMPT_SHIELD);
    return container.fetch(request);
  },
};

interface Env {
  PROMPT_SHIELD: DurableObjectNamespace<PromptShield>;
  SPIDER_PROMPT_SHIELD_MODEL: string;
  HF_TOKEN: string;
}
