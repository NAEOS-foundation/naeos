import { describe, it, expect } from "vitest";
import { Handler } from "./handler";
import { DefaultService } from "./service";

describe("Handler", () => {
  it("should handle request", () => {
    const service = new DefaultService();
    const handler = new Handler(service);
    expect(handler.handle()).toBe("processed");
  });
});
