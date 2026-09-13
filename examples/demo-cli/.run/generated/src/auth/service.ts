export interface Service {
  process(): string;
}

export class DefaultService implements Service {
  process(): string {
    return "processed";
  }
}
