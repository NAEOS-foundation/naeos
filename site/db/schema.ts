import { pgTable, serial, text, timestamp } from "drizzle-orm/pg-core";

export const newsletterSubscribers = pgTable("newsletter_subscribers", {
  id: serial().primaryKey(),
  email: text().notNull().unique(),
  locale: text().notNull().default("en"),
  source: text().notNull().default("footer"),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});
