import { drizzle } from "drizzle-orm/netlify-db";

import * as schema from "./schema";

// @ts-expect-error netlify-db types expect a connection string but drizzle accepts schema-only for Netlify
export const db = drizzle({ schema });
