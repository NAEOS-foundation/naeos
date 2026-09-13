import { Service } from "./service";

export class Handler {
  constructor(private service: Service) {}

  handle(): string {
    return this.service.process();
  }
}
